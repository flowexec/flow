package internal

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/atotto/clipboard"
	"github.com/flowexec/tuikit/views"
	"github.com/spf13/cobra"

	"github.com/flowexec/flow/cmd/internal/flags"
	"github.com/flowexec/flow/internal/context"
	"github.com/flowexec/flow/internal/io"
	"github.com/flowexec/flow/internal/io/secret"
	secretV2 "github.com/flowexec/flow/internal/io/secret/v2"
	"github.com/flowexec/flow/internal/logger"
	"github.com/flowexec/flow/internal/utils"
	envUtils "github.com/flowexec/flow/internal/utils/env"
	"github.com/flowexec/flow/internal/vault"
	vaultV2 "github.com/flowexec/flow/internal/vault/v2"
	"github.com/flowexec/flow/types/config"
)

func RegisterSecretCmd(ctx *context.Context, rootCmd *cobra.Command) {
	secretCmd := &cobra.Command{
		Use:     "secret",
		Aliases: []string{"scrt", "secrets"},
		Short:   "Manage secrets stored in a vault.",
	}
	registerSetSecretCmd(ctx, secretCmd)
	registerListSecretCmd(ctx, secretCmd)
	registerGetSecretCmd(ctx, secretCmd)
	registerRemoveSecretCmd(ctx, secretCmd)
	rootCmd.AddCommand(secretCmd)
}

func registerRemoveSecretCmd(ctx *context.Context, secretCmd *cobra.Command) {
	removeCmd := &cobra.Command{
		Use:     "remove NAME",
		Aliases: []string{"delete", "rm"},
		Short:   "Remove a secret from the vault.",
		Args:    cobra.ExactArgs(1),
		Run:     func(cmd *cobra.Command, args []string) { removeSecretFunc(ctx, cmd, args) },
	}
	secretCmd.AddCommand(removeCmd)
}

func removeSecretFunc(ctx *context.Context, _ *cobra.Command, args []string) {
	reference := args[0]

	form, err := views.NewForm(
		io.Theme(ctx.Config.Theme.String()),
		ctx.StdIn(),
		ctx.StdOut(),
		&views.FormField{
			Key:   "confirm",
			Type:  views.PromptTypeConfirm,
			Title: fmt.Sprintf("Are you sure you want to remove the secret '%s'?", reference),
		})
	if err != nil {
		logger.Log().FatalErr(err)
	}
	if err := form.Run(ctx); err != nil {
		logger.Log().FatalErr(err)
	}
	resp := form.FindByKey("confirm").Value()
	if truthy, _ := strconv.ParseBool(resp); !truthy {
		logger.Log().Warnf("Aborting")
		return
	}

	if currentVault(ctx.Config) == vaultV2.LegacyVaultReservedName {
		logger.Log().Warnf("Using deprecated vault. Consider creating a new vault with 'flow vault create' command.")
		v := vault.NewVault()
		if err = v.DeleteSecret(reference); err != nil {
			logger.Log().FatalErr(err)
		}
	} else {
		_, v, err := vaultV2.VaultFromName(currentVault(ctx.Config))
		defer v.Close()

		if err != nil {
			logger.Log().FatalErr(err)
		}
		if err = v.DeleteSecret(reference); err != nil {
			logger.Log().FatalErr(err)
		}
	}

	logger.Log().PlainTextSuccess(fmt.Sprintf("Secret '%s' deleted from vault", reference))
}

func registerSetSecretCmd(ctx *context.Context, secretCmd *cobra.Command) {
	setCmd := &cobra.Command{
		Use:     "set NAME [VALUE]",
		Aliases: []string{"new", "create", "update"},
		Short:   "Set a secret in the current vault. If no value is provided, you will be prompted to enter one.",
		Args:    cobra.MinimumNArgs(1),
		Run:     func(cmd *cobra.Command, args []string) { setSecretFunc(ctx, cmd, args) },
	}
	RegisterFlag(ctx, setCmd, *flags.SecretFromFile)
	secretCmd.AddCommand(setCmd)
}

func setSecretFunc(ctx *context.Context, cmd *cobra.Command, args []string) {
	reference := args[0]
	filename := flags.ValueFor[string](cmd, *flags.SecretFromFile, false)

	var value string
	switch {
	case filename != "" && len(args) >= 2:
		logger.Log().FatalErr(errors.New("must specify either a filename OR a value as an argument"))
	case filename != "":
		wd, err := os.Getwd()
		if err != nil {
			logger.Log().FatalErr(err)
		}
		expanded := utils.ExpandPath(filename, wd, envUtils.EnvListToEnvMap(os.Environ()))
		data, err := os.ReadFile(expanded)
		if err != nil {
			logger.Log().FatalErr(err)
		}
		value = string(data)
	case len(args) == 1:
		form, err := views.NewForm(
			io.Theme(ctx.Config.Theme.String()),
			ctx.StdIn(),
			ctx.StdOut(),
			&views.FormField{
				Key:   "value",
				Type:  views.PromptTypeMasked,
				Title: "Enter the secret value",
			})
		if err != nil {
			logger.Log().FatalErr(err)
		}
		if err := form.Run(ctx); err != nil {
			logger.Log().FatalErr(err)
		}
		value = form.FindByKey("value").Value()
	case len(args) == 2:
		value = args[1]
	default:
		logger.Log().Warnx("merging multiple arguments into a single value", "count", len(args))
		value = strings.Join(args[1:], " ")
	}

	sv := vault.SecretValue(value)
	vaultName := currentVault(ctx.Config)
	if vaultName == vaultV2.LegacyVaultReservedName {
		logger.Log().Warnf(
			"Using deprecated vault '%s'. Consider creating a new vault with 'flow vault create' command.",
			vaultName,
		)
		v := vault.NewVault()
		err := v.SetSecret(reference, sv)
		if err != nil {
			logger.Log().FatalErr(err)
		}
	} else {
		_, v, err := vaultV2.VaultFromName(vaultName)
		defer v.Close()

		if err != nil {
			logger.Log().FatalErr(err)
		}
		if err = v.SetSecret(reference, vaultV2.NewSecretValue([]byte(value))); err != nil {
			logger.Log().FatalErr(err)
		}
	}

	logger.Log().PlainTextSuccess(fmt.Sprintf("Secret %s set in vault", reference))
}

