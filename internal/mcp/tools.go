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
