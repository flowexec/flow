package internal

import (
	"errors"
	"fmt"
	"os"
	osExec "os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/google/uuid"

	tuikitIO "github.com/flowexec/tuikit/io"
	"github.com/flowexec/tuikit/views"
	"github.com/gen2brain/beeep"
	"github.com/spf13/cobra"

	errhandler "github.com/flowexec/flow/cmd/internal/errors"
	"github.com/flowexec/flow/cmd/internal/flags"
	"github.com/flowexec/flow/internal/io"
	"github.com/flowexec/flow/internal/runner"
	"github.com/flowexec/flow/internal/runner/engine"
	"github.com/flowexec/flow/internal/runner/exec"
	"github.com/flowexec/flow/internal/runner/launch"
	"github.com/flowexec/flow/internal/runner/parallel"
	"github.com/flowexec/flow/internal/runner/render"
	"github.com/flowexec/flow/internal/runner/request"
	"github.com/flowexec/flow/internal/runner/serial"
	"github.com/flowexec/flow/internal/utils/env"
	"github.com/flowexec/flow/pkg/context"
	flowErrors "github.com/flowexec/flow/pkg/errors"
	"github.com/flowexec/flow/pkg/filesystem"
	"github.com/flowexec/flow/pkg/logger"
	"github.com/flowexec/flow/pkg/store"
	"github.com/flowexec/flow/types/executable"
	"github.com/flowexec/flow/types/workspace"
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
		Long: execDocumentation + fmt.Sprintf(
			"\n\nSee %s for more information on executable verbs and "+
				"%s for more information on executable IDs.\n\n%s",
			io.TypesDocsURL("flowfile", "executableverb"),
			io.TypesDocsURL("flowfile", "executableref"),
			execExamples,
		),
		Args: cobra.ArbitraryArgs,
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

	var ref executable.Ref
	if len(args) == 0 {
		ref = context.ExpandRef(ctx, executable.NewRef("", verb))
	} else {
		idArg := args[0]
		ref = context.ExpandRef(ctx, executable.NewRef(idArg, verb))
	}

	e, err := ctx.ExecutableCache.GetExecutableByRef(ref)
	if err != nil && errors.Is(flowErrors.NewExecutableNotFoundError(ref.String()), err) {
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

	if !e.IsExecutableFromWorkspace(ctx.CurrentWorkspace.AssignedName()) {
		errhandler.HandleFatal(ctx, cmd, fmt.Errorf(
			"executable '%s' cannot be executed from workspace %s",
			ref,
			ctx.Config.CurrentWorkspace,
		))
	}

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
	eng := engine.NewExecEngine()
	runErr := runner.Exec(ctx, e, eng, envMap, execArgs)
	dur := time.Since(startTime)

	cleanupProcessStore(ctx)
	recordExecution(ctx, ref, startTime, dur, runErr)

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
	child.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
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
	run.LogArchiveID = findArchiveByID(ctx.LogArchiveID)
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

func recordExecution(ctx *context.Context, ref executable.Ref, startTime time.Time, dur time.Duration, runErr error) {
	record := store.ExecutionRecord{
		Ref:       ref.String(),
		StartedAt: startTime,
		Duration:  dur,
	}
	if runErr != nil {
		record.ExitCode = 1
		record.Error = runErr.Error()
	}
	record.LogArchiveID = findArchiveByID(ctx.LogArchiveID)
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

//nolint:gocognit
func pendingFormFields(
	ctx *context.Context, rootExec *executable.Executable, envMap map[string]string,
) []*views.FormField {
	pending := make([]*views.FormField, 0)
	switch {
	case rootExec.Exec != nil:
		for _, param := range rootExec.Exec.Params {
			_, exists := envMap[param.EnvKey]
			if param.Prompt != "" && !exists {
				pending = append(pending, &views.FormField{Key: param.EnvKey, Title: param.Prompt})
			}
		}
	case rootExec.Launch != nil:
		for _, param := range rootExec.Launch.Params {
			_, exists := envMap[param.EnvKey]
			if param.Prompt != "" && !exists {
				pending = append(pending, &views.FormField{Key: param.EnvKey, Title: param.Prompt})
			}
		}
	case rootExec.Request != nil:
		for _, param := range rootExec.Request.Params {
			_, exists := envMap[param.EnvKey]
			if param.Prompt != "" && !exists {
				pending = append(pending, &views.FormField{Key: param.EnvKey, Title: param.Prompt})
			}
		}
	case rootExec.Render != nil:
		for _, param := range rootExec.Render.Params {
			_, exists := envMap[param.EnvKey]
			if param.Prompt != "" && !exists {
				pending = append(pending, &views.FormField{Key: param.EnvKey, Title: param.Prompt})
			}
		}
	case rootExec.Serial != nil:
		for _, param := range rootExec.Serial.Params {
			_, exists := envMap[param.EnvKey]
			if param.Prompt != "" && !exists {
				pending = append(pending, &views.FormField{Key: param.EnvKey, Title: param.Prompt})
			}
		}
		for _, child := range rootExec.Serial.Execs {
			if child.Ref != "" {
				childExec, err := ctx.ExecutableCache.GetExecutableByRef(child.Ref)
				if err != nil {
					continue
				}
				childPending := pendingFormFields(ctx, childExec, envMap)
				pending = append(pending, childPending...)
			}
		}
	case rootExec.Parallel != nil:
		for _, param := range rootExec.Parallel.Params {
			if param.Prompt != "" {
				pending = append(pending, &views.FormField{Key: param.EnvKey, Title: param.Prompt})
			}
		}
		for _, child := range rootExec.Parallel.Execs {
			if child.Ref != "" {
				childExec, err := ctx.ExecutableCache.GetExecutableByRef(child.Ref)
				if err != nil {
					continue
				}
				childPending := pendingFormFields(ctx, childExec, envMap)
				pending = append(pending, childPending...)
			}
		}
	}
	return pending
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
	//nolint:lll
	execDocumentation = `
Execute an executable where EXECUTABLE_ID is the target executable's ID in the form of 'ws/ns:name'.
The flow subcommand used should match the target executable's verb or one of its aliases.

If the target executable accepts arguments, use '--' to separate flow flags from executable arguments.
Flag arguments use standard '--flag=value' or '--flag value' syntax. Boolean flags can omit the value (e.g., '--verbose' implies true).
Positional arguments are specified as values without any prefix.
`
	execExamples = `
#### Examples
**Execute a nameless flow in the current workspace with the 'install' verb**

flow install

**Execute a nameless flow in the 'ws' workspace with the 'test' verb**

flow test ws/

**Execute the 'build' flow in the current workspace and namespace**

flow exec build

flow run build  (Equivalent to the above since 'run' is an alias for the 'exec' verb)

**Execute the 'docs' flow with the 'show' verb in the current workspace and namespace**

flow show docs

**Execute the 'build' flow in the 'ws' workspace and 'ns' namespace**

flow exec ws/ns:build

**Execute the 'build' flow in the 'ws' workspace and 'ns' namespace with flag and positional arguments**

flow exec ws/ns:build -- --flag1=value1 --flag2=value2 value3 value4
`
)
