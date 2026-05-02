package mcp

import (
	_ "embed"
	"os"

	"github.com/mark3labs/mcp-go/server"

	"github.com/flowexec/flow/v2/internal/io"
)

//go:embed resources/server-instructions.md
var serverInstructions string

type Server struct {
	srv      *server.MCPServer
	executor CommandExecutor
}

func NewServer(executor CommandExecutor) *Server {
	srv := server.NewMCPServer(
		"Flow",
		"1.0.0",
		server.WithToolCapabilities(true),
		server.WithPromptCapabilities(false),
		server.WithResourceCapabilities(true, true),
		server.WithInstructions(serverInstructions),
	)
	addServerTools(srv, executor)
	addServerPrompts(srv)
	addServerResources(srv)

	return &Server{srv: srv, executor: executor}
}

func (s *Server) Run() error {
	_ = os.Setenv(io.DisableInteractiveEnvKey, "true")

	return server.ServeStdio(s.srv)
}

// GetMCPServer returns the underlying MCP server for testing purposes
func (s *Server) GetMCPServer() *server.MCPServer {
	return s.srv
}
