package internal

import (
	"fmt"
	"path/filepath"
	"slices"
	"strings"

	"github.com/spf13/cobra"

	errhandler "github.com/flowexec/flow/cmd/internal/errors"
	"github.com/flowexec/flow/cmd/internal/flags"
	"github.com/flowexec/flow/cmd/internal/response"
	"github.com/flowexec/flow/internal/validation"
	"github.com/flowexec/flow/pkg/context"
	flowerrors "github.com/flowexec/flow/pkg/errors"
)

func RegisterSchemaCmd(ctx *context.Context, rootCmd *cobra.Command) {
	schemaCmd := &cobra.Command{
		Use:   "schema",
		Short: "Validate flowfiles and workspace configs against their schemas.",
		Long:  schemaLong,
	}

	registerSchemaValidateCmd(ctx, schemaCmd)
	rootCmd.AddCommand(schemaCmd)
}

func registerSchemaValidateCmd(ctx *context.Context, parent *cobra.Command) {
	validateCmd := &cobra.Command{
		Use:   "validate FILE...",
		Short: "Validate flow files and workspace configs against their schemas.",
		Long: "Validate one or more flow files or workspace configuration files against their JSON schemas. " +
			"File type is auto-detected from the filename (*.flow for flow files, flow.yaml for workspace configs). " +
			"Use --type to override auto-detection. Use --strict to also check for unknown keys.",
		Args: cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			schemaValidateFunc(ctx, cmd, args)
		},
	}

	RegisterFlag(ctx, validateCmd, *flags.StrictFlag)
	RegisterFlag(ctx, validateCmd, *flags.FileTypeFlag)
	RegisterFlag(ctx, validateCmd, *flags.OutputFormatFlag)
	parent.AddCommand(validateCmd)
}

type fileResult struct {
	File   string             `json:"file"`
	Type   string             `json:"type"`
	Valid  bool               `json:"valid"`
	Errors []validation.Issue `json:"errors,omitempty"`
}

func schemaValidateFunc(ctx *context.Context, cmd *cobra.Command, args []string) {
	strict := flags.ValueFor[bool](cmd, *flags.StrictFlag, false)
	typeOverride := flags.ValueFor[string](cmd, *flags.FileTypeFlag, false)

	if typeOverride != "" {
		validTypes := validation.ValidFileTypes()
		if !slices.Contains(validTypes, typeOverride) {
			errhandler.HandleUsage(ctx, cmd, "invalid --type %q; must be one of: %s",
				typeOverride, strings.Join(validTypes, ", "))
		}
	}

	results := validateFiles(ctx, cmd, args, validation.FileType(typeOverride), strict)

	anyInvalid := false
	for _, r := range results {
		if !r.Valid {
			anyInvalid = true
			break
		}
	}

	if anyInvalid {
		handleValidationFailure(ctx, cmd, results)
		return
	}

	filesSummary := make([]any, 0, len(results))
	for _, r := range results {
		filesSummary = append(filesSummary, map[string]any{"file": r.File, "type": r.Type})
	}
	response.HandleSuccess(ctx, cmd, fmt.Sprintf("%d file(s) valid", len(results)), map[string]any{
		"files": filesSummary,
	})
}

func validateFiles(
	ctx *context.Context,
	cmd *cobra.Command,
	args []string,
	ft validation.FileType,
	strict bool,
) []fileResult {
	var results []fileResult
	for _, arg := range args {
		abs, err := filepath.Abs(arg)
		if err != nil {
			errhandler.HandleFatal(ctx, cmd, fmt.Errorf("resolving path %q: %w", arg, err))
		}

		fileFT := ft
		if fileFT == "" {
			detected, err := validation.DetectFileType(abs)
			if err != nil {
				errhandler.HandleFatal(ctx, cmd, err)
			}
			fileFT = detected
		}

		result, err := validation.ValidateFile(abs, fileFT, strict)
		if err != nil {
			errhandler.HandleFatal(ctx, cmd, err)
		}

		fr := fileResult{File: abs, Type: string(fileFT), Valid: result.Valid}
		if !result.Valid {
			fr.Errors = result.Errors
		}
		results = append(results, fr)
	}
	return results
}

const schemaLong = `Utilities for working with flow YAML schemas. Use these commands to validate flowfiles
and workspace configs against their JSON schemas — useful in CI pipelines and pre-commit hooks.`

func handleValidationFailure(ctx *context.Context, cmd *cobra.Command, results []fileResult) {
	var msgs []string
	filesWithErrors := make([]map[string]any, 0)
	for _, r := range results {
		if r.Valid {
			continue
		}
		var errStrs []string
		for _, e := range r.Errors {
			errStrs = append(errStrs, e.String())
		}
		msgs = append(msgs, fmt.Sprintf("%s:\n  %s", r.File, strings.Join(errStrs, "\n  ")))

		issueList := make([]any, 0, len(r.Errors))
		for _, e := range r.Errors {
			issueList = append(issueList, map[string]any{"path": e.Path, "message": e.Message})
		}
		filesWithErrors = append(filesWithErrors, map[string]any{
			"file":   r.File,
			"type":   r.Type,
			"errors": issueList,
		})
	}

	errhandler.HandleFatal(ctx, cmd, flowerrors.NewValidationError(
		fmt.Sprintf("validation failed for %d file(s):\n%s", len(filesWithErrors), strings.Join(msgs, "\n")),
		map[string]any{"files": filesWithErrors},
	))
}
