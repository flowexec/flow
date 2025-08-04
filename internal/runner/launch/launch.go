package launch

import (
	"os"
	"strings"

	"github.com/pkg/errors"

	"github.com/flowexec/flow/internal/context"
	"github.com/flowexec/flow/internal/runner"
	"github.com/flowexec/flow/internal/runner/engine"
	"github.com/flowexec/flow/internal/services/open"
	"github.com/flowexec/flow/internal/utils"
	"github.com/flowexec/flow/internal/utils/env"
	"github.com/flowexec/flow/types/executable"
)

type launchRunner struct{}

func NewRunner() runner.Runner {
	return &launchRunner{}
}

func (r *launchRunner) Name() string {
	return "launch"
}

func (r *launchRunner) IsCompatible(executable *executable.Executable) bool {
	if executable == nil || executable.Launch == nil {
		return false
	}
	return true
}

func (r *launchRunner) Exec(
	ctx *context.Context,
	e *executable.Executable,
	_ engine.Engine,
	inputEnv map[string]string,
	inputArgs []string,
) error {
	launchSpec := e.Launch
	envMap, err := env.BuildEnvMap(
		ctx.Config.CurrentVaultName(),
		e.Env(),
		inputArgs,
		inputEnv,
		env.DefaultEnv(ctx, e),
	)
	if err != nil {
		return errors.Wrap(err, "unable to set parameters to env")
	}
	if err := env.SetEnv(ctx.Config.CurrentVaultName(), e.Env(), inputArgs, envMap); err != nil {
		return errors.Wrap(err, "unable to set parameters to env")
	}

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

	launchSpec.URI = os.ExpandEnv(launchSpec.URI)
	targetURI := launchSpec.URI
	if !strings.HasPrefix(targetURI, "http") {
		targetURI = utils.ExpandDirectory(
			launchSpec.URI,
			e.WorkspacePath(),
			e.FlowFilePath(),
			envMap,
		)
	}

	if launchSpec.App != "" {
		return open.OpenWith(launchSpec.App, targetURI)
	}
	return open.Open(targetURI)
}
