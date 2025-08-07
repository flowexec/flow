//nolint:nilerr
package mcp

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/pkg/errors"

	"github.com/flowexec/flow/internal/filesystem"
	"github.com/flowexec/flow/types/executable"
)

var (
	//go:embed resources/concepts-guide.md
	conceptsMD string

	//go:embed resources/file-types-guide.md
	fileTypesMD string

	// The below schemas are updated by the docsgen tool. We embed instead of fetching to avoid unnecessary network
	// requests and to ensure that the MCP server always has the schema needed for the current CLI version.

	//go:embed resources/flowfile_schema.json
	flowFileSchema string

	//go:embed resources/template_schema.json
	templateSchema string

	//go:embed resources/workspace_schema.json
	workspaceSchema string
)

//nolint:funlen
func addServerTools(srv *server.MCPServer, executor CommandExecutor) {
	// Ideally this information would just be exposed via resources but many MCP clients don't support resources.
	// This implementation should be revisited in the future.
	// See https://modelcontextprotocol.io/clients
	getFlowInfo := mcp.NewTool("get_info",
		mcp.WithDescription(
			"Get information about flow, it's usage, and the current workflow execution context. "+
				"This includes file JSON schemas for flow executable, template, and workspace files, concepts guides, "+
				"and the current user configuration and state details."))
	getFlowInfo.Annotations = mcp.ToolAnnotation{
		Title:           "Get flow information and current context",
		DestructiveHint: boolPtr(false), ReadOnlyHint: boolPtr(true),
		IdempotentHint: boolPtr(false), OpenWorldHint: boolPtr(false),
	}
	srv.AddTool(getFlowInfo, getInfoHandler)

	getWorkspace := mcp.NewTool("get_workspace",
		mcp.WithString("workspace_name", mcp.Required(), mcp.Description("Registered workspace name")),
		mcp.WithDescription("Get details about a registered flow workspaces"),
	)
	getWorkspace.Annotations = mcp.ToolAnnotation{
		Title:           "Get a specific workspace by name",
		DestructiveHint: boolPtr(false), ReadOnlyHint: boolPtr(true),
		IdempotentHint: boolPtr(true), OpenWorldHint: boolPtr(true),
	}
	srv.AddTool(getWorkspace, getWorkspaceHandler(executor))

	listWorkspaces := mcp.NewTool("list_workspaces",
		mcp.WithDescription("List all registered flow workspaces"),
	)
	listWorkspaces.Annotations = mcp.ToolAnnotation{
		Title:           "List workspaces",
		DestructiveHint: boolPtr(false), ReadOnlyHint: boolPtr(true),
		IdempotentHint: boolPtr(true), OpenWorldHint: boolPtr(true),
	}
	srv.AddTool(listWorkspaces, listWorkspacesHandler(executor))

	switchWorkspace := mcp.NewTool("switch_workspace",
		mcp.WithString("workspace_name", mcp.Required(), mcp.Description("Registered workspace name")),
		mcp.WithDescription("Change the current workspace"),
	)
	switchWorkspace.Annotations = mcp.ToolAnnotation{
		Title:           "Change the current workspace",
		DestructiveHint: boolPtr(false), ReadOnlyHint: boolPtr(false),
		IdempotentHint: boolPtr(true), OpenWorldHint: boolPtr(false),
	}
	srv.AddTool(switchWorkspace, switchWorkspaceHandler(executor))

	getExecutable := mcp.NewTool("get_executable",
		mcp.WithDescription("Get detailed information about an executable"),
		mcp.WithString("executable_verb", mcp.Required(),
			mcp.Enum(executable.SortedValidVerbs()...),
			mcp.Description("Executable verb")),
		mcp.WithString("executable_id",
			mcp.Pattern(`^([a-zA-Z0-9_-]+(/[a-zA-Z0-9_-]+)?:)?[a-zA-Z0-9_-]+$`),
			mcp.Description("Executable ID (workspace/namespace:name or just name if using the current workspace/namespace)")),
	)
	getExecutable.Annotations = mcp.ToolAnnotation{
		Title:           "Get a specific executable by reference",
		DestructiveHint: boolPtr(false), ReadOnlyHint: boolPtr(true),
		IdempotentHint: boolPtr(true), OpenWorldHint: boolPtr(true),
	}
	srv.AddTool(getExecutable, getExecutableHandler(executor))

	listExecutables := mcp.NewTool("list_executables",
		mcp.WithDescription("List and filter executables across all workspaces"),
		mcp.WithString("workspace", mcp.Description("Workspace name (optional)")),
		mcp.WithString("namespace", mcp.Description("Namespace filter (optional)")),
		mcp.WithString("verb", mcp.Description("Verb filter (optional)")),
		mcp.WithString("keyword", mcp.Description("Keyword filter (optional)")),
		mcp.WithString("tag", mcp.Description("Tag filter (optional)")),
	)
	listExecutables.Annotations = mcp.ToolAnnotation{
		Title:           "List executables",
		DestructiveHint: boolPtr(false), ReadOnlyHint: boolPtr(true),
		IdempotentHint: boolPtr(true), OpenWorldHint: boolPtr(true),
	}
	srv.AddTool(listExecutables, listExecutablesHandler(executor))

	executeFlow := mcp.NewTool("execute",
		mcp.WithDescription("Execute a flow executable"),
		mcp.WithString("executable_verb", mcp.Required(),
			mcp.Enum(executable.SortedValidVerbs()...),
			mcp.Description("Executable verb")),
		mcp.WithString("executable_id",
			mcp.Pattern(`^([a-zA-Z0-9_-]+(/[a-zA-Z0-9_-]+)?:)?[a-zA-Z0-9_-]+$`),
			mcp.Description(
				"Executable ID (workspace/namespace:name or just name if using the current workspace/namespace). "+
					"If the executable does not have a name, you can specify just the workspace (`ws/`), namespace (`ns:`) "+
					"both (`ws/ns:`) or neither if the current workspace/namespace should be used.")),
		mcp.WithString("args", mcp.Description("Arguments to pass")),
		mcp.WithBoolean("sync", mcp.Description("Sync executable changes before execution")),
	)
	executeFlow.Annotations = mcp.ToolAnnotation{
		Title:        "Execute executable",
		ReadOnlyHint: boolPtr(false), DestructiveHint: boolPtr(true),
		IdempotentHint: boolPtr(false), OpenWorldHint: boolPtr(true),
	}
	srv.AddTool(executeFlow, executeFlowHandler(executor))

	getExecutionLogs := mcp.NewTool("get_execution_logs",
		mcp.WithDescription("Get a list of the recent flow execution logs"),
		mcp.WithBoolean("last", mcp.Description("Get only the last execution logs")))
	getExecutionLogs.Annotations = mcp.ToolAnnotation{
		Title:           "Get execution logs",
		DestructiveHint: boolPtr(false), ReadOnlyHint: boolPtr(true),
		IdempotentHint: boolPtr(true), OpenWorldHint: boolPtr(true),
	}
	srv.AddTool(getExecutionLogs, getExecutionLogsHandler(executor))

	sync := mcp.NewTool("sync_executables",
		mcp.WithDescription("Sync the flow workspace and executable state"))
	sync.Annotations = mcp.ToolAnnotation{
		Title:           "Sync executable and workspace state",
		DestructiveHint: boolPtr(false), ReadOnlyHint: boolPtr(false),
		IdempotentHint: boolPtr(false), OpenWorldHint: boolPtr(true),
	}
	srv.AddTool(sync, syncStateHandler(executor))
}

