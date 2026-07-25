package internal

import (
	"errors"
	"fmt"
	stdio "io"
	"os"
	osExec "os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"gopkg.in/yaml.v3"

	tuikitIO "github.com/flowexec/tuikit/io"
	"github.com/flowexec/tuikit/views"
	"github.com/gen2brain/beeep"
	"github.com/spf13/cobra"

	errhandler "github.com/flowexec/flow/v2/cmd/internal/errors"
	"github.com/flowexec/flow/v2/cmd/internal/flags"
	"github.com/flowexec/flow/v2/internal/io"
	"github.com/flowexec/flow/v2/internal/runner"
	"github.com/flowexec/flow/v2/internal/runner/engine"
	"github.com/flowexec/flow/v2/internal/runner/exec"
	"github.com/flowexec/flow/v2/internal/runner/launch"
	"github.com/flowexec/flow/v2/internal/runner/parallel"
	"github.com/flowexec/flow/v2/internal/runner/render"
	"github.com/flowexec/flow/v2/internal/runner/request"
	"github.com/flowexec/flow/v2/internal/runner/serial"
	"github.com/flowexec/flow/v2/internal/utils/env"
	"github.com/flowexec/flow/v2/pkg/context"
	flowErrors "github.com/flowexec/flow/v2/pkg/errors"
	"github.com/flowexec/flow/v2/pkg/filesystem"
	"github.com/flowexec/flow/v2/pkg/logger"
	"github.com/flowexec/flow/v2/pkg/store"
	"github.com/flowexec/flow/v2/types/executable"
	"github.com/flowexec/flow/v2/types/workspace"
)

const (
	// backgroundRunIDEnv is set on child processes spawned by --background.
	backgroundRunIDEnv = "FLOW_BACKGROUND_RUN_ID"
)

func RegisterExecCmd(ctx *context.Context, rootCmd *cobra.Command) {
	subCmd := &cobra.Command{
		Use:     "exec EXECUTABLE_ID [-- args...]",
		Aliases: executable.SortedValidVerbs(),
		Short:   "Execute any executable by reference.",
		Long:    execLong,
		Example: execExamples,
		Args:    cobra.ArbitraryArgs,
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			verbStr := cmd.CalledAs()
			verb := executable.Verb(verbStr)
			execList, err := ctx.ExecutableCache.GetExecutableList()
			if err != nil {
				return nil, cobra.ShellCompDirectiveError
			}
			execIDs := make([]string, 0, len(execList))
			for _, e := range execList {
				if e.Verb.Equals(verb) {
					execIDs = append(execIDs, e.ID())
				}
			}
			return execIDs, cobra.ShellCompDirectiveNoFileComp
		},
		PreRun: func(cmd *cobra.Command, args []string) {
			logMode := flags.ValueFor[string](cmd, *flags.LogModeFlag, false)
			if err := tuikitIO.LogMode(logMode).Validate(); err != nil {
				errhandler.HandleFatal(ctx, cmd, err)
			}
			execPreRun(ctx, cmd, args)
		},
		Run: func(cmd *cobra.Command, args []string) {
			verbStr := cmd.CalledAs()
			verb := executable.Verb(verbStr)
			execFunc(ctx, cmd, verb, args)
		},
	}
	RegisterFlag(ctx, subCmd, *flags.ParameterValueFlag)
	RegisterFlag(ctx, subCmd, *flags.LogModeFlag)
	RegisterFlag(ctx, subCmd, *flags.BackgroundFlag)
	RegisterFlag(ctx, subCmd, *flags.CmdFlag)
	RegisterFlag(ctx, subCmd, *flags.CmdModeFlag)
	RegisterFlag(ctx, subCmd, *flags.LabelFlag)
	RegisterFlag(ctx, subCmd, *flags.CmdDirFlag)
	RegisterFlag(ctx, subCmd, *flags.SpecFlag)
	RegisterFlag(ctx, subCmd, *flags.RunWorkspaceFlag)
	rootCmd.AddCommand(subCmd)
}

