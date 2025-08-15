package main

import (
	"encoding/json"
	"net/url"
	"os"
	"path/filepath"
	"slices"

	"github.com/flowexec/flow/tools/docsgen/schema"
)

const (
	schemaDir    = "docs/schemas"
	mcpSchemaDir = "internal/mcp/resources"
	idBase       = "https://flowexec.io/schemas"
)

// The JSON schema that's bundled in to the MCP server should always match the schemas that are provided via the docs
// site. Not all schema are needed so below is just an allowlist of schemas that we embed.
var mcpSchemaResources = []string{
	schema.WorkspaceDefinitionTitle,
	schema.FlowfileDefinitionTitle,
	schema.TemplateDefinitionTitle,
}

func generateJSONSchemas() {
	sm := schema.RegisteredSchemaMap()
	for fn, s := range sm {
		//nolint:nestif
		if slices.Contains(TopLevelPages, fn.Title()) {
			updateFileID(s, fn)
			for key, value := range s.Properties {
				if !value.Ext.IsExported() {
					delete(s.Properties, key)
					continue
				}
				schema.MergeSchemas(s, value, fn, sm)
			}
			for _, value := range s.Definitions {
				schema.MergeSchemas(s, value, fn, sm)
			}

			s.Title = fn.Title()
			schemaJSON, err := json.MarshalIndent(s, "", "  ")
			if err != nil {
				panic(err)
			}
			docsPath := filepath.Join(rootDir(), schemaDir, fn.JSONSchemaFile())
			if err := writeSchemaFile(string(schemaJSON), docsPath); err != nil {
				panic(err)
			}
			if slices.Contains(mcpSchemaResources, fn.Title()) {
				mcpPath := filepath.Join(rootDir(), mcpSchemaDir, fn.JSONSchemaFile())
				if err := writeSchemaFile(string(schemaJSON), mcpPath); err != nil {
					panic(err)
				}
			}
		}
	}
}

func writeSchemaFile(content, path string) error {
	filePath := filepath.Clean(path)
	file, err := os.Create(filePath)
	if err != nil {
		panic(err)
	}
	defer file.Close()
	if _, err := file.WriteString(content); err != nil {
		panic(err)
	}
	return nil
}

func updateFileID(s *schema.JSONSchema, file schema.FileName) {
	s.ID, _ = url.JoinPath(idBase, file.JSONSchemaFile())
}