func getInfoHandler(_ context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	cfg, err := filesystem.LoadConfig()
	if err != nil {
		return nil, errors.Wrap(err, "failed to load user config")
	}
	cfg.SetDefaults()

	wsName, err := cfg.CurrentWorkspaceName()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get current workspace name")
	}

	infoData := map[string]interface{}{
		"usageGuides": map[string]interface{}{
			"concepts":  conceptsMD,
			"fileTypes": fileTypesMD,
		},
		"schemas": map[string]interface{}{
			"flowFileSchema":        flowFileSchema,
			"workspaceConfigSchema": workspaceSchema,
			"templateFileSchema":    templateSchema,
		},
		"currentContext": map[string]interface{}{
			"workspace":     wsName,
			"namespace":     cfg.CurrentNamespace,
			"vault":         cfg.CurrentVault,
			"workspaceMode": cfg.WorkspaceMode,
			"workspacePath": cfg.Workspaces[cfg.CurrentWorkspace],
		},
	}
	jsonData, err := json.Marshal(infoData)
	if err != nil {
		return nil, err
	}

	return mcp.NewToolResultText(string(jsonData)), nil
}

func getWorkspaceHandler(executor CommandExecutor) server.ToolHandlerFunc {
	return func(_ context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		wsName, err := request.RequireString("workspace_name")
		if err != nil {
			return mcp.NewToolResultError("workspace_name is required"), nil
		}

		output, err := executor.Execute("workspace", "get", wsName, "--output", "json")
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to get workspaces %s: %s", wsName, output)), nil
		}

		return mcp.NewToolResultText(output), nil
	}
}

