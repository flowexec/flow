//nolint:nilerr
package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"gopkg.in/yaml.v3"

	"github.com/flowexec/flow/v2/internal/validation"
	"github.com/flowexec/flow/v2/pkg/filesystem"
	"github.com/flowexec/flow/v2/types/executable"
)

// currentExecutableID names the nameless executable in the current workspace and namespace. The CLI
// reads its first positional as the executable ID, so args given without an ID need this in front.
const currentExecutableID = ":"

// plainLogArgs makes a run's output the command's own, without flow's timestamp and level prefixes.
// An executable that sets its own logMode still gets it.
var plainLogArgs = []string{"--log-mode", "text"}

const runWorkspaceDescription = "Workspace whose environment to use. Default: the one containing the " +
	"working directory, else the current one. Doesn't change the current workspace."

func addExecutableTools(srv *server.MCPServer, executor CommandExecutor) {
	getExecutable := mcp.NewTool("get_executable",
		mcp.WithDescription("Get one executable's full definition: what it runs, its args, params and secrets. "+
			"Use before running an unfamiliar executable or when debugging a failure."),
		mcp.WithString("executable_verb", mcp.Required(),
			mcp.Description("Executable verb, e.g. run, build, test, deploy.")),
		mcp.WithString("executable_id",
			mcp.Pattern(executable.ExecutableIDPattern),
			mcp.Description("Executable ID: `ws/ns:name`, or `name` in the current workspace and namespace. "+
				"Nested namespaces need the workspace: `ws/parent/child:name` (`./` for the current one).")),
	)
	getExecutable.Annotations = mcp.ToolAnnotation{
		Title:           "Get a specific executable by reference",
		DestructiveHint: boolPtr(false), ReadOnlyHint: boolPtr(true),
		IdempotentHint: boolPtr(true), OpenWorldHint: boolPtr(true),
	}
	srv.AddTool(getExecutable, getExecutableHandler(executor))

	listExecutables := mcp.NewTool("list_executables",
		mcp.WithDescription("List executables across workspaces. Check here before running a shell command for "+
			"a dev task (build, test, lint, deploy): flow may already have one. Returns names, verbs, short "+
			"descriptions and tags. Set include_args to also get each one's args and prompt params; "+
			"get_executable has the full definition."),
		mcp.WithString("workspace", mcp.Description("Workspace name")),
		mcp.WithString("namespace", mcp.Description("Namespace filter")),
		mcp.WithString("verb", mcp.Description("Verb filter")),
		mcp.WithString("keyword", mcp.Description("Keyword filter")),
		mcp.WithString("tag", mcp.Description("Tag filter")),
		mcp.WithString("cursor", mcp.Description("Next-page cursor")),
		mcp.WithString("dir", mcp.Description(
			"List the nearest workspace at or above this directory, registered or not. Set it to your "+
				"working directory.")),
		mcp.WithBoolean("include_args", mcp.Description(
			"Add each executable's `inputs`: positional args in order, flags with type and default, which "+
				"are required, and prompt params to pass in execute's `params`. Default false.")),
	)
	listExecutables.Annotations = mcp.ToolAnnotation{
		Title:           "List executables",
		DestructiveHint: boolPtr(false), ReadOnlyHint: boolPtr(true),
		IdempotentHint: boolPtr(true), OpenWorldHint: boolPtr(true),
	}
	srv.AddTool(listExecutables, listExecutablesHandler(executor))

	addExecuteTool(srv, executor)
	addRunCommandTool(srv, executor)
	addRunPythonTool(srv, executor)
	addRunExecutableTool(srv, executor)

	writeFlowfile := mcp.NewTool("write_flowfile",
		mcp.WithDescription("Create or update a .flow file. Validates it against the schema before writing, then "+
			"syncs so list_executables and get_executable see the change. Prefer this over writing flow YAML "+
			"directly."),
		mcp.WithString("path", mcp.Required(),
			mcp.Description("Absolute or workspace-relative path ending in .flow or .flow.yaml")),
		mcp.WithString("content", mcp.Required(),
			mcp.Description("The flowfile's full YAML")),
		mcp.WithBoolean("overwrite",
			mcp.Description("Replace an existing file (default false)")),
	)
	writeFlowfile.Annotations = mcp.ToolAnnotation{
		Title:           "Write a flow file",
		DestructiveHint: boolPtr(true), ReadOnlyHint: boolPtr(false),
		IdempotentHint: boolPtr(false), OpenWorldHint: boolPtr(false),
	}
	srv.AddTool(writeFlowfile, writeFlowfileHandler(srv, executor))
}

