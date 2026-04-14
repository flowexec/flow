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
	schemaDir = "docs/public/schemas"
	idBase    = "https://flowexec.io/schemas"
)

// The JSON schemas are published at https://flowexec.io/schemas/*.
func generateJSONSchemas() {
	sm := schema.RegisteredSchemaMap()
	for fn, s := range sm {
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
