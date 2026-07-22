package exec

import (
	"github.com/pkg/errors"

	"github.com/flowexec/flow/v2/internal/runner"
	"github.com/flowexec/flow/v2/internal/runner/engine"
	"github.com/flowexec/flow/v2/internal/services/run"
	"github.com/flowexec/flow/v2/internal/utils/env"
	"github.com/flowexec/flow/v2/pkg/context"
	"github.com/flowexec/flow/v2/pkg/logger"
	"github.com/flowexec/flow/v2/types/executable"
)

// Test seams: swapped by tests to avoid spawning real subshells.
var (
	runCmdFn         = run.RunCmd
	runFileFn        = run.RunFile
	runContainerFn   = run.RunContainer
	resolveRuntimeFn = run.ResolveRuntime
)

type execRunner struct{}

func NewRunner() runner.Runner {
	return &execRunner{}
}

func (r *execRunner) Name() string {
	return "exec"
}

func (r *execRunner) IsCompatible(executable *executable.Executable) bool {
	if executable == nil || executable.Exec == nil {
		return false
	}
	return true
}

func (r *execRunner) Exec(
	ctx *context.Context,
	e *executable.Executable,
	_ engine.Engine,
	inputEnv map[string]string,
	inputArgs []string,
) error {
	execSpec := e.Exec
	defaultEnv := env.DefaultEnv(ctx, e)
	envMap, err := env.BuildEnvMap(ctx.Config.CurrentVaultName(), e.Env(), inputArgs, inputEnv, defaultEnv)
	if err != nil {
		return errors.Wrap(err, "unable to set parameters to env")
	}
	envList := env.EnvMapToEnvList(envMap)

	if cb, err := env.CreateTempEnvFiles(
		ctx.Config.CurrentVaultName(),
		e.FlowFilePath(),
		e.WorkspacePath(),
		e.Env(),
		inputArgs,
		envMap,
	); err != nil {
		ctx.AddCallback(cb)
		return errors.Wrap(err, "unable to create temporary env files")
	} else {
		ctx.AddCallback(cb)
	}

	targetDir, isTmp, err := execSpec.Dir.ExpandDirectory(
		e.WorkspacePath(),
		e.FlowFilePath(),
		ctx.ProcessTmpDir,
		envMap,
	)
	if err != nil {
		return errors.Wrap(err, "unable to expand directory")
	} else if isTmp {
		ctx.ProcessTmpDir = targetDir
	}

	logMode := execSpec.LogMode
	logFields := execSpec.GetLogFields()

	if execSpec.Cmd == "" && execSpec.File == "" {
		return errors.New("either cmd or file must be specified")
	}
	if execSpec.Cmd != "" && execSpec.File != "" {
		return errors.New("cannot set both cmd and file")
	}

	if execSpec.Container != nil {
		spec, err := buildContainerSpec(e, targetDir, envMap)
		if err != nil {
			return err
		}
		// Register cleanup before launch so an orphaned container is removed even
		// if the run is abandoned by the timeout goroutine in the runner.
		ctx.AddCallback(func(*context.Context) error {
			return run.ForceRemoveContainer(spec.Runtime, spec.Name)
		})
		return runContainerFn(ctx, spec, logMode, logger.Log(), ctx.StdIn(), logFields, ctx.CurrentTask)
	}

	switch {
	case execSpec.Cmd != "":
		return runCmdFn(execSpec.Cmd, targetDir, envList, logMode, logger.Log(), ctx.StdIn(), logFields, ctx.CurrentTask)
	case execSpec.File != "":
		return runFileFn(execSpec.File, targetDir, envList, logMode, logger.Log(), ctx.StdIn(), logFields, ctx.CurrentTask)
	default:
		return errors.New("unable to determine how e should be run")
	}
}
