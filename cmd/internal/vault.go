package internal

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/flowexec/tuikit/views"
	extvault "github.com/flowexec/vault"
	"github.com/spf13/cobra"

	errhandler "github.com/flowexec/flow/v2/cmd/internal/errors"
	"github.com/flowexec/flow/v2/cmd/internal/flags"
	"github.com/flowexec/flow/v2/cmd/internal/response"
	vaultIO "github.com/flowexec/flow/v2/internal/io/vault"
	"github.com/flowexec/flow/v2/internal/utils"
	"github.com/flowexec/flow/v2/internal/vault"
	"github.com/flowexec/flow/v2/pkg/context"
	"github.com/flowexec/flow/v2/pkg/filesystem"
	"github.com/flowexec/flow/v2/pkg/logger"
)

func RegisterVaultCmd(ctx *context.Context, rootCmd *cobra.Command) {
	vaultCmd := &cobra.Command{
		Use:     "vault",
		Aliases: []string{"vlt", "vaults"},
		Short:   "Manage sensitive secret stores.",
		Long:    vaultLong,
		Args:    cobra.NoArgs,
	}
	registerCreateVaultCmd(ctx, vaultCmd)
	registerGetVaultCmd(ctx, vaultCmd)
	registerListVaultCmd(ctx, vaultCmd)
	registerSwitchVaultCmd(ctx, vaultCmd)
	registerRemoveVaultCmd(ctx, vaultCmd)
	registerEditVaultCmd(ctx, vaultCmd)
	// TODO: add command for testing vault connectivity
	rootCmd.AddCommand(vaultCmd)
}

func registerCreateVaultCmd(ctx *context.Context, vaultCmd *cobra.Command) {
	createCmd := &cobra.Command{
		Use:     "create NAME",
		Aliases: []string{"new", "add"},
		Short:   "Create a new vault.",
		Example: vaultCreateExamples,
		Args:    cobra.ExactArgs(1),
		PreRun: func(cmd *cobra.Command, args []string) {
			vaultName := args[0]
			if vaultName == vault.DemoVaultReservedName {
				errhandler.HandleUsage(ctx, cmd, "create is unsupported for the reserved vaults")
			} else if err := vault.ValidateIdentifier(vaultName); err != nil {
				errhandler.HandleUsage(ctx, cmd, "invalid vault name '%s': %v", vaultName, err)
			}

			if vault.VaultExists(vaultName) {
				errhandler.HandleUsage(ctx, cmd, "vault %s already exists", vaultName)
			}
		},
		Run: func(cmd *cobra.Command, args []string) { createVaultFunc(ctx, cmd, args) },
	}

	RegisterFlag(ctx, createCmd, *flags.VaultTypeFlag)
	RegisterFlag(ctx, createCmd, *flags.VaultPathFlag)
	RegisterFlag(ctx, createCmd, *flags.VaultSetFlag)
	RegisterFlag(ctx, createCmd, *flags.VaultFromFileFlag)
	RegisterFlag(ctx, createCmd, *flags.OutputFormatFlag)
	// AES flags
	RegisterFlag(ctx, createCmd, *flags.VaultKeyEnvFlag)
	RegisterFlag(ctx, createCmd, *flags.VaultKeyFileFlag)
	// Age flags
	RegisterFlag(ctx, createCmd, *flags.VaultRecipientsFlag)
	RegisterFlag(ctx, createCmd, *flags.VaultIdentityEnvFlag)
	RegisterFlag(ctx, createCmd, *flags.VaultIdentityFileFlag)

	vaultCmd.AddCommand(createCmd)
}

// vaultWorkspacePath returns the directory a workspace-relative vault path resolves against. A
// vault outlives the command that created it, so anchoring one in a directory flow only found by
// walking up from here is almost never what the user meant — warn before doing it.
func vaultWorkspacePath(ctx *context.Context, cmd *cobra.Command) string {
	if ctx.CurrentWorkspace == nil {
		return ""
	}
	path := ctx.CurrentWorkspace.Location()
	// Machine-readable output is parsed by the caller, so a loose warning line would corrupt it.
	outputFormat := flags.ValueFor[string](cmd, *flags.OutputFormatFlag, false)
	if !ctx.WorkspaceIsRegistered() && outputFormat == "" {
		logger.Log().Warnf(
			"workspace '%s' is not registered; run 'flow workspace add %s %s' to keep this vault reachable",
			ctx.CurrentWorkspaceName(), ctx.CurrentWorkspaceName(), path,
		)
	}
	return path
}

