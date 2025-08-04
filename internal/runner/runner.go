package runner

import (
	"fmt"
	"time"

	"github.com/flowexec/flow/internal/context"
	"github.com/flowexec/flow/internal/runner/engine"
	"github.com/flowexec/flow/types/executable"
)

//go:generate mockgen -destination=mocks/mock_runner.go -package=mocks github.com/flowexec/flow/internal/runner Runner
type Runner interface {
	Name() string
	Exec(
		ctx *context.Context,
		e *executable.Executable,
		eng engine.Engine,
		inputEnv map[string]string,
		inputArgs []string,
	) error
	IsCompatible(executable *executable.Executable) bool
}

var registeredRunners []Runner

func init() {
	registeredRunners = make([]Runner, 0)
}

func RegisterRunner(runner Runner) {
	registeredRunners = append(registeredRunners, runner)
}

func Exec(
	ctx *context.Context,
	executable *executable.Executable,
	eng engine.Engine,
	inputEnv map[string]string,
	inputArgs []string,
) error {
	var assignedRunner Runner
	for _, runner := range registeredRunners {
		if runner.IsCompatible(executable) {
			assignedRunner = runner
			break
		}
	}
	if assignedRunner == nil {
		return fmt.Errorf("compatible runner not found for executable %s", executable.ID())
	}
	ctx.RootExecutable = executable

	if executable.Timeout == nil {
		return assignedRunner.Exec(ctx, executable, eng, inputEnv, inputArgs)
	}

	done := make(chan error, 1)
	go func() {
		done <- assignedRunner.Exec(ctx, executable, eng, inputEnv, inputArgs)
	}()

	select {
	case err := <-done:
		return err
	case <-time.After(*executable.Timeout):
		return fmt.Errorf("timeout after %v", *executable.Timeout)
	}
}

func Reset() {
	registeredRunners = make([]Runner, 0)
}
