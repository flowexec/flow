package mcp

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"sync"

	"github.com/google/uuid"
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

// runDirCtxKey keys the working directory the flow subprocess should run in.
type runDirCtxKey struct{}

// withRunDir returns a context carrying the directory to launch the flow subprocess from.
//
// This matters because flow resolves its workspace by walking up from the working directory. The
// server process inherits whatever directory the MCP client was started in, which is routinely
// not where the caller is working — an agent in a git worktree, say. Without this, every run
// would silently execute against the server's directory instead of the caller's.
func withRunDir(ctx context.Context, dir string) context.Context {
	if dir == "" {
		return ctx
	}
	return context.WithValue(ctx, runDirCtxKey{}, dir)
}

func runDirFromContext(ctx context.Context) string {
	dir, _ := ctx.Value(runDirCtxKey{}).(string)
	return dir
}

// stdioSessionID is what mcp-go reports as the session ID for every stdio connection — a
// package constant, not a per-connection value (see mcp-go server/stdio.go). Taken at face
// value it collapses every run from every client into one "session".
const stdioSessionID = "stdio"

// processSessionID identifies this server process as a single client session. flow's MCP server
// is stdio-only and each client spawns its own `flow mcp` process, so one process is exactly one
// client connection — the session boundary mcp-go's constant fails to draw. It is resolved once
// and never changes, which is what makes it a usable grouping key in execution history.
//
// It is the last resort: a caller that supplies its own identifier knows better than we do what
// belongs together.
var processSessionID = sync.OnceValue(uuid.NewString)

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
	// A transport that issues genuine per-connection IDs keeps them; only substitute when it
	// gave us nothing usable.
	if prov.Session == "" || prov.Session == stdioSessionID {
		// An inherited session beats one we invent. A harness that tags its environment wants
		// the runs it makes over MCP grouped with the ones it makes by shelling out to the CLI
		// — minting our own here would split a single conversation across two IDs.
		if inherited := store.RunEnvValue(store.RunSessionEnv); inherited != "" {
			prov.Session = inherited
		} else {
			prov.Session = processSessionID()
		}
	}
	// Same reasoning for the client, for a transport that does not report one.
	if prov.Client == "" {
		prov.Client = store.RunEnvValue(store.RunClientEnv)
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
	cmd.Dir = runDirFromContext(ctx)
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