func createVaultFunc(ctx *context.Context, cmd *cobra.Command, args []string) {
	vaultName := args[0]
	vaultType := flags.ValueFor[string](cmd, *flags.VaultTypeFlag, false)
	vaultPath := flags.ValueFor[string](cmd, *flags.VaultPathFlag, false)
	setVault := flags.ValueFor[bool](cmd, *flags.VaultSetFlag, false)

	var result *vault.CreateResult
	var err error

	switch strings.ToLower(vaultType) {
	case "unencrypted":
		result, err = vault.NewUnencryptedVault(vaultName, vaultPath)
	case "aes256":
		keyEnv := flags.ValueFor[string](cmd, *flags.VaultKeyEnvFlag, false)
		keyFile := flags.ValueFor[string](cmd, *flags.VaultKeyFileFlag, false)
		result, err = vault.NewAES256Vault(vaultName, vaultPath, keyEnv, keyFile)
	case "age":
		recipients := flags.ValueFor[string](cmd, *flags.VaultRecipientsFlag, false)
		identityEnv := flags.ValueFor[string](cmd, *flags.VaultIdentityEnvFlag, false)
		identityFile := flags.ValueFor[string](cmd, *flags.VaultIdentityFileFlag, false)
		result, err = vault.NewAgeVault(vaultName, vaultPath, recipients, identityEnv, identityFile)
	case "keyring":
		result, err = vault.NewKeyringVault(vaultName)
	case "external":
		cfgFile := flags.ValueFor[string](cmd, *flags.VaultFromFileFlag, false)
		if cfgFile == "" {
			errhandler.HandleUsage(ctx, cmd, "external vault requires a configuration file to be specified with --config")
		}
		result, err = vault.NewExternalVault(cfgFile)
	default:
		errhandler.HandleUsage(
			ctx, cmd,
			"unsupported vault type: %s - must be one of 'aes256', 'age', 'unencrypted', 'keyring', or 'external'",
			vaultType,
		)
	}
	if err != nil {
		errhandler.HandleFatal(ctx, cmd, err)
	}

	if ctx.Config.Vaults == nil {
		ctx.Config.Vaults = make(map[string]string)
	}

	vaultPath = utils.ExpandDirectory(
		vaultPath, vaultWorkspacePath(ctx, cmd), vault.CacheDirectory(vaultName), nil,
	)

	ctx.Config.Vaults[vaultName] = vaultPath
	if setVault {
		ctx.Config.CurrentVault = &vaultName
		logger.Log().Infof("Vault '%s' set as current vault", vaultName)
	}
	if err := filesystem.WriteConfig(ctx.Config); err != nil {
		errhandler.HandleFatal(ctx, cmd, fmt.Errorf("unable to save user configuration: %w", err))
	}

	outputFormat := flags.ValueFor[string](cmd, *flags.OutputFormatFlag, false)
	logLevel := flags.ValueFor[string](cmd, *flags.LogLevel, false)

	// In plain-text mode, print the generated key so it can be captured by
	// scripts or seen by the user before it scrolls away.
	if result.GeneratedKey != "" && outputFormat == "" {
		if logLevel == "fatal" {
			// Bare key output for scripting (e.g. key=$(flow vault create ...))
			logger.Log().Print(result.GeneratedKey)
		} else {
			keyEnv := flags.ValueFor[string](cmd, *flags.VaultKeyEnvFlag, false)
			if keyEnv == "" {
				keyEnv = vault.DefaultVaultKeyEnv
			}
			logger.Log().Println(fmt.Sprintf("Your vault encryption key is: %s", result.GeneratedKey))
			logger.Log().PlainTextInfo(fmt.Sprintf(
				"You will need this key to modify your vault data. Store it somewhere safe!\n"+
					"Set this value to the %s environment variable to access the vault in the future.\n",
				keyEnv,
			))
		}
	}

	data := map[string]any{
		"name": result.Name,
		"type": result.Type,
	}
	if result.GeneratedKey != "" {
		data["generatedKey"] = result.GeneratedKey
	}
	response.HandleSuccess(ctx, cmd, fmt.Sprintf("Vault '%s' created", result.Name), data)
}