func execPreRun(_ *context.Context, _ *cobra.Command, _ []string) {
	runner.RegisterRunner(exec.NewRunner())
	runner.RegisterRunner(launch.NewRunner())
	runner.RegisterRunner(request.NewRunner())
	runner.RegisterRunner(render.NewRunner())
	runner.RegisterRunner(serial.NewRunner())
	runner.RegisterRunner(parallel.NewRunner())
}

func execFunc(ctx *context.Context, cmd *cobra.Command, verb executable.Verb, args []string) {
	logMode := flags.ValueFor[string](cmd, *flags.LogModeFlag, false)
	if logMode != "" {
		logger.Log().SetMode(tuikitIO.LogMode(logMode))
	}

	if err := verb.Validate(); err != nil {
		errhandler.HandleFatal(ctx, cmd, err)
	}

	// Ad-hoc / transient modes: run something not resolved from the executable cache.
	adhocCmds := flags.ValueFor[[]string](cmd, *flags.CmdFlag, false)
	spec := flags.ValueFor[string](cmd, *flags.SpecFlag, false)
	switch {
	case len(adhocCmds) > 0 && spec != "":
		errhandler.HandleUsage(ctx, cmd, "--cmd and --spec are mutually exclusive")
		return
	case len(adhocCmds) > 0:
		execAdHoc(ctx, cmd, verb, adhocCmds)
		return
	case spec != "":
		execTransientSpec(ctx, cmd, verb, spec)
		return
	}

	e, ref := resolveExecutableForRun(ctx, cmd, verb, args)

	// Handle --background: spawn a detached child process and return immediately.
	background := flags.ValueFor[bool](cmd, *flags.BackgroundFlag, false)
	if background {
		launchBackground(ctx, ref, verb, args)
		return
	}

	// If this is a background child process, eagerly record the log archive path
	// so that `logs --running` can stream output while we're still executing.
	bgRunID := os.Getenv(backgroundRunIDEnv)
	if bgRunID != "" {
		linkBackgroundArchive(ctx, bgRunID)
	}

	if ctx.DataStore != nil {
		if err := ctx.DataStore.CreateProcessBucket(ref.String()); err != nil {
			errhandler.HandleFatal(ctx, cmd, err)
		}
		_ = os.Setenv(store.BucketEnv, ref.String())
	}

	envMap := buildExecEnv(ctx, cmd, e)

	var execArgs []string
	if len(args) >= 2 {
		execArgs = args[1:]
	}

	startTime := time.Now()
	prov := runProvenanceFromEnv()
	recordRunStart(ctx, ref, startTime, prov, "", "")

	eng := engine.NewExecEngine()
	runErr := runner.Exec(ctx, e, eng, envMap, execArgs)
	dur := time.Since(startTime)

	cleanupProcessStore(ctx)
	recordExecution(ctx, ref, startTime, dur, runErr, prov, "", "")

	// Update background run record if this is a child process.
	if bgRunID != "" {
		finalizeBackgroundRun(ctx, bgRunID, runErr)
	}

	if runErr != nil {
		errhandler.HandleFatal(ctx, cmd, runErr)
	}
	logger.Log().Debug(fmt.Sprintf("%s flow completed", ref), "Elapsed", dur.Round(time.Millisecond))
	sendCompletionNotifications(ctx, cmd, dur)
}

// resolveExecutableForRun resolves the target executable and ref from the verb and args, syncing the
// cache on a miss and validating workspace membership. Fatal on any resolution/validation failure.
func resolveExecutableForRun(
	ctx *context.Context, cmd *cobra.Command, verb executable.Verb, args []string,
) (*executable.Executable, executable.Ref) {
	var ref executable.Ref
	if len(args) == 0 {
		ref = context.ExpandRef(ctx, executable.NewRef("", verb))
	} else {
		ref = context.ExpandRef(ctx, executable.NewRef(args[0], verb))
	}

	e, err := ctx.ExecutableCache.GetExecutableByRef(ref)
	if err != nil && errors.Is(err, flowErrors.NewExecutableNotFoundError(ref.String())) {
		logger.Log().Debugf("Executable %s not found in cache, syncing cache", ref)
		if err := ctx.ExecutableCache.Update(); err != nil {
			errhandler.HandleFatal(ctx, cmd, err)
		}
		e, err = ctx.ExecutableCache.GetExecutableByRef(ref)
	}
	if err != nil {
		errhandler.HandleFatal(ctx, cmd, err)
	}

	if err := e.Validate(); err != nil {
		errhandler.HandleFatal(ctx, cmd, err)
	}

	if ctx.CurrentWorkspace != nil && !e.IsExecutableFromWorkspace(ctx.CurrentWorkspace.AssignedName()) {
		errhandler.HandleFatal(ctx, cmd, fmt.Errorf(
			"executable '%s' belongs to workspace '%s' and cannot be run from the current workspace '%s'",
			ref,
			e.Workspace(),
			ctx.Config.CurrentWorkspace,
		))
	}
	return e, ref
}

