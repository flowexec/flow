package internal

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	tuikitIO "github.com/flowexec/tuikit/io"
	"github.com/flowexec/tuikit/views"
	"github.com/gen2brain/beeep"
	"github.com/spf13/cobra"

	"github.com/flowexec/flow/cmd/internal/flags"
	"github.com/flowexec/flow/internal/cache"
	"github.com/flowexec/flow/internal/context"
	"github.com/flowexec/flow/internal/io"
	"github.com/flowexec/flow/internal/logger"
	"github.com/flowexec/flow/internal/runner"
	"github.com/flowexec/flow/internal/runner/engine"
	"github.com/flowexec/flow/internal/runner/exec"
	"github.com/flowexec/flow/internal/runner/launch"
	"github.com/flowexec/flow/internal/runner/parallel"
	"github.com/flowexec/flow/internal/runner/render"
	"github.com/flowexec/flow/internal/runner/request"
	"github.com/flowexec/flow/internal/runner/serial"
	"github.com/flowexec/flow/internal/services/store"
	"github.com/flowexec/flow/internal/vault"
	vaultV2 "github.com/flowexec/flow/internal/vault/v2"
	"github.com/flowexec/flow/types/executable"
)

func RegisterExecCmd(ctx *context.Context, rootCmd *cobra.Command) {
	subCmd := &cobra.Command{
		Use:     "exec EXECUTABLE_ID [args...]",
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
			execList, err := ctx.ExecutableCache.GetExecutableList()
			if err != nil {
				return nil, cobra.ShellCompDirectiveError
			}
			execIDs := make([]string, 0, len(execList))
			for _, e := range execList {
				execIDs = append(execIDs, e.ID())
			}
			return execIDs, cobra.ShellCompDirectiveNoFileComp
		},
		PreRun: func(cmd *cobra.Command, args []string) {
			logMode := flags.ValueFor[string](cmd, *flags.LogModeFlag, false)
			if err := tuikitIO.LogMode(logMode).Validate(); err != nil {
				logger.Log().FatalErr(err)
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

// TODO: refactor this function to simplify the logic
//
//nolint:funlen,gocognit
func execFunc(ctx *context.Context, cmd *cobra.Command, verb executable.Verb, args []string) {
	logMode := flags.ValueFor[string](cmd, *flags.LogModeFlag, false)
	if logMode != "" {
		logger.Log().SetMode(tuikitIO.LogMode(logMode))
	}

	if err := verb.Validate(); err != nil {
		logger.Log().FatalErr(err)
	}

	var ref executable.Ref
	if len(args) == 0 {
		ref = context.ExpandRef(ctx, executable.NewRef("", verb))
	} else {
		idArg := args[0]
		ref = context.ExpandRef(ctx, executable.NewRef(idArg, verb))
	}

	e, err := ctx.ExecutableCache.GetExecutableByRef(ref)
	if err != nil && errors.Is(cache.NewExecutableNotFoundError(ref.String()), err) {
		logger.Log().Debugf("Executable %s not found in cache, syncing cache", ref)
		if err := ctx.ExecutableCache.Update(); err != nil {
			logger.Log().FatalErr(err)
		}
		e, err = ctx.ExecutableCache.GetExecutableByRef(ref)
	}
	if err != nil {
		logger.Log().FatalErr(err)
	}

	if err := e.Validate(); err != nil {
		logger.Log().FatalErr(err)
	}

	if !e.IsExecutableFromWorkspace(ctx.CurrentWorkspace.AssignedName()) {
		logger.Log().FatalErr(fmt.Errorf(
			"e '%s' cannot be executed from workspace %s",
			ref,
			ctx.Config.CurrentWorkspace,
		))
	}

	s, err := store.NewStore(store.Path())
	if err != nil {
		logger.Log().FatalErr(err)
	}
	if _, err = s.CreateAndSetBucket(ref.String()); err != nil {
		logger.Log().FatalErr(err)
	}
	_ = s.Close()

	// add --param overrides to the env map
	envMap := make(map[string]string)
	paramOverrides := flags.ValueFor[[]string](cmd, *flags.ParameterValueFlag, false)
	applyParameterOverrides(paramOverrides, envMap)

	// add values from the prompt param type to the env map
	textInputs := pendingFormFields(ctx, e, envMap)
	if len(textInputs) > 0 {
		form, err := views.NewForm(io.Theme(ctx.Config.Theme.String()), ctx.StdIn(), ctx.StdOut(), textInputs...)
		if err != nil {
			logger.Log().FatalErr(err)
		}
		if err := form.Run(ctx.Ctx); err != nil {
			logger.Log().FatalErr(err)
		}
		for key, val := range form.ValueMap() {
			envMap[key] = fmt.Sprintf("%v", val)
		}
	}

	if ctx.Config.CurrentVault == nil || *ctx.Config.CurrentVault == vaultV2.LegacyVaultReservedName {
		setAuthEnv(ctx, cmd, e, false)
	}
	startTime := time.Now()
	eng := engine.NewExecEngine()

	var execArgs []string
	if len(args) >= 2 {
		execArgs = args[1:]
	}

	if err := runner.Exec(ctx, e, eng, envMap, execArgs); err != nil {
		logger.Log().FatalErr(err)
	}
	dur := time.Since(startTime)
	processStore, err := store.NewStore(store.Path())
	if err != nil {
		logger.Log().Errorf("failed clearing process store\n%v", err)
	}
	if processStore != nil {
		if err = processStore.DeleteBucket(store.EnvironmentBucket()); err != nil {
			logger.Log().Errorf("failed clearing process store\n%v", err)
		}
		_ = processStore.Close()
	}
	logger.Log().Debugx(fmt.Sprintf("%s flow completed", ref), "Elapsed", dur.Round(time.Millisecond))
	if TUIEnabled(ctx, cmd) {
		if dur > 1*time.Minute && ctx.Config.SendSoundNotification() {
			_ = beeep.Beep(beeep.DefaultFreq, beeep.DefaultDuration)
		}
		if dur > 1*time.Minute && ctx.Config.SendTextNotification() {
			_ = beeep.Notify("Flow", "Flow completed", "")
		}
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
	ctx.CancelFunc()
	return nil
}

func setAuthEnv(ctx *context.Context, _ *cobra.Command, executable *executable.Executable, force bool) {
	if authRequired(ctx, executable) || force {
		form, err := views.NewForm(
			io.Theme(ctx.Config.Theme.String()),
			ctx.StdIn(),
			ctx.StdOut(),
			&views.FormField{
				Key:   vault.EncryptionKeyEnvVar,
				Title: "Enter vault encryption key",
				Type:  views.PromptTypeMasked,
			})
		if err != nil {
			logger.Log().FatalErr(err)
		}
		if err := form.Run(ctx.Ctx); err != nil {
			logger.Log().FatalErr(err)
		}
		val := form.FindByKey(vault.EncryptionKeyEnvVar).Value()
		if val == "" {
			logger.Log().FatalErr(fmt.Errorf("vault encryption key required"))
		}
		if err := os.Setenv(vault.EncryptionKeyEnvVar, val); err != nil {
			logger.Log().FatalErr(fmt.Errorf("failed to set vault encryption key\n%w", err))
		}
	}
}

// TODO: refactor this function to simplify the logic
//
//nolint:all
func authRequired(ctx *context.Context, rootExec *executable.Executable) bool {
	if os.Getenv(vault.EncryptionKeyEnvVar) != "" {
		return false
	}
	switch {
	case rootExec.Exec != nil:
		for _, param := range rootExec.Exec.Params {
			if param.SecretRef != "" {
				return true
			}
		}
	case rootExec.Launch != nil:
		for _, param := range rootExec.Launch.Params {
			if param.SecretRef != "" {
				return true
			}
		}
	case rootExec.Request != nil:
		for _, param := range rootExec.Request.Params {
			if param.SecretRef != "" {
				return true
			}
		}
	case rootExec.Render != nil:
		for _, param := range rootExec.Render.Params {
			if param.SecretRef != "" {
				return true
			}
		}
	case rootExec.Serial != nil:
		for _, param := range rootExec.Serial.Params {
			if param.SecretRef != "" {
				return true
			}
		}
		for _, e := range rootExec.Serial.Execs {
			if e.Ref != "" {
				childExec, err := ctx.ExecutableCache.GetExecutableByRef(e.Ref)
				if err != nil {
					continue
				}
				if authRequired(ctx, childExec) {
					return true
				}
			}
		}
	case rootExec.Parallel != nil:
		for _, param := range rootExec.Parallel.Params {
			if param.SecretRef != "" {
				return true
			}
		}
		for _, e := range rootExec.Parallel.Execs {
			if e.Ref != "" {
				childExec, err := ctx.ExecutableCache.GetExecutableByRef(e.Ref)
				if err != nil {
					continue
				}
				if authRequired(ctx, childExec) {
					return true
				}
			}
		}
	}
	return false
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

If the target executable accept arguments, they can be passed in the form of flag or positional arguments.
Flag arguments are specified with the format 'flag=value' and positional arguments are specified as values without any prefix.
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

flow exec ws/ns:build flag1=value1 flag2=value2 value3 value4
`
)
