package mcp

import (
	"context"

	"github.com/mark3labs/mcp-go/server"
)

func addServerTools(srv *server.MCPServer, executor CommandExecutor) {
	addSystemTools(srv, executor)
	addWorkspaceTools(srv, executor)
	addExecutableTools(srv, executor)
}

func boolPtr(b bool) *bool {
	return &b
}

// maxExecutionOutputBytes caps captured subprocess output returned by execute/run_command/
// run_executable so a large run (e.g. `validate`, which chains generate+lint+test+e2e) can't
// return a single unbounded JSON-RPC message. Mirrors the cap get_execution_logs already
// enforces (see maxLogContentMaxBytes) so no MCP tool response is unbounded.
const maxExecutionOutputBytes = 200_000

// capOutput truncates output to at most maxExecutionOutputBytes, keeping the tail since that's
// where errors and failures surface. Returns the (possibly truncated) output and whether
// truncation happened; the full, untruncated output remains available via get_execution_logs.
func capOutput(output string) (string, bool) {
	if len(output) <= maxExecutionOutputBytes {
		return output, false
	}
	return output[len(output)-maxExecutionOutputBytes:], true
}

// sendProgress sends a progress notification to the client if a progress token was provided.
// It silently ignores errors (e.g., no active session in test contexts).
func sendProgress(srv *server.MCPServer, ctx context.Context, token any, progress, total float64, message string) {
	if token == nil || srv == nil {
		return
	}
	// Recover from panics in case the session context is not available (e.g., in-process test clients).
	defer func() { _ = recover() }()
	_ = srv.SendNotificationToClient(ctx, "notifications/progress", map[string]any{
		"progressToken": token,
		"progress":      progress,
		"total":         total,
		"message":       message,
	})
}
