//nolint:nilerr
package mcp

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func addSecretTools(srv *server.MCPServer, executor CommandExecutor) {
	listSecrets := mcp.NewTool("list_secrets",
		mcp.WithDescription("List secret names in the current vault. Returns only names and metadata — never "+
			"values. Use this to discover which secrets are available for executables that reference vault entries. "+
			"Secrets are referenced in flowfiles via `$s(NAME)` or as environment variables."),
		mcp.WithString("vault", mcp.Description(
			"Vault name (default: the current vault from get_info's context)")),
		mcp.WithString("cursor", mcp.Description("Next-page cursor")),
	)
	listSecrets.Annotations = mcp.ToolAnnotation{
		Title:           "List secret names",
		DestructiveHint: boolPtr(false), ReadOnlyHint: boolPtr(true),
		IdempotentHint: boolPtr(true), OpenWorldHint: boolPtr(false),
	}
	srv.AddTool(listSecrets, listSecretsHandler(executor))
}

type secretEntry struct {
	Key       string `json:"key"`
	Vault     string `json:"vault"`
	Reference string `json:"reference,omitempty"`
}

type secretListOutput struct {
	Secrets    []secretEntry `json:"secrets"`
	NextCursor string        `json:"nextCursor,omitempty"`
	TotalCount int           `json:"totalCount"`
}

func listSecretsHandler(executor CommandExecutor) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		cursor := request.GetString("cursor", "")

		args := []string{"secret", "list", outputFlag, outputJSON}
		if vault := request.GetString("vault", ""); vault != "" {
			args = append(args, "--vault", vault)
		}

		output, err := executor.Execute(args...)
		if err != nil {
			return toolError(ErrCodeExecutionFailed, fmt.Sprintf("Failed to list secrets: %s", output)), nil
		}

		// The CLI returns an array of {vault, key, value, reference}. Strip the value
		// field — MCP consumers see names and references only.
		var rawSecrets []map[string]any
		if err := json.Unmarshal([]byte(output), &rawSecrets); err != nil {
			return mcp.NewToolResultText(output), nil
		}

		entries := make([]secretEntry, 0, len(rawSecrets))
		for _, raw := range rawSecrets {
			entry := secretEntry{}
			if v, ok := raw["key"].(string); ok {
				entry.Key = v
			}
			if v, ok := raw["vault"].(string); ok {
				entry.Vault = v
			}
			if v, ok := raw["reference"].(string); ok {
				entry.Reference = v
			}
			if entry.Key != "" {
				entries = append(entries, entry)
			}
		}

		page, nextCursor, totalCount, err := paginate(entries, cursor, defaultPageSize)
		if err != nil {
			return toolError(ErrCodeInvalidInput, err.Error()), nil
		}

		result := secretListOutput{
			Secrets:    page,
			NextCursor: nextCursor,
			TotalCount: totalCount,
		}
		jsonData, _ := json.Marshal(result)
		return mcp.NewToolResultStructured(result, string(jsonData)), nil
	}
}
