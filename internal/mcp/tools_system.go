//nolint:nilerr
package mcp

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"github.com/flowexec/flow/pkg/filesystem"
)

// Documentation URLs — kept as constants so a get_info call returns pointers rather
// than embedding the full content (which would balloon the response to ~20KB).
const (
	docsBaseURL    = "https://flowexec.io"
	docsLLMsTxtURL = docsBaseURL + "/llms.txt"

	flowInfoSummary = "Flow is a local automation platform. " +
		"Executables (tasks) are declared in *.flow YAML files; workspaces group them by project " +
		"and are rooted at a flow.yaml config. Templates (*.flow.tmpl) generate new workflows. " +
		"Secrets live in vaults. Use the `get_executable`, `list_executables`, and `execute` tools to " +
		"explore and run; use `write_flowfile` to author new files. Refer to llms.txt for full docs."
)

func addSystemTools(srv *server.MCPServer, executor CommandExecutor) {
	getFlowInfo := mcp.NewTool("get_info",
		mcp.WithDescription(
			"Get information about flow, it's usage, and the current workflow execution context. "+
				"This includes file JSON schemas for flow executable, template, and workspace files, concepts guides, "+
				"and the current user configuration and state details."),
		mcp.WithOutputSchema[FlowInfoOutput](),
	)
	getFlowInfo.Annotations = mcp.ToolAnnotation{
		Title:           "Get flow information and current context",
		DestructiveHint: boolPtr(false), ReadOnlyHint: boolPtr(true),
		IdempotentHint: boolPtr(false), OpenWorldHint: boolPtr(false),
	}
	srv.AddTool(getFlowInfo, getInfoHandler)

	getExecutionLogs := mcp.NewTool("get_execution_logs",
		mcp.WithDescription("Get a list of the recent flow execution logs"),
		mcp.WithBoolean("last", mcp.Description("Get only the last execution logs")),
		mcp.WithString("cursor", mcp.Description("Pagination cursor for next page of results")),
		mcp.WithOutputSchema[LogListOutput](),
	)
	getExecutionLogs.Annotations = mcp.ToolAnnotation{
		Title:           "Get execution logs",
		DestructiveHint: boolPtr(false), ReadOnlyHint: boolPtr(true),
		IdempotentHint: boolPtr(true), OpenWorldHint: boolPtr(true),
	}
	srv.AddTool(getExecutionLogs, getExecutionLogsHandler(executor))

	sync := mcp.NewTool("sync_executables",
		mcp.WithDescription("Sync the flow workspace and executable state"),
		mcp.WithOutputSchema[SyncOutput](),
	)
	sync.Annotations = mcp.ToolAnnotation{
		Title:           "Sync executable and workspace state",
		DestructiveHint: boolPtr(false), ReadOnlyHint: boolPtr(false),
		IdempotentHint: boolPtr(false), OpenWorldHint: boolPtr(true),
	}
	srv.AddTool(sync, syncStateHandler(srv, executor))
}

func getInfoHandler(_ context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	cfg, err := filesystem.LoadConfig()
	if err != nil {
		return toolError(ErrCodeInternal, fmt.Sprintf("failed to load user config: %s", err)), nil
	}
	cfg.SetDefaults()

	wsName, err := cfg.CurrentWorkspaceName()
	if err != nil {
		return toolError(ErrCodeInternal, fmt.Sprintf("failed to get current workspace name: %s", err)), nil
	}

	output := FlowInfoOutput{
		CurrentContext: CurrentContext{
			Workspace:     wsName,
			Namespace:     cfg.CurrentNamespace,
			Vault:         cfg.CurrentVaultName(),
			WorkspaceMode: string(cfg.WorkspaceMode),
			WorkspacePath: cfg.Workspaces[cfg.CurrentWorkspace],
		},
		Summary:    flowInfoSummary,
		DocsURL:    docsBaseURL,
		LLMsTxtURL: docsLLMsTxtURL,
		SchemaURLs: SchemaURLs{
			FlowFile:  docsBaseURL + "/schemas/flowfile_schema.json",
			Workspace: docsBaseURL + "/schemas/workspace_schema.json",
			Template:  docsBaseURL + "/schemas/template_schema.json",
			Config:    docsBaseURL + "/schemas/config_schema.json",
		},
		GuideURLs: map[string]string{
			"concepts":      docsBaseURL + "/guides/concepts",
			"fileTypes":     docsBaseURL + "/guides/executables",
			"firstWorkflow": docsBaseURL + "/guides/first-workflow",
			"workspaces":    docsBaseURL + "/guides/workspaces",
			"templates":     docsBaseURL + "/guides/templating",
			"secrets":       docsBaseURL + "/guides/secrets",
		},
	}

	jsonData, err := json.Marshal(output)
	if err != nil {
		return toolError(ErrCodeInternal, fmt.Sprintf("failed to marshal response: %s", err)), nil
	}

	return mcp.NewToolResultText(string(jsonData)), nil
}

func getExecutionLogsHandler(executor CommandExecutor) server.ToolHandlerFunc {
	return func(_ context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		last := request.GetBool("last", false)
		cursor := request.GetString("cursor", "")

		cmdArgs := []string{"logs", "--output", "json"}
		if last {
			cmdArgs = append(cmdArgs, "--last")
		}

		output, err := executor.Execute(cmdArgs...)
		if err != nil {
			return toolError(ErrCodeExecutionFailed, fmt.Sprintf("Failed to get flow execution logs: %s", output)), nil
		}

		// If requesting last log only, no pagination needed — wrap in list output.
		if last {
			var entry LogEntry
			if err := json.Unmarshal([]byte(output), &entry); err != nil {
				// Return raw output if we can't parse it.
				return mcp.NewToolResultText(output), nil
			}
			result := LogListOutput{
				History:    []LogEntry{entry},
				TotalCount: 1,
			}
			jsonData, _ := json.Marshal(result)
			return mcp.NewToolResultText(string(jsonData)), nil
		}

		// Parse the CLI list output and apply pagination.
		var cliOutput struct {
			History []LogEntry `json:"history"`
		}
		if err := json.Unmarshal([]byte(output), &cliOutput); err != nil {
			return mcp.NewToolResultText(output), nil
		}

		page, nextCursor, totalCount, err := paginate(cliOutput.History, cursor, defaultPageSize)
		if err != nil {
			return toolError(ErrCodeInvalidInput, err.Error()), nil
		}

		result := LogListOutput{
			History:    page,
			NextCursor: nextCursor,
			TotalCount: totalCount,
		}
		jsonData, _ := json.Marshal(result)
		return mcp.NewToolResultText(string(jsonData)), nil
	}
}

func syncStateHandler(srv *server.MCPServer, executor CommandExecutor) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		var progressToken any
		if request.Params.Meta != nil {
			progressToken = request.Params.Meta.ProgressToken
		}

		sendProgress(srv, ctx, progressToken, 0, 1, "Syncing state")
		output, err := executor.ExecuteContext(ctx, "sync")

		if ctx.Err() != nil {
			return toolError(ErrCodeCancelled, "sync was cancelled"), nil
		}

		if err != nil {
			return toolError(ErrCodeExecutionFailed, fmt.Sprintf("Failed to sync flow's state: %s", output)), nil
		}

		sendProgress(srv, ctx, progressToken, 1, 1, "Complete")
		srv.SendNotificationToAllClients("notifications/resources/list_changed", nil)

		result := SyncOutput{Output: output}
		jsonData, _ := json.Marshal(result)
		return mcp.NewToolResultText(string(jsonData)), nil
	}
}
