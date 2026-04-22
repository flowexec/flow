package validation

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/santhosh-tekuri/jsonschema/v6"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
	"gopkg.in/yaml.v3"

	"github.com/flowexec/flow/types/executable"
)

// FileType identifies which schema to validate against.
type FileType string

const (
	FileTypeFlowFile  FileType = "flowfile"
	FileTypeWorkspace FileType = "workspace"
	FileTypeConfig    FileType = "config"
	FileTypeTemplate  FileType = "template"
)

// Issue represents a single validation error.
type Issue struct {
	Path    string `json:"path"`
	Message string `json:"message"`
}

func (i Issue) String() string {
	if i.Path == "" {
		return i.Message
	}
	return fmt.Sprintf("%s: %s", i.Path, i.Message)
}

// Result holds the outcome of a validation.
type Result struct {
	Valid  bool    `json:"valid"`
	Errors []Issue `json:"errors,omitempty"`
}

// ValidFileTypes returns the list of accepted --type values.
func ValidFileTypes() []string {
	return []string{
		string(FileTypeFlowFile),
		string(FileTypeWorkspace),
		string(FileTypeConfig),
		string(FileTypeTemplate),
	}
}

// DetectFileType infers the schema type from the file's name.
func DetectFileType(filePath string) (FileType, error) {
	base := filepath.Base(filePath)
	switch {
	case base == "flow.yaml":
		return FileTypeWorkspace, nil
	case executable.HasFlowFileExt(base):
		return FileTypeFlowFile, nil
	default:
		return "", fmt.Errorf(
			"cannot auto-detect file type for %q; use --type (%s)",
			base, strings.Join(ValidFileTypes(), ", "),
		)
	}
}

// ValidateFile reads a file from disk and validates it against the appropriate schema.
func ValidateFile(filePath string, ft FileType, strict bool) (*Result, error) {
	data, err := os.ReadFile(filepath.Clean(filePath))
	if err != nil {
		return nil, fmt.Errorf("reading file: %w", err)
	}
	return ValidateBytes(data, ft, strict)
}

// ValidateBytes validates raw YAML content against the schema for the given file type.
func ValidateBytes(data []byte, ft FileType, strict bool) (*Result, error) {
	var doc any
	if err := yaml.Unmarshal(data, &doc); err != nil {
		return &Result{
			Valid:  false,
			Errors: []Issue{{Message: fmt.Sprintf("invalid YAML: %s", err)}},
		}, nil
	}

	schemaJSON, err := loadSchema(ft, strict)
	if err != nil {
		return nil, fmt.Errorf("loading schema: %w", err)
	}

	compiler := jsonschema.NewCompiler()
	if err := compiler.AddResource(schemaFileName(ft), schemaJSON); err != nil {
		return nil, fmt.Errorf("adding schema resource: %w", err)
	}
	schema, err := compiler.Compile(schemaFileName(ft))
	if err != nil {
		return nil, fmt.Errorf("compiling schema: %w", err)
	}

	validationErr := schema.Validate(doc)
	if validationErr == nil {
		return &Result{Valid: true}, nil
	}

	verr := &jsonschema.ValidationError{}
	ok := errors.As(validationErr, &verr)
	if !ok {
		return nil, fmt.Errorf("unexpected validation error type: %w", validationErr)
	}

	issues := collectIssues(verr)
	return &Result{Valid: false, Errors: issues}, nil
}

func schemaFileName(ft FileType) string {
	return string(ft) + "_schema.json"
}

func loadSchema(ft FileType, strict bool) (any, error) {
	raw, err := schemas.ReadFile(schemaFileName(ft))
	if err != nil {
		return nil, fmt.Errorf("reading embedded schema %s: %w", schemaFileName(ft), err)
	}

	var schemaMap map[string]any
	if err := json.Unmarshal(raw, &schemaMap); err != nil {
		return nil, fmt.Errorf("parsing schema JSON: %w", err)
	}

	if strict {
		injectNoAdditionalProperties(schemaMap)
	}

	return schemaMap, nil
}

// injectNoAdditionalProperties recursively sets additionalProperties: false
// on object types that have "properties" defined but no existing
// "additionalProperties" setting. This preserves intentional map-typed objects
// (like Annotations, VerbAliases) that already set additionalProperties.
func injectNoAdditionalProperties(node map[string]any) {
	typ, _ := node["type"].(string)
	_, hasProps := node["properties"]
	_, hasAdditional := node["additionalProperties"]

	if typ == "object" && hasProps && !hasAdditional {
		node["additionalProperties"] = false
	}

	if props, ok := node["properties"].(map[string]any); ok {
		for _, v := range props {
			if sub, ok := v.(map[string]any); ok {
				injectNoAdditionalProperties(sub)
			}
		}
	}
	if defs, ok := node["definitions"].(map[string]any); ok {
		for _, v := range defs {
			if sub, ok := v.(map[string]any); ok {
				injectNoAdditionalProperties(sub)
			}
		}
	}
	if items, ok := node["items"].(map[string]any); ok {
		injectNoAdditionalProperties(items)
	}
}

// printer is a non-nil message.Printer required by ErrorKind.LocalizedString.
var printer = message.NewPrinter(language.English)

// collectIssues flattens a jsonschema.ValidationError tree into a list of issues.
func collectIssues(verr *jsonschema.ValidationError) []Issue {
	var issues []Issue
	if len(verr.Causes) == 0 {
		path := "/" + strings.Join(verr.InstanceLocation, "/")
		issues = append(issues, Issue{
			Path:    path,
			Message: verr.ErrorKind.LocalizedString(printer),
		})
		return issues
	}
	for _, cause := range verr.Causes {
		issues = append(issues, collectIssues(cause)...)
	}
	return issues
}
