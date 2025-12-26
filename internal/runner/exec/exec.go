package exec

import (
	"github.com/pkg/errors"

	"github.com/flowexec/flow/internal/runner"
	"github.com/flowexec/flow/internal/runner/engine"
	"github.com/flowexec/flow/internal/services/run"
	"github.com/flowexec/flow/internal/utils/env"
	"github.com/flowexec/flow/pkg/context"
	"github.com/flowexec/flow/pkg/logger"
	"github.com/flowexec/flow/types/executable"
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

	switch {
	case execSpec.Cmd == "" && execSpec.File == "":
		return errors.New("either cmd or file must be specified")
	case execSpec.Cmd != "" && execSpec.File != "":
		return errors.New("cannot set both cmd and file")
	case execSpec.Cmd != "":
		return run.RunCmd(execSpec.Cmd, targetDir, envList, logMode, logger.Log(), ctx.StdIn(), logFields)
	case execSpec.File != "":
		return run.RunFile(execSpec.File, targetDir, envList, logMode, logger.Log(), ctx.StdIn(), logFields)
	default:
		return errors.New("unable to determine how e should be run")
	}
}