func addExecuteTool(srv *server.MCPServer, executor CommandExecutor) {
	executeFlow := mcp.NewTool("execute",
		mcp.WithDescription("Run an executable by verb and optional ID. Prefer it over raw shell commands for dev "+
			"tasks: flow supplies the environment and secrets and records the run. A run that finishes returns "+
			"its output and exitCode, non-zero included; an error means it never started."),
		mcp.WithString("executable_verb", mcp.Required(),
			mcp.Description("Executable verb, e.g. run, build, test, deploy.")),
		mcp.WithString("executable_id",
			mcp.Pattern(executable.ExecutableIDPattern),
			mcp.Description(
				"Executable ID: `ws/ns:name`, or `name` in the current workspace and namespace. For a nameless "+
					"executable give `ws/`, `ns:`, `ws/ns:` or nothing. Nested namespaces need the workspace: "+
					"`ws/parent/child:name` (`./` for the current one).")),
		mcp.WithString("args", mcp.Description(
			"The executable's arguments as on a command line, e.g. `--node workbox 'uname -a'`. Shell-style "+
				"quoting, no expansion. All go to the executable, never to flow. A leading `--` is optional; a "+
				"later one makes the rest positional. Undefined arguments fail before anything runs.")),
		mcp.WithObject("params",
			mcp.AdditionalProperties(map[string]any{"type": "string"}),
			mcp.Description("The executable's `params` by EnvKey. Set as environment variables even if undeclared, "+
				"and win over `args` with the same EnvKey. Answer prompt params here: a prompt without a value "+
				"fails before anything runs (list_executables include_args lists them). Passed on flow's "+
				"command line, so never put secrets here; use a vault.")),
		mcp.WithString("dir", mcp.Description(
			"Directory to run in; flow uses the nearest workspace at or above it, registered or not. Set it to "+
				"your working directory (a worktree or fresh clone): the default is the server's own.")),
		mcp.WithString("workspace", mcp.Description(
			"Workspace to run in, by name or path. Overrides `dir`; doesn't change the current workspace.")),
		mcp.WithBoolean("sync", mcp.Description(
			"Sync before running: rescan executables and pull git-sourced workspaces from their remotes.")),
		mcp.WithOutputSchema[ExecutionOutput](),
	)
	executeFlow.Annotations = mcp.ToolAnnotation{
		Title:        "Execute executable",
		ReadOnlyHint: boolPtr(false), DestructiveHint: boolPtr(true),
		IdempotentHint: boolPtr(false), OpenWorldHint: boolPtr(true),
	}
	srv.AddTool(executeFlow, executeFlowHandler(srv, executor))
}

