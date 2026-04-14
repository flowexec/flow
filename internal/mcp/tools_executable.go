//nolint:nilerr
package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"gopkg.in/yaml.v3"

	"github.com/flowexec/flow/pkg/filesystem"
	"github.com/flowexec/flow/types/executable"
)

func addExecutableTools(srv *server.MCPServer, executor CommandExecutor) {
	getExecutable := mcp.NewTool("get_executable",
		mcp.WithDescription("Get detailed information about an executable"),
		mcp.WithString("executable_verb", mcp.Required(),
			mcp.Enum(executable.SortedValidVerbs()...),
			mcp.Description("Executable verb")),
		mcp.WithString("executable_id",
			mcp.Pattern(`^([a-zA-Z0-9_-]+(/[a-zA-Z0-9_-]+)?:)?[a-zA-Z0-9_-]+$`),
			mcp.Description("Executable ID (workspace/namespace:name or just name if using the current workspace/namespace)")),
		mcp.WithOutputSchema[ExecutableOutput](),
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
		mcp.WithString("cursor", mcp.Description("Pagination cursor for next page of results")),
		mcp.WithOutputSchema[ExecutableListOutput](),
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
		mcp.WithOutputSchema[ExecutionOutput](),
	)
	executeFlow.Annotations = mcp.ToolAnnotation{
		Title:        "Execute executable",
		ReadOnlyHint: boolPtr(false), DestructiveHint: boolPtr(true),
		IdempotentHint: boolPtr(false), OpenWorldHint: boolPtr(true),
	}
	srv.AddTool(executeFlow, executeFlowHandler(srv, executor))

	writeFlowfile := mcp.NewTool("write_flowfile",
		mcp.WithDescription("Create or update a flow file with YAML content. Validates the YAML before writing."),
		mcp.WithString("path", mcp.Required(),
			mcp.Description("Absolute or workspace-relative path for the flowfile (must end in .flow or .flow.yaml)")),
		mcp.WithString("content", mcp.Required(),
			mcp.Description("Full YAML content of the flowfile")),
		mcp.WithBoolean("overwrite",
			mcp.Description("Whether to overwrite an existing file (default: false)")),
		mcp.WithOutputSchema[WriteFlowFileOutput](),
	)
	writeFlowfile.Annotations = mcp.ToolAnnotation{
		Title:           "Write a flow file",
		DestructiveHint: boolPtr(true), ReadOnlyHint: boolPtr(false),
		IdempotentHint: boolPtr(false), OpenWorldHint: boolPtr(false),
	}
	srv.AddTool(writeFlowfile, writeFlowfileHandler(srv))
}

func getExecutableHandler(executor CommandExecutor) server.ToolHandlerFunc {
	return func(_ context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		executableVerb, err := request.RequireString("executable_verb")
		if err != nil {
			return toolError(ErrCodeInvalidInput, "executable_verb is required"), nil
		}
		executableID := request.GetString("executable_id", "")

		cmdArgs := []string{"browse", "--output", "json", executableVerb}
		if executableID != "" {
			cmdArgs = append(cmdArgs, executableID)
		}

		output, err := executor.Execute(cmdArgs...)
		if err != nil {
			ref := strings.Join([]string{executableVerb, executableID}, " ")
			return toolError(ErrCodeNotFound, fmt.Sprintf("Failed to get executable %s: %s", ref, output)), nil
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
		cursor := request.GetString("cursor", "")

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
			return toolError(ErrCodeExecutionFailed, fmt.Sprintf("Failed to list executables: %s", output)), nil
		}

		var cliOutput struct {
			Executables []ExecutableOutput `json:"executables"`
		}
		if err := json.Unmarshal([]byte(output), &cliOutput); err != nil {
			return mcp.NewToolResultText(output), nil
		}

		page, nextCursor, totalCount, err := paginate(cliOutput.Executables, cursor, defaultPageSize)
		if err != nil {
			return toolError(ErrCodeInvalidInput, err.Error()), nil
		}

		result := ExecutableListOutput{
			Executables: page,
			NextCursor:  nextCursor,
			TotalCount:  totalCount,
		}
		jsonData, _ := json.Marshal(result)
		return mcp.NewToolResultText(string(jsonData)), nil
	}
}

