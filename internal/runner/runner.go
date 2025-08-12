package runner

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/jahvon/expression"

	"github.com/flowexec/flow/internal/context"
	"github.com/flowexec/flow/internal/logger"
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

type CtxData struct {
	Workspace     string `expr:"workspace"`
	Namespace     string `expr:"namespace"`
	WorkspacePath string `expr:"workspacePath"`
	FlowFileName  string `expr:"flowFileName"`
	FlowFilePath  string `expr:"flowFilePath"`
	FlowFileDir   string `expr:"flowFileDir"`
}

func ExpressionEnv(
	ctx *context.Context,
	executable *executable.Executable,
	dataMap, envMap map[string]string,
) expression.Data {
	fn := filepath.Base(filepath.Base(executable.FlowFilePath()))
	data, err := expression.BuildData(
		ctx,
		envMap,
		"store", dataMap,
		"ctx", &CtxData{
			Workspace:     ctx.CurrentWorkspace.AssignedName(),
			Namespace:     ctx.Config.CurrentNamespace,
			WorkspacePath: executable.WorkspacePath(),
			FlowFileName:  fn,
			FlowFilePath:  executable.FlowFilePath(),
			FlowFileDir:   filepath.Dir(executable.FlowFilePath()),
		},
	)
	if err != nil {
		logger.Log().Errorf("failed to build expression data: %v", err)
		return nil
	}
	return data
}
