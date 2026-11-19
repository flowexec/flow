//nolint:nilerr
package mcp

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func addTemplateTools(srv *server.MCPServer, executor CommandExecutor) {
	listTemplates := mcp.NewTool("list_templates",
		mcp.WithDescription("List flowfile templates available for scaffolding new executables and workflows. "+
			"Templates are either registered globally or discovered in workspaces (*.flow.tmpl files). "+
			"Each entry includes the template's name, description, form fields, and location."),
		mcp.WithString("cursor", mcp.Description("Next-page cursor")),
	)
	listTemplates.Annotations = mcp.ToolAnnotation{
		Title:           "List flowfile templates",
		DestructiveHint: boolPtr(false), ReadOnlyHint: boolPtr(true),
		IdempotentHint: boolPtr(true), OpenWorldHint: boolPtr(true),
	}
	srv.AddTool(listTemplates, listTemplatesHandler(executor))

	getTemplate := mcp.NewTool("get_template",
		mcp.WithDescription("Get a template's full definition: its form fields, artifacts, pre/post-run hooks, "+
			"and the template body. Use before generate_template to understand what fields are needed."),
		mcp.WithString("name", mcp.Required(), mcp.Description(
			"Template name, as returned by list_templates")),
	)
	getTemplate.Annotations = mcp.ToolAnnotation{
		Title:           "Get template details",
		DestructiveHint: boolPtr(false), ReadOnlyHint: boolPtr(true),
		IdempotentHint: boolPtr(true), OpenWorldHint: boolPtr(true),
	}
	srv.AddTool(getTemplate, getTemplateHandler(executor))

	generateTemplate := mcp.NewTool("generate_template",
		mcp.WithDescription("Scaffold a new flowfile and its artifacts from a template. The template's form "+
			"fields are answered via the fields parameter; unanswered required fields fail before writing."),
		mcp.WithString("name", mcp.Required(), mcp.Description(
			"Template name, as returned by list_templates")),
		mcp.WithString("flowfile_name", mcp.Description(
			"Name for the generated flowfile (defaults to the template's own name)")),
		mcp.WithString("output_dir", mcp.Description(
			"Directory within the workspace to write the flowfile and artifacts into")),
		mcp.WithString("workspace", mcp.Description(
			"Workspace to generate into (default: current workspace)")),
		mcp.WithObject("fields",
			mcp.AdditionalProperties(map[string]any{"type": "string"}),
			mcp.Description("Template form field values as name→value pairs. "+
				"Get the field names from get_template or list_templates.")),
	)
	generateTemplate.Annotations = mcp.ToolAnnotation{
		Title:           "Generate from template",
		DestructiveHint: boolPtr(true), ReadOnlyHint: boolPtr(false),
		IdempotentHint: boolPtr(false), OpenWorldHint: boolPtr(false),
	}
	srv.AddTool(generateTemplate, generateTemplateHandler(srv, executor))
}

type templateEntry struct {
	Name        string         `json:"name"`
	Description string         `json:"description,omitempty"`
	Location    string         `json:"location,omitempty"`
	Form        []templateForm `json:"form,omitempty"`
}

type templateForm struct {
	Key         string `json:"key"`
	Prompt      string `json:"prompt,omitempty"`
	Description string `json:"description,omitempty"`
	Default     string `json:"default,omitempty"`
	Type        string `json:"type,omitempty"`
}

type templateListOutput struct {
	Templates  []templateEntry `json:"templates"`
	NextCursor string          `json:"nextCursor,omitempty"`
	TotalCount int             `json:"totalCount"`
}

func listTemplatesHandler(executor CommandExecutor) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		cursor := request.GetString("cursor", "")

		output, err := executor.Execute("template", "list", outputFlag, outputJSON)
		if err != nil {
			return toolError(ErrCodeExecutionFailed, fmt.Sprintf("Failed to list templates: %s", output)), nil
		}

		var rawTemplates []json.RawMessage
		if err := json.Unmarshal([]byte(output), &rawTemplates); err != nil {
			return mcp.NewToolResultText(output), nil
		}

		entries := make([]templateEntry, 0, len(rawTemplates))
		for _, raw := range rawTemplates {
			var full map[string]json.RawMessage
			if json.Unmarshal(raw, &full) != nil {
				continue
			}
			entry := templateEntry{}
			if v, ok := full["name"]; ok {
				_ = json.Unmarshal(v, &entry.Name)
			}
			if v, ok := full["description"]; ok {
				_ = json.Unmarshal(v, &entry.Description)
			}
			if v, ok := full["location"]; ok {
				_ = json.Unmarshal(v, &entry.Location)
			}
			if v, ok := full["form"]; ok {
				_ = json.Unmarshal(v, &entry.Form)
			}
			if entry.Name != "" {
				entries = append(entries, entry)
			}
		}

		page, nextCursor, totalCount, err := paginate(entries, cursor, defaultPageSize)
		if err != nil {
			return toolError(ErrCodeInvalidInput, err.Error()), nil
		}

		result := templateListOutput{
			Templates:  page,
			NextCursor: nextCursor,
			TotalCount: totalCount,
		}
		jsonData, _ := json.Marshal(result)
		return mcp.NewToolResultStructured(result, string(jsonData)), nil
	}
}

func getTemplateHandler(executor CommandExecutor) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		name, err := request.RequireString("name")
		if err != nil {
			return toolError(ErrCodeInvalidInput, "name is required"), nil
		}

		output, err := executor.Execute("template", "get", "--template", name, outputFlag, outputJSON)
		if err != nil {
			return toolError(ErrCodeNotFound, fmt.Sprintf("Failed to get template %s: %s", name, output)), nil
		}

		return mcp.NewToolResultText(output), nil
	}
}

type generateTemplateOutput struct {
	FlowfileName string            `json:"flowfileName"`
	FlowfilePath string            `json:"flowfilePath"`
	OutputDir    string            `json:"outputDir,omitempty"`
	FormValues   map[string]string `json:"formValues,omitempty"`
	Artifacts    []string          `json:"artifacts,omitempty"`
}

func generateTemplateHandler(srv *server.MCPServer, executor CommandExecutor) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		name, err := request.RequireString("name")
		if err != nil {
			return toolError(ErrCodeInvalidInput, "name is required"), nil
		}

		args := []string{"template", "generate", "--template", name, outputFlag, outputJSON}

		if flowfileName := request.GetString("flowfile_name", ""); flowfileName != "" {
			args = append(args, flowfileName)
		}
		if outputDir := request.GetString("output_dir", ""); outputDir != "" {
			args = append(args, "--dir", outputDir)
		}
		if ws := request.GetString("workspace", ""); ws != "" {
			args = append(args, "--workspace", ws)
		}

		// Pass form field overrides as --field key=value pairs.
		if fieldsRaw := request.GetArguments()["fields"]; fieldsRaw != nil {
			if fields, ok := fieldsRaw.(map[string]any); ok {
				for k, v := range fields {
					args = append(args, "--field", fmt.Sprintf("%s=%v", k, v))
				}
			}
		}

		output, err := executor.Execute(args...)
		if err != nil {
			return toolError(ErrCodeExecutionFailed, fmt.Sprintf("Failed to generate from template: %s", output)), nil
		}

		result, parseErr := extractEnvelopeData[generateTemplateOutput](output)
		if parseErr != nil {
			return mcp.NewToolResultText(output), nil
		}

		srv.SendNotificationToAllClients("notifications/resources/list_changed", nil)

		jsonData, _ := json.Marshal(result)
		return mcp.NewToolResultStructured(result, string(jsonData)), nil
	}
}
