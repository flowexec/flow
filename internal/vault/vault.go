package vault

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/flowexec/vault"

	"github.com/flowexec/flow/v2/internal/utils"
	"github.com/flowexec/flow/v2/pkg/filesystem"
	"github.com/flowexec/flow/v2/pkg/logger"
)

const (
	DefaultVaultKeyEnv      = "FLOW_VAULT_KEY"
	DefaultVaultIdentityEnv = "FLOW_VAULT_IDENTITY"

	v2CacheDataDir = "vaults"
	keyringService = "io.flowexec.flow"
)

type Vault = vault.Provider
type VaultConfig = vault.Config

// ReferenceVault is implemented by vaults that link a key to a secret kept in
// another system rather than storing the secret themselves -- currently the
// external provider. Callers type-assert a Vault to reach it.
type ReferenceVault = vault.ReferenceVault

// Re-exported so command code can recognise these without importing the
// library alongside this package.
var (
	ErrReadOnly         = vault.ErrReadOnly
	ErrSecretNotFound   = vault.ErrSecretNotFound
	ErrInvalidReference = vault.ErrInvalidReference
)

// AsReferenceVault reports whether a vault holds links rather than secrets.
//
// Read-through vaults answer "remove" by forgetting where something is, not by
// destroying it, so callers phrase their prompts and messages differently. This
// is the one check that difference hangs on.
func AsReferenceVault(v Vault) (ReferenceVault, bool) {
	ref, ok := v.(ReferenceVault)
	return ref, ok
}

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

	// The ID comes from an external file and is used to derive the stored config path,
	// so validate it before it reaches the filesystem.
	if err := ValidateIdentifier(cfg.ID); err != nil {
		return nil, fmt.Errorf("invalid vault name %q in config: %w", cfg.ID, err)
	}

	applyExternalDefaults(&cfg)

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

	// Validate before deriving a filesystem path so a crafted name (e.g. containing
	// path separators or "..") cannot be used to read arbitrary files as vault configs.
	if err := ValidateIdentifier(name); err != nil {
		return nil, nil, fmt.Errorf("invalid vault name: %w", err)
	}

	cfgPath := ConfigFilePath(name)
	cfg, err := vault.LoadConfigJSON(cfgPath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to load vault config: %w", err)
	}

	applyExternalDefaults(&cfg)

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

// configsDir is the directory holding per-vault configuration files. It is the source
// of truth for which vaults exist: a vault is usable iff its <name>.json lives here.
func configsDir() string {
	return CacheDirectory("configs")
}

func ConfigFilePath(vaultName string) string {
	return filepath.Join(configsDir(), fmt.Sprintf("%s.json", vaultName))
}

// ListVaultNames returns the names of all configured vaults by reading the vault config
// directory — the same source VaultFromName loads from — so listings never drift from
// what is actually openable. The reserved demo vault is not included (it has no config
// file); callers that want it should add it explicitly. Names are sorted.
func ListVaultNames() ([]string, error) {
	entries, err := os.ReadDir(configsDir())
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("unable to read vault config directory: %w", err)
	}

	names := make([]string, 0, len(entries))
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".json") {
			continue
		}
		names = append(names, strings.TrimSuffix(e.Name(), ".json"))
	}
	sort.Strings(names)
	return names, nil
}

// VaultExists reports whether a vault with the given name is available — either the
// reserved demo vault or a vault whose config file exists on disk.
func VaultExists(name string) bool {
	if strings.EqualFold(name, DemoVaultReservedName) {
		return true
	}
	if ValidateIdentifier(name) != nil {
		return false
	}
	_, err := os.Stat(ConfigFilePath(name))
	return err == nil
}

// RemoveVaultConfig deletes a vault's configuration file. The vault's encrypted secret
// data, stored separately at the vault's storage path, is left untouched. Removing a
// non-existent config is not an error.
func RemoveVaultConfig(name string) error {
	if err := ValidateIdentifier(name); err != nil {
		return fmt.Errorf("invalid vault name: %w", err)
	}
	if err := os.Remove(ConfigFilePath(name)); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("unable to remove vault config: %w", err)
	}
	return nil
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

// applyExternalDefaults fills in settings a config authored elsewhere cannot
// know. External vaults created before the link registry existed have no
// storage path, so this has to run when opening a vault as well as creating one.
func applyExternalDefaults(cfg *VaultConfig) {
	if cfg.External != nil && cfg.External.StoragePath == "" {
		cfg.External.StoragePath = CacheDirectory(cfg.ID)
	}
}
