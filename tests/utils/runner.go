package utils

import (
	"errors"
	"fmt"
	"os/exec"

	"github.com/flowexec/flow/cmd"
	"github.com/flowexec/flow/pkg/context"
)

type CommandRunner struct{}

func NewE2ECommandRunner() *CommandRunner {
	return &CommandRunner{}
}

func (r *CommandRunner) Run(ctx *context.Context, args ...string) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic occurred: %v", r)
		}
	}()
	rootCmd := cmd.NewRootCmd(ctx)
	rootCmd.SetArgs(args)
	rootCmd.SetIn(ctx.StdIn())
	rootCmd.SetOut(ctx.StdOut())
	if err = cmd.Execute(ctx, rootCmd); err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			err = fmt.Errorf("exit code: %d", exitErr.ExitCode())
		}
		return err
	}
	return
}
