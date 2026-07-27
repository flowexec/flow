package secret

import (
	"errors"
	"fmt"

	"github.com/flowexec/tuikit"
	"github.com/flowexec/tuikit/themes"
	"github.com/flowexec/tuikit/types"
	"github.com/flowexec/tuikit/views"

	ioCommon "github.com/flowexec/flow/v2/internal/io/common"
	"github.com/flowexec/flow/v2/internal/vault"
	"github.com/flowexec/flow/v2/pkg/context"
)

// Views for read-through vaults, which hold links rather than secret material.
//
// The difference from the ordinary secret views is not cosmetic. A linked secret
// lives in another tool, so showing one costs a provider command -- and for
// 1Password, potentially a biometric prompt. Anything that would resolve every
// key at once has to be avoided, and anything that would write has to be
// replaced by something that changes the link instead.

// brokenLinkBody is shown in place of a value when a link no longer resolves.
func brokenLinkBody(err error) string {
	if errors.Is(err, vault.ErrSecretNotFound) {
		return "This link no longer resolves -- the secret it points at has been " +
			"removed or renamed in the provider.\n\nPress 'l' to re-link it, or 'x' to remove the link."
	}
	return fmt.Sprintf("Could not read this secret from its provider:\n\n%v", err)
}

// linkedMetadata describes a link, including where it points.
func linkedMetadata(vlt vault.Vault, links vault.ReferenceVault, ref vault.SecretRef) []views.DetailField {
	reference, err := links.Reference(ref.Key())
	if err != nil {
		reference = fmt.Sprintf("<unresolved: %v>", err)
	}
	return []views.DetailField{
		{Key: "Name", Value: ref.Key()},
		{Key: "Reference", Value: reference},
		{Key: "Vault", Value: vlt.ID()},
	}
}

// linkedSecretCallbacks replaces rename and edit -- both of which wrote secret
// material -- with re-linking, and phrases removal as unlinking.
func linkedSecretCallbacks(
	ctx *context.Context,
	container *tuikit.Container,
	links vault.ReferenceVault,
	ref vault.SecretRef,
	secret vault.Secret,
	loadSecretList func(),
) []types.KeyCallback {
	return []types.KeyCallback{
		{
			Key: "l", Label: "re-link",
			Callback: func() error {
				form, err := views.NewFormView(
					container.RenderState(),
					&views.FormField{
						Key:   "value",
						Type:  views.PromptTypeText,
						Title: "Enter the new reference",
					})
				if err != nil {
					container.HandleError(fmt.Errorf("encountered error creating the form: %w", err))
					return nil
				}
				if err := ctx.SetView(form); err != nil {
					container.HandleError(fmt.Errorf("unable to set view: %w", err))
					return nil
				}
				if err := links.Link(ref.Key(), form.FindByKey("value").Value()); err != nil {
					container.HandleError(fmt.Errorf("unable to re-link: %w", err))
					return nil
				}
				loadSecretList()
				container.SetNotice("link updated", themes.OutputLevelInfo)
				return nil
			},
		},
		{
			Key: "c", Label: "copy",
			Callback: func() error {
				ioCommon.CopyToClipboard(container, secret.PlainTextString(), "secret copied to clipboard")
				return nil
			},
		},
		{
			// Deliberately labelled "unlink" rather than "delete": this removes
			// the link and leaves the secret where it is.
			Key: "x", Label: "unlink",
			Callback: func() error {
				if err := links.Unlink(ref.Key()); err != nil {
					container.HandleError(fmt.Errorf("unable to unlink: %w", err))
					return nil
				}
				loadSecretList()
				container.SetNotice("link removed (the secret itself was not deleted)", themes.OutputLevelInfo)
				return nil
			},
		},
	}
}

// linkedListView lists a read-through vault from its registry, showing where
// each key points instead of a column of identical masks.
func linkedListView(
	ctx *context.Context,
	vlt vault.Vault,
	links vault.ReferenceVault,
	keys []string,
	asPlainText bool,
) tuikit.View {
	container := ctx.TUIContainer()

	if len(keys) == 0 {
		container.HandleError(fmt.Errorf(
			"no secrets linked in vault '%s' yet -- link one with `flow secret link NAME REFERENCE`",
			vlt.ID()))
	}

	columns := []views.TableColumn{
		{Title: fmt.Sprintf("Secrets (%d)", len(keys)), Percentage: 35},
		{Title: "Reference", Percentage: 45},
		{Title: "Vault", Percentage: 20},
	}

	rows := make([]views.TableRow, 0, len(keys))
	for _, key := range keys {
		reference, err := links.Reference(key)
		if err != nil {
			reference = fmt.Sprintf("<unresolved: %v>", err)
		}
		rows = append(rows, views.TableRow{Data: []string{key, reference, vlt.ID()}})
	}

	table := views.NewTable(container.RenderState(), columns, rows, views.TableDisplayMini)
	table.SetOnSelect(func(_ int) error {
		row := table.GetSelectedRow()
		if row == nil || len(row.Data()) < 1 {
			return fmt.Errorf("no secret selected")
		}
		ref := vault.SecretRef(fmt.Sprintf("%s/%s", vlt.ID(), row.Data()[0]))
		view := NewSecretView(ctx, vlt, ref, asPlainText)
		if view == nil {
			// NewSecretView has already reported why.
			return nil
		}
		return container.SetView(view)
	})
	return table
}
