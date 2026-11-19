package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	tuikitIO "github.com/flowexec/tuikit/io"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"github.com/flowexec/flow/v2/pkg/filesystem"
	"github.com/flowexec/flow/v2/types/config"
)

const mimeJSON = "application/json"

func addServerResources(srv *server.MCPServer, executor CommandExecutor) {
	// Static resource: server instructions.
	instructionsResource := mcp.NewResource(
		"flow://server/instructions",
		"Server Instructions",
		mcp.WithResourceDescription("The MCP server's usage instructions — the same text sent as "+
			"the server instructions on connect. Re-read after switching workspaces or worktrees."),
		mcp.WithMIMEType("text/markdown"),
	)
	srv.AddResource(instructionsResource, instructionsResourceHandler)

	// Resource template: workspace by name
	workspaceTemplate := mcp.NewResourceTemplate(
		"flow://workspace/{name}",
		"Workspace Configuration",
		mcp.WithTemplateDescription("Workspace metadata and configuration as JSON"),
		mcp.WithTemplateMIMEType(mimeJSON),
	)
	srv.AddResourceTemplate(workspaceTemplate, workspaceResourceHandler)

	// Resource template: executable by workspace/namespace/name
	executableTemplate := mcp.NewResourceTemplate(
		executableURITemplate,
		"Executable Definition",
		mcp.WithTemplateDescription("Executable definition and metadata as JSON"),
		mcp.WithTemplateMIMEType(mimeJSON),
	)
	srv.AddResourceTemplate(executableTemplate, executableResourceHandler)

	// Resource template: flowfile by path
	flowfileTemplate := mcp.NewResourceTemplate(
		"flow://flowfile/{+path}",
		"Flow File",
		mcp.WithTemplateDescription("Raw flowfile YAML content"),
		mcp.WithTemplateMIMEType("text/yaml"),
	)
	srv.AddResourceTemplate(flowfileTemplate, flowfileResourceHandler)

	// Resource template: execution log by run ID
	logsTemplate := mcp.NewResourceTemplate(
		"flow://logs/{run_id}",
		"Execution Log",
		mcp.WithTemplateDescription("Output of a specific execution run as plain text"),
		mcp.WithTemplateMIMEType("text/plain"),
	)
	srv.AddResourceTemplate(logsTemplate, logsResourceHandler)

	// Static resource: recent failures.
	failuresResource := mcp.NewResource(
		"flow://failures/recent",
		"Recent Failures",
		mcp.WithResourceDescription("The last few failed executions with error output, "+
			"for picking up where a previous session left off."),
		mcp.WithMIMEType(mimeJSON),
	)
	srv.AddResource(failuresResource, recentFailuresHandler(executor))
}

const recentFailuresLimit = 5

func recentFailuresHandler(executor CommandExecutor) server.ResourceHandlerFunc {
	return func(_ context.Context, _ mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
		output, err := executor.Execute(
			"logs", "--status", "failed", "--content", "--tail", "30",
			"--max-bytes", "10000", outputFlag, outputJSON,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to query logs: %w", err)
		}

		var cliOutput struct {
			History []LogEntry `json:"history"`
		}
		if err := json.Unmarshal([]byte(output), &cliOutput); err != nil {
			return nil, fmt.Errorf("failed to parse logs: %w", err)
		}

		entries := cliOutput.History
		if len(entries) > recentFailuresLimit {
			entries = entries[len(entries)-recentFailuresLimit:]
		}

		result := struct {
			Failures []LogEntry `json:"failures"`
			Count    int        `json:"count"`
		}{
			Failures: entries,
			Count:    len(entries),
		}
		jsonData, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			return nil, fmt.Errorf("failed to marshal failures: %w", err)
		}

		return []mcp.ResourceContents{
			mcp.TextResourceContents{
				URI:      "flow://failures/recent",
				MIMEType: mimeJSON,
				Text:     string(jsonData),
			},
		}, nil
	}
}

func instructionsResourceHandler(_ context.Context, _ mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
	return []mcp.ResourceContents{
		mcp.TextResourceContents{
			URI:      "flow://server/instructions",
			MIMEType: "text/markdown",
			Text:     serverInstructions,
		},
	}, nil
}

