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

// fakeSessionWithInfo also reports a client name, as a transport that completed the initialize
// handshake does.
type fakeSessionWithInfo struct {
	fakeSession
	name string
}

func (f fakeSessionWithInfo) GetClientInfo() mcp.Implementation {
	return mcp.Implementation{Name: f.name}
}
func (f fakeSessionWithInfo) SetClientInfo(mcp.Implementation) {}
func (f fakeSessionWithInfo) GetClientCapabilities() mcp.ClientCapabilities {
	return mcp.ClientCapabilities{}
}
func (f fakeSessionWithInfo) SetClientCapabilities(mcp.ClientCapabilities) {}

func TestMCPProvenance(t *testing.T) {
	srv := server.NewMCPServer("test", "1.0.0")

	// Start from an environment that inherits nothing. mcpProvenance deliberately prefers an
	// inherited session over one it mints, so a developer running these tests from inside a
	// harness that exports FLOW_RUN_SESSION — an agent session, or CI — would otherwise see the
	// fallback cases fail on their machine and pass on everyone else's. Subtests that want an
	// inherited value set it themselves.
	t.Setenv(store.RunSessionEnv, "")
	t.Setenv(store.RunClientEnv, "")

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

	t.Run("prefers an inherited session over one it mints", func(t *testing.T) {
		// A harness that tags its environment wants its MCP runs grouped with the ones it
		// makes by shelling out; minting our own would split one conversation across two IDs.
		t.Setenv(store.RunSessionEnv, "harness-conversation-9")
		ctx := srv.WithContext(context.Background(), fakeSession{id: stdioSessionID})
		if prov := mcpProvenance(ctx); prov.Session != "harness-conversation-9" {
			t.Errorf("expected the inherited session, got %q", prov.Session)
		}
	})

	t.Run("resolves an inherited session given as a ${NAME} reference", func(t *testing.T) {
		t.Setenv("HARNESS_SESSION_ID", "resolved-conversation")
		t.Setenv(store.RunSessionEnv, "${HARNESS_SESSION_ID}")
		if prov := mcpProvenance(context.Background()); prov.Session != "resolved-conversation" {
			t.Errorf("expected the referenced value, got %q", prov.Session)
		}
	})

	t.Run("falls back to the client the environment names", func(t *testing.T) {
		t.Setenv(store.RunClientEnv, "some-harness")
		// No transport session, so nothing reports a client name.
		if prov := mcpProvenance(context.Background()); prov.Client != "some-harness" {
			t.Errorf("expected the inherited client, got %q", prov.Client)
		}
	})

	t.Run("keeps the transport's client over the environment's", func(t *testing.T) {
		t.Setenv(store.RunClientEnv, "stale-value")
		sess := fakeSessionWithInfo{fakeSession: fakeSession{id: "real-7"}, name: "cursor"}
		ctx := srv.WithContext(context.Background(), sess)
		if prov := mcpProvenance(ctx); prov.Client != "cursor" {
			t.Errorf("expected the transport's client, got %q", prov.Client)
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
