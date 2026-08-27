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

	"github.com/flowexec/flow/v2/internal/validation"
	"github.com/flowexec/flow/v2/pkg/filesystem"
	"github.com/flowexec/flow/v2/types/executable"
)

func addExecutableTools(srv *server.MCPServer, executor CommandExecutor) {
	getExecutable := mcp.NewTool("get_executable",
		mcp.WithDescription("Get the full definition of a specific flow workflow — what command it runs, "+
			"what parameters it accepts, and what secrets it requires. "+
			"Use before executing an unfamiliar executable or when debugging a failure."),
		mcp.WithString("executable_verb", mcp.Required(),
			mcp.Description("Executable verb (e.g. run, exec, build, test, deploy). "+
				"Validated server-side; see the flow docs for the full verb list.")),
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
		mcp.WithDescription("Discover available workflows across all workspaces. Call this before running shell "+
			"commands for build, test, deploy, lint, or any dev task — to check whether a flow executable "+
			"already handles it. Returns names, verbs, descriptions, and tags."),
		mcp.WithString("workspace", mcp.Description("Workspace name (optional)")),
		mcp.WithString("namespace", mcp.Description("Namespace filter (optional)")),
		mcp.WithString("verb", mcp.Description("Verb filter (optional)")),
		mcp.WithString("keyword", mcp.Description("Keyword filter (optional)")),
		mcp.WithString("tag", mcp.Description("Tag filter (optional)")),
		mcp.WithString("cursor", mcp.Description("Pagination cursor for next page of results")),
		mcp.WithString("dir", mcp.Description(
			"Directory to resolve the workspace from. Set this to the directory you are working in to see the "+
				"executables of the nearest workspace at or above it, even one that is not registered.")),
	)
	listExecutables.Annotations = mcp.ToolAnnotation{
		Title:           "List executables",
		DestructiveHint: boolPtr(false), ReadOnlyHint: boolPtr(true),
		IdempotentHint: boolPtr(true), OpenWorldHint: boolPtr(true),
	}
	srv.AddTool(listExecutables, listExecutablesHandler(executor))

	executeFlow := mcp.NewTool("execute",
		mcp.WithDescription("Run a flow workflow by verb and optional ID. Prefer this over direct shell commands "+
			"for build, test, deploy, lint, generate, and other dev tasks — flow handles environment setup, "+
			"secret injection, retries, and logging automatically."),
		mcp.WithString("executable_verb", mcp.Required(),
			mcp.Description("Executable verb (e.g. run, exec, build, test, deploy). "+
				"Validated server-side; see the flow docs for the full verb list.")),
		mcp.WithString("executable_id",
			mcp.Pattern(`^([a-zA-Z0-9_-]+(/[a-zA-Z0-9_-]+)?:)?[a-zA-Z0-9_-]+$`),
			mcp.Description(
				"Executable ID (workspace/namespace:name or just name if using the current workspace/namespace). "+
					"If the executable does not have a name, you can specify just the workspace (`ws/`), namespace (`ns:`) "+
					"both (`ws/ns:`) or neither if the current workspace/namespace should be used.")),
		mcp.WithString("args", mcp.Description("Arguments to pass")),
		mcp.WithString("dir", mcp.Description(
			"Directory to resolve the workspace from and run in. Set this to the directory you are working in "+
				"— a git worktree or a freshly cloned repo — and flow uses the nearest workspace at or above it, "+
				"registered or not. Defaults to the server's own directory, which is often not yours.")),
		mcp.WithString("workspace", mcp.Description(
			"Workspace to run in, by registered name or by path. Overrides `dir` resolution. Does not change "+
				"the global current workspace.")),
		mcp.WithBoolean("sync", mcp.Description("Sync executable changes before execution")),
		mcp.WithOutputSchema[ExecutionOutput](),
	)
	executeFlow.Annotations = mcp.ToolAnnotation{
		Title:        "Execute executable",
		ReadOnlyHint: boolPtr(false), DestructiveHint: boolPtr(true),
		IdempotentHint: boolPtr(false), OpenWorldHint: boolPtr(true),
	}
	srv.AddTool(executeFlow, executeFlowHandler(srv, executor))

	addRunCommandTool(srv, executor)
	addRunPythonTool(srv, executor)
	addRunExecutableTool(srv, executor)

	writeFlowfile := mcp.NewTool("write_flowfile",
		mcp.WithDescription("Create or update a .flow workflow file. Use when the user wants to add or modify "+
			"automation — builds, tests, deploys, scripts. Validates the YAML against the schema before "+
			"writing, then refreshes flow's executable cache so the change is immediately visible to "+
			"list_executables/get_executable. Prefer this over writing YAML files directly."),
		mcp.WithString("path", mcp.Required(),
			mcp.Description("Absolute or workspace-relative path for the flowfile (must end in .flow or .flow.yaml)")),
		mcp.WithString("content", mcp.Required(),
			mcp.Description("Full YAML content of the flowfile")),
		mcp.WithBoolean("overwrite",
			mcp.Description("Whether to overwrite an existing file (default: false)")),
	)
	writeFlowfile.Annotations = mcp.ToolAnnotation{
		Title:           "Write a flow file",
		DestructiveHint: boolPtr(true), ReadOnlyHint: boolPtr(false),
		IdempotentHint: boolPtr(false), OpenWorldHint: boolPtr(false),
	}
	srv.AddTool(writeFlowfile, writeFlowfileHandler(srv, executor))
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
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
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

		output, err := executor.ExecuteContext(withRunDir(ctx, request.GetString("dir", "")), cmdArgs...)
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
		return mcp.NewToolResultStructured(result, string(jsonData)), nil
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
		if ws := request.GetString("workspace", ""); ws != "" {
			cmdArgs = append(cmdArgs, "--workspace", ws)
		}
		if syncFlag {
			cmdArgs = append(cmdArgs, "--sync")
		}

		// Capture the MCP caller's identity so the resulting execution record records who ran it.
		ctx = withProvenance(ctx, mcpProvenance(ctx))
		// The subprocess resolves its workspace from its working directory, so the caller's
		// directory has to reach it — the server's own is rarely the right one.
		ctx = withRunDir(ctx, request.GetString("dir", ""))

		sendProgress(srv, ctx, progressToken, 0, 2, "Preparing execution")
		output, err := executor.ExecuteContext(ctx, cmdArgs...)

		if ctx.Err() != nil {
			return toolError(ErrCodeCancelled, "execution was cancelled"), nil
		}

		sendProgress(srv, ctx, progressToken, 1, 2, "Processing result")

		if err != nil {
			ref := strings.Join([]string{executableVerb, executableID}, " ")
			capped, _ := capOutput(output)
			return toolError(ErrCodeExecutionFailed, fmt.Sprintf("%s execution failed: %s", ref, capped)), nil
		}

		sendProgress(srv, ctx, progressToken, 2, 2, "Complete")

		capped, truncated := capOutput(output)
		result := ExecutionOutput{Output: capped, Truncated: truncated}
		jsonData, _ := json.Marshal(result)
		return mcp.NewToolResultStructured(result, string(jsonData)), nil
	}
}

