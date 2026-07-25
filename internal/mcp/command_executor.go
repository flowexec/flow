package mcp

import (
	"context"
	"fmt"
	"os"
	"os/exec"

	"github.com/mark3labs/mcp-go/server"
	"github.com/pkg/errors"

	"github.com/flowexec/flow/v2/pkg/store"
)

const cliBinaryEnvKey = "FLOW_CLI_BINARY"

// provenanceCtxKey keys the run provenance carried on a context into the flow subprocess env.
type provenanceCtxKey struct{}

// runProvenance identifies who/what is launching a run, propagated to the flow subprocess as env vars.
type runProvenance struct {
	Source  string
	Client  string
	Session string
}

// withProvenance returns a context carrying run provenance for ExecuteContext to forward as env vars.
func withProvenance(ctx context.Context, p runProvenance) context.Context {
	return context.WithValue(ctx, provenanceCtxKey{}, p)
}

func provenanceFromContext(ctx context.Context) (runProvenance, bool) {
	p, ok := ctx.Value(provenanceCtxKey{}).(runProvenance)
	return p, ok
}

// mcpProvenance builds run provenance for an MCP-originated command, capturing the calling client's
// name and session ID from the MCP session (when the transport exposes them).
func mcpProvenance(ctx context.Context) runProvenance {
	prov := runProvenance{Source: store.RunSourceMCP}
	if sess := server.ClientSessionFromContext(ctx); sess != nil {
		prov.Session = sess.SessionID()
		if withInfo, ok := sess.(server.SessionWithClientInfo); ok {
			prov.Client = withInfo.GetClientInfo().Name
		}
	}
	return prov
}

//go:generate mockgen -destination=mocks/command_executor.go -package=mocks . CommandExecutor
type CommandExecutor interface {
	Execute(args ...string) (string, error)
	ExecuteContext(ctx context.Context, args ...string) (string, error)
}

// FlowCLIExecutor runs the flow CLI with provided arguments. The CLI is being executed instead of importing the
// internal flow package directly to avoid duplicating the code that's defined in the cmd package, to make testing
// easier, and to avoid having to refactor the Context obj which is not currently designed in a way to be copied/reused
// across "executions". Maybe consider refactoring this when the context is refactored.
//
// The binary name can be overridden by setting the FLOW_CLI_BINARY environment variable.
type FlowCLIExecutor struct{}

func (c *FlowCLIExecutor) Execute(args ...string) (string, error) {
	return c.ExecuteContext(context.Background(), args...)
}

func (c *FlowCLIExecutor) ExecuteContext(ctx context.Context, args ...string) (string, error) {
	name := "flow"
	if envName := os.Getenv(cliBinaryEnvKey); envName != "" {
		name = envName
	}
	cmd := exec.CommandContext(ctx, name, args...) // #nosec G204,G702
	if p, ok := provenanceFromContext(ctx); ok {
		cmd.Env = append(os.Environ(),
			fmt.Sprintf("%s=%s", store.RunSourceEnv, p.Source),
			fmt.Sprintf("%s=%s", store.RunClientEnv, p.Client),
			fmt.Sprintf("%s=%s", store.RunSessionEnv, p.Session),
		)
	}
	output, err := cmd.CombinedOutput()
	if err != nil {
		// Only return an error if it's not an exit error.
		var exitErr exec.ExitError
		if !errors.As(err, &exitErr) {
			return string(output), err
		}
	}
	return string(output), nil
}
