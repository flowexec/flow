package templates

import (
	"fmt"

	"github.com/flowexec/tuikit/views"

	"github.com/flowexec/flow/pkg/context"
	"github.com/flowexec/flow/pkg/logger"
	"github.com/flowexec/flow/types/executable"
)

func showForm(ctx *context.Context, fields executable.FormFields) error {
	if len(fields) == 0 {
		return nil
	}
	in := ctx.StdIn()
	out := ctx.StdOut()

	if err := fields.Validate(); err != nil {
		return fmt.Errorf("invalid form fields: %w", err)
	}
	var ff []*views.FormField
	for _, f := range fields {
		var t views.FormFieldType
		switch f.Type {
		case executable.FieldTypeMasked:
			t = views.PromptTypeMasked
		case executable.FieldTypeMultiline:
			t = views.PromptTypeMultiline
		case executable.FieldTypeConfirm:
			t = views.PromptTypeConfirm
		case executable.FieldTypeText:
			fallthrough
		default:
			t = views.PromptTypeText
		}
		ff = append(ff, &views.FormField{
			Key:            f.Key,
			Type:           t,
			Group:          uint(f.Group),
			Description:    f.Description,
			Default:        f.Default,
			Title:          f.Prompt,
			Placeholder:    f.Default,
			Required:       f.Required,
			ValidationExpr: f.Validate,
		})
	}
	form, err := views.NewForm(logger.Theme(ctx.Config.Theme.String()), in, out, ff...)
	if err != nil {
		return fmt.Errorf("encountered form init error: %w", err)
	}
	if err = form.Run(ctx); err != nil {
		return fmt.Errorf("encountered form run error: %w", err)
	}
	for _, f := range fields {
		v, ok := form.ValueMap()[f.Key]
		if !ok {
			continue
		}
		f.Set(fmt.Sprintf("%v", v))
	}
	return nil
}
