package vault

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/flowexec/vault"

	"github.com/flowexec/flow/internal/utils"
	"github.com/flowexec/flow/pkg/filesystem"
	"github.com/flowexec/flow/pkg/logger"
)

const (
	DefaultVaultKeyEnv      = "FLOW_VAULT_KEY"
	DefaultVaultIdentityEnv = "FLOW_VAULT_IDENTITY"

	v2CacheDataDir = "vaults"
	keyringService = "io.flowexec.flow"
)

type Vault = vault.Provider
type VaultConfig = vault.Config

// CreateResult contains metadata about a newly created vault.
type CreateResult struct {
	Name         string `json:"name"`
	Type         string `json:"type"`
	GeneratedKey string `json:"generatedKey,omitempty"`
}

func NewAES256Vault(name, storagePath, keyEnv, keyFile string) (*CreateResult, error) {
	if keyEnv == "" {
		logger.Log().Debugf("no AES key provided, using default environment variable %s", DefaultVaultKeyEnv)
		keyEnv = DefaultVaultKeyEnv
	} else {
		logger.Log().Debugf("using AES key from environment variable %s", keyEnv)
	}

	var generatedKey string
	key := os.Getenv(keyEnv)
	if key == "" {
		var err error
		key, err = vault.GenerateEncryptionKey()
		if err != nil {
			return nil, fmt.Errorf("unable to generate encryption key: %w", err)
		}
		generatedKey = key
		// this key needs to be set when initializing the vault
		if err := os.Setenv(keyEnv, key); err != nil {
			return nil, fmt.Errorf("unable to set environment variable %s: %w", keyEnv, err)
		}
	} else {
		logger.Log().Debugf("using existing AES key from environment variable %s", keyEnv)
	}

	storagePath = utils.ExpandPath(storagePath, CacheDirectory(""), nil)
	if storagePath == "" {
		return nil, fmt.Errorf("unable to expand storage path")
	}

	opts := []vault.Option{
		vault.WithAESPath(storagePath),
		vault.WithProvider(vault.ProviderTypeAES256),
		vault.WithAESKeyFromEnv(keyEnv),
	}

	if keyFile != "" {
		keyFile = utils.ExpandPath(keyFile, CacheDirectory(""), nil)
		if keyFile == "" {
			return nil, fmt.Errorf("unable to expand key file path")
		}
		opts = append(opts, vault.WithAESKeyFromFile(keyFile))
		if err := writeKeyToFile(key, keyFile); err != nil {
			logger.Log().Warn("unable to write key to file", "err", err)
		}
	}

	v, cfg, err := vault.New(name, opts...)
	if err != nil {
		return nil, err
	}

	cfgPath := ConfigFilePath(v.ID())
	if err = vault.SaveConfigJSON(*cfg, cfgPath); err != nil {
		return nil, fmt.Errorf("unable to save vault config: %w", err)
	}

	return &CreateResult{Name: v.ID(), Type: "aes256", GeneratedKey: generatedKey}, nil
}

func NewUnencryptedVault(name, storagePath string) (*CreateResult, error) {
	storagePath = utils.ExpandPath(storagePath, CacheDirectory(""), nil)
	if storagePath == "" {
		return nil, fmt.Errorf("unable to expand storage path")
	}

	opts := []vault.Option{vault.WithUnencryptedPath(storagePath), vault.WithProvider(vault.ProviderTypeUnencrypted)}

	v, cfg, err := vault.New(name, opts...)
	if err != nil {
		return nil, err
	}

	cfgPath := ConfigFilePath(v.ID())
	if err = vault.SaveConfigJSON(*cfg, cfgPath); err != nil {
		return nil, fmt.Errorf("unable to save vault config: %w", err)
	}

	return &CreateResult{Name: v.ID(), Type: "unencrypted"}, nil
}

func NewAgeVault(name, storagePath, recipients, identityKey, identityFile string) (*CreateResult, error) {
	storagePath = utils.ExpandPath(storagePath, CacheDirectory(""), nil)
	if storagePath == "" {
		return nil, fmt.Errorf("unable to expand storage path")
	}

	opts := []vault.Option{vault.WithAgePath(storagePath), vault.WithProvider(vault.ProviderTypeAge)}
	if recipients != "" {
		opts = append(opts, vault.WithAgeRecipients(strings.Split(recipients, ",")...))
	}
	if identityKey != "" {
		opts = append(opts, vault.WithAgeIdentityFromEnv(identityKey))
	}
	if identityFile != "" {
		identityFile = utils.ExpandPath(identityFile, CacheDirectory(""), nil)
		opts = append(opts, vault.WithAgeIdentityFromFile(identityFile))
	}

	if identityKey == "" && identityFile == "" {
		logger.Log().Debugf("no Age identity provided, using default environment variable %s", DefaultVaultIdentityEnv)
		opts = append(opts, vault.WithAgeIdentityFromEnv(DefaultVaultIdentityEnv))
	}

	v, cfg, err := vault.New(name, opts...)
	if err != nil {
		return nil, err
	}

	cfgPath := ConfigFilePath(v.ID())
	if err = vault.SaveConfigJSON(*cfg, cfgPath); err != nil {
		return nil, fmt.Errorf("unable to save vault config: %w", err)
	}

	return &CreateResult{Name: v.ID(), Type: "age"}, nil
}