func addRunCommandTool(srv *server.MCPServer, executor CommandExecutor) {
	runCommand := mcp.NewTool("run_command",
		mcp.WithDescription("Run one or more shell commands through flow instead of a raw shell tool. Commands run "+
			"with the current workspace's environment and secrets, output is captured to flow's logs, and each run "+
			"is recorded in execution history with provenance (visible via get_execution_logs / `flow logs`). "+
			"Prefer this for build, test, run, and one-off commands so the work is observable and reproducible. "+
			"Pass `commands` (with `mode`) to run several commands in a single call."),
		mcp.WithString("command",
			mcp.Description("A single shell command to run. Provide this or `commands`.")),
		mcp.WithArray("commands", mcp.WithStringItems(),
			mcp.Description("Multiple shell commands to run in one call (as a serial or parallel batch — see `mode`). "+
				"Provide this or `command`.")),
		mcp.WithString("mode",
			mcp.Description("How to run `commands`: 'serial' (default, stop on first failure) or 'parallel'.")),
		mcp.WithString("label",
			mcp.Description("Short human-readable label describing what the command(s) do (recorded in history).")),
		mcp.WithString("dir",
			mcp.Description("Working directory for the command (defaults to the current directory).")),
		mcp.WithString("workspace",
			mcp.Description("Workspace whose environment to use for this run. Defaults to the workspace containing "+
				"the working directory, then the current workspace. Does not change the global current workspace.")),
		mcp.WithBoolean("sync", mcp.Description("Sync flow cache and workspaces before running.")),
		mcp.WithOutputSchema[ExecutionOutput](),
	)
	runCommand.Annotations = mcp.ToolAnnotation{
		Title:        "Run ad-hoc shell command(s) through flow",
		ReadOnlyHint: boolPtr(false), DestructiveHint: boolPtr(true),
		IdempotentHint: boolPtr(false), OpenWorldHint: boolPtr(true),
	}
	srv.AddTool(runCommand, runCommandHandler(srv, executor))
}