func registerGetVaultCmd(ctx *context.Context, vaultCmd *cobra.Command) {
	getCmd := &cobra.Command{
		Use:     "get NAME",
		Aliases: []string{"view", "show"},
		Short:   "Get the details of a vault.",
		Args:    cobra.MaximumNArgs(1),
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			return vaultNames(), cobra.ShellCompDirectiveNoFileComp
		},
		PreRun: func(cmd *cobra.Command, args []string) {
			var vaultName string
			if len(args) == 0 {
				vaultName = ctx.Config.CurrentVaultName()
			} else {
				vaultName = args[0]
			}

			if err := vault.ValidateIdentifier(vaultName); err != nil {
				errhandler.HandleUsage(ctx, cmd, "invalid vault name '%s': %v", vaultName, err)
			}

			StartTUI(ctx, cmd)
		},
		PostRun: func(cmd *cobra.Command, args []string) { WaitForTUI(ctx, cmd) },
		Run:     func(cmd *cobra.Command, args []string) { getVaultFunc(ctx, cmd, args) },
	}
	RegisterFlag(ctx, getCmd, *flags.OutputFormatFlag)
	vaultCmd.AddCommand(getCmd)
}

func getVaultFunc(ctx *context.Context, cmd *cobra.Command, args []string) {
	outputFormat := flags.ValueFor[string](cmd, *flags.OutputFormatFlag, false)

	var vaultName string
	if len(args) == 0 {
		vaultName = ctx.Config.CurrentVaultName()
	} else {
		vaultName = args[0]
	}

	if TUIEnabled(ctx, cmd) {
		view := vaultIO.NewVaultView(ctx.TUIContainer(), vaultName)
		SetView(ctx, cmd, view)
	} else {
		vaultIO.PrintVault(outputFormat, vaultName)
	}
}

func registerListVaultCmd(ctx *context.Context, vaultCmd *cobra.Command) {
	listCmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List all available vaults.",
		Args:    cobra.NoArgs,
		PreRun: func(cmd *cobra.Command, args []string) {
			StartTUI(ctx, cmd)
		},
		PostRun: func(cmd *cobra.Command, args []string) { WaitForTUI(ctx, cmd) },
		Run:     func(cmd *cobra.Command, args []string) { listVaultsFunc(ctx, cmd, args) },
	}
	RegisterFlag(ctx, listCmd, *flags.OutputFormatFlag)
	vaultCmd.AddCommand(listCmd)
}

func listVaultsFunc(ctx *context.Context, cmd *cobra.Command, _ []string) {
	outputFormat := flags.ValueFor[string](cmd, *flags.OutputFormatFlag, false)

	names, err := vault.ListVaultNames()
	if err != nil {
		errhandler.HandleFatal(ctx, cmd, err)
	}
	if TUIEnabled(ctx, cmd) {
		view := vaultIO.NewVaultListView(ctx.TUIContainer(), names)
		SetView(ctx, cmd, view)
	} else {
		vaultIO.PrintVaultList(outputFormat, names)
	}
}

func registerRemoveVaultCmd(ctx *context.Context, vaultCmd *cobra.Command) {
	removeCmd := &cobra.Command{
		Use:     "remove NAME",
		Aliases: []string{"rm", "delete"},
		Short:   "Remove an existing vault.",
		Long: "Remove an existing vault by its name. The vault's encrypted secret data remains on disk " +
			"at its storage path, but its configuration is deleted so flow no longer tracks it.\n" +
			"Note: You cannot remove the current vault.",
		Args: cobra.ExactArgs(1),
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			return vaultNames(), cobra.ShellCompDirectiveNoFileComp
		},
		Run: func(cmd *cobra.Command, args []string) { removeVaultFunc(ctx, cmd, args) },
	}
	RegisterFlag(ctx, removeCmd, *flags.OutputFormatFlag)
	RegisterFlag(ctx, removeCmd, *flags.YesFlag)
	vaultCmd.AddCommand(removeCmd)
}

