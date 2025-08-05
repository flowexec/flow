package mcp

import (
	"os"
	"os/exec"
)

const cliBinaryEnvKey = "FLOW_CLI_BINARY"

//go:generate mockgen -destination=mocks/command_executor.go -package=mocks . CommandExecutor
type CommandExecutor interface {
	Execute(args ...string) (string, error)
}

// FlowCLIExecutor runs the flow CLI with provided arguments. The CLI is being executed instead of importing the
// internal flow package directly to avoid duplicating the code that's defined in the cmd package and to make testing
// easier.
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
		return "", err
	}
	return string(output), nil
}