func listWorkspacesHandler(executor CommandExecutor) server.ToolHandlerFunc {
	return func(_ context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		output, err := executor.Execute("workspace", "list", "--output", "json")
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to list workspaces: %s", output)), nil
		}

		return mcp.NewToolResultText(output), nil
	}
}

func switchWorkspaceHandler(executor CommandExecutor) server.ToolHandlerFunc {
	return func(_ context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		wsName, err := request.RequireString("workspace_name")
		if err != nil {
			return mcp.NewToolResultError("workspace_name is required"), nil
		}

		output, err := executor.Execute("workspace", "switch", wsName)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to switch workspace to %s: %s", wsName, output)), nil
		}

		return mcp.NewToolResultText(output), nil
	}
}

func getExecutableHandler(executor CommandExecutor) server.ToolHandlerFunc {
	return func(_ context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		executableVerb, err := request.RequireString("executable_verb")
		if err != nil {
			return mcp.NewToolResultError("executable_verb is required"), nil
		}
		executableID := request.GetString("executable_id", "")

		cmdArgs := []string{"browse", "--output", "json", executableVerb}
		if executableID != "" {
			cmdArgs = append(cmdArgs, executableID)
		}

		output, err := executor.Execute(cmdArgs...)
		if err != nil {
			ref := strings.Join([]string{executableVerb, executableID}, " ")
			return mcp.NewToolResultError(fmt.Sprintf("Failed to get executable %s: %s", ref, output)), nil
		}

		return mcp.NewToolResultText(output), nil
	}
}

func listExecutablesHandler(executor CommandExecutor) server.ToolHandlerFunc {
	return func(_ context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		wsFilter := request.GetString("workspace", executable.WildcardWorkspace)
		nsFilter := request.GetString("namespace", executable.WildcardNamespace)
		verbFilter := request.GetString("verb", "")
		keywordFilter := request.GetString("keyword", "")
		tagFilter := request.GetString("tag", "")

		cmdArgs := []string{"browse", "--output", "json", "--workspace", wsFilter, "--namespace", nsFilter}
		if verbFilter != "" {
			cmdArgs = append(cmdArgs, "--verb", verbFilter)
		}
		if keywordFilter != "" {
			cmdArgs = append(cmdArgs, "--filter", keywordFilter)
		}
		if tagFilter != "" {
			cmdArgs = append(cmdArgs, "--tag", tagFilter)
		}

		output, err := executor.Execute(cmdArgs...)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to list executables: %s", output)), nil
		}

		return mcp.NewToolResultText(output), nil
	}
}

func executeFlowHandler(executor CommandExecutor) server.ToolHandlerFunc {
	return func(_ context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		executableVerb, err := request.RequireString("executable_verb")
		if err != nil {
			return mcp.NewToolResultError("executable_verb is required"), nil
		}
		executableID := request.GetString("executable_id", "")

		args := request.GetString("args", "")
		sync := request.GetBool("sync", false)

		cmdArgs := []string{executableVerb}
		if executableID != "" {
			cmdArgs = append(cmdArgs, executableID)
		}
		if args != "" {
			cmdArgs = append(cmdArgs, strings.Fields(args)...)
		}
		if sync {
			cmdArgs = append(cmdArgs, "--sync")
		}

		output, _ := executor.Execute(cmdArgs...)
		// if err != nil {
		// 	ref := strings.Join([]string{executableVerb, executableID}, " ")
		// 	return mcp.NewToolResultError(fmt.Sprintf("%s execution failed: %s", ref, output)), nil
		// }

		return mcp.NewToolResultText(output), nil
	}
}

func getExecutionLogsHandler(executor CommandExecutor) server.ToolHandlerFunc {
	return func(_ context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		last := request.GetBool("last", false)
		cmdArgs := []string{"logs", "--output", "json"}
		if last {
			cmdArgs = append(cmdArgs, "--last")
		}

		output, err := executor.Execute(cmdArgs...)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to get flow execution logs: %s", output)), nil
		}

		return mcp.NewToolResultText(output), nil
	}
}

func syncStateHandler(executor CommandExecutor) server.ToolHandlerFunc {
	return func(_ context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		output, err := executor.Execute("sync")
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to sync flow's state: %s", output)), nil
		}

		return mcp.NewToolResultText(output), nil
	}
}

func boolPtr(b bool) *bool {
	if b {
		return &b
	}
	return nil
}