// execAdHoc runs one or more arbitrary shell commands through flow as a transient, in-memory
// executable. Nothing is written to disk as a flowfile; the run flows through the normal engine
// (workspace env, vault secrets, logging) and is recorded in history with the command text and an
// optional label so it shows up in `flow logs` like a named executable. A single command becomes an
// `exec` executable; multiple commands become a `serial` or `parallel` executable (per --mode).
func execAdHoc(ctx *context.Context, cmd *cobra.Command, verb executable.Verb, commands []string) {
	label := flags.ValueFor[string](cmd, *flags.LabelFlag, false)
	dir := flags.ValueFor[string](cmd, *flags.CmdDirFlag, false)
	if dir == "" {
		if wd, err := os.Getwd(); err == nil {
			dir = wd
		}
	}

	// Default ad-hoc output to raw command output (Text mode routes stdout/stderr straight through
	// without flow's timestamp/level/color decoration), so callers — especially agents via the MCP
	// run_command tool — get the command's actual output. An explicit --log-mode still wins.
	logMode := tuikitIO.Text
	if lm := flags.ValueFor[string](cmd, *flags.LogModeFlag, false); lm != "" {
		logMode = tuikitIO.LogMode(lm)
	}

	joined := strings.Join(commands, "\n")
	e := &executable.Executable{
		Verb:        verb,
		Name:        adHocName(label, joined),
		Description: label,
	}
	if len(commands) == 1 {
		e.Exec = &executable.ExecExecutableType{
			Cmd:     commands[0],
			Dir:     executable.Directory(dir),
			LogMode: logMode,
		}
	} else {
		steps := make(executable.SerialRefConfigList, len(commands))
		for i, c := range commands {
			steps[i] = executable.SerialRefConfig{Cmd: c}
		}
		mode := flags.ValueFor[string](cmd, *flags.CmdModeFlag, false)
		if mode == "parallel" {
			pSteps := make(executable.ParallelRefConfigList, len(commands))
			for i, c := range commands {
				pSteps[i] = executable.ParallelRefConfig{Cmd: c}
			}
			e.Parallel = &executable.ParallelExecutableType{Execs: pSteps}
		} else {
			e.Serial = &executable.SerialExecutableType{Execs: steps}
		}
	}
	setTransientContext(ctx, cmd, e, dir)
	runTransientExecutable(ctx, cmd, e, joined, label)
}

// execTransientSpec runs a transient executable parsed from an inline definition (--spec). Unlike
// --cmd, the spec can be any executable type (exec, serial, parallel, request, render, launch); it
// is never written to disk but runs through the normal engine and is recorded in history.
func execTransientSpec(ctx *context.Context, cmd *cobra.Command, verb executable.Verb, spec string) {
	content, err := resolveSpecContent(spec)
	if err != nil {
		errhandler.HandleFatal(ctx, cmd, err)
	}

	e := &executable.Executable{}
	if err := yaml.Unmarshal([]byte(content), e); err != nil {
		errhandler.HandleUsage(ctx, cmd, "invalid executable spec: %v", err)
		return
	}
	runDir, _ := os.Getwd()
	setTransientContext(ctx, cmd, e, runDir)
	e.SetDefaults()
	if e.Verb == "" {
		e.Verb = verb
	}
	if err := e.Validate(); err != nil {
		errhandler.HandleUsage(ctx, cmd, "invalid executable spec: %v", err)
		return
	}

	label := flags.ValueFor[string](cmd, *flags.LabelFlag, false)
	if label == "" {
		label = e.Name
	}
	runTransientExecutable(ctx, cmd, e, "", label)
}

