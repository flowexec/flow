package mcp

import (
	"os"
	"os/exec"

	"github.com/pkg/errors"
)

const cliBinaryEnvKey = "FLOW_CLI_BINARY"

//go:generate mockgen -destination=mocks/command_executor.go -package=mocks . CommandExecutor
type CommandExecutor interface {
	Execute(args ...string) (string, error)
}

// FlowCLIExecutor runs the flow CLI with provided arguments. The CLI is being executed instead of importing the
// internal flow package directly to avoid duplicating the code that's defined in the cmd package, to make testing
// easier, and to avoid having to refactor the Context obj which is not currently designed in a way to be copied/reused
// across "executions". Maybe consider refactoring this when the context is refactored.
//
// The binary name can be overridden by setting the FLOW_CLI_BINARY environment variable.
type FlowCLIExecutor struct{}

func (c *FlowCLIExecutor) Execute(args ...string) (string, error) {
	name := "flow"
	if envName := os.Getenv(cliBinaryEnvKey); envName != "" {
		name = envName
	}
	cmd := exec.Command(name, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		// Only return an error if it's not an exit error.
		exitErr := &exec.ExitError{}
		if !errors.As(err, &exitErr) {
			return string(output), err
		}
	}
	return string(output), nil
}