func runCommandHandler(srv *server.MCPServer, executor CommandExecutor) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		var commands []string
		if c := request.GetString("command", ""); c != "" {
			commands = append(commands, c)
		}
		commands = append(commands, request.GetStringSlice("commands", nil)...)
		if len(commands) == 0 {
			return toolError(ErrCodeInvalidInput, "command or commands is required"), nil
		}

		cmdArgs := []string{"exec"}
		for _, c := range commands {
			cmdArgs = append(cmdArgs, "--cmd", c)
		}
		if mode := request.GetString("mode", ""); mode != "" {
			cmdArgs = append(cmdArgs, "--mode", mode)
		}
		if label := request.GetString("label", ""); label != "" {
			cmdArgs = append(cmdArgs, "--label", label)
		}
		if dir := request.GetString("dir", ""); dir != "" {
			cmdArgs = append(cmdArgs, "--dir", dir)
		}
		if ws := request.GetString("workspace", ""); ws != "" {
			cmdArgs = append(cmdArgs, "--workspace", ws)
		}
		if request.GetBool("sync", false) {
			cmdArgs = append(cmdArgs, "--sync")
		}

		return runTransientTool(ctx, srv, request, executor, cmdArgs, "command failed")
	}
}

func addRunPythonTool(srv *server.MCPServer, executor CommandExecutor) {
	runPython := mcp.NewTool("run_python",
		mcp.WithDescription("Run Python code through flow instead of a raw shell tool or a scratch file. "+
			"The script runs with the current workspace's environment and secrets, output is captured to "+
			"flow's logs, and the run is recorded in execution history with provenance (visible via "+
			"get_execution_logs / `flow logs`). Prefer this over shelling out to `python -c` or writing a "+
			"temporary .py file. flow picks the workspace's virtualenv when there is one, so imports resolve "+
			"against the project's installed dependencies."),
		mcp.WithString("code", mcp.Required(),
			mcp.Description("The Python source to run. Multi-line scripts are fine — the code runs from a "+
				"file, so tracebacks report real line numbers. Read parameters and secrets from os.environ.")),
		mcp.WithString("label",
			mcp.Description("Short human-readable label describing what the script does (recorded in history).")),
		mcp.WithString("dir",
			mcp.Description("Working directory for the script (defaults to the current directory). This also "+
				"determines which workspace's environment and virtualenv are used.")),
		mcp.WithString("workspace",
			mcp.Description("Workspace whose environment to use for this run. Defaults to the workspace containing "+
				"the working directory, then the current workspace. Does not change the global current workspace.")),
		mcp.WithBoolean("sync", mcp.Description("Sync flow cache and workspaces before running.")),
		mcp.WithOutputSchema[ExecutionOutput](),
	)
	runPython.Annotations = mcp.ToolAnnotation{
		Title:        "Run Python through flow",
		ReadOnlyHint: boolPtr(false), DestructiveHint: boolPtr(true),
		IdempotentHint: boolPtr(false), OpenWorldHint: boolPtr(true),
	}
	srv.AddTool(runPython, runPythonHandler(srv, executor))
}

func runPythonHandler(srv *server.MCPServer, executor CommandExecutor) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		code, err := request.RequireString("code")
		if err != nil || code == "" {
			return toolError(ErrCodeInvalidInput, "code is required"), nil
		}

		cmdArgs := []string{"exec", "--interpreter", "python", "--cmd", code}
		if label := request.GetString("label", ""); label != "" {
			cmdArgs = append(cmdArgs, "--label", label)
		}
		if dir := request.GetString("dir", ""); dir != "" {
			cmdArgs = append(cmdArgs, "--dir", dir)
		}
		if ws := request.GetString("workspace", ""); ws != "" {
			cmdArgs = append(cmdArgs, "--workspace", ws)
		}
		if request.GetBool("sync", false) {
			cmdArgs = append(cmdArgs, "--sync")
		}

		return runTransientTool(ctx, srv, request, executor, cmdArgs, "python script failed")
	}
}

