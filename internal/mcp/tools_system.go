//nolint:nilerr
package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"github.com/flowexec/flow/v2/pkg/filesystem"
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
			"Bootstrap context about the flow environment. Returns the current workspace, "+
				"schema URLs for authoring .flow files, and the docs index (llms.txt). "+
				"Call this at the start of a session to understand the project's automation setup, "+
				"or whenever you need schema URLs to author or validate flow configuration."),
	)
	getFlowInfo.Annotations = mcp.ToolAnnotation{
		Title:           "Get flow information and current context",
		DestructiveHint: boolPtr(false), ReadOnlyHint: boolPtr(true),
		IdempotentHint: boolPtr(false), OpenWorldHint: boolPtr(false),
	}
	srv.AddTool(getFlowInfo, getInfoHandler)

	getExecutionLogs := mcp.NewTool("get_execution_logs",
		mcp.WithDescription("Retrieve metadata and output from recent flow executions. "+
			"Use when debugging a failed run or when the user asks about the results of a previous task. "+
			"By default only run metadata is returned (ref, status, exitCode, error) — set `tail`, `grep`, "+
			"or `content` to include the captured log output. Output is read live, so it works on "+
			"still-running executions too (you get a snapshot of what has been written so far). "+
			"Set `mine` to see only what this session has run so far."),
		mcp.WithBoolean("last", mcp.Description("Get only the last execution logs")),
		mcp.WithBoolean("mine", mcp.Description("Only return runs launched by this MCP session "+
			"(useful for reviewing what you have run so far).")),
		mcp.WithString("source", mcp.Description("Filter by run origin: 'cli' or 'mcp'.")),
		mcp.WithString("session", mcp.Description("Filter to a single provenance session ID.")),
		mcp.WithString("status", mcp.Description("Filter by status: running, completed, or failed.")),
		mcp.WithString("cursor", mcp.Description("Pagination cursor for next page of results")),
		mcp.WithBoolean("content", mcp.Description("Include each run's captured log output. Implied by "+
			"`tail` or `grep`. Output is always capped (see `max_bytes`) to protect the context window.")),
		mcp.WithNumber("tail", mcp.Description("Include only the last N lines of log output. Recommended "+
			"for debugging since errors surface at the end of a log. Implies `content`.")),
		mcp.WithString("grep", mcp.Description("Include only log lines matching this regular expression "+
			"(applied before `tail`). Use to pull just the error lines out of a large log. Implies `content`.")),
		mcp.WithNumber("max_bytes", mcp.Description(fmt.Sprintf(
			"Cap the log output returned per run to the last N bytes. Defaults to %d and is hard-capped "+
				"at %d to protect the context window.", defaultLogContentMaxBytes, maxLogContentMaxBytes))),
	)
	getExecutionLogs.Annotations = mcp.ToolAnnotation{
		Title:           "Get execution logs",
		DestructiveHint: boolPtr(false), ReadOnlyHint: boolPtr(true),
		IdempotentHint: boolPtr(true), OpenWorldHint: boolPtr(true),
	}
	srv.AddTool(getExecutionLogs, getExecutionLogsHandler(executor))

	sync := mcp.NewTool("sync_executables",
		mcp.WithDescription("Refresh the cached workspace and executable state. "+
			"Use when executables seem out of date or after adding new .flow files."),
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

	var wsName, wsPath string
	if len(cfg.Workspaces) > 0 {
		wsName, err = cfg.CurrentWorkspaceName()
		if err != nil {
			return toolError(ErrCodeInternal, fmt.Sprintf("failed to get current workspace name: %s", err)), nil
		}
		wsPath = cfg.Workspaces[wsName]
	}

	output := FlowInfoOutput{
		CurrentContext: CurrentContext{
			Workspace:     wsName,
			Namespace:     cfg.CurrentNamespace,
			Vault:         cfg.CurrentVaultName(),
			WorkspaceMode: string(cfg.WorkspaceMode),
			WorkspacePath: wsPath,
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

	return mcp.NewToolResultStructured(output, string(jsonData)), nil
}

// Log-content byte caps enforced server-side so an under-specified request can't flood the
// context window with an unbounded log dump.
const (
	defaultLogContentMaxBytes = 50_000
	maxLogContentMaxBytes     = 200_000
)

func getExecutionLogsHandler(executor CommandExecutor) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		last := request.GetBool("last", false)
		cursor := request.GetString("cursor", "")
		source := request.GetString("source", "")
		session := request.GetString("session", "")
		status := request.GetString("status", "")

		// `mine` scopes results to the calling MCP session's own runs so an agent can review
		// exactly what it launched — resolved from the live session, not client-supplied.
		if request.GetBool("mine", false) {
			prov := mcpProvenance(ctx)
			source = prov.Source
			if prov.Session != "" {
				session = prov.Session
			}
		}

		cmdArgs := []string{"logs", "--output", "json"}
		if last {
			cmdArgs = append(cmdArgs, "--last")
		}
		if source != "" {
			cmdArgs = append(cmdArgs, "--source", source)
		}
		if session != "" {
			cmdArgs = append(cmdArgs, "--session", session)
		}
		if status != "" {
			cmdArgs = append(cmdArgs, "--status", status)
		}
		cmdArgs = append(cmdArgs, logContentArgs(request)...)

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
			return mcp.NewToolResultStructured(result, string(jsonData)), nil
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
		return mcp.NewToolResultStructured(result, string(jsonData)), nil
	}
}

// logContentArgs translates the content-related request parameters into `flow logs` flags.
// It returns no flags unless content was requested (via content/tail/grep), and always
// enforces a byte cap — clamped into [1, maxLogContentMaxBytes] — so an under-specified
// request can't flood the context window with an unbounded log dump.
func logContentArgs(request mcp.CallToolRequest) []string {
	tail := request.GetInt("tail", 0)
	grep := request.GetString("grep", "")
	if !request.GetBool("content", false) && tail <= 0 && grep == "" {
		return nil
	}

	maxBytes := request.GetInt("max_bytes", 0)
	if maxBytes <= 0 {
		maxBytes = defaultLogContentMaxBytes
	}
	if maxBytes > maxLogContentMaxBytes {
		maxBytes = maxLogContentMaxBytes
	}

	args := []string{"--content", "--max-bytes", strconv.Itoa(maxBytes)}
	if tail > 0 {
		args = append(args, "--tail", strconv.Itoa(tail))
	}
	if grep != "" {
		args = append(args, "--grep", grep)
	}
	return args
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
		return mcp.NewToolResultStructured(result, string(jsonData)), nil
	}
}