func workspaceResourceHandler(_ context.Context, request mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
	uri := request.Params.URI
	name := extractURIParam(uri, "flow://workspace/")
	if name == "" {
		return nil, fmt.Errorf("workspace name is required in URI")
	}

	cfg, err := filesystem.LoadConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	wsPath, ok := cfg.Workspaces[name]
	if !ok {
		return nil, fmt.Errorf("workspace %q not found", name)
	}

	ws, err := filesystem.LoadWorkspaceConfig(name, wsPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load workspace config: %w", err)
	}

	output := WorkspaceOutput{
		Name:              ws.AssignedName(),
		Path:              ws.Location(),
		DisplayName:       ws.DisplayName,
		Description:       ws.Description,
		Tags:              ws.Tags,
		Annotations:       map[string]string(ws.Annotations),
		AgentInstructions: ws.ResolvedAgentInstructions(),
	}
	if ws.Executables != nil {
		output.Executables = &ExecutableFilter{
			Included: ws.Executables.Included,
			Excluded: ws.Executables.Excluded,
		}
	}

	jsonData, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal workspace: %w", err)
	}

	return []mcp.ResourceContents{
		mcp.TextResourceContents{
			URI:      uri,
			MIMEType: mimeJSON,
			Text:     string(jsonData),
		},
	}, nil
}

func executableResourceHandler(_ context.Context, request mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
	uri := request.Params.URI
	// Parse flow://executable/{workspace}/{namespace}/{name}.
	// Empty workspace or namespace segments resolve to the current context
	// (workspace required; namespace may itself be empty = no-namespace executable).
	parts := extractExecutableURIParts(uri)
	if parts.name == "" {
		return nil, fmt.Errorf("invalid executable URI: name is required")
	}

	cfg, err := filesystem.LoadConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	resolved, err := filesystem.ResolveWorkspace(cfg, filesystem.ResolveOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to resolve current workspace: %w", err)
	}
	if parts.workspace == "" {
		if resolved == nil {
			return nil, fmt.Errorf("workspace is empty in URI and no current workspace is set")
		}
		parts.workspace = resolved.Name
	}
	if parts.namespace == "" {
		parts.namespace = cfg.CurrentNamespace
	}

	wsPath, ok := cfg.Workspaces[parts.workspace]
	if !ok {
		// The named workspace may be one discovered from the working directory, which is absent
		// from the config but still perfectly runnable.
		if resolved == nil || resolved.Name != parts.workspace {
			return nil, fmt.Errorf("workspace %q not found", parts.workspace)
		}
		wsPath = resolved.Path
	}

	ws, err := filesystem.LoadWorkspaceConfig(parts.workspace, wsPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load workspace config: %w", err)
	}

	flowFiles, err := filesystem.LoadWorkspaceFlowFiles(ws)
	if err != nil {
		return nil, fmt.Errorf("failed to load workspace flow files: %w", err)
	}

	for _, ff := range flowFiles {
		for _, exec := range ff.Executables {
			if exec.Name == parts.name && ff.Namespace == parts.namespace {
				visibility := ""
				if exec.Visibility != nil {
					visibility = string(*exec.Visibility)
				}
				output := ExecutableOutput{
					ID:          exec.ID(),
					Ref:         exec.Ref().String(),
					Name:        exec.Name,
					Namespace:   ff.Namespace,
					Workspace:   parts.workspace,
					FlowFile:    ff.ConfigPath(),
					Description: exec.Description,
					Verb:        string(exec.Verb),
					Visibility:  visibility,
					Tags:        exec.Tags,
					Aliases:     exec.Aliases,
					Annotations: map[string]string(exec.Annotations),
				}

				jsonData, err := json.MarshalIndent(output, "", "  ")
				if err != nil {
					return nil, fmt.Errorf("failed to marshal executable: %w", err)
				}

				return []mcp.ResourceContents{
					mcp.TextResourceContents{
						URI:      uri,
						MIMEType: mimeJSON,
						Text:     string(jsonData),
					},
				}, nil
			}
		}
	}

	return nil, fmt.Errorf("executable %s/%s:%s not found", parts.workspace, parts.namespace, parts.name)
}

