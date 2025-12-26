package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"text/template"

	"golang.org/x/exp/maps"

	"github.com/flowexec/flow/tools/docsgen/schema"
)

const (
	mdDir = "types"
)

var (
	TopLevelPages = []string{
		schema.FlowfileDefinitionTitle,
		schema.ConfigDefinitionTitle,
		schema.WorkspaceDefinitionTitle,
		schema.TemplateDefinitionTitle,
	}
)

type topLevelPage struct {
	Title       string
	Description string
	Required    []string
	Filename    string

	Properties      map[schema.FieldKey]*schema.JSONSchema
	Definitions     map[schema.FieldKey]*schema.JSONSchema
	PropertyOrder   []schema.FieldKey
	DefinitionOrder []schema.FieldKey
}

//nolint:nestif
func generateMarkdownDocs() {
	dir := filepath.Join(rootDir(), DocsDir, mdDir)
	typeTemplate := templateFileData("type.md.tmpl")
	sm := schema.RegisteredSchemaMap()
	for fn, s := range sm {
		if slices.Contains(TopLevelPages, fn.Title()) {
			page := newTopLevelPage(fn, s, sm)
			if page == nil {
				continue
			}
			var buf bytes.Buffer
			tmpl, err := template.New(page.Filename).
				Funcs(template.FuncMap{
					"TypeStr":    typeStr,
					"OneLine":    removeNewlines,
					"IsRequired": requiredStr,
					"DefaultStr": defaultStr,
				}).
				Parse(typeTemplate)
			if err != nil {
				panic(err)
			}

			if err := tmpl.Execute(&buf, page); err != nil {
				panic(err)
			}
			markdown := buf.String()
			filePath := filepath.Clean(filepath.Join(dir, page.Filename))
			file, err := os.Create(filePath)
			if err != nil {
				panic(err)
			}
			defer file.Close()
			if _, err := file.WriteString(markdown); err != nil {
				panic(err)
			}
		}
	}
}

func newTopLevelPage(f schema.FileName, s *schema.JSONSchema, sm map[schema.FileName]*schema.JSONSchema) *topLevelPage {
	if s.Type != "object" {
		return nil
	}
	pOrder := maps.Keys(s.Properties)
	slices.Sort(pOrder)
	dOrder := maps.Keys(s.Definitions)
	slices.Sort(dOrder)

	for key, value := range s.Properties {
		if !value.Ext.IsExported() {
			delete(s.Properties, key)
			continue
		}
		schema.MergeSchemas(s, value, f, sm)
	}
	for _, value := range s.Definitions {
		schema.MergeSchemas(s, value, f, sm)
	}

	return &topLevelPage{
		Title:           f.Title(),
		Description:     s.Description,
		Required:        s.Required,
		Filename:        f.MarkdownFile(),
		Properties:      s.Properties,
		Definitions:     s.Definitions,
		PropertyOrder:   pOrder,
		DefinitionOrder: dOrder,
	}
}

func templateFileData(filename string) string {
	p := filepath.Clean(filepath.Join(rootDir(), "tools/docsgen", filename))
	f, err := os.ReadFile(p)
	if err != nil {
		panic(err)
	}
	return string(f)
}

func typeStr(s *schema.JSONSchema) string {
	name := s.Type
	if s.Ref.String() != "" {
		name = string(s.Ref.Key())
	}
	if strings.Contains(name, ".") {
		parts := strings.Split(name, ".")
		name = ""
		for _, part := range parts {
			name += schema.TitleCase(part)
		}
	}
	standard := []string{"string", "integer", "number", "boolean", "object"}
	switch {
	case name == "array":
		return fmt.Sprintf("`array` (%s)", typeStr(s.Items))
	case name == "object" && s.AdditionalProperties != nil:
		// TODO: look into if this is correct
		return fmt.Sprintf("`map` (`string` -> %s)", typeStr(s.AdditionalProperties))
	case slices.Contains(standard, name) ||
		strings.HasPrefix(name, "map[") ||
		strings.HasPrefix(name, "[]"):
		return fmt.Sprintf("`%s`", name)
	default:
		return fmt.Sprintf("[%s](#%s)", name, strings.ToLower(name))
	}
}

func requiredStr(list []string, key schema.FieldKey) string {
	for _, k := range list {
		if strings.ToLower(k) == key.Lower() {
			return "✘"
		}
	}
	return ""
}

func removeNewlines(s string) string {
	return strings.ReplaceAll(s, "\n", " ")
}

func defaultStr(val interface{}) string {
	if val == nil {
		return ""
	}
	str := fmt.Sprintf("%v", val)
	// Escape HTML-like tags to prevent VitePress parsing issues
	str = strings.ReplaceAll(str, "<", "\\<")
	str = strings.ReplaceAll(str, ">", "\\>")
	return str
}