func executeFlowHandler(srv *server.MCPServer, executor CommandExecutor) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		executableVerb, err := request.RequireString("executable_verb")
		if err != nil {
			return toolError(ErrCodeInvalidInput, "executable_verb is required"), nil
		}
		executableID := request.GetString("executable_id", "")

		args := request.GetString("args", "")
		syncFlag := request.GetBool("sync", false)
		var progressToken any
		if request.Params.Meta != nil {
			progressToken = request.Params.Meta.ProgressToken
		}

		cmdArgs := []string{executableVerb}
		if executableID != "" {
			cmdArgs = append(cmdArgs, executableID)
		}
		if args != "" {
			cmdArgs = append(cmdArgs, strings.Fields(args)...)
		}
		if syncFlag {
			cmdArgs = append(cmdArgs, "--sync")
		}

		sendProgress(srv, ctx, progressToken, 0, 2, "Preparing execution")
		output, err := executor.ExecuteContext(ctx, cmdArgs...)

		if ctx.Err() != nil {
			return toolError(ErrCodeCancelled, "execution was cancelled"), nil
		}

		sendProgress(srv, ctx, progressToken, 1, 2, "Processing result")

		if err != nil {
			ref := strings.Join([]string{executableVerb, executableID}, " ")
			return toolError(ErrCodeExecutionFailed, fmt.Sprintf("%s execution failed: %s", ref, output)), nil
		}

		sendProgress(srv, ctx, progressToken, 2, 2, "Complete")

		result := ExecutionOutput{Output: output}
		jsonData, _ := json.Marshal(result)
		return mcp.NewToolResultText(string(jsonData)), nil
	}
}

func writeFlowfileHandler(srv *server.MCPServer) server.ToolHandlerFunc {
	return func(_ context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		path, err := request.RequireString("path")
		if err != nil {
			return toolError(ErrCodeInvalidInput, "path is required"), nil
		}
		content, err := request.RequireString("content")
		if err != nil {
			return toolError(ErrCodeInvalidInput, "content is required"), nil
		}
		overwrite := request.GetBool("overwrite", false)

		// Validate file extension
		if !strings.HasSuffix(path, ".flow") && !strings.HasSuffix(path, ".flow.yaml") {
			return toolError(ErrCodeValidationFailed, "path must end in .flow or .flow.yaml"), nil
		}

		// Resolve absolute path
		absPath := path
		if !filepath.IsAbs(path) {
			cfg, err := filesystem.LoadConfig()
			if err != nil {
				return toolError(ErrCodeInternal, fmt.Sprintf("failed to load config: %s", err)), nil
			}
			if cfg.CurrentWorkspace != "" {
				if wsPath, ok := cfg.Workspaces[cfg.CurrentWorkspace]; ok {
					absPath = filepath.Join(wsPath, path)
				}
			}
		}

		// Check if file exists when overwrite is false
		if !overwrite {
			if _, err := os.Stat(absPath); err == nil {
				msg := fmt.Sprintf("file already exists at %s (use overwrite=true to replace)", absPath)
				return toolError(ErrCodeValidationFailed, msg), nil
			}
		}

		// Validate YAML content by parsing into FlowFile
		var flowFile executable.FlowFile
		if err := yaml.Unmarshal([]byte(content), &flowFile); err != nil {
			return toolError(ErrCodeValidationFailed, fmt.Sprintf("invalid flowfile YAML: %s", err)), nil
		}

		// Write the flowfile
		if err := filesystem.WriteFlowFile(absPath, &flowFile); err != nil {
			return toolError(ErrCodeInternal, fmt.Sprintf("failed to write flowfile: %s", err)), nil
		}

		// Collect executable names for the summary
		var execNames []string
		for _, exec := range flowFile.Executables {
			execNames = append(execNames, exec.Name)
		}

		srv.SendNotificationToAllClients("notifications/resources/list_changed", nil)

		output := WriteFlowFileOutput{
			Path:        absPath,
			Executables: execNames,
			Overwritten: overwrite,
		}
		jsonData, _ := json.Marshal(output)
		return mcp.NewToolResultText(string(jsonData)), nil
	}
}