// resolveSpecContent resolves a --spec value into raw definition content: '-' reads stdin,
// a leading '@' reads the named file, and anything else is treated as inline content.
func resolveSpecContent(spec string) (string, error) {
	switch {
	case spec == "-":
		data, err := stdio.ReadAll(os.Stdin)
		if err != nil {
			return "", fmt.Errorf("unable to read spec from stdin: %w", err)
		}
		return string(data), nil
	case strings.HasPrefix(spec, "@"):
		data, err := os.ReadFile(filepath.Clean(spec[1:]))
		if err != nil {
			return "", fmt.Errorf("unable to read spec file: %w", err)
		}
		return string(data), nil
	default:
		return spec, nil
	}
}

// setTransientContext anchors a transient executable to a resolved workspace so it runs with that
// workspace's environment (and secrets), even though it is not backed by a flowfile. The workspace
// is resolved per-run (explicit --workspace, then the workspace containing runDir, then the global
// current) WITHOUT ever mutating the global current workspace. When the resolved workspace differs
// from the global current, a notice is emitted so the run's context is visible/auditable.
func setTransientContext(ctx *context.Context, cmd *cobra.Command, e *executable.Executable, runDir string) {
	wsName, wsPath := resolveRunWorkspace(ctx, cmd, runDir)

	currentName := ""
	if ctx.CurrentWorkspace != nil {
		currentName = ctx.CurrentWorkspace.AssignedName()
	}
	if wsName != "" && wsName != currentName {
		logger.Log().Infof("Running in workspace '%s' (current workspace is '%s')", wsName, currentName)
	}

	var flowFilePath string
	if wsPath != "" {
		flowFilePath = filepath.Join(wsPath, "flow.yaml")
	}
	e.SetContext(wsName, wsPath, ctx.Config.CurrentNamespace, flowFilePath)
}

// resolveRunWorkspace picks the workspace a transient run should use, in priority order:
// an explicit --workspace flag, then the workspace whose location contains runDir (longest match),
// then the global current workspace. It never changes the global current workspace.
func resolveRunWorkspace(ctx *context.Context, cmd *cobra.Command, runDir string) (name, path string) {
	wsList, err := ctx.WorkspacesCache.GetWorkspaceConfigList()
	if err != nil {
		logger.Log().Debugf("unable to load workspaces for transient run resolution: %v", err)
	}

	if explicit := flags.ValueFor[string](cmd, *flags.RunWorkspaceFlag, false); explicit != "" {
		if ws := wsList.FindByName(explicit); ws != nil {
			return ws.AssignedName(), ws.Location()
		}
		errhandler.HandleUsage(ctx, cmd, "unknown workspace %q", explicit)
	}

	if ws := workspaceForPath(wsList, runDir); ws != nil {
		return ws.AssignedName(), ws.Location()
	}

	if ctx.CurrentWorkspace != nil {
		return ctx.CurrentWorkspace.AssignedName(), ctx.CurrentWorkspace.Location()
	}
	return "", ""
}

// workspaceForPath returns the workspace whose location is the longest prefix of dir, or nil.
func workspaceForPath(wsList workspace.WorkspaceList, dir string) *workspace.Workspace {
	if dir == "" {
		return nil
	}
	abs, err := filepath.Abs(dir)
	if err != nil {
		abs = dir
	}
	var best *workspace.Workspace
	for _, ws := range wsList {
		loc := ws.Location()
		if loc == "" {
			continue
		}
		if abs == loc || strings.HasPrefix(abs, loc+string(filepath.Separator)) {
			if best == nil || len(loc) > len(best.Location()) {
				best = ws
			}
		}
	}
	return best
}

