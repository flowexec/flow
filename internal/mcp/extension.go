package mcp

import (
	"fmt"
	"strings"

	"github.com/mark3labs/mcp-go/server"
)

// reservedURIScheme is the resource URI scheme owned by flow's built-in resources. Extensions may
// not register under it so a future built-in resource can never collide with an extension's.
const reservedURIScheme = "flow://"

// Extension adds capabilities to the flow MCP server. Everything an extension registers is added
// after flow's built-ins. A tool, prompt, or resource that collides with a built-in or with another
// extension fails server construction rather than silently replacing it.
type Extension struct {
	// Name identifies the extension in error messages.
	Name string

	Tools             []server.ServerTool
	Prompts           []server.ServerPrompt
	Resources         []server.ServerResource
	ResourceTemplates []server.ServerResourceTemplate

	// Instructions is appended to the server instructions sent to clients on initialize.
	Instructions string

	// ServerOptions are applied after flow's own options when the underlying server is created,
	// e.g. server.WithToolHandlerMiddleware or server.WithHooks. They apply to built-ins too and
	// can override flow's options, so use them for cross-cutting concerns only.
	ServerOptions []server.ServerOption
}

func extensionServerOptions(exts []Extension) []server.ServerOption {
	var opts []server.ServerOption
	for _, ext := range exts {
		opts = append(opts, ext.ServerOptions...)
	}
	return opts
}

func extensionInstructions(base string, exts []Extension) string {
	parts := []string{base}
	for _, ext := range exts {
		if s := strings.TrimSpace(ext.Instructions); s != "" {
			parts = append(parts, s)
		}
	}
	return strings.Join(parts, "\n\n")
}

// registry tracks names and URIs already claimed on the server.
type registry struct {
	tools, prompts, resources, templates map[string]bool
}

func newRegistry(srv *server.MCPServer) *registry {
	r := &registry{
		tools:     map[string]bool{},
		prompts:   map[string]bool{},
		resources: map[string]bool{},
		templates: map[string]bool{},
	}
	for name := range srv.ListTools() {
		r.tools[name] = true
	}
	for name := range srv.ListPrompts() {
		r.prompts[name] = true
	}
	for uri := range srv.ListResources() {
		r.resources[uri] = true
	}
	return r
}

func (r *registry) claimAll(ext Extension) error {
	for _, t := range ext.Tools {
		if err := claim(r.tools, t.Tool.Name, t.Handler != nil); err != nil {
			return fmt.Errorf("tool %w", err)
		}
	}
	for _, p := range ext.Prompts {
		if err := claim(r.prompts, p.Prompt.Name, p.Handler != nil); err != nil {
			return fmt.Errorf("prompt %w", err)
		}
	}
	for _, res := range ext.Resources {
		if err := claimURI(r.resources, res.Resource.URI, res.Handler != nil); err != nil {
			return fmt.Errorf("resource %w", err)
		}
	}
	for _, rt := range ext.ResourceTemplates {
		uri := ""
		if rt.Template.URITemplate != nil {
			uri = rt.Template.URITemplate.Raw()
		}
		if err := claimURI(r.templates, uri, rt.Handler != nil); err != nil {
			return fmt.Errorf("resource template %w", err)
		}
	}
	return nil
}

// applyExtensions validates every extension against the already-registered built-ins and each
// other, then registers them. Nothing is registered if any extension is invalid.
func applyExtensions(srv *server.MCPServer, exts []Extension) error {
	reg := newRegistry(srv)
	for i, ext := range exts {
		if err := reg.claimAll(ext); err != nil {
			label := ext.Name
			if label == "" {
				label = fmt.Sprintf("#%d", i)
			}
			return fmt.Errorf("mcp extension %s: %w", label, err)
		}
	}

	for _, ext := range exts {
		if len(ext.Tools) > 0 {
			srv.AddTools(ext.Tools...)
		}
		if len(ext.Prompts) > 0 {
			srv.AddPrompts(ext.Prompts...)
		}
		if len(ext.Resources) > 0 {
			srv.AddResources(ext.Resources...)
		}
		if len(ext.ResourceTemplates) > 0 {
			srv.AddResourceTemplates(ext.ResourceTemplates...)
		}
	}
	return nil
}

func claim(seen map[string]bool, name string, hasHandler bool) error {
	switch {
	case name == "":
		return fmt.Errorf("has no name")
	case !hasHandler:
		return fmt.Errorf("%q has no handler", name)
	case seen[name]:
		return fmt.Errorf("%q is already registered", name)
	}
	seen[name] = true
	return nil
}

func claimURI(seen map[string]bool, uri string, hasHandler bool) error {
	if strings.HasPrefix(uri, reservedURIScheme) {
		return fmt.Errorf("%q uses the reserved %s scheme", uri, reservedURIScheme)
	}
	return claim(seen, uri, hasHandler)
}