func registerListSecretCmd(ctx *context.Context, secretCmd *cobra.Command) {
	listCmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List secrets stored in the current vault.",
		Args:    cobra.NoArgs,
		PreRun:  func(cmd *cobra.Command, args []string) { StartTUI(ctx, cmd) },
		PostRun: func(cmd *cobra.Command, args []string) { WaitForTUI(ctx, cmd) },
		Run:     func(cmd *cobra.Command, args []string) { listSecretFunc(ctx, cmd, args) },
	}
	RegisterFlag(ctx, listCmd, *flags.OutputSecretAsPlainTextFlag)
	RegisterFlag(ctx, listCmd, *flags.OutputFormatFlag)
	secretCmd.AddCommand(listCmd)
}

func listSecretFunc(ctx *context.Context, cmd *cobra.Command, _ []string) {
	asPlainText := flags.ValueFor[bool](cmd, *flags.OutputSecretAsPlainTextFlag, false)
	outputFormat := flags.ValueFor[string](cmd, *flags.OutputFormatFlag, false)

	//nolint:nestif
	if currentVault(ctx.Config) == vaultV2.LegacyVaultReservedName {
		v := vault.NewVault()
		secrets, err := v.GetAllSecrets()
		if err != nil {
			logger.Log().FatalErr(err)
		}

		interactiveUI := TUIEnabled(ctx, cmd)
		if interactiveUI {
			secret.LoadSecretListView(ctx, asPlainText)
		} else {
			secret.PrintSecrets(ctx, secrets, outputFormat, asPlainText)
		}
	} else {
		name := currentVault(ctx.Config)
		interactiveUI := TUIEnabled(ctx, cmd)

		_, v, err := vaultV2.VaultFromName(name)
		defer func() {
			// Don't close the vault prematurely if we're in interactive mode
			go func() {
				if interactiveUI {
					ctx.TUIContainer.WaitForExit()
				}
				_ = v.Close()
			}()
		}()

		if err != nil {
			logger.Log().FatalErr(err)
		}

		if interactiveUI {
			view := secretV2.NewSecretListView(ctx, v, asPlainText)
			SetView(ctx, cmd, view)
		} else {
			secretV2.PrintSecrets(ctx, name, v, outputFormat, asPlainText)
		}
	}
}

func registerGetSecretCmd(ctx *context.Context, secretCmd *cobra.Command) {
	getCmd := &cobra.Command{
		Use:     "get REFERENCE",
		Aliases: []string{"show", "view"},
		Short:   "Get the value of a secret in the current vault.",
		Args:    cobra.ExactArgs(1),
		Run:     func(cmd *cobra.Command, args []string) { getSecretFunc(ctx, cmd, args) },
	}
	RegisterFlag(ctx, getCmd, *flags.OutputSecretAsPlainTextFlag)
	RegisterFlag(ctx, getCmd, *flags.CopyFlag)
	secretCmd.AddCommand(getCmd)
}

func getSecretFunc(ctx *context.Context, cmd *cobra.Command, args []string) {
	reference := args[0]
	asPlainText := flags.ValueFor[bool](cmd, *flags.OutputSecretAsPlainTextFlag, false)
	copyValue := flags.ValueFor[bool](cmd, *flags.CopyFlag, false)

	//nolint:nestif
	if currentVault(ctx.Config) == vaultV2.LegacyVaultReservedName {
		logger.Log().Warnf("Using deprecated vault. Consider creating a new vault with 'flow vault create' command.")
		v := vault.NewVault()
		s, err := v.GetSecret(reference)
		if err != nil {
			logger.Log().FatalErr(err)
		}

		if asPlainText {
			logger.Log().PlainTextInfo(s.PlainTextString())
		} else {
			logger.Log().PlainTextInfo(s.String())
		}

		if copyValue {
			if err := clipboard.WriteAll(s.PlainTextString()); err != nil {
				logger.Log().Error(err, "\nunable to copy secret value to clipboard")
			} else {
				logger.Log().PlainTextSuccess("\ncopied secret value to clipboard")
			}
		}
	} else {
		rVault, key, err := vaultV2.RefToParts(vaultV2.SecretRef(reference))
		if err != nil {
			logger.Log().FatalErr(err)
		}
		if rVault == "" {
			rVault = currentVault(ctx.Config)
		}
		_, v, err := vaultV2.VaultFromName(rVault)
		defer v.Close()

		if err != nil {
			logger.Log().FatalErr(err)
		}
		s, err := v.GetSecret(key)
		if err != nil {
			logger.Log().FatalErr(err)
		}

		if asPlainText {
			logger.Log().PlainTextInfo(s.PlainTextString())
		} else {
			logger.Log().PlainTextInfo(s.String())
		}
		if copyValue {
			if err := clipboard.WriteAll(s.PlainTextString()); err != nil {
				logger.Log().Error(err, "\nunable to copy secret value to clipboard")
			} else {
				logger.Log().PlainTextSuccess("\ncopied secret value to clipboard")
			}
		}
	}
}

func currentVault(cfg *config.Config) string {
	if cfg.CurrentVault == nil || *cfg.CurrentVault == "" {
		return vaultV2.LegacyVaultReservedName
	}
	return *cfg.CurrentVault
}