func addRunExecutableTool(srv *server.MCPServer, executor CommandExecutor) {
	runExecutable := mcp.NewTool("run_executable",
		mcp.WithDescription("Run a transient executable of ANY type from an inline definition, without saving a "+
			".flow file. Use this when a single command (run_command) isn't enough — e.g. a `serial`/`parallel` "+
			"batch of steps, an HTTP `request`, or a `render`/`launch`. The spec is the same shape as one entry "+
			"under a flowfile's `executables:` list. Runs with workspace env/secrets and is recorded in history. "+
			"Author non-trivial specs against the flowfile schema (see get_info schemaUrls.flowFile)."),
		mcp.WithString("spec", mcp.Required(),
			mcp.Description("A single executable definition as YAML or JSON (e.g. {\"verb\":\"run\",\"serial\":"+
				"{\"execs\":[{\"cmd\":\"...\"},{\"cmd\":\"...\"}]}}).")),
		mcp.WithString("label",
			mcp.Description("Short human-readable label recorded in history (defaults to the executable's name).")),
		mcp.WithString("dir",
			mcp.Description("Directory to resolve the workspace from and run in (defaults to the server's directory). "+
				"Set this to the directory you are working in — flow uses the nearest workspace at or above it.")),
		mcp.WithString("workspace",
			mcp.Description("Workspace whose environment to use for this run. Defaults to the workspace containing "+
				"the working directory, then the current workspace. Does not change the global current workspace.")),
		mcp.WithBoolean("sync", mcp.Description("Sync flow cache and workspaces before running.")),
		mcp.WithOutputSchema[ExecutionOutput](),
	)
	runExecutable.Annotations = mcp.ToolAnnotation{
		Title:        "Run a transient executable of any type",
		ReadOnlyHint: boolPtr(false), DestructiveHint: boolPtr(true),
		IdempotentHint: boolPtr(false), OpenWorldHint: boolPtr(true),
	}
	srv.AddTool(runExecutable, runExecutableHandler(srv, executor))
}

func runExecutableHandler(srv *server.MCPServer, executor CommandExecutor) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		spec, err := request.RequireString("spec")
		if err != nil {
			return toolError(ErrCodeInvalidInput, "spec is required"), nil
		}
		cmdArgs := []string{"exec", "--spec", spec}
		if label := request.GetString("label", ""); label != "" {
			cmdArgs = append(cmdArgs, "--label", label)
		}
		if ws := request.GetString("workspace", ""); ws != "" {
			cmdArgs = append(cmdArgs, "--workspace", ws)
		}
		if request.GetBool("sync", false) {
			cmdArgs = append(cmdArgs, "--sync")
		}
		return runTransientTool(ctx, srv, request, executor, cmdArgs, "executable failed")
	}
}

// runTransientTool shells out to `flow exec` for the transient run tools (run_command, run_executable),
// tagging the run with MCP provenance and returning a structured result.
func runTransientTool(
	ctx context.Context, srv *server.MCPServer, request mcp.CallToolRequest,
	executor CommandExecutor, cmdArgs []string, failMsg string,
) (*mcp.CallToolResult, error) {
	var progressToken any
	if request.Params.Meta != nil {
		progressToken = request.Params.Meta.ProgressToken
	}

	// Tag the run as MCP-originated with the caller's identity.
	ctx = withProvenance(ctx, mcpProvenance(ctx))
	// `dir` is forwarded as --dir for the command itself; it also has to be the subprocess's
	// working directory so workspace resolution starts from where the caller is working.
	ctx = withRunDir(ctx, request.GetString("dir", ""))

	sendProgress(srv, ctx, progressToken, 0, 2, "Preparing execution")
	output, err := executor.ExecuteContext(ctx, cmdArgs...)

	if ctx.Err() != nil {
		return toolError(ErrCodeCancelled, "execution was cancelled"), nil
	}

	sendProgress(srv, ctx, progressToken, 1, 2, "Processing result")

	if err != nil {
		capped, _ := capOutput(output)
		return toolError(ErrCodeExecutionFailed, fmt.Sprintf("%s: %s", failMsg, capped)), nil
	}

	sendProgress(srv, ctx, progressToken, 2, 2, "Complete")

	capped, truncated := capOutput(output)
	result := ExecutionOutput{Output: capped, Truncated: truncated}
	jsonData, _ := json.Marshal(result)
	return mcp.NewToolResultStructured(result, string(jsonData)), nil
}