// runTransientExecutable runs an already-constructed, in-memory executable through the normal engine
// and records it in history with the given command/label provenance. Shared by --cmd and --spec.
func runTransientExecutable(
	ctx *context.Context, cmd *cobra.Command, e *executable.Executable, command, label string,
) {
	ref := e.Ref()

	if ctx.DataStore != nil {
		if err := ctx.DataStore.CreateProcessBucket(ref.String()); err != nil {
			errhandler.HandleFatal(ctx, cmd, err)
		}
		_ = os.Setenv(store.BucketEnv, ref.String())
	}

	envMap := buildExecEnv(ctx, cmd, e)

	startTime := time.Now()
	prov := runProvenanceFromEnv()
	recordRunStart(ctx, ref, startTime, prov, command, label)

	eng := engine.NewExecEngine()
	runErr := runner.Exec(ctx, e, eng, envMap, nil)
	dur := time.Since(startTime)

	cleanupProcessStore(ctx)
	recordExecution(ctx, ref, startTime, dur, runErr, prov, command, label)

	if runErr != nil {
		errhandler.HandleFatal(ctx, cmd, runErr)
	}
	logger.Log().Debug(fmt.Sprintf("transient executable completed: %s", ref), "Elapsed", dur.Round(time.Millisecond))
	sendCompletionNotifications(ctx, cmd, dur)
}

// adHocName derives a short, ref-safe name for a transient command run, preferring the label and
// falling back to the command's first token.
func adHocName(label, command string) string {
	base := label
	if base == "" {
		if fields := strings.Fields(command); len(fields) > 0 {
			base = fields[0]
		}
	}
	slug := slugify(base)
	if slug == "" {
		slug = "command"
	}
	return "adhoc-" + slug
}

