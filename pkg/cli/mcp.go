package cli

import "github.com/flowexec/flow/v2/internal/mcp"

// MCPExtension adds tools, prompts, resources, instructions, or server options to the MCP server
// started by the mcp command. Registrations are added after flow's built-ins; a name or URI that
// collides with a built-in or another extension makes the mcp command fail at startup instead of
// replacing the existing registration. Resource URIs under flow:// are reserved for flow.
//
// Prefix tool and prompt names with your CLI's name so a tool added to flow later can't collide.
type MCPExtension = mcp.Extension

// RegisterOption configures RegisterAllCommands.
type RegisterOption func(*registerConfig)

type registerConfig struct {
	mcpOptions []mcp.Option
}

// WithMCPExtensions registers extensions on the MCP server started by the mcp command. It can be
// passed more than once; extensions are applied in order.
func WithMCPExtensions(exts ...MCPExtension) RegisterOption {
	return func(c *registerConfig) {
		c.mcpOptions = append(c.mcpOptions, mcp.WithExtensions(exts...))
	}
}

// WithMCPServerInfo overrides the name and version the MCP server reports to clients. An empty
// value keeps flow's default.
func WithMCPServerInfo(name, version string) RegisterOption {
	return func(c *registerConfig) {
		c.mcpOptions = append(c.mcpOptions, mcp.WithServerInfo(name, version))
	}
}
