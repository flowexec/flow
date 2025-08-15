package vault

import (
	"github.com/flowexec/flow/internal/io/common"
	"github.com/flowexec/flow/internal/logger"
)

func PrintVault(format, vaultName string) {
	if vaultName == "" {
		logger.Log().Fatalf("Vault name was not provided")
	}

	vault, err := vaultFromName(vaultName)
	if err != nil {
		logger.Log().Fatalf("Failed to get vault %s - %v", vaultName, err)
	}

	switch common.NormalizeFormat(format) {
	case common.YAMLFormat:
		str, err := vault.YAML()
		if err != nil {
			logger.Log().Fatalf("Failed to marshal user config - %v", err)
		}
		logger.Log().Println(str)
	case common.JSONFormat:
		str, err := vault.JSON()
		if err != nil {
			logger.Log().Fatalf("Failed to marshal user config - %v", err)
		}
		logger.Log().Println(str)
	default:
		logger.Log().Fatalf("Unsupported output format %s", format)
	}
}

func PrintVaultList(format string, vaultNames []string) {
	vaults := vaultCollection{Vaults: make([]*vaultEntity, 0, len(vaultNames))}
	for _, name := range vaultNames {
		vault, err := vaultFromName(name)
		if err != nil {
			logger.Log().Fatalf("Vault error %s - %v", name, err)
		}
		vaults.Vaults = append(vaults.Vaults, vault)
	}
	logger.Log().Debugf("listing %d vaults", len(vaults.Vaults))

	switch common.NormalizeFormat(format) {
	case common.YAMLFormat:
		str, err := vaults.YAML()
		if err != nil {
			logger.Log().Fatalf("Failed to marshal vault list - %v", err)
		}
		logger.Log().Println(str)
	case common.JSONFormat:
		str, err := vaults.JSON()
		if err != nil {
			logger.Log().Fatalf("Failed to marshal vault list - %v", err)
		}
		logger.Log().Println(str)
	default:
		logger.Log().Fatalf("Unsupported output format %s", format)
	}
}
