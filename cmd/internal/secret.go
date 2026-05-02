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

	errhandler "github.com/flowexec/flow/v2/cmd/internal/errors"
	"github.com/flowexec/flow/v2/cmd/internal/flags"
	"github.com/flowexec/flow/v2/cmd/internal/response"
	"github.com/flowexec/flow/v2/internal/io/secret"
	"github.com/flowexec/flow/v2/internal/utils"
	envUtils "github.com/flowexec/flow/v2/internal/utils/env"
	"github.com/flowexec/flow/v2/internal/vault"
	"github.com/flowexec/flow/v2/pkg/context"
	"github.com/flowexec/flow/v2/pkg/logger"
	"github.com/flowexec/flow/v2/types/config"
)

func RegisterSecretCmd(ctx *context.Context, rootCmd *cobra.Command) {
	secretCmd := &cobra.Command{
		Use:     "secret",
		Aliases: []string{"scrt", "secrets"},
		Short:   "Manage secrets stored in a vault.",
		Long:    secretLong,
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
	RegisterFlag(ctx, removeCmd, *flags.VaultNameFlag)
	RegisterFlag(ctx, removeCmd, *flags.OutputFormatFlag)
	RegisterFlag(ctx, removeCmd, *flags.YesFlag)
	secretCmd.AddCommand(removeCmd)
}

func removeSecretFunc(ctx *context.Context, cmd *cobra.Command, args []string) {
	reference := args[0]

	skipConfirm := flags.ValueFor[bool](cmd, *flags.YesFlag, false)
	if !skipConfirm {
		form, err := views.NewForm(
			logger.Theme(ctx.Config.Theme.String()),
			ctx.StdIn(),
			ctx.StdOut(),
			&views.FormField{
				Key:   "confirm",
				Type:  views.PromptTypeConfirm,
				Title: fmt.Sprintf("Are you sure you want to remove the secret '%s'?", reference),
			})
		if err != nil {
			errhandler.HandleFatal(ctx, cmd, err)
		}
		if err := form.Run(ctx); err != nil {
			errhandler.HandleFatal(ctx, cmd, err)
		}
		resp := form.FindByKey("confirm").Value()
		if truthy, _ := strconv.ParseBool(resp); !truthy {
			logger.Log().Warnf("Aborting")
			return
		}
	}

	_, v, err := vault.VaultFromName(effectiveVault(cmd, ctx.Config))
	if err != nil {
		errhandler.HandleFatal(ctx, cmd, err)
		return
	}
	defer v.Close()

	if err = v.DeleteSecret(reference); err != nil {
		errhandler.HandleFatal(ctx, cmd, err)
	}

	response.HandleSuccess(ctx, cmd, fmt.Sprintf("Secret '%s' deleted from vault", reference), map[string]any{
		"name": reference,
	})
}

func registerSetSecretCmd(ctx *context.Context, secretCmd *cobra.Command) {
	setCmd := &cobra.Command{
		Use:     "set NAME [VALUE]",
		Aliases: []string{"new", "create", "update"},
		Short:   "Set a secret in the current vault. If no value is provided, you will be prompted to enter one.",
		Example: secretSetExamples,
		Args:    cobra.MinimumNArgs(1),
		Run:     func(cmd *cobra.Command, args []string) { setSecretFunc(ctx, cmd, args) },
	}
	RegisterFlag(ctx, setCmd, *flags.VaultNameFlag)
	RegisterFlag(ctx, setCmd, *flags.SecretFromFile)
	RegisterFlag(ctx, setCmd, *flags.OutputFormatFlag)
	secretCmd.AddCommand(setCmd)
}

func setSecretFunc(ctx *context.Context, cmd *cobra.Command, args []string) {
	reference := args[0]
	filename := flags.ValueFor[string](cmd, *flags.SecretFromFile, false)

	var value string
	switch {
	case filename != "" && len(args) >= 2:
		errhandler.HandleFatal(ctx, cmd, errors.New("must specify either a filename OR a value as an argument"))
	case filename != "":
		wd, err := os.Getwd()
		if err != nil {
			errhandler.HandleFatal(ctx, cmd, err)
		}
		expanded := utils.ExpandPath(filename, wd, envUtils.EnvListToEnvMap(os.Environ()))
		data, err := os.ReadFile(expanded)
		if err != nil {
			errhandler.HandleFatal(ctx, cmd, err)
		}
		value = string(data)
	case len(args) == 1:
		form, err := views.NewForm(
			logger.Theme(ctx.Config.Theme.String()),
			ctx.StdIn(),
			ctx.StdOut(),
			&views.FormField{
				Key:   "value",
				Type:  views.PromptTypeMasked,
				Title: "Enter the secret value",
			})
		if err != nil {
			errhandler.HandleFatal(ctx, cmd, err)
		}
		if err := form.Run(ctx); err != nil {
			errhandler.HandleFatal(ctx, cmd, err)
		}
		value = form.FindByKey("value").Value()
	case len(args) == 2:
		value = args[1]
	default:
		logger.Log().Warn("merging multiple arguments into a single value", "count", len(args))
		value = strings.Join(args[1:], " ")
	}

	vaultName := effectiveVault(cmd, ctx.Config)
	_, v, err := vault.VaultFromName(vaultName)
	if err != nil {
		errhandler.HandleFatal(ctx, cmd, err)
		return
	}
	defer v.Close()

	if err = v.SetSecret(reference, vault.NewSecretValue([]byte(value))); err != nil {
		errhandler.HandleFatal(ctx, cmd, err)
	}

	response.HandleSuccess(ctx, cmd, fmt.Sprintf("Secret %s set in vault", reference), map[string]any{
		"name": reference,
	})
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
	RegisterFlag(ctx, listCmd, *flags.VaultNameFlag)
	RegisterFlag(ctx, listCmd, *flags.OutputSecretAsPlainTextFlag)
	RegisterFlag(ctx, listCmd, *flags.OutputFormatFlag)
	secretCmd.AddCommand(listCmd)
}

func listSecretFunc(ctx *context.Context, cmd *cobra.Command, _ []string) {
	asPlainText := flags.ValueFor[bool](cmd, *flags.OutputSecretAsPlainTextFlag, false)
	outputFormat := flags.ValueFor[string](cmd, *flags.OutputFormatFlag, false)

	name := effectiveVault(cmd, ctx.Config)
	interactiveUI := TUIEnabled(ctx, cmd)

	_, v, err := vault.VaultFromName(name)
	defer func() {
		// Don't close the vault prematurely if we're in interactive mode
		go func() {
			if interactiveUI {
				ctx.TUIContainer().WaitForExit()
			}
			_ = v.Close()
		}()
	}()

	if err != nil {
		errhandler.HandleFatal(ctx, cmd, err)
	}

	if interactiveUI {
		view := secret.NewSecretListView(ctx, v, asPlainText)
		SetView(ctx, cmd, view)
	} else {
		secret.PrintSecrets(ctx, name, v, outputFormat, asPlainText)
	}
}

func registerGetSecretCmd(ctx *context.Context, secretCmd *cobra.Command) {
	getCmd := &cobra.Command{
		Use:     "get REFERENCE",
		Aliases: []string{"show", "view"},
		Short:   "Get the value of a secret in the current vault.",
		Example: secretGetExamples,
		Args:    cobra.ExactArgs(1),
		Run:     func(cmd *cobra.Command, args []string) { getSecretFunc(ctx, cmd, args) },
	}
	RegisterFlag(ctx, getCmd, *flags.VaultNameFlag)
	RegisterFlag(ctx, getCmd, *flags.OutputSecretAsPlainTextFlag)
	RegisterFlag(ctx, getCmd, *flags.CopyFlag)
	RegisterFlag(ctx, getCmd, *flags.OutputFormatFlag)
	secretCmd.AddCommand(getCmd)
}

func getSecretFunc(ctx *context.Context, cmd *cobra.Command, args []string) {
	reference := args[0]
	asPlainText := flags.ValueFor[bool](cmd, *flags.OutputSecretAsPlainTextFlag, false)
	copyValue := flags.ValueFor[bool](cmd, *flags.CopyFlag, false)

	rVault, key, err := vault.RefToParts(vault.SecretRef(reference))
	if err != nil {
		errhandler.HandleFatal(ctx, cmd, err)
	}
	if rVault == "" {
		rVault = effectiveVault(cmd, ctx.Config)
	}
	_, v, err := vault.VaultFromName(rVault)
	if err != nil {
		errhandler.HandleFatal(ctx, cmd, err)
		return
	}
	defer v.Close()

	s, err := v.GetSecret(key)
	if err != nil {
		errhandler.HandleFatal(ctx, cmd, err)
	}

	outputFormat := flags.ValueFor[string](cmd, *flags.OutputFormatFlag, false)
	switch outputFormat {
	case "json", "yaml", "yml":
		value := s.String()
		if asPlainText {
			value = s.PlainTextString()
		}
		data := map[string]any{
			"name":  reference,
			"value": value,
		}
		if copyValue {
			if err := clipboard.WriteAll(s.PlainTextString()); err != nil {
				data["copyError"] = err.Error()
			} else {
				data["copied"] = true
			}
		}
		response.HandleSuccess(ctx, cmd, fmt.Sprintf("Secret '%s' retrieved", reference), data)
	default:
		if asPlainText {
			logger.Log().PlainTextInfo(s.PlainTextString())
		} else {
			logger.Log().PlainTextInfo(s.String())
		}
		if copyValue {
			if err := clipboard.WriteAll(s.PlainTextString()); err != nil {
				logger.Log().WrapError(err, "\nunable to copy secret value to clipboard")
			} else {
				logger.Log().PlainTextSuccess("\ncopied secret value to clipboard")
			}
		}
	}
}

const (
	secretLong = `Manage secrets stored in the active vault. Secrets are encrypted key-value pairs that
can be referenced inside flowfiles using the secret reference syntax (e.g. ${secret:MY_KEY}).

The active vault is used by default; pass --vault to target a different one. Use
'vault' subcommands to create and manage vaults.`

	//nolint:gosec // example strings, not real credentials
	secretSetExamples = `
  flow secret set MY_TOKEN              # prompted securely
  flow secret set MY_TOKEN s3cr3t       # inline value
  flow secret set MY_TOKEN --from-file ./token.txt
`

	//nolint:gosec // example strings, not real credentials
	secretGetExamples = `
  flow secret get MY_TOKEN
  flow secret get MY_TOKEN --as-plain-text
  flow secret get MY_TOKEN --copy
`
)

func effectiveVault(cmd *cobra.Command, cfg *config.Config) string {
	if v := flags.ValueFor[string](cmd, *flags.VaultNameFlag, false); v != "" {
		return v
	}
	if cfg.CurrentVault == nil {
		return ""
	}
	return *cfg.CurrentVault
}