func writeFlowfileHandler(srv *server.MCPServer, executor CommandExecutor) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
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

		absPath, toolErr := resolveFlowfilePath(path)
		if toolErr != nil {
			return toolErr, nil
		}

		if toolErr := checkFlowfileOverwrite(absPath, overwrite); toolErr != nil {
			return toolErr, nil
		}

		if toolErr := validateFlowfileContent(content); toolErr != nil {
			return toolErr, nil
		}

		// Parse into typed struct for writing.
		var flowFile executable.FlowFile
		if err := yaml.Unmarshal([]byte(content), &flowFile); err != nil {
			return toolError(ErrCodeValidationFailed, fmt.Sprintf("invalid flowfile YAML: %s", err)), nil
		}

		if err := filesystem.WriteFlowFile(absPath, &flowFile); err != nil {
			return toolError(ErrCodeInternal, fmt.Sprintf("failed to write flowfile: %s", err)), nil
		}

		var execNames []string
		for _, exec := range flowFile.Executables {
			execNames = append(execNames, exec.Name)
		}

		// write_flowfile is the one executable-mutating tool that writes directly rather than
		// shelling to the CLI, so nothing else refreshes the persisted executable cache — without
		// this, a list_executables/get_executable call right after would still see stale state.
		// Best-effort: the file is already written and valid, so a sync failure here doesn't
		// invalidate the write itself.
		syncFailed := false
		if _, err := executor.ExecuteContext(ctx, "sync"); err != nil {
			syncFailed = true
		}

		srv.SendNotificationToAllClients("notifications/resources/list_changed", nil)

		output := WriteFlowFileOutput{
			Path:        absPath,
			Executables: execNames,
			Overwritten: overwrite,
			SyncFailed:  syncFailed,
		}
		jsonData, _ := json.Marshal(output)
		return mcp.NewToolResultStructured(output, string(jsonData)), nil
	}
}

func resolveFlowfilePath(path string) (string, *mcp.CallToolResult) {
	if filepath.IsAbs(path) {
		return path, nil
	}
	cfg, err := filesystem.LoadConfig()
	if err != nil {
		return "", toolError(ErrCodeInternal, fmt.Sprintf("failed to load config: %s", err))
	}
	// Resolve against the workspace flow would actually run in, which may be one discovered from
	// the working directory rather than the one persisted in the config.
	res, err := filesystem.ResolveWorkspace(cfg, filesystem.ResolveOptions{})
	if err != nil {
		return "", toolError(ErrCodeInvalidInput, err.Error())
	}
	if res != nil {
		return filepath.Join(res.Path, path), nil
	}
	return path, nil
}

func checkFlowfileOverwrite(absPath string, overwrite bool) *mcp.CallToolResult {
	if overwrite {
		return nil
	}
	if _, err := os.Stat(absPath); err == nil {
		return toolError(ErrCodeValidationFailed,
			fmt.Sprintf("file already exists at %s (use overwrite=true to replace)", absPath))
	}
	return nil
}

func validateFlowfileContent(content string) *mcp.CallToolResult {
	result, err := validation.ValidateBytes([]byte(content), validation.FileTypeFlowFile, false)
	if err != nil {
		return toolError(ErrCodeInternal, fmt.Sprintf("validation error: %s", err))
	}
	if !result.Valid {
		issueStrs := make([]string, 0, len(result.Errors))
		for _, e := range result.Errors {
			issueStrs = append(issueStrs, e.String())
		}
		return toolError(ErrCodeValidationFailed,
			fmt.Sprintf("invalid flowfile: %s", strings.Join(issueStrs, "; ")))
	}
	return nil
}
