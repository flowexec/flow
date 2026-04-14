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

	"github.com/flowexec/flow/pkg/filesystem"
)

func addServerResources(srv *server.MCPServer) {
	// Resource template: workspace by name
	workspaceTemplate := mcp.NewResourceTemplate(
		"flow://workspace/{name}",
		"Workspace Configuration",
		mcp.WithTemplateDescription("Workspace metadata and configuration as JSON"),
		mcp.WithTemplateMIMEType("application/json"),
	)
	srv.AddResourceTemplate(workspaceTemplate, workspaceResourceHandler)

	// Resource template: executable by workspace/namespace/name
	executableTemplate := mcp.NewResourceTemplate(
		"flow://executable/{workspace}/{namespace}/{name}",
		"Executable Definition",
		mcp.WithTemplateDescription("Executable definition and metadata as JSON"),
		mcp.WithTemplateMIMEType("application/json"),
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
		Name:        ws.AssignedName(),
		Path:        ws.Location(),
		DisplayName: ws.DisplayName,
		Description: ws.Description,
		Tags:        ws.Tags,
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
			MIMEType: "application/json",
			Text:     string(jsonData),
		},
	}, nil
}

func executableResourceHandler(_ context.Context, request mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
	uri := request.Params.URI
	// Parse flow://executable/{workspace}/{namespace}/{name}
	parts := extractExecutableURIParts(uri)
	if parts.workspace == "" || parts.namespace == "" || parts.name == "" {
		return nil, fmt.Errorf("invalid executable URI: expected flow://executable/{workspace}/{namespace}/{name}")
	}

	cfg, err := filesystem.LoadConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	wsPath, ok := cfg.Workspaces[parts.workspace]
	if !ok {
		return nil, fmt.Errorf("workspace %q not found", parts.workspace)
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
				}

				jsonData, err := json.MarshalIndent(output, "", "  ")
				if err != nil {
					return nil, fmt.Errorf("failed to marshal executable: %w", err)
				}

				return []mcp.ResourceContents{
					mcp.TextResourceContents{
						URI:      uri,
						MIMEType: "application/json",
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

	absPath := path
	if !filepath.IsAbs(path) {
		// Try to resolve relative to current workspace
		if cfg.CurrentWorkspace != "" {
			if wsPath, ok := cfg.Workspaces[cfg.CurrentWorkspace]; ok {
				absPath = filepath.Join(wsPath, path)
			}
		}
	}

	// Security check: path must be within a registered workspace
	if !isPathInWorkspace(absPath, cfg.Workspaces) {
		return nil, fmt.Errorf("path %q is not within a registered workspace", path)
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

type executableURIParts struct {
	workspace string
	namespace string
	name      string
}

// extractExecutableURIParts parses flow://executable/{workspace}/{namespace}/{name}
func extractExecutableURIParts(uri string) executableURIParts {
	const prefix = "flow://executable/"
	trimmed := strings.TrimPrefix(uri, prefix)
	parts := strings.SplitN(trimmed, "/", 3)
	if len(parts) != 3 {
		return executableURIParts{}
	}
	return executableURIParts{
		workspace: parts[0],
		namespace: parts[1],
		name:      parts[2],
	}
}

// isPathInWorkspace checks if an absolute path is within any registered workspace.
func isPathInWorkspace(absPath string, workspaces map[string]string) bool {
	for _, wsPath := range workspaces {
		if strings.HasPrefix(absPath, wsPath) {
			return true
		}
	}
	return false
}
