package mcp

import (
	"context"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"github.com/flowexec/flow/v2/pkg/store"
)

// fakeSession is a minimal ClientSession reporting a fixed ID, standing in for a transport.
type fakeSession struct{ id string }

func (f fakeSession) Initialize()       {}
func (f fakeSession) Initialized() bool { return true }
func (f fakeSession) NotificationChannel() chan<- mcp.JSONRPCNotification {
	return nil
}
func (f fakeSession) SessionID() string { return f.id }

func TestMCPProvenance(t *testing.T) {
	srv := server.NewMCPServer("test", "1.0.0")

	t.Run("replaces mcp-go's stdio constant with the process session ID", func(t *testing.T) {
		ctx := srv.WithContext(context.Background(), fakeSession{id: stdioSessionID})
		prov := mcpProvenance(ctx)
		if prov.Session == stdioSessionID {
			t.Fatalf("expected %q to be replaced; every client would share one session", stdioSessionID)
		}
		if prov.Session != processSessionID() {
			t.Errorf("expected the process session ID, got %q", prov.Session)
		}
	})

	t.Run("replaces an absent session ID", func(t *testing.T) {
		if prov := mcpProvenance(context.Background()); prov.Session != processSessionID() {
			t.Errorf("expected the process session ID, got %q", prov.Session)
		}
	})

	t.Run("keeps a genuine per-connection session ID", func(t *testing.T) {
		ctx := srv.WithContext(context.Background(), fakeSession{id: "real-session-7"})
		if prov := mcpProvenance(ctx); prov.Session != "real-session-7" {
			t.Errorf("expected the transport's session ID, got %q", prov.Session)
		}
	})

	t.Run("always marks the run as MCP-sourced", func(t *testing.T) {
		if prov := mcpProvenance(context.Background()); prov.Source != store.RunSourceMCP {
			t.Errorf("expected source %q, got %q", store.RunSourceMCP, prov.Source)
		}
	})
}

func TestProcessSessionID_StableAndNonEmpty(t *testing.T) {
	first := processSessionID()
	if first == "" {
		t.Fatal("expected a non-empty process session ID")
	}
	if second := processSessionID(); second != first {
		t.Errorf("expected a stable ID across calls, got %q then %q", first, second)
	}
}
