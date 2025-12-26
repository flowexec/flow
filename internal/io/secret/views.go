//nolint:cyclop,funlen
package secret

import (
	"fmt"

	"github.com/flowexec/tuikit"
	"github.com/flowexec/tuikit/themes"
	"github.com/flowexec/tuikit/types"
	"github.com/flowexec/tuikit/views"

	vault2 "github.com/flowexec/flow/internal/vault"
	"github.com/flowexec/flow/pkg/context"
	"github.com/flowexec/flow/pkg/logger"
)

func NewSecretView(
	ctx *context.Context,
	vlt vault2.Vault,
	ref vault2.SecretRef,
	asPlainText bool,
) tuikit.View {
	container := ctx.TUIContainer
	if ref.Vault() != vlt.ID() {
		err := fmt.Errorf(
			"failure while initializing the secret view secret: vault mismatch -expected %s, got %s",
			vlt.ID(),
			ref.Vault(),
		)
		container.HandleError(err)
		return nil
	}

	s, err := vlt.GetSecret(ref.Key())
	if err != nil {
		container.HandleError(fmt.Errorf("failure while initializing the secret view secret: %w", err))
		return nil
	}

	secret, err := vault2.NewSecret(vlt.ID(), ref.Key(), s)
	if err != nil {
		container.HandleError(fmt.Errorf("failure while initializing the secret view secret: %w", err))
		return nil
	}
	if asPlainText {
		secret = secret.AsPlaintext()
	} else {
		secret = secret.AsObfuscatedText()
	}

	loadSecretList := func() {
		view := NewSecretListView(ctx, vlt, asPlainText)
		if err := ctx.SetView(view); err != nil {
			logger.Log().FatalErr(err)
		}
	}

	var secretKeyCallbacks = []types.KeyCallback{
		{
			Key: "r", Label: "rename",
			Callback: func() error {
				form, err := views.NewFormView(
					container.RenderState(),
					&views.FormField{
						Key:   "value",
						Type:  views.PromptTypeText,
						Title: "Enter the new secret name",
					})
				if err != nil {
					container.HandleError(fmt.Errorf("encountered error creating the form: %w", err))
					return nil
				}
				if err := ctx.SetView(form); err != nil {
					container.HandleError(fmt.Errorf("unable to set view: %w", err))
					return nil
				}
				newName := form.FindByKey("value").Value()
				if err := vlt.SetSecret(newName, secret); err != nil {
					container.HandleError(fmt.Errorf("unable to set secret with new name: %w", err))
					return nil
				}
				if err := vlt.DeleteSecret(ref.Key()); err != nil {
					container.HandleError(fmt.Errorf("unable to delete old secret: %w", err))
					return nil
				}
				loadSecretList()
				container.SetNotice("secret renamed", themes.OutputLevelInfo)
				return nil
			},
		},
		{
			Key: "e", Label: "edit",
			Callback: func() error {
				form, err := views.NewFormView(
					container.RenderState(),
					&views.FormField{
						Key:   "value",
						Type:  views.PromptTypeMasked,
						Title: "Enter the new secret value",
					})
				if err != nil {
					container.HandleError(fmt.Errorf("encountered error creating the form: %w", err))
					return nil
				}
				if err := ctx.SetView(form); err != nil {
					container.HandleError(fmt.Errorf("unable to set view: %w", err))
					return nil
				}
				newValue := form.FindByKey("value").Value()
				secretValue := vault2.NewSecretValue([]byte(newValue))
				if err := vlt.SetSecret(ref.Key(), secretValue); err != nil {
					container.HandleError(fmt.Errorf("unable to edit secret: %w", err))
					return nil
				}
				loadSecretList()
				container.SetNotice("secret value updated", themes.OutputLevelInfo)
				return nil
			},
		},
		{
			Key: "x", Label: "delete",
			Callback: func() error {
				if err := vlt.DeleteSecret(ref.Key()); err != nil {
					container.HandleError(fmt.Errorf("unable to delete secret: %w", err))
					return nil
				}
				loadSecretList()
				container.SetNotice("secret deleted", themes.OutputLevelInfo)
				return nil
			},
		},
	}

	return views.NewEntityView(container.RenderState(), secret, types.EntityFormatDocument, secretKeyCallbacks...)
}

func NewSecretListView(
	ctx *context.Context,
	vlt vault2.Vault,
	asPlainText bool,
) tuikit.View {
	container := ctx.TUIContainer

	keys, err := vlt.ListSecrets()
	if err != nil {
		container.HandleError(fmt.Errorf("failed to list secrets: %w", err))
		return nil
	}
	secrets := make(vault2.SecretList, 0, len(keys))
	for _, key := range keys {
		s, err := vlt.GetSecret(key)
		if err != nil {
			container.HandleError(fmt.Errorf("failed to get secret %s: %w", key, err))
			continue
		}
		secret, err := vault2.NewSecret(vlt.ID(), key, s)
		if err != nil {
			container.HandleError(fmt.Errorf("failed to create secret object for %s: %w", key, err))
			continue
		}
		if asPlainText {
			secret = secret.AsPlaintext()
		} else {
			secret = secret.AsObfuscatedText()
		}
		secrets = append(secrets, secret)
	}

	if len(secrets.Items()) == 0 {
		container.HandleError(fmt.Errorf("no secrets found"))
	}

	selectFunc := func(filterVal string) error {
		var secret vault2.Secret
		var found bool
		for _, s := range secrets {
			if s == nil {
				continue
			}
			if string(s.Ref()) == filterVal {
				secret = s
				found = true
				break
			}
		}
		if !found || secret == nil {
			return fmt.Errorf("secret not found")
		}

		return container.SetView(NewSecretView(ctx, vlt, secret.Ref(), asPlainText))
	}

	return views.NewCollectionView(container.RenderState(), secrets, types.CollectionFormatList, selectFunc)
}