// slugify reduces a string to lowercase alphanumerics separated by single dashes, capped at 40 chars.
func slugify(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	var b strings.Builder
	prevDash := false
	for _, r := range s {
		switch {
		case (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9'):
			b.WriteRune(r)
			prevDash = false
		default:
			if !prevDash && b.Len() > 0 {
				b.WriteByte('-')
				prevDash = true
			}
		}
	}
	out := strings.Trim(b.String(), "-")
	if len(out) > 40 {
		out = strings.Trim(out[:40], "-")
	}
	return out
}

// launchBackground spawns a detached flow process for the given executable and returns immediately.
func launchBackground(ctx *context.Context, ref executable.Ref, verb executable.Verb, args []string) {
	runID := uuid.New().String()[:8]

	// Build the child command: same verb + args. Stdout/stderr are set to nil so
	// Go redirects them to /dev/null — terminal output is suppressed but the tuikit
	// archive handler still writes to the log file normally.
	childArgs := []string{string(verb)}
	if len(args) > 0 {
		childArgs = append(childArgs, args...)
	}

	flowBin, err := os.Executable()
	if err != nil {
		logger.Log().FatalErr(fmt.Errorf("unable to find flow binary: %w", err))
	}

	child := osExec.Command(flowBin, childArgs...)
	child.Env = append(os.Environ(), fmt.Sprintf("%s=%s", backgroundRunIDEnv, runID))
	setSysProcAttr(child)
	child.Stdout = nil
	child.Stderr = nil
	child.Stdin = nil

	if err := child.Start(); err != nil {
		logger.Log().FatalErr(fmt.Errorf("failed to start background process: %w", err))
	}

	run := store.BackgroundRun{
		ID:        runID,
		PID:       child.Process.Pid,
		Ref:       ref.String(),
		StartedAt: time.Now(),
		Status:    store.BackgroundRunning,
	}
	if ctx.DataStore != nil {
		if err := ctx.DataStore.SaveBackgroundRun(run); err != nil {
			logger.Log().Errorf("failed to save background run record: %v", err)
		}
	}

	// Release the child process so it survives parent exit.
	_ = child.Process.Release()

	logger.Log().Println(fmt.Sprintf("Started background run %s (PID %d) for %s", runID, run.PID, ref))
}

// linkBackgroundArchive eagerly writes the log archive path into the background run
// record so that `logs attach` can stream output while the child is still executing.
// Unlike findArchiveByID, this scans the log directory directly without skipping empty
// files — the archive file exists at startup but may not have content yet.
func linkBackgroundArchive(ctx *context.Context, runID string) {
	if ctx.DataStore == nil || ctx.LogArchiveID == "" {
		return
	}
	archivePath := findArchiveFileByID(ctx.LogArchiveID)
	if archivePath == "" {
		return
	}
	run, err := ctx.DataStore.GetBackgroundRun(runID)
	if err != nil {
		return
	}
	run.LogArchiveID = archivePath
	_ = ctx.DataStore.SaveBackgroundRun(run)
}

// findArchiveFileByID scans the logs directory for a file whose name starts with the
// given archive ID. Unlike ListArchiveEntries, this does not skip empty files.
func findArchiveFileByID(archiveID string) string {
	logsDir := filesystem.LogsDir()
	files, err := os.ReadDir(logsDir)
	if err != nil {
		return ""
	}
	for _, f := range files {
		if f.IsDir() {
			continue
		}
		if strings.HasPrefix(f.Name(), archiveID) {
			return filepath.Join(logsDir, f.Name())
		}
	}
	return ""
}

// finalizeBackgroundRun updates the background run record with the final status.
func finalizeBackgroundRun(ctx *context.Context, runID string, runErr error) {
	if ctx.DataStore == nil {
		return
	}
	run, err := ctx.DataStore.GetBackgroundRun(runID)
	if err != nil {
		logger.Log().Debug("failed to load background run for finalization", "err", err)
		return
	}
	now := time.Now()
	run.CompletedAt = &now
	if archivePath := findArchiveByID(ctx.LogArchiveID); archivePath != "" {
		run.LogArchiveID = archivePath
	}
	if runErr != nil {
		run.Status = store.BackgroundFailed
		run.Error = runErr.Error()
	} else {
		run.Status = store.BackgroundCompleted
	}
	if err := ctx.DataStore.SaveBackgroundRun(run); err != nil {
		logger.Log().Debug("failed to finalize background run", "err", err)
	}
}

func buildExecEnv(ctx *context.Context, cmd *cobra.Command, e *executable.Executable) map[string]string {
	envMap := make(map[string]string)
	if wsData, err := ctx.WorkspacesCache.GetWorkspaceConfigList(); err != nil {
		logger.Log().Errorf("failed to get workspace cache data, skipping env file resolution: %v", err)
	} else {
		if wsCfg := wsData.FindByName(e.Workspace()); wsCfg == nil {
			logger.Log().Warnf("workspace %s not found in cache, skipping env file resolution", e.Workspace())
		} else {
			applyWorkspaceParameterOverrides(wsCfg, envMap)
		}
	}

	paramOverrides := flags.ValueFor[[]string](cmd, *flags.ParameterValueFlag, false)
	applyParameterOverrides(paramOverrides, envMap)

	textInputs := pendingFormFields(ctx, e, envMap)
	if len(textInputs) > 0 {
		form, err := views.NewForm(logger.Theme(ctx.Config.Theme.String()), ctx.StdIn(), ctx.StdOut(), textInputs...)
		if err != nil {
			logger.Log().FatalErr(err)
		}
		if err := form.Run(ctx); err != nil {
			logger.Log().FatalErr(err)
		}
		for key, val := range form.ValueMap() {
			envMap[key] = fmt.Sprintf("%v", val)
		}
	}
	return envMap
}

func cleanupProcessStore(ctx *context.Context) {
	if ctx.DataStore != nil {
		if err := ctx.DataStore.DeleteProcessBucket(store.EnvironmentBucket()); err != nil {
			logger.Log().Errorf("failed clearing process store\n%v", err)
		}
	}
}

// provenance bundles who/what launched a run, recorded on its execution record.
type provenance struct {
	source, client, session string
}

// runProvenanceFromEnv resolves run provenance from environment variables set by the caller
// (e.g. the MCP server). Source defaults to "cli".
func runProvenanceFromEnv() provenance {
	source := os.Getenv(store.RunSourceEnv)
	if source == "" {
		source = store.RunSourceCLI
	}
	return provenance{
		source:  source,
		client:  os.Getenv(store.RunClientEnv),
		session: os.Getenv(store.RunSessionEnv),
	}
}

// recordRunStart writes an in-progress ("running") execution record before the run begins, so that
// `flow logs` (from any process, via the shared store) can show the run as active. It is keyed by
// the run's log archive ID so recordExecution can upsert it into its terminal state on completion.
// No-op when there is no stable ID (legacy fallback: only the terminal record is written).
// command/label are set for ad-hoc runs and empty for named executables.
func recordRunStart(
	ctx *context.Context, ref executable.Ref, startTime time.Time, prov provenance, command, label string,
) {
	if ctx.DataStore == nil || ctx.LogArchiveID == "" {
		return
	}
	record := store.ExecutionRecord{
		ID:         ctx.LogArchiveID,
		Ref:        ref.String(),
		StartedAt:  startTime,
		Status:     store.RunRunning,
		PID:        os.Getpid(),
		Source:     prov.source,
		ClientName: prov.client,
		SessionID:  prov.session,
		Command:    command,
		Label:      label,
	}
	if recErr := ctx.DataStore.RecordExecution(record); recErr != nil {
		logger.Log().Debug("failed to record run start", "err", recErr)
	}
}

func recordExecution(
	ctx *context.Context, ref executable.Ref, startTime time.Time, dur time.Duration, runErr error,
	prov provenance, command, label string,
) {
	now := time.Now()
	record := store.ExecutionRecord{
		ID:          ctx.LogArchiveID,
		Ref:         ref.String(),
		StartedAt:   startTime,
		CompletedAt: &now,
		Duration:    dur,
		Status:      store.RunCompleted,
		PID:         os.Getpid(),
		Source:      prov.source,
		ClientName:  prov.client,
		SessionID:   prov.session,
		Command:     command,
		Label:       label,
	}
	if runErr != nil {
		record.ExitCode = 1
		record.Error = runErr.Error()
		record.Status = store.RunFailed
	}
	if archivePath := findArchiveByID(ctx.LogArchiveID); archivePath != "" {
		record.LogArchiveID = archivePath
	} else {
		record.LogArchiveID = findArchiveFileByID(ctx.LogArchiveID)
	}
	if ctx.DataStore != nil {
		if recErr := ctx.DataStore.RecordExecution(record); recErr != nil {
			logger.Log().Debug("failed to record execution history", "err", recErr)
		}
	}
}

// findArchiveByID searches log archive entries for one matching the given ID.
// Returns the entry's path if found, empty string otherwise.
func findArchiveByID(archiveID string) string {
	if archiveID == "" {
		return ""
	}
	entries, err := tuikitIO.ListArchiveEntries(filesystem.LogsDir())
	if err != nil {
		return ""
	}
	for _, e := range entries {
		if e.ID == archiveID {
			return e.Path
		}
	}
	return ""
}

func sendCompletionNotifications(ctx *context.Context, cmd *cobra.Command, dur time.Duration) {
	if !TUIEnabled(ctx, cmd) || dur <= 1*time.Minute {
		return
	}
	if ctx.Config.SendSoundNotification() {
		_ = beeep.Beep(beeep.DefaultFreq, beeep.DefaultDuration)
	}
	if ctx.Config.SendTextNotification() {
		_ = beeep.Notify("Flow", "Flow completed", "")
	}
}

func runByRef(ctx *context.Context, cmd *cobra.Command, argsStr string) error {
	s := strings.Split(argsStr, " ")
	if len(s) != 2 {
		return fmt.Errorf("invalid reference string %s", argsStr)
	}
	verbStr := s[0]
	verb := executable.Verb(verbStr)
	id := s[1]

	cmds := cmd.Root().Commands()
	var execCmd *cobra.Command
	for _, c := range cmds {
		if c.Name() == "exec" {
			execCmd = c
			break
		}
	}

	if execCmd == nil {
		return errors.New("exec command not found")
	}
	execCmd.SetArgs([]string{verbStr, id})
	execCmd.SetOut(ctx.StdOut())
	execCmd.SetErr(ctx.StdOut())
	execCmd.SetIn(ctx.StdIn())
	execPreRun(ctx, execCmd, []string{id})
	execFunc(ctx, execCmd, verb, []string{id})
	ctx.Cancel()
	return nil
}

func pendingFormFields(
	ctx *context.Context, rootExec *executable.Executable, envMap map[string]string,
) []*views.FormField {
	var pending []*views.FormField

	if env := rootExec.Env(); env != nil {
		pending = append(pending, pendingFieldsFromParams(env.Params, envMap)...)
	}

	var childRefs []executable.Ref
	switch {
	case rootExec.Serial != nil:
		for _, child := range rootExec.Serial.Execs {
			if child.Ref != "" {
				childRefs = append(childRefs, child.Ref)
			}
		}
	case rootExec.Parallel != nil:
		for _, child := range rootExec.Parallel.Execs {
			if child.Ref != "" {
				childRefs = append(childRefs, child.Ref)
			}
		}
	}
	for _, ref := range childRefs {
		childExec, err := ctx.ExecutableCache.GetExecutableByRef(ref)
		if err != nil {
			continue
		}
		pending = append(pending, pendingFormFields(ctx, childExec, envMap)...)
	}

	return pending
}

func pendingFieldsFromParams(params executable.ParameterList, envMap map[string]string) []*views.FormField {
	var fields []*views.FormField
	for _, param := range params {
		_, exists := envMap[param.EnvKey]
		if param.Prompt != "" && !exists {
			fields = append(fields, &views.FormField{Key: param.EnvKey, Title: param.Prompt})
		}
	}
	return fields
}

//nolint:nestif
func applyWorkspaceParameterOverrides(ws *workspace.Workspace, envMap map[string]string) {
	if len(ws.EnvFiles) > 0 {
		loaded, err := env.LoadEnvFromFiles(ws.EnvFiles, ws.Location())
		if err != nil {
			logger.Log().Errorf("failed loading env files for workspace %s: %v", ws.AssignedName(), err)
		}
		for k, v := range loaded {
			envMap[k] = v
		}
	} else {
		rootEnvFile := filepath.Join(ws.Location(), ".env")
		if _, err := os.Stat(rootEnvFile); err == nil {
			loaded, err := env.LoadEnvFromFiles([]string{rootEnvFile}, ws.Location())
			if err != nil {
				logger.Log().Errorf("failed loading root env file %s: %v", rootEnvFile, err)
			} else {
				for k, v := range loaded {
					envMap[k] = v
				}
			}
		}
	}
}

func applyParameterOverrides(overrides []string, envMap map[string]string) {
	for _, override := range overrides {
		parts := strings.SplitN(override, "=", 2)
		if len(parts) != 2 {
			continue // skip invalid overrides
		}
		key, value := parts[0], parts[1]
		envMap[key] = value
	}
}

var (
	execExamples = `
  # Execute a nameless flow in the current workspace with the 'install' verb
  flow install

  # Execute a nameless flow in the 'ws' workspace with the 'test' verb
  flow test ws/

  # Execute the 'build' flow in the current workspace and namespace
  flow exec build
  flow run build   # 'run' is an alias for the 'exec' verb

  # Execute the 'docs' flow with the 'show' verb
  flow show docs

  # Execute in a specific workspace and namespace
  flow exec ws/ns:build

  # Pass flag and positional arguments to the executable
  flow exec ws/ns:build -- --flag1=value1 --flag2=value2 value3 value4
`
)

var execLong = fmt.Sprintf(
	"Execute an executable where EXECUTABLE_ID is the target executable's ID in the form of 'ws/ns:name'.\n"+
		"The flow subcommand used should match the target executable's verb or one of its aliases.\n\n"+
		"If the target executable accepts arguments, use '--' to separate flow flags from executable arguments.\n"+
		"Flag arguments use standard '--flag=value' or '--flag value' syntax. "+
		"Boolean flags can omit the value (e.g. '--verbose' implies true).\n"+
		"Positional arguments are specified as values without any prefix.\n\n"+
		"See %s for more information on executable verbs.\n"+
		"See %s for more information on executable IDs.",
	io.TypesDocsURL("flowfile", "executableverb"),
	io.TypesDocsURL("flowfile", "executableref"),
)
