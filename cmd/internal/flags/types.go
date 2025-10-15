//nolint:lll
package flags

type Metadata struct {
	Name      string
	Shorthand string
	Usage     string
	Default   interface{}
	Required  bool
}

var LogLevel = &Metadata{
	Name:      "log-level",
	Usage:     "Log verbosity level (debug, info, fatal)",
	Shorthand: "L",
	Default:   "info",
	Required:  false,
}

var LogModeFlag = &Metadata{
	Name:      "log-mode",
	Shorthand: "m",
	Usage:     "Log mode (text, logfmt, json, hidden)",
	Default:   "",
	Required:  false,
}

var SyncCacheFlag = &Metadata{
	Name:     "sync",
	Usage:    "Sync flow cache and workspaces",
	Default:  false,
	Required: false,
}

var FilterExecSubstringFlag = &Metadata{
	Name:      "filter",
	Shorthand: "f",
	Usage:     "Filter executable by reference substring.",
	Default:   "",
	Required:  false,
}

var FilterWorkspaceFlag = &Metadata{
	Name:      "workspace",
	Shorthand: "w",
	Usage:     "Filter executables by workspace.",
	Default:   "",
	Required:  false,
}

var AllNamespacesFlag = &Metadata{
	Name:      "all",
	Shorthand: "a",
	Usage:     "List from all namespaces.",
	Default:   false,
	Required:  false,
}

var FilterNamespaceFlag = &Metadata{
	Name:      "namespace",
	Shorthand: "n",
	Usage:     "Filter executables by namespace.",
	Default:   "",
	Required:  false,
}

var FilterVerbFlag = &Metadata{
	Name:      "verb",
	Shorthand: "v",
	Usage:     "Filter executables by verb.",
	Default:   "",
	Required:  false,
}

var VisibilityFlag = &Metadata{
	Name:     "visibility",
	Usage:    "Filter by visibility level (hierarchical). Valid: public, private, internal, hidden. Default: private",
	Default:  "",
	Required: false,
}

var FilterTagFlag = &Metadata{
	Name:      "tag",
	Shorthand: "t",
	Usage:     "Filter by tags.",
	Default:   []string{},
	Required:  false,
}

var OutputFormatFlag = &Metadata{
	Name:      "output",
	Shorthand: "o",
	Usage:     "Output format. One of: yaml, json, or tui.",
	Default:   "tui",
	Required:  false,
}

var OutputSecretAsPlainTextFlag = &Metadata{
	Name:      "plaintext",
	Shorthand: "p",
	Usage:     "Output the secret value as plain text instead of an obfuscated string",
	Default:   false,
	Required:  false,
}

var SetAfterCreateFlag = &Metadata{
	Name:      "set",
	Shorthand: "s",
	Usage:     "Set the newly created workspace as the current workspace",
	Default:   false,
	Required:  false,
}

var FixedWsModeFlag = &Metadata{
	Name:      "fixed",
	Shorthand: "f",
	Usage:     "Set the workspace mode to fixed",
	Default:   false,
	Required:  false,
}

var ListFlag = &Metadata{
	Name:      "list",
	Shorthand: "l",
	Usage:     "Show a simple list view of executables instead of interactive discovery.",
	Default:   false,
	Required:  false,
}

var CopyFlag = &Metadata{
	Name:     "copy",
	Usage:    "Copy the secret value to the clipboard",
	Default:  false,
	Required: false,
}

var SecretFromFile = &Metadata{
	Name:     "file",
	Usage:    "File to read the secret's value from",
	Default:  "",
	Required: false,
}

var LastLogEntryFlag = &Metadata{
	Name:     "last",
	Usage:    "Print the last execution's logs",
	Default:  false,
	Required: false,
}

var TemplateWorkspaceFlag = &Metadata{
	Name:      "workspace",
	Shorthand: "w",
	Usage:     "Workspace to create the flow file and its artifacts. Defaults to the current workspace.",
	Default:   "",
	Required:  false,
}

var TemplateOutputPathFlag = &Metadata{
	Name:      "output",
	Shorthand: "o",
	Usage:     "Output directory (within the workspace) to create the flow file and its artifacts. If the directory does not exist, it will be created.",
	Default:   "",
	Required:  false,
}

var TemplateFlag = &Metadata{
	Name:      "template",
	Shorthand: "t",
	Usage:     "Registered template name. Templates can be registered in the flow configuration file or with `flow set template`.",
	Default:   "",
	Required:  false,
}

var TemplateFilePathFlag = &Metadata{
	Name:      "file",
	Shorthand: "f",
	Usage:     "Path to the template file. It must be a valid flow file template.",
	Default:   "",
	Required:  false,
}

var SetSoundNotificationFlag = &Metadata{
	Name:    "sound",
	Usage:   "Update completion sound notification setting",
	Default: false,
}

var StoreAllFlag = &Metadata{
	Name:    "all",
	Usage:   "Force clear all stored data",
	Default: false,
}

var ParameterValueFlag = &Metadata{
	Name:      "param",
	Shorthand: "p",
	Usage: "Set a parameter value by env key. (i.e. KEY=value) Use multiple times to set multiple parameters." +
		"This will override any existing parameter values defined for the executable.",
	Default: []string{},
}

var VaultSetFlag = &Metadata{
	Name:      "set",
	Shorthand: "s",
	Usage:     "Set the newly created vault as the current vault",
	Default:   false,
}

var VaultTypeFlag = &Metadata{
	Name:      "type",
	Shorthand: "t",
	Usage:     "Vault type. Either unencrypted, age, aes256, keyring, or external",
	Default:   "aes256",
	Required:  false,
}

var VaultPathFlag = &Metadata{
	Name:      "path",
	Shorthand: "p",
	Usage:     "Directory that the vault will use to store its data. If not set, the vault will be stored in the flow cache directory.",
	Default:   "",
	Required:  false,
}

var VaultKeyEnvFlag = &Metadata{
	Name:     "key-env",
	Usage:    "Environment variable name for the vault encryption key. Only used for AES256 vaults.",
	Default:  "",
	Required: false,
}

var VaultKeyFileFlag = &Metadata{
	Name:     "key-file",
	Usage:    "File path for the vault encryption key. An absolute path is recommended. Only used for AES256 vaults.",
	Default:  "",
	Required: false,
}

var VaultRecipientsFlag = &Metadata{
	Name:     "recipients",
	Usage:    "Comma-separated list of recipient keys for the vault. Only used for Age vaults.",
	Default:  "",
	Required: false,
}

var VaultIdentityEnvFlag = &Metadata{
	Name:     "identity-env",
	Usage:    "Environment variable name for the Age vault identity. Only used for Age vaults.",
	Default:  "",
	Required: false,
}

var VaultIdentityFileFlag = &Metadata{
	Name:     "identity-file",
	Usage:    "File path for the Age vault identity. An absolute path is recommended. Only used for Age vaults.",
	Default:  "",
	Required: false,
}

var VaultFromFileFlag = &Metadata{
	Name:      "config",
	Shorthand: "c",
	Usage:     "File path to read the external vault's configuration from. The file must be a valid vault configuration file.",
	Default:   "",
	Required:  false,
}

var GlobalCacheFlag = &Metadata{
	Name:      "global",
	Shorthand: "g",
	Usage:     "Force use of the global cache scope, even when called from within an executable",
	Default:   false,
	Required:  false,
}
