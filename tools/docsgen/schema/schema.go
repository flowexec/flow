package schema

import (
	"strings"
)

type FieldKey string

func (f FieldKey) String() string {
	return string(f)
}

func (f FieldKey) Lower() string {
	return FieldCase(string(f))
}

func (f FieldKey) Title() string {
	return TitleCase(string(f))
}

//nolint:lll
type JSONSchema struct {
	Schema               string                   `json:"$schema,omitempty"              yaml:"$schema,omitempty"`
	Ref                  Ref                      `json:"$ref,omitempty"                 yaml:"$ref,omitempty"`
	ID                   string                   `json:"$id,omitempty"                  yaml:"$id,omitempty"`
	Title                string                   `json:"title,omitempty"                yaml:"title,omitempty"`
	Description          string                   `json:"description,omitempty"          yaml:"description,omitempty"`
	Type                 string                   `json:"type,omitempty"                 yaml:"type,omitempty"`
	Required             []string                 `json:"required,omitempty"             yaml:"required,omitempty"`
	Default              interface{}              `json:"default,omitempty"              yaml:"default,omitempty"`
	Enum                 []string                 `json:"enum,omitempty"                 yaml:"enum,omitempty"`
	Pattern              string                   `json:"pattern,omitempty"              yaml:"pattern,omitempty"`
	AnyOf                []*JSONSchema            `json:"anyOf,omitempty"                yaml:"anyOf,omitempty"`
	Definitions          map[FieldKey]*JSONSchema `json:"definitions,omitempty"          yaml:"definitions,omitempty"`
	Properties           map[FieldKey]*JSONSchema `json:"properties,omitempty"           yaml:"properties,omitempty"`
	AdditionalProperties *JSONSchema              `json:"additionalProperties,omitempty" yaml:"additionalProperties,omitempty"`
	Items                *JSONSchema              `json:"items,omitempty"                yaml:"items,omitempty"`
	Ext                  SchemaExt                `json:"-"                              yaml:"goJSONSchema,omitempty"`
	Docsgen              DocsgenExt               `json:"-"                              yaml:"docsgen,omitempty"`
}

type SchemaExt struct {
	Type       string `json:"-" yaml:"type"`
	Identifier string `json:"-" yaml:"identifier"`
}

func (e SchemaExt) IsExported() bool {
	return e.Identifier == "" || (e.Identifier[0] >= 'A' && e.Identifier[0] <= 'Z')
}

//nolint:all
func MergeSchemas(dst, src *JSONSchema, dstFile FileName, schemaMap map[FileName]*JSONSchema) {
	if src.Items != nil {
		MergeSchemas(dst, src.Items, dstFile, schemaMap)
	}
	for _, value := range src.Definitions {
		if value.Ref.IsRoot() {
			continue
		}
		MergeSchemas(dst, value, dstFile, schemaMap)
	}
	for key, value := range src.Properties {
		if !value.Ext.IsExported() {
			delete(src.Properties, FieldKey(value.Ext.Identifier))
			continue
		}
		MergeSchemas(dst, value, dstFile, schemaMap)
		src.Properties[key].Ref = convertToLocalSchemaRef(value.Ref, dstFile)
	}

	var match *JSONSchema
	switch {
	case src.Ref.String() == "":
		// the source is not a reference
		if src.Items != nil {
			MergeSchemas(dst, src.Items, dstFile, schemaMap)
		}
		for key, value := range src.Properties {
			if !value.Ext.IsExported() {
				delete(src.Properties, key)
				continue
			}
			MergeSchemas(dst, value, dstFile, schemaMap)
			src.Properties[key].Ref = convertToLocalSchemaRef(value.Ref, dstFile)
		}
		for i, value := range src.Definitions {
			MergeSchemas(dst, value, dstFile, schemaMap)
			src.Definitions[i].Ref = convertToLocalSchemaRef(value.Ref, dstFile)
		}
	case src.Ref.ExternalFile() == "" && dst.Definitions[src.Ref.Key()] == nil:
		// the ref is a local definition but doesn't exist in the destination schema
		// should never happen if defined correctly
		match = &JSONSchema{
			Type:        src.Type,
			Description: src.Description,
			Required:    src.Required,
			Default:     src.Default,
			Enum:        src.Enum,
		}
	case src.Ref.ExternalFile() == "":
		// the ref is a local definition and exists in the destination schema
		for key, value := range src.Properties {
			if !value.Ext.IsExported() {
				delete(src.Properties, key)
				continue
			}
			MergeSchemas(dst, value, dstFile, schemaMap)
		}
	default:
		// the ref is an external definition
		for fn, s := range schemaMap {
			if src.Ref.ExternalFile().Title() == fn.Title() {
				if FieldKey(fn.Title()) == src.Ref.Key() {
					// root level reference
					match = s
					match.ID = ""
				} else {
					def, found := s.Definitions[src.Ref.Key()]
					if !found {
						continue
					}
					match = def
				}
				if match.Items != nil {
					match.Items.Ref = expandLocalSchemaRef(match.Items.Ref, fn)
					MergeSchemas(dst, match.Items, dstFile, schemaMap)
					match.Items.Ref = convertToLocalSchemaRef(match.Items.Ref, dstFile)
				}
				for _, value := range match.Properties {
					if !value.Ext.IsExported() {
						continue
					}
					value.Ref = expandLocalSchemaRef(value.Ref, fn)
					MergeSchemas(dst, value, dstFile, schemaMap)
				}
				break
			}
		}
	}

	if match == nil {
		return
	}

	src.Ref = convertToLocalSchemaRef(src.Ref, dstFile)
	if _, found := dst.Definitions[src.Ref.Key()]; !found {
		var d JSONSchema
		d = *match
		d.Schema = ""
		d.Definitions = nil
		dst.Definitions[src.Ref.Key()] = &d
	}
	for key, value := range match.Properties {
		if !value.Ext.IsExported() {
			delete(match.Properties, key)
			continue
		}
		MergeSchemas(dst, value, dstFile, schemaMap)
	}
	if match.Items != nil {
		MergeSchemas(dst, match.Items, dstFile, schemaMap)
	}
	for _, value := range match.Definitions {
		MergeSchemas(dst, value, dstFile, schemaMap)
	}
}

func FieldCase(s string) string {
	if s == "" {
		return ""
	}
	firstLetter := strings.ToLower(s[:1])
	return firstLetter + s[1:]
}

func TitleCase(s string) string {
	if s == "" {
		return ""
	} else if s == strings.ToLower(FlowfileDefinitionTitle) {
		return "FlowFile"
	}
	firstLetter := strings.ToUpper(s[:1])
	return firstLetter + s[1:]
}

type DocsgenExt struct {
	// OpenPattern publishes an enum as "one of the enum values, or any string matching this
	// pattern", so editors still suggest the enum values without rejecting others.
	OpenPattern string `json:"-" yaml:"openPattern"`
}

func (s *JSONSchema) ApplyOpenPattern() {
	if s.Docsgen.OpenPattern == "" || len(s.Enum) == 0 {
		return
	}
	s.AnyOf = []*JSONSchema{{Enum: s.Enum}, {Pattern: s.Docsgen.OpenPattern}}
	s.Enum = nil
}