func removeVaultFunc(ctx *context.Context, cmd *cobra.Command, args []string) {
	vaultName := args[0]

	if vaultName == vault.DemoVaultReservedName {
		errhandler.HandleUsage(ctx, cmd, "remove is unsupported for the reserved vault")
	}

	skipConfirm := flags.ValueFor[bool](cmd, *flags.YesFlag, false)
	if !skipConfirm {
		form, err := views.NewForm(
			logger.Theme(ctx.Config.Theme.String()),
			ctx.StdIn(),
			ctx.StdOut(),
			&views.FormField{
				Key:   "confirm",
				Type:  views.PromptTypeConfirm,
				Title: fmt.Sprintf("Are you sure you want to remove the vault '%s'?", vaultName),
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

	userConfig := ctx.Config
	if userConfig.CurrentVault != nil && vaultName == *userConfig.CurrentVault {
		errhandler.HandleUsage(ctx, cmd, "cannot remove the current vault")
	}
	if !vault.VaultExists(vaultName) {
		errhandler.HandleFatal(ctx, cmd, fmt.Errorf("vault %s was not found", vaultName))
	}

	// Delete the vault's config file (the source of truth). Encrypted secret data at the
	// vault's storage path is intentionally preserved.
	if err := vault.RemoveVaultConfig(vaultName); err != nil {
		errhandler.HandleFatal(ctx, cmd, err)
	}
	// Also drop the legacy config-map entry if present, keeping the two in sync.
	delete(userConfig.Vaults, vaultName)
	if err := filesystem.WriteConfig(userConfig); err != nil {
		errhandler.HandleFatal(ctx, cmd, err)
	}

	response.HandleSuccess(ctx, cmd, fmt.Sprintf("Vault '%s' deleted", vaultName), map[string]any{
		"name": vaultName,
	})
}

func registerSwitchVaultCmd(ctx *context.Context, vaultCmd *cobra.Command) {
	switchCmd := &cobra.Command{
		Use:     "switch NAME",
		Aliases: []string{"use", "set"},
		Short:   "Switch the active vault.",
		Args:    cobra.ExactArgs(1),
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			return vaultNames(), cobra.ShellCompDirectiveNoFileComp
		},
		PreRun: func(cmd *cobra.Command, args []string) {
			vaultName := args[0]
			if vaultName == vault.DemoVaultReservedName {
				return
			}
			if !vault.VaultExists(vaultName) {
				errhandler.HandleFatal(ctx, cmd, fmt.Errorf("vault %s not found", vaultName))
			}
		},
		Run: func(cmd *cobra.Command, args []string) { switchVaultFunc(ctx, cmd, args) },
	}
	RegisterFlag(ctx, switchCmd, *flags.OutputFormatFlag)
	vaultCmd.AddCommand(switchCmd)
}

func switchVaultFunc(ctx *context.Context, cmd *cobra.Command, args []string) {
	vaultName := args[0]
	userConfig := ctx.Config
	userConfig.CurrentVault = &vaultName

	if err := filesystem.WriteConfig(userConfig); err != nil {
		errhandler.HandleFatal(ctx, cmd, err)
	}
	response.HandleSuccess(ctx, cmd, "Vault set to "+vaultName, map[string]any{
		"name": vaultName,
	})
}

func registerEditVaultCmd(ctx *context.Context, vaultCmd *cobra.Command) {
	editCmd := &cobra.Command{
		Use:     "edit NAME",
		Aliases: []string{"update", "modify"},
		Short:   "Edit the configuration of an existing vault.",
		Long: "Edit the configuration of an existing vault. " +
			"Note: You cannot change the vault type after creation.",
		Args: cobra.ExactArgs(1),
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			return vaultNames(), cobra.ShellCompDirectiveNoFileComp
		},
		PreRun: func(cmd *cobra.Command, args []string) {
			vaultName := args[0]
			if vaultName == vault.DemoVaultReservedName {
				errhandler.HandleUsage(ctx, cmd, "edit is unsupported for the reserved vault")
			} else if err := vault.ValidateIdentifier(vaultName); err != nil {
				errhandler.HandleUsage(ctx, cmd, "invalid vault name '%s': %v", vaultName, err)
			}

			if !vault.VaultExists(vaultName) {
				errhandler.HandleFatal(ctx, cmd, fmt.Errorf("vault %s not found", vaultName))
			}
		},
		Run: func(cmd *cobra.Command, args []string) { editVaultFunc(ctx, cmd, args) },
	}

	RegisterFlag(ctx, editCmd, *flags.VaultPathFlag)
	RegisterFlag(ctx, editCmd, *flags.OutputFormatFlag)
	// AES flags
	RegisterFlag(ctx, editCmd, *flags.VaultKeyEnvFlag)
	RegisterFlag(ctx, editCmd, *flags.VaultKeyFileFlag)
	// Age flags
	RegisterFlag(ctx, editCmd, *flags.VaultRecipientsFlag)
	RegisterFlag(ctx, editCmd, *flags.VaultIdentityEnvFlag)
	RegisterFlag(ctx, editCmd, *flags.VaultIdentityFileFlag)

	vaultCmd.AddCommand(editCmd)
}

func editVaultFunc(ctx *context.Context, cmd *cobra.Command, args []string) {
	vaultName := args[0]

	vaultPath := flags.ValueFor[string](cmd, *flags.VaultPathFlag, false)
	keyEnv := flags.ValueFor[string](cmd, *flags.VaultKeyEnvFlag, false)
	keyFile := flags.ValueFor[string](cmd, *flags.VaultKeyFileFlag, false)
	recipients := flags.ValueFor[string](cmd, *flags.VaultRecipientsFlag, false)
	identityEnv := flags.ValueFor[string](cmd, *flags.VaultIdentityEnvFlag, false)
	identityFile := flags.ValueFor[string](cmd, *flags.VaultIdentityFileFlag, false)

	cfgPath := vault.ConfigFilePath(vaultName)
	existingCfg, err := extvault.LoadConfigJSON(cfgPath)
	if err != nil {
		errhandler.HandleFatal(ctx, cmd, fmt.Errorf("failed to load vault configuration: %w", err))
	}

	// TODO: add support for appending KeySources and IdentitySources instead of overwriting them
	switch existingCfg.Type {
	case extvault.ProviderTypeAES256:
		if vaultPath != "" {
			existingCfg.Aes.StoragePath = vaultPath
		}
		if keyEnv != "" {
			existingCfg.Aes.KeySource = []extvault.KeySource{{
				Type: "env",
				Name: keyEnv,
			}}
		}
		if keyFile != "" {
			existingCfg.Aes.KeySource = []extvault.KeySource{{
				Type: "file",
				Path: keyFile,
			}}
		}
	case extvault.ProviderTypeAge:
		if vaultPath != "" {
			existingCfg.Age.StoragePath = vaultPath
		}
		if recipients != "" {
			existingCfg.Age.Recipients = strings.Split(recipients, ",")
		}
		if identityEnv != "" {
			existingCfg.Age.IdentitySources = []extvault.IdentitySource{{
				Type: "env",
				Name: identityEnv,
			}}
		}
		if identityFile != "" {
			existingCfg.Age.IdentitySources = []extvault.IdentitySource{{
				Type: "file",
				Path: identityFile,
			}}
		}
	default:
		errhandler.HandleUsage(ctx, cmd, "unsupported vault type: %s", existingCfg.Type)
	}

	if err = extvault.SaveConfigJSON(existingCfg, cfgPath); err != nil {
		errhandler.HandleFatal(ctx, cmd, fmt.Errorf("failed to save vault configuration: %w", err))
	}

	response.HandleSuccess(
		ctx,
		cmd,
		fmt.Sprintf("Vault '%s' configuration updated successfully", vaultName),
		map[string]any{
			"name": vaultName,
		},
	)
}

func vaultNames() []string {
	names := []string{vault.DemoVaultReservedName}
	discovered, err := vault.ListVaultNames()
	if err != nil {
		return names
	}
	return append(names, discovered...)
}

const (
	vaultLong = `Manage secret stores (vaults). A vault is an encrypted key-value store that holds secrets
referenced by your executables. Multiple vault types are supported (e.g. age encryption,
AES-256, system keyring, or environment-variable passthrough).

One vault is active at a time; use 'vault switch' to change the active vault. Secrets
within a vault are managed with the 'secret' subcommands.`

	vaultCreateExamples = `
  flow vault create myvault --type age --set
  flow vault create myvault --type aes256 --key-env VAULT_KEY
  flow vault create myvault --type keyring
`
)