func flowfileResourceHandler(_ context.Context, request mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
	uri := request.Params.URI
	path := extractURIParam(uri, "flow://flowfile/")
	if path == "" {
		return nil, fmt.Errorf("path is required in URI")
	}

	// Validate path is within a registered workspace
	cfg, err := filesystem.LoadConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	resolved, err := filesystem.ResolveWorkspace(cfg, filesystem.ResolveOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to resolve current workspace: %w", err)
	}

	absPath := path
	if !filepath.IsAbs(path) && resolved != nil {
		absPath = filepath.Join(resolved.Path, path)
	}

	// Security check: path must be within a workspace flow can see from here
	if !isPathInWorkspace(absPath, accessibleWorkspaceRoots(cfg, resolved)) {
		return nil, fmt.Errorf("path %q is not within a known workspace", path)
	}

	data, err := os.ReadFile(filepath.Clean(absPath))
	if err != nil {
		return nil, fmt.Errorf("failed to read flowfile: %w", err)
	}

	return []mcp.ResourceContents{
		mcp.TextResourceContents{
			URI:      uri,
			MIMEType: "text/yaml",
			Text:     string(data),
		},
	}, nil
}

func logsResourceHandler(_ context.Context, request mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
	uri := request.Params.URI
	runID := extractURIParam(uri, "flow://logs/")
	if runID == "" {
		return nil, fmt.Errorf("run_id is required in URI")
	}

	logsDir := filesystem.LogsDir()
	entries, err := tuikitIO.ListArchiveEntries(logsDir)
	if err != nil {
		return nil, fmt.Errorf("failed to list log entries: %w", err)
	}

	for _, entry := range entries {
		if entry.ID == runID {
			content, err := entry.Read()
			if err != nil {
				return nil, fmt.Errorf("failed to read log entry: %w", err)
			}
			return []mcp.ResourceContents{
				mcp.TextResourceContents{
					URI:      uri,
					MIMEType: "text/plain",
					Text:     content,
				},
			}, nil
		}
	}

	return nil, fmt.Errorf("log entry with run ID %q not found", runID)
}

// extractURIParam extracts the value after a prefix from a URI.
func extractURIParam(uri, prefix string) string {
	if !strings.HasPrefix(uri, prefix) {
		return ""
	}
	return strings.TrimPrefix(uri, prefix)
}

// executableURITemplate uses reserved expansion for the namespace so nested namespaces
// (which contain `/`) still match.
const executableURITemplate = "flow://executable/{workspace}/{+namespace}/{name}"

type executableURIParts struct {
	workspace string
	namespace string
	name      string
}

// extractExecutableURIParts parses flow://executable/{workspace}/{namespace}/{name}. The namespace
// may itself contain `/` (nested namespaces), so it is everything between the first and last segment.
func extractExecutableURIParts(uri string) executableURIParts {
	const prefix = "flow://executable/"
	trimmed := strings.TrimPrefix(uri, prefix)
	ws, rest, ok := strings.Cut(trimmed, "/")
	if !ok {
		return executableURIParts{}
	}
	i := strings.LastIndex(rest, "/")
	if i < 0 {
		return executableURIParts{}
	}
	return executableURIParts{
		workspace: ws,
		namespace: rest[:i],
		name:      rest[i+1:],
	}
}

// isPathInWorkspace checks if an absolute path is within any registered workspace.
// isPathInWorkspace reports whether absPath lies inside one of the given workspace roots. The
// comparison is path-segment aware, so "/src/wsX" is not treated as being inside "/src/ws".
func isPathInWorkspace(absPath string, roots []string) bool {
	for _, root := range roots {
		if filesystem.IsPathWithin(absPath, root) {
			return true
		}
	}
	return false
}

// accessibleWorkspaceRoots returns every directory a resource may be read from: the registered
// workspaces plus the workspace resolved from the working directory, which may be a discovered
// one that appears nowhere in the config.
func accessibleWorkspaceRoots(cfg *config.Config, resolved *filesystem.ResolvedWorkspace) []string {
	roots := make([]string, 0, len(cfg.Workspaces)+1)
	for _, wsPath := range cfg.Workspaces {
		roots = append(roots, wsPath)
	}
	if resolved != nil {
		roots = append(roots, resolved.Path)
	}
	return roots
}