func NewKeyringVault(name string) (*CreateResult, error) {
	opts := []vault.Option{
		vault.WithKeyringService(fmt.Sprintf("%s.%s", keyringService, name)),
		vault.WithProvider(vault.ProviderTypeKeyring)}
	v, cfg, err := vault.New(name, opts...)
	if err != nil {
		return nil, err
	}

	cfgPath := ConfigFilePath(v.ID())
	if err = vault.SaveConfigJSON(*cfg, cfgPath); err != nil {
		return nil, fmt.Errorf("unable to save vault config: %w", err)
	}

	return &CreateResult{Name: v.ID(), Type: "keyring"}, nil
}

func NewExternalVault(providerConfigFile string) (*CreateResult, error) {
	if providerConfigFile == "" {
		return nil, fmt.Errorf("provider config file path cannot be empty")
	}

	providerConfigFile = utils.ExpandPath(providerConfigFile, CacheDirectory(""), nil)
	if providerConfigFile == "" {
		return nil, fmt.Errorf("unable to expand provider config file path")
	}

	cfg, err := vault.LoadConfigJSON(providerConfigFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load vault config: %w", err)
	}

	v, _, err := vault.New(cfg.ID, vault.WithExternalConfig(cfg.External))
	if err != nil {
		return nil, err
	}

	cfgPath := ConfigFilePath(v.ID())
	if err = vault.SaveConfigJSON(cfg, cfgPath); err != nil {
		return nil, fmt.Errorf("unable to save vault config: %w", err)
	}

	return &CreateResult{Name: v.ID(), Type: "external"}, nil
}

func VaultFromName(name string) (*VaultConfig, Vault, error) {
	if name == "" {
		return nil, nil, fmt.Errorf("vault name cannot be empty")
	} else if strings.ToLower(name) == DemoVaultReservedName {
		return newDemoVaultConfig(), newDemoVault(), nil
	}

	cfgPath := ConfigFilePath(name)
	cfg, err := vault.LoadConfigJSON(cfgPath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to load vault config: %w", err)
	}

	switch cfg.Type {
	case vault.ProviderTypeAge:
		provider, err := vault.NewAgeVault(&cfg)
		return &cfg, provider, err
	case vault.ProviderTypeAES256:
		provider, err := vault.NewAES256Vault(&cfg)
		return &cfg, provider, err
	case vault.ProviderTypeUnencrypted:
		provider, err := vault.NewUnencryptedVault(&cfg)
		return &cfg, provider, err
	case vault.ProviderTypeKeyring:
		provider, err := vault.NewKeyringVault(&cfg)
		return &cfg, provider, err
	case vault.ProviderTypeExternal:
		// todo: rename this func in the vault pkg
		provider, err := vault.NewExternalVaultProvider(&cfg)
		return &cfg, provider, err
	default:
		return nil, nil, fmt.Errorf("unsupported vault type: %s", cfg.Type)
	}
}

func CacheDirectory(subPath string) string {
	return filepath.Join(filesystem.CachedDataDirPath(), v2CacheDataDir, subPath)
}

func ConfigFilePath(vaultName string) string {
	return filepath.Join(
		filesystem.CachedDataDirPath(),
		v2CacheDataDir,
		fmt.Sprintf("configs/%s.json", vaultName),
	)
}

func writeKeyToFile(key, filePath string) error {
	if key == "" {
		return nil
	}
	if filePath == "" {
		return fmt.Errorf("no file path provided to write key")
	}

	if _, err := os.Stat(filePath); err == nil {
		logger.Log().Debugf("key file already exists at %s, skipping write", filePath)
		return nil
	}

	if err := os.MkdirAll(filepath.Dir(filePath), 0750); err != nil {
		return fmt.Errorf("unable to create directory for key file: %w", err)
	}

	if err := os.WriteFile(filePath, []byte(key), 0600); err != nil { // #nosec G703
		return fmt.Errorf("unable to write key to file: %w", err)
	}
	logger.Log().Infof("Key written to file: %s", filePath)

	return nil
}
