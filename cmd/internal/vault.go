package internal

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/flowexec/tuikit/views"
	extvault "github.com/flowexec/vault"
	"github.com/spf13/cobra"
	"golang.org/x/exp/maps"

	errhandler "github.com/flowexec/flow/cmd/internal/errors"
	"github.com/flowexec/flow/cmd/internal/flags"
	"github.com/flowexec/flow/cmd/internal/response"
	vaultIO "github.com/flowexec/flow/internal/io/vault"
	"github.com/flowexec/flow/internal/utils"
	"github.com/flowexec/flow/internal/vault"
	"github.com/flowexec/flow/pkg/context"
	"github.com/flowexec/flow/pkg/filesystem"
	"github.com/flowexec/flow/pkg/logger"
	"github.com/flowexec/flow/types/config"
)

func RegisterVaultCmd(ctx *context.Context, rootCmd *cobra.Command) {
	vaultCmd := &cobra.Command{
		Use:     "vault",
		Aliases: []string{"vlt", "vaults"},
		Short:   "Manage sensitive secret stores.",
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
		Args:    cobra.ExactArgs(1),
		PreRun: func(cmd *cobra.Command, args []string) {
			vaultName := args[0]
			if vaultName == vault.DemoVaultReservedName {
				errhandler.HandleUsage(ctx, cmd, "create is unsupported for the reserved vaults")
			} else if err := vault.ValidateIdentifier(vaultName); err != nil {
				errhandler.HandleUsage(ctx, cmd, "invalid vault name '%s': %v", vaultName, err)
			}

			if _, found := ctx.Config.Vaults[vaultName]; found {
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

	curWs := ctx.Config.CurrentWorkspace
	vaultPath = utils.ExpandDirectory(
		vaultPath, ctx.Config.Workspaces[curWs], vault.CacheDirectory(vaultName), nil,
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
			return vaultNames(ctx.Config), cobra.ShellCompDirectiveNoFileComp
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

	cfg := ctx.Config
	if TUIEnabled(ctx, cmd) {
		view := vaultIO.NewVaultListView(ctx.TUIContainer(), maps.Keys(cfg.Vaults))
		SetView(ctx, cmd, view)
	} else {
		vaultIO.PrintVaultList(outputFormat, maps.Keys(cfg.Vaults))
	}
}

func registerRemoveVaultCmd(ctx *context.Context, vaultCmd *cobra.Command) {
	removeCmd := &cobra.Command{
		Use:     "remove NAME",
		Aliases: []string{"rm", "delete"},
		Short:   "Remove an existing vault.",
		Long: "Remove an existing vault by its name. The vault data will remain in it's original location, " +
			"but the vault will be unlinked from the global configuration.\nNote: You cannot remove the current vault.",
		Args: cobra.ExactArgs(1),
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			return vaultNames(ctx.Config), cobra.ShellCompDirectiveNoFileComp
		},
		PreRun: func(cmd *cobra.Command, args []string) { validateVaults(ctx, cmd) },
		Run:    func(cmd *cobra.Command, args []string) { removeVaultFunc(ctx, cmd, args) },
	}
	RegisterFlag(ctx, removeCmd, *flags.OutputFormatFlag)
	vaultCmd.AddCommand(removeCmd)
}

func removeVaultFunc(ctx *context.Context, cmd *cobra.Command, args []string) {
	vaultName := args[0]

	if vaultName == vault.DemoVaultReservedName {
		errhandler.HandleUsage(ctx, cmd, "remove is unsupported for the reserved vault")
	}

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

	userConfig := ctx.Config
	if userConfig.CurrentVault != nil && vaultName == *userConfig.CurrentVault {
		errhandler.HandleUsage(ctx, cmd, "cannot remove the current vault")
	}
	if _, found := userConfig.Vaults[vaultName]; !found {
		errhandler.HandleFatal(ctx, cmd, fmt.Errorf("vault %s was not found", vaultName))
	}

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
			return vaultNames(ctx.Config), cobra.ShellCompDirectiveNoFileComp
		},
		PreRun: func(cmd *cobra.Command, args []string) {
			vaultName := args[0]
			reservedName := vaultName == vault.DemoVaultReservedName
			if reservedName {
				return
			}
			validateVaults(ctx, cmd)
			if _, found := ctx.Config.Vaults[vaultName]; !found {
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
			return vaultNames(ctx.Config), cobra.ShellCompDirectiveNoFileComp
		},
		PreRun: func(cmd *cobra.Command, args []string) {
			validateVaults(ctx, cmd)
			vaultName := args[0]
			if vaultName == vault.DemoVaultReservedName {
				errhandler.HandleUsage(ctx, cmd, "edit is unsupported for the reserved vault")
			} else if err := vault.ValidateIdentifier(vaultName); err != nil {
				errhandler.HandleUsage(ctx, cmd, "invalid vault name '%s': %v", vaultName, err)
			}

			userConfig := ctx.Config
			if _, found := userConfig.Vaults[vaultName]; !found {
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

func vaultNames(cfg *config.Config) []string {
	names := []string{vault.DemoVaultReservedName}
	if cfg == nil || cfg.Vaults == nil {
		return nil
	}
	for name := range cfg.Vaults {
		names = append(names, name)
	}
	return names
}

func validateVaults(ctx *context.Context, cmd *cobra.Command) {
	if ctx.Config == nil || ctx.Config.Vaults == nil {
		errhandler.HandleUsage(ctx, cmd, "no vaults configured")
	}
}
