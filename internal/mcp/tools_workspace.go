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

func addWorkspaceTools(srv *server.MCPServer, executor CommandExecutor) {
	getWorkspace := mcp.NewTool("get_workspace",
		mcp.WithString("workspace_name", mcp.Required(), mcp.Description("Registered workspace name")),
		mcp.WithDescription("Get details about a registered flow workspaces"),
		mcp.WithOutputSchema[WorkspaceOutput](),
	)
	getWorkspace.Annotations = mcp.ToolAnnotation{
		Title:           "Get a specific workspace by name",
		DestructiveHint: boolPtr(false), ReadOnlyHint: boolPtr(true),
		IdempotentHint: boolPtr(true), OpenWorldHint: boolPtr(true),
	}
	srv.AddTool(getWorkspace, getWorkspaceHandler(executor))

	listWorkspaces := mcp.NewTool("list_workspaces",
		mcp.WithDescription("List all registered flow workspaces"),
		mcp.WithString("cursor", mcp.Description("Pagination cursor for next page of results")),
		mcp.WithOutputSchema[WorkspaceListOutput](),
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
		mcp.WithOutputSchema[SwitchWorkspaceOutput](),
	)
	switchWorkspace.Annotations = mcp.ToolAnnotation{
		Title:           "Change the current workspace",
		DestructiveHint: boolPtr(false), ReadOnlyHint: boolPtr(false),
		IdempotentHint: boolPtr(true), OpenWorldHint: boolPtr(false),
	}
	srv.AddTool(switchWorkspace, switchWorkspaceHandler(srv, executor))

	getWorkspaceConfig := mcp.NewTool("get_workspace_config",
		mcp.WithString("name", mcp.Required(), mcp.Description("Workspace name")),
		mcp.WithDescription("Get the full workspace configuration as structured JSON"),
		mcp.WithOutputSchema[WorkspaceConfigOutput](),
	)
	getWorkspaceConfig.Annotations = mcp.ToolAnnotation{
		Title:           "Get workspace configuration",
		DestructiveHint: boolPtr(false), ReadOnlyHint: boolPtr(true),
		IdempotentHint: boolPtr(true), OpenWorldHint: boolPtr(true),
	}
	srv.AddTool(getWorkspaceConfig, getWorkspaceConfigHandler())
}

func getWorkspaceHandler(executor CommandExecutor) server.ToolHandlerFunc {
	return func(_ context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		wsName, err := request.RequireString("workspace_name")
		if err != nil {
			return toolError(ErrCodeInvalidInput, "workspace_name is required"), nil
		}

		output, err := executor.Execute("workspace", "get", wsName, "--output", "json")
		if err != nil {
			return toolError(ErrCodeNotFound, fmt.Sprintf("Failed to get workspace %s: %s", wsName, output)), nil
		}

		// Validate output parses as workspace JSON; return as-is if valid.
		var ws WorkspaceOutput
		if err := json.Unmarshal([]byte(output), &ws); err != nil {
			return mcp.NewToolResultText(output), nil
		}

		jsonData, _ := json.Marshal(ws)
		return mcp.NewToolResultText(string(jsonData)), nil
	}
}

func listWorkspacesHandler(executor CommandExecutor) server.ToolHandlerFunc {
	return func(_ context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		cursor := request.GetString("cursor", "")

		output, err := executor.Execute("workspace", "list", "--output", "json")
		if err != nil {
			return toolError(ErrCodeExecutionFailed, fmt.Sprintf("Failed to list workspaces: %s", output)), nil
		}

		var cliOutput struct {
			Workspaces []WorkspaceOutput `json:"workspaces"`
		}
		if err := json.Unmarshal([]byte(output), &cliOutput); err != nil {
			return mcp.NewToolResultText(output), nil
		}

		page, nextCursor, totalCount, err := paginate(cliOutput.Workspaces, cursor, defaultPageSize)
		if err != nil {
			return toolError(ErrCodeInvalidInput, err.Error()), nil
		}

		result := WorkspaceListOutput{
			Workspaces: page,
			NextCursor: nextCursor,
			TotalCount: totalCount,
		}
		jsonData, _ := json.Marshal(result)
		return mcp.NewToolResultText(string(jsonData)), nil
	}
}

func switchWorkspaceHandler(srv *server.MCPServer, executor CommandExecutor) server.ToolHandlerFunc {
	return func(_ context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		wsName, err := request.RequireString("workspace_name")
		if err != nil {
			return toolError(ErrCodeInvalidInput, "workspace_name is required"), nil
		}

		output, err := executor.Execute("workspace", "switch", wsName)
		if err != nil {
			return toolError(ErrCodeNotFound, fmt.Sprintf("Failed to switch workspace to %s: %s", wsName, output)), nil
		}

		srv.SendNotificationToAllClients("notifications/resources/list_changed", nil)

		result := SwitchWorkspaceOutput{Output: output}
		jsonData, _ := json.Marshal(result)
		return mcp.NewToolResultText(string(jsonData)), nil
	}
}

func getWorkspaceConfigHandler() server.ToolHandlerFunc {
	return func(_ context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		name, err := request.RequireString("name")
		if err != nil {
			return toolError(ErrCodeInvalidInput, "name is required"), nil
		}

		cfg, err := filesystem.LoadConfig()
		if err != nil {
			return toolError(ErrCodeInternal, fmt.Sprintf("failed to load config: %s", err)), nil
		}

		wsPath, ok := cfg.Workspaces[name]
		if !ok {
			return toolError(ErrCodeNotFound, fmt.Sprintf("workspace %q not found", name)), nil
		}

		ws, err := filesystem.LoadWorkspaceConfig(name, wsPath)
		if err != nil {
			return toolError(ErrCodeNotFound, fmt.Sprintf("failed to load workspace config: %s", err)), nil
		}

		output := WorkspaceConfigOutput{
			WorkspaceOutput: WorkspaceOutput{
				Name:        ws.AssignedName(),
				Path:        ws.Location(),
				DisplayName: ws.DisplayName,
				Description: ws.Description,
				Tags:        ws.Tags,
				Annotations: map[string]string(ws.Annotations),
			},
		}
		if ws.Executables != nil {
			output.Executables = &ExecutableFilter{
				Included: ws.Executables.Included,
				Excluded: ws.Executables.Excluded,
			}
		}

		jsonData, _ := json.Marshal(output)
		return mcp.NewToolResultText(string(jsonData)), nil
	}
}
