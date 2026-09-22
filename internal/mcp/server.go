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

const (
	defaultServerName    = "Flow"
	defaultServerVersion = "1.0.0"
)

type serverConfig struct {
	name       string
	version    string
	extensions []Extension
}

type Option func(*serverConfig)

func WithServerInfo(name, version string) Option {
	return func(c *serverConfig) {
		if name != "" {
			c.name = name
		}
		if version != "" {
			c.version = version
		}
	}
}

// WithExtensions registers extensions on the server. It can be passed more than once; extensions
// are applied in order.
func WithExtensions(exts ...Extension) Option {
	return func(c *serverConfig) {
		c.extensions = append(c.extensions, exts...)
	}
}

// NewServer builds the flow MCP server with its built-in tools, prompts, and resources, plus any
// extensions. It returns an error if an extension is invalid or collides with a registration.
func NewServer(executor CommandExecutor, opts ...Option) (*Server, error) {
	cfg := &serverConfig{name: defaultServerName, version: defaultServerVersion}
	for _, opt := range opts {
		opt(cfg)
	}
	exts := cfg.extensions

	srvOpts := append([]server.ServerOption{
		server.WithToolCapabilities(true),
		server.WithPromptCapabilities(false),
		server.WithResourceCapabilities(true, true),
		server.WithInstructions(extensionInstructions(serverInstructions, exts)),
	}, extensionServerOptions(exts)...)
	srv := server.NewMCPServer(cfg.name, cfg.version, srvOpts...)
	addServerTools(srv, executor)
	addServerPrompts(srv)
	addServerResources(srv)

	if err := applyExtensions(srv, exts); err != nil {
		return nil, err
	}
	return &Server{srv: srv, executor: executor}, nil
}

func (s *Server) Run() error {
	_ = os.Setenv(io.DisableInteractiveEnvKey, "true")

	return server.ServeStdio(s.srv)
}

// GetMCPServer returns the underlying MCP server for testing purposes
func (s *Server) GetMCPServer() *server.MCPServer {
	return s.srv
}