func getExecutableHandler(executor CommandExecutor) server.ToolHandlerFunc {
	return func(_ context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		executableVerb, err := request.RequireString("executable_verb")
		if err != nil {
			return toolError(ErrCodeInvalidInput, "executable_verb is required"), nil
		}
		executableID := request.GetString("executable_id", "")
		if err := validateExecutableRef(executableVerb, executableID); err != nil {
			return toolError(ErrCodeInvalidInput, err.Error()), nil
		}

		cmdArgs := []string{"browse", "--output", "json", executableVerb}
		if executableID != "" {
			cmdArgs = append(cmdArgs, executableID)
		}

		output, err := executor.Execute(cmdArgs...)
		if err != nil {
			ref := strings.Join([]string{executableVerb, executableID}, " ")
			return toolError(ErrCodeNotFound, fmt.Sprintf("Failed to get executable %s: %s", ref, output)), nil
		}

		return mcp.NewToolResultText(dedupeExecutableDescription(output)), nil
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
		if request.GetBool("include_args", false) {
			addExecutableInputs(cliOutput.Executables, output)
		}

		page, nextCursor, totalCount, err := paginate(cliOutput.Executables, cursor, defaultPageSize)
		if err != nil {
			return toolError(ErrCodeInvalidInput, err.Error()), nil
		}
		for i := range page {
			page[i].FullDescription = ""
			page[i].Description = summarizeDescription(page[i].Description)
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

		if err := validateExecutableRef(executableVerb, executableID); err != nil {
			return toolError(ErrCodeInvalidInput, err.Error()), nil
		}

		execArgs, err := executableArgs(args)
		if err != nil {
			return toolError(ErrCodeInvalidInput, fmt.Sprintf("invalid args: %s", err)), nil
		}
		paramFlags, err := paramOverrideArgs(request.GetArguments()["params"])
		if err != nil {
			return toolError(ErrCodeInvalidInput, fmt.Sprintf("invalid params: %s", err)), nil
		}

		cmdArgs := []string{executableVerb}
		if executableID != "" {
			cmdArgs = append(cmdArgs, executableID)
		} else if len(execArgs) > 0 {
			cmdArgs = append(cmdArgs, currentExecutableID)
		}
		if ws := request.GetString("workspace", ""); ws != "" {
			cmdArgs = append(cmdArgs, "--workspace", ws)
		}
		if syncFlag {
			cmdArgs = append(cmdArgs, "--sync")
		}
		cmdArgs = append(cmdArgs, paramFlags...)
		cmdArgs = append(cmdArgs, plainLogArgs...)
		if len(execArgs) > 0 {
			// Past the separator nothing can be taken for one of flow's own flags.
			cmdArgs = append(cmdArgs, "--")
			cmdArgs = append(cmdArgs, execArgs...)
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

		sendProgress(srv, ctx, progressToken, 2, 2, "Complete")
		ref := strings.Join([]string{executableVerb, executableID}, " ")
		return executionResult(ref+" execution failed", output, err), nil
	}
}

func addRunCommandTool(srv *server.MCPServer, executor CommandExecutor) {
	runCommand := mcp.NewTool("run_command",
		mcp.WithDescription("Run shell commands through flow instead of a raw shell: with the workspace's "+
			"environment and secrets, and recorded in history (get_execution_logs, `flow logs`). Pass `commands` "+
			"with `mode` to run several in one call."),
		mcp.WithString("command",
			mcp.Description("One shell command. Give this or `commands`.")),
		mcp.WithArray("commands", mcp.WithStringItems(),
			mcp.Description("Several shell commands, run as one batch per `mode`. Give this or `command`.")),
		mcp.WithString("mode",
			mcp.Description("`serial` (default; stops at the first failure) or `parallel`")),
		mcp.WithString("label",
			mcp.Description("Short label for the run's history entry")),
		mcp.WithString("dir",
			mcp.Description("Working directory (default: the current directory)")),
		mcp.WithString("workspace",
			mcp.Description(runWorkspaceDescription)),
		mcp.WithBoolean("sync", mcp.Description(
			"Sync before running: rescan executables and pull git-sourced workspaces from their remotes.")),
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
		mcp.WithDescription("Run Python code through flow instead of `python -c` or a scratch file: with the "+
			"workspace's environment and secrets, its virtualenv when there is one, and recorded in history."),
		mcp.WithString("code", mcp.Required(),
			mcp.Description("Python source; multi-line is fine, and tracebacks show real line numbers. "+
				"file, so tracebacks report real line numbers. Read parameters and secrets from os.environ.")),
		mcp.WithString("label",
			mcp.Description("Short label for the run's history entry")),
		mcp.WithString("dir",
			mcp.Description("Working directory (default: the current directory); also picks the workspace "+
				"environment and virtualenv")),
		mcp.WithString("workspace",
			mcp.Description(runWorkspaceDescription)),
		mcp.WithBoolean("sync", mcp.Description(
			"Sync before running: rescan executables and pull git-sourced workspaces from their remotes.")),
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
		mcp.WithDescription("Run a transient executable of any type from an inline spec, without saving a .flow "+
			"file: a serial or parallel batch, an HTTP request, a render or a launch. Use it when one command "+
			"isn't enough. The spec is one entry of a flowfile's `executables:`; author non-trivial ones against "+
			"get_info's schemaUrls.flowFile. Runs with the workspace's environment and secrets, recorded in history."),
		mcp.WithString("spec", mcp.Required(),
			mcp.Description("One executable definition as YAML or JSON, e.g. "+
				"{\"verb\":\"run\",\"serial\":{\"execs\":[{\"cmd\":\"...\"}]}}")),
		mcp.WithString("label",
			mcp.Description("Short label for the run's history entry (default: the executable's name)")),
		mcp.WithString("dir",
			mcp.Description("Directory to run in; flow uses the nearest workspace at or above it. Set it to your "+
				"working directory: the default is the server's own.")),
		mcp.WithString("workspace",
			mcp.Description(runWorkspaceDescription)),
		mcp.WithBoolean("sync", mcp.Description(
			"Sync before running: rescan executables and pull git-sourced workspaces from their remotes.")),
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
	output, err := executor.ExecuteContext(ctx, append(cmdArgs, plainLogArgs...)...)

	if ctx.Err() != nil {
		return toolError(ErrCodeCancelled, "execution was cancelled"), nil
	}

	sendProgress(srv, ctx, progressToken, 1, 2, "Processing result")

	sendProgress(srv, ctx, progressToken, 2, 2, "Complete")
	return executionResult(failMsg, output, err), nil
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

// addExecutableInputs fills each executable's Inputs from the type definitions in the CLI's JSON,
// which the summary fields don't carry.
func addExecutableInputs(execs []ExecutableOutput, cliJSON string) {
	type environment struct {
		Args   executable.ArgumentList  `json:"args"`
		Params executable.ParameterList `json:"params"`
	}
	var typed struct {
		Executables []map[string]json.RawMessage `json:"executables"`
	}
	if err := json.Unmarshal([]byte(cliJSON), &typed); err != nil || len(typed.Executables) != len(execs) {
		return
	}
	for i, raw := range typed.Executables {
		for _, kind := range []string{"exec", "serial", "parallel", "launch", "request", "render"} {
			body, ok := raw[kind]
			if !ok {
				continue
			}
			var env environment
			if json.Unmarshal(body, &env) != nil {
				break
			}
			execs[i].Inputs = summarizeInputs(env.Args, env.Params)
			break
		}
	}
}

func summarizeInputs(args executable.ArgumentList, params executable.ParameterList) *ExecutableInputs {
	inputs := &ExecutableInputs{}
	for _, arg := range args {
		if arg.Pos == nil && arg.Flag == "" {
			continue
		}
		summary := ArgSummary{
			Pos: arg.Pos, Flag: arg.Flag, EnvKey: arg.EnvKey, Default: arg.Default, Required: arg.Required,
		}
		if arg.Type != "" && arg.Type != executable.ArgumentTypeString {
			summary.Type = string(arg.Type)
		}
		inputs.Args = append(inputs.Args, summary)
	}
	slices.SortStableFunc(inputs.Args, func(a, b ArgSummary) int {
		switch {
		case a.Pos != nil && b.Pos != nil:
			return *a.Pos - *b.Pos
		case a.Pos != nil:
			return -1
		case b.Pos != nil:
			return 1
		}
		return 0
	})
	for _, param := range params {
		if param.Prompt != "" && param.EnvKey != "" {
			inputs.Prompts = append(inputs.Prompts, PromptSummary{EnvKey: param.EnvKey, Prompt: param.Prompt})
		}
	}
	return inputs
}
