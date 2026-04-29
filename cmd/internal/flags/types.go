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

var FilterAnnotationFlag = &Metadata{
	Name: "annotation",
	Usage: "Filter by annotations. Format: 'key=value' for exact value match, " +
		"or 'key' for presence regardless of value. Repeat the flag for multiple selectors; " +
		"all selectors must match (AND).",
	Default:  []string{},
	Required: false,
}

var OutputFormatFlag = &Metadata{
	Name:      "output",
	Shorthand: "o",
	Usage:     "Output format. One of: yaml, json, or tui.",
	Default:   "",
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

var GitBranchFlag = &Metadata{
	Name:      "branch",
	Shorthand: "b",
	Usage:     "Git branch to checkout when cloning a git workspace",
	Default:   "",
	Required:  false,
}

var GitTagFlag = &Metadata{
	Name:    "tag",
	Usage:   "Git tag to checkout when cloning a git workspace",
	Default: "",
}

var GitDepthFlag = &Metadata{
	Name:    "depth",
	Usage:   "Git clone depth (0 for full history)",
	Default: 0,
}

var GitPullFlag = &Metadata{
	Name:      "git",
	Shorthand: "g",
	Usage:     "Pull latest changes for all git-sourced workspaces before syncing",
	Default:   false,
	Required:  false,
}

var ForceFlag = &Metadata{
	Name:    "force",
	Usage:   "Force update by discarding local changes (hard reset to remote)",
	Default: false,
}

var YesFlag = &Metadata{
	Name:      "yes",
	Shorthand: "y",
	Usage:     "Skip confirmation prompts",
	Default:   false,
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

var LogFilterWorkspaceFlag = &Metadata{
	Name:      "workspace",
	Shorthand: "w",
	Usage:     "Filter history by workspace name.",
	Default:   "",
	Required:  false,
}

var LogFilterStatusFlag = &Metadata{
	Name:     "status",
	Usage:    "Filter history by status (success or failure).",
	Default:  "",
	Required: false,
}

var LogFilterSinceFlag = &Metadata{
	Name:     "since",
	Usage:    "Filter history to entries after a duration (e.g. 1h, 30m, 7d).",
	Default:  "",
	Required: false,
}

var LogFilterLimitFlag = &Metadata{
	Name:     "limit",
	Usage:    "Maximum number of records to display.",
	Default:  0,
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
	Name:      "dir",
	Shorthand: "d",
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

var BackgroundFlag = &Metadata{
	Name:      "background",
	Shorthand: "b",
	Usage:     "Run the executable in the background and return a run ID immediately.",
	Default:   false,
	Required:  false,
}

var RunningFlag = &Metadata{
	Name:     "running",
	Usage:    "Show only active background processes.",
	Default:  false,
	Required: false,
}

var ParameterValueFlag = &Metadata{
	Name:      "param",
	Shorthand: "p",
	Usage: "Set a parameter value by env key. (i.e. KEY=value) Use multiple times to set multiple parameters. " +
		"This will override any existing parameter values defined for the executable.",
	Default: []string{},
}

var TemplateFieldFlag = &Metadata{
	Name:      "set",
	Shorthand: "s",
	Usage: "Set a form field value by key (KEY=value). Repeat for multiple fields. " +
		"Fields with a matching key will skip the interactive prompt.",
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

var VaultNameFlag = &Metadata{
	Name:      "vault",
	Shorthand: "V",
	Usage:     "Vault name to use instead of the current vault.",
	Default:   "",
	Required:  false,
}

var CLIVersionFlag = &Metadata{
	Name:    "version",
	Usage:   "Target version to install (e.g. v2.1.0). Defaults to the latest release.",
	Default: "",
}

var StrictFlag = &Metadata{
	Name:    "strict",
	Usage:   "Also check for unknown keys not defined in the schema",
	Default: false,
}

var FileTypeFlag = &Metadata{
	Name:    "type",
	Usage:   "File type to validate as (flowfile, workspace, config, template). Auto-detected if omitted.",
	Default: "",
}

var GlobalCacheFlag = &Metadata{
	Name:      "global",
	Shorthand: "g",
	Usage:     "Force use of the global cache scope, even when called from within an executable",
	Default:   false,
	Required:  false,
}
