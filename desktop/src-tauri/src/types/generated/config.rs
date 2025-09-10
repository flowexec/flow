#![allow(clippy::redundant_closure_call)]
#![allow(clippy::needless_lifetimes)]
#![allow(clippy::match_single_binding)]
#![allow(clippy::clone_on_copy)]

#[doc = r" Error types."]
pub mod error {
    #[doc = r" Error from a `TryFrom` or `FromStr` implementation."]
    pub struct ConversionError(::std::borrow::Cow<'static, str>);
    impl ::std::error::Error for ConversionError {}
    impl ::std::fmt::Display for ConversionError {
        fn fmt(&self, f: &mut ::std::fmt::Formatter<'_>) -> Result<(), ::std::fmt::Error> {
            ::std::fmt::Display::fmt(&self.0, f)
        }
    }
    impl ::std::fmt::Debug for ConversionError {
        fn fmt(&self, f: &mut ::std::fmt::Formatter<'_>) -> Result<(), ::std::fmt::Error> {
            ::std::fmt::Debug::fmt(&self.0, f)
        }
    }
    impl From<&'static str> for ConversionError {
        fn from(value: &'static str) -> Self {
            Self(value.into())
        }
    }
    impl From<String> for ConversionError {
        fn from(value: String) -> Self {
            Self(value.into())
        }
    }
}
#[doc = "The color palette for the interactive UI.\nThe colors can be either an ANSI 16, ANSI 256, or TrueColor (hex) value.\nIf unset, the default color for the current theme will be used.\n"]
#[doc = r""]
#[doc = r" <details><summary>JSON schema</summary>"]
#[doc = r""]
#[doc = r" ```json"]
#[doc = "{"]
#[doc = "  \"description\": \"The color palette for the interactive UI.\\nThe colors can be either an ANSI 16, ANSI 256, or TrueColor (hex) value.\\nIf unset, the default color for the current theme will be used.\\n\","]
#[doc = "  \"type\": \"object\","]
#[doc = "  \"properties\": {"]
#[doc = "    \"black\": {"]
#[doc = "      \"type\": \"string\""]
#[doc = "    },"]
#[doc = "    \"body\": {"]
#[doc = "      \"type\": \"string\""]
#[doc = "    },"]
#[doc = "    \"border\": {"]
#[doc = "      \"type\": \"string\""]
#[doc = "    },"]
#[doc = "    \"codeStyle\": {"]
#[doc = "      \"description\": \"The style of the code block. For example, `monokai`, `dracula`, `github`, etc.\\nSee [chroma styles](https://github.com/alecthomas/chroma/tree/master/styles) for available style names.\\n\","]
#[doc = "      \"type\": \"string\""]
#[doc = "    },"]
#[doc = "    \"emphasis\": {"]
#[doc = "      \"type\": \"string\""]
#[doc = "    },"]
#[doc = "    \"error\": {"]
#[doc = "      \"type\": \"string\""]
#[doc = "    },"]
#[doc = "    \"gray\": {"]
#[doc = "      \"type\": \"string\""]
#[doc = "    },"]
#[doc = "    \"info\": {"]
#[doc = "      \"type\": \"string\""]
#[doc = "    },"]
#[doc = "    \"primary\": {"]
#[doc = "      \"type\": \"string\""]
#[doc = "    },"]
#[doc = "    \"secondary\": {"]
#[doc = "      \"type\": \"string\""]
#[doc = "    },"]
#[doc = "    \"success\": {"]
#[doc = "      \"type\": \"string\""]
#[doc = "    },"]
#[doc = "    \"tertiary\": {"]
#[doc = "      \"type\": \"string\""]
#[doc = "    },"]
#[doc = "    \"warning\": {"]
#[doc = "      \"type\": \"string\""]
#[doc = "    },"]
#[doc = "    \"white\": {"]
#[doc = "      \"type\": \"string\""]
#[doc = "    }"]
#[doc = "  }"]
#[doc = "}"]
#[doc = r" ```"]
#[doc = r" </details>"]
#[derive(:: serde :: Deserialize, :: serde :: Serialize, Clone, Debug)]
pub struct ColorPalette {
    #[serde(default, skip_serializing_if = "::std::option::Option::is_none")]
    pub black: ::std::option::Option<::std::string::String>,
    #[serde(default, skip_serializing_if = "::std::option::Option::is_none")]
    pub body: ::std::option::Option<::std::string::String>,
    #[serde(default, skip_serializing_if = "::std::option::Option::is_none")]
    pub border: ::std::option::Option<::std::string::String>,
    #[doc = "The style of the code block. For example, `monokai`, `dracula`, `github`, etc.\nSee [chroma styles](https://github.com/alecthomas/chroma/tree/master/styles) for available style names.\n"]
    #[serde(
        rename = "codeStyle",
        default,
        skip_serializing_if = "::std::option::Option::is_none"
    )]
    pub code_style: ::std::option::Option<::std::string::String>,
    #[serde(default, skip_serializing_if = "::std::option::Option::is_none")]
    pub emphasis: ::std::option::Option<::std::string::String>,
    #[serde(default, skip_serializing_if = "::std::option::Option::is_none")]
    pub error: ::std::option::Option<::std::string::String>,
    #[serde(default, skip_serializing_if = "::std::option::Option::is_none")]
    pub gray: ::std::option::Option<::std::string::String>,
    #[serde(default, skip_serializing_if = "::std::option::Option::is_none")]
    pub info: ::std::option::Option<::std::string::String>,
    #[serde(default, skip_serializing_if = "::std::option::Option::is_none")]
    pub primary: ::std::option::Option<::std::string::String>,
    #[serde(default, skip_serializing_if = "::std::option::Option::is_none")]
    pub secondary: ::std::option::Option<::std::string::String>,
    #[serde(default, skip_serializing_if = "::std::option::Option::is_none")]
    pub success: ::std::option::Option<::std::string::String>,
    #[serde(default, skip_serializing_if = "::std::option::Option::is_none")]
    pub tertiary: ::std::option::Option<::std::string::String>,
    #[serde(default, skip_serializing_if = "::std::option::Option::is_none")]
    pub warning: ::std::option::Option<::std::string::String>,
    #[serde(default, skip_serializing_if = "::std::option::Option::is_none")]
    pub white: ::std::option::Option<::std::string::String>,
}
impl ::std::convert::From<&ColorPalette> for ColorPalette {
    fn from(value: &ColorPalette) -> Self {
        value.clone()
    }
}
impl ::std::default::Default for ColorPalette {
    fn default() -> Self {
        Self {
            black: Default::default(),
            body: Default::default(),
            border: Default::default(),
            code_style: Default::default(),
            emphasis: Default::default(),
            error: Default::default(),
            gray: Default::default(),
            info: Default::default(),
            primary: Default::default(),
            secondary: Default::default(),
            success: Default::default(),
            tertiary: Default::default(),
            warning: Default::default(),
            white: Default::default(),
        }
    }
}
impl ColorPalette {
    pub fn builder() -> builder::ColorPalette {
        Default::default()
    }
}
#[doc = "User Configuration for the Flow CLI.\nIncludes configurations for workspaces, templates, I/O, and other settings for the CLI.\n\nIt is read from the user's flow config directory:\n- **MacOS**: `$HOME/Library/Application Support/flow`\n- **Linux**: `$HOME/.config/flow`\n- **Windows**: `%APPDATA%\\flow`\n\nAlternatively, a custom path can be set using the `FLOW_CONFIG_PATH` environment variable.\n"]
#[doc = r""]
#[doc = r" <details><summary>JSON schema</summary>"]
#[doc = r""]
#[doc = r" ```json"]
#[doc = "{"]
#[doc = "  \"$id\": \"https://flowexec.io/schemas/config_schema.json\","]
#[doc = "  \"title\": \"Config\","]
#[doc = "  \"description\": \"User Configuration for the Flow CLI.\\nIncludes configurations for workspaces, templates, I/O, and other settings for the CLI.\\n\\nIt is read from the user's flow config directory:\\n- **MacOS**: `$HOME/Library/Application Support/flow`\\n- **Linux**: `$HOME/.config/flow`\\n- **Windows**: `%APPDATA%\\\\flow`\\n\\nAlternatively, a custom path can be set using the `FLOW_CONFIG_PATH` environment variable.\\n\","]
#[doc = "  \"type\": \"object\","]
#[doc = "  \"required\": ["]
#[doc = "    \"currentWorkspace\","]
#[doc = "    \"workspaces\""]
#[doc = "  ],"]
#[doc = "  \"properties\": {"]
#[doc = "    \"colorOverride\": {"]
#[doc = "      \"description\": \"Override the default color palette for the interactive UI.\\nThis can be used to customize the colors of the UI.\\n\","]
#[doc = "      \"$ref\": \"#/definitions/ColorPalette\""]
#[doc = "    },"]
#[doc = "    \"currentNamespace\": {"]
#[doc = "      \"description\": \"The name of the current namespace.\\n\\nNamespaces are used to reference executables in the CLI using the format `workspace:namespace/name`.\\nIf the namespace is not set, only executables defined without a namespace will be discovered.\\n\","]
#[doc = "      \"default\": \"\","]
#[doc = "      \"type\": \"string\""]
#[doc = "    },"]
#[doc = "    \"currentVault\": {"]
#[doc = "      \"description\": \"The name of the current vault. This should match a key in the `vaults` map.\","]
#[doc = "      \"type\": \"string\""]
#[doc = "    },"]
#[doc = "    \"currentWorkspace\": {"]
#[doc = "      \"description\": \"The name of the current workspace. This should match a key in the `workspaces` or `remoteWorkspaces` map.\","]
#[doc = "      \"default\": \"\","]
#[doc = "      \"type\": \"string\""]
#[doc = "    },"]
#[doc = "    \"defaultLogMode\": {"]
#[doc = "      \"description\": \"The default log mode to use when running executables.\\nThis can either be `hidden`, `json`, `logfmt` or `text`\\n\\n`hidden` will not display any logs.\\n`json` will display logs in JSON format.\\n`logfmt` will display logs with a log level, timestamp, and message.\\n`text` will just display the log message.\\n\","]
#[doc = "      \"default\": \"logfmt\","]
#[doc = "      \"type\": \"string\""]
#[doc = "    },"]
#[doc = "    \"defaultTimeout\": {"]
#[doc = "      \"description\": \"The default timeout to use when running executables.\\nThis should be a valid duration string.\\n\","]
#[doc = "      \"default\": \"30m\","]
#[doc = "      \"type\": \"string\""]
#[doc = "    },"]
#[doc = "    \"interactive\": {"]
#[doc = "      \"$ref\": \"#/definitions/Interactive\""]
#[doc = "    },"]
#[doc = "    \"templates\": {"]
#[doc = "      \"description\": \"A map of flowfile template names to their paths.\","]
#[doc = "      \"default\": {},"]
#[doc = "      \"type\": \"object\","]
#[doc = "      \"additionalProperties\": {"]
#[doc = "        \"type\": \"string\""]
#[doc = "      }"]
#[doc = "    },"]
#[doc = "    \"theme\": {"]
#[doc = "      \"description\": \"The theme of the interactive UI.\","]
#[doc = "      \"default\": \"default\","]
#[doc = "      \"type\": \"string\","]
#[doc = "      \"enum\": ["]
#[doc = "        \"default\","]
#[doc = "        \"everforest\","]
#[doc = "        \"dark\","]
#[doc = "        \"light\","]
#[doc = "        \"dracula\","]
#[doc = "        \"tokyo-night\""]
#[doc = "      ]"]
#[doc = "    },"]
#[doc = "    \"vaults\": {"]
#[doc = "      \"description\": \"A map of vault names to their paths. The path should be a valid absolute path to the vault file created by flow.\","]
#[doc = "      \"type\": \"object\","]
#[doc = "      \"additionalProperties\": {"]
#[doc = "        \"type\": \"string\""]
#[doc = "      }"]
#[doc = "    },"]
#[doc = "    \"workspaceMode\": {"]
#[doc = "      \"description\": \"The mode of the workspace. This can be either `fixed` or `dynamic`.\\nIn `fixed` mode, the current workspace used at runtime is always the one set in the currentWorkspace config field.\\nIn `dynamic` mode, the current workspace used at runtime is determined by the current directory.\\nIf the current directory is within a workspace, that workspace is used.\\n\","]
#[doc = "      \"default\": \"dynamic\","]
#[doc = "      \"type\": \"string\","]
#[doc = "      \"enum\": ["]
#[doc = "        \"fixed\","]
#[doc = "        \"dynamic\""]
#[doc = "      ]"]
#[doc = "    },"]
#[doc = "    \"workspaces\": {"]
#[doc = "      \"description\": \"Map of workspace names to their paths. The path should be a valid absolute path to the workspace directory.\\n\","]
#[doc = "      \"type\": \"object\","]
#[doc = "      \"additionalProperties\": {"]
#[doc = "        \"type\": \"string\""]
#[doc = "      }"]
#[doc = "    }"]
#[doc = "  }"]
#[doc = "}"]
#[doc = r" ```"]
#[doc = r" </details>"]
#[derive(:: serde :: Deserialize, :: serde :: Serialize, Clone, Debug)]
pub struct Config {
    #[doc = "Override the default color palette for the interactive UI.\nThis can be used to customize the colors of the UI.\n"]
    #[serde(
        rename = "colorOverride",
        default,
        skip_serializing_if = "::std::option::Option::is_none"
    )]
    pub color_override: ::std::option::Option<ColorPalette>,
    #[doc = "The name of the current namespace.\n\nNamespaces are used to reference executables in the CLI using the format `workspace:namespace/name`.\nIf the namespace is not set, only executables defined without a namespace will be discovered.\n"]
    #[serde(rename = "currentNamespace", default)]
    pub current_namespace: ::std::string::String,
    #[doc = "The name of the current vault. This should match a key in the `vaults` map."]
    #[serde(
        rename = "currentVault",
        default,
        skip_serializing_if = "::std::option::Option::is_none"
    )]
    pub current_vault: ::std::option::Option<::std::string::String>,
    #[doc = "The name of the current workspace. This should match a key in the `workspaces` or `remoteWorkspaces` map."]
    #[serde(rename = "currentWorkspace")]
    pub current_workspace: ::std::string::String,
    #[doc = "The default log mode to use when running executables.\nThis can either be `hidden`, `json`, `logfmt` or `text`\n\n`hidden` will not display any logs.\n`json` will display logs in JSON format.\n`logfmt` will display logs with a log level, timestamp, and message.\n`text` will just display the log message.\n"]
    #[serde(
        rename = "defaultLogMode",
        default = "defaults::config_default_log_mode"
    )]
    pub default_log_mode: ::std::string::String,
    #[doc = "The default timeout to use when running executables.\nThis should be a valid duration string.\n"]
    #[serde(
        rename = "defaultTimeout",
        default = "defaults::config_default_timeout"
    )]
    pub default_timeout: ::std::string::String,
    #[serde(default, skip_serializing_if = "::std::option::Option::is_none")]
    pub interactive: ::std::option::Option<Interactive>,
    #[doc = "A map of flowfile template names to their paths."]
    #[serde(
        default,
        skip_serializing_if = ":: std :: collections :: HashMap::is_empty"
    )]
    pub templates: ::std::collections::HashMap<::std::string::String, ::std::string::String>,
    #[doc = "The theme of the interactive UI."]
    #[serde(default = "defaults::config_theme")]
    pub theme: ConfigTheme,
    #[doc = "A map of vault names to their paths. The path should be a valid absolute path to the vault file created by flow."]
    #[serde(
        default,
        skip_serializing_if = ":: std :: collections :: HashMap::is_empty"
    )]
    pub vaults: ::std::collections::HashMap<::std::string::String, ::std::string::String>,
    #[doc = "The mode of the workspace. This can be either `fixed` or `dynamic`.\nIn `fixed` mode, the current workspace used at runtime is always the one set in the currentWorkspace config field.\nIn `dynamic` mode, the current workspace used at runtime is determined by the current directory.\nIf the current directory is within a workspace, that workspace is used.\n"]
    #[serde(rename = "workspaceMode", default = "defaults::config_workspace_mode")]
    pub workspace_mode: ConfigWorkspaceMode,
    #[doc = "Map of workspace names to their paths. The path should be a valid absolute path to the workspace directory.\n"]
    pub workspaces: ::std::collections::HashMap<::std::string::String, ::std::string::String>,
}
impl ::std::convert::From<&Config> for Config {
    fn from(value: &Config) -> Self {
        value.clone()
    }
}
impl Config {
    pub fn builder() -> builder::Config {
        Default::default()
    }
}
#[doc = "The theme of the interactive UI."]
#[doc = r""]
#[doc = r" <details><summary>JSON schema</summary>"]
#[doc = r""]
#[doc = r" ```json"]
#[doc = "{"]
#[doc = "  \"description\": \"The theme of the interactive UI.\","]
#[doc = "  \"default\": \"default\","]
#[doc = "  \"type\": \"string\","]
#[doc = "  \"enum\": ["]
#[doc = "    \"default\","]
#[doc = "    \"everforest\","]
#[doc = "    \"dark\","]
#[doc = "    \"light\","]
#[doc = "    \"dracula\","]
#[doc = "    \"tokyo-night\""]
#[doc = "  ]"]
#[doc = "}"]
#[doc = r" ```"]
#[doc = r" </details>"]
#[derive(
    :: serde :: Deserialize,
    :: serde :: Serialize,
    Clone,
    Copy,
    Debug,
    Eq,
    Hash,
    Ord,
    PartialEq,
    PartialOrd,
)]
pub enum ConfigTheme {
    #[serde(rename = "default")]
    Default,
    #[serde(rename = "everforest")]
    Everforest,
    #[serde(rename = "dark")]
    Dark,
    #[serde(rename = "light")]
    Light,
    #[serde(rename = "dracula")]
    Dracula,
    #[serde(rename = "tokyo-night")]
    TokyoNight,
}
impl ::std::convert::From<&Self> for ConfigTheme {
    fn from(value: &ConfigTheme) -> Self {
        value.clone()
    }
}
impl ::std::fmt::Display for ConfigTheme {
    fn fmt(&self, f: &mut ::std::fmt::Formatter<'_>) -> ::std::fmt::Result {
        match *self {
            Self::Default => f.write_str("default"),
            Self::Everforest => f.write_str("everforest"),
            Self::Dark => f.write_str("dark"),
            Self::Light => f.write_str("light"),
            Self::Dracula => f.write_str("dracula"),
            Self::TokyoNight => f.write_str("tokyo-night"),
        }
    }
}
impl ::std::str::FromStr for ConfigTheme {
    type Err = self::error::ConversionError;
    fn from_str(value: &str) -> ::std::result::Result<Self, self::error::ConversionError> {
        match value {
            "default" => Ok(Self::Default),
            "everforest" => Ok(Self::Everforest),
            "dark" => Ok(Self::Dark),
            "light" => Ok(Self::Light),
            "dracula" => Ok(Self::Dracula),
            "tokyo-night" => Ok(Self::TokyoNight),
            _ => Err("invalid value".into()),
        }
    }
}
impl ::std::convert::TryFrom<&str> for ConfigTheme {
    type Error = self::error::ConversionError;
    fn try_from(value: &str) -> ::std::result::Result<Self, self::error::ConversionError> {
        value.parse()
    }
}
impl ::std::convert::TryFrom<&::std::string::String> for ConfigTheme {
    type Error = self::error::ConversionError;
    fn try_from(
        value: &::std::string::String,
    ) -> ::std::result::Result<Self, self::error::ConversionError> {
        value.parse()
    }
}
impl ::std::convert::TryFrom<::std::string::String> for ConfigTheme {
    type Error = self::error::ConversionError;
    fn try_from(
        value: ::std::string::String,
    ) -> ::std::result::Result<Self, self::error::ConversionError> {
        value.parse()
    }
}
impl ::std::default::Default for ConfigTheme {
    fn default() -> Self {
        ConfigTheme::Default
    }
}
#[doc = "The mode of the workspace. This can be either `fixed` or `dynamic`.\nIn `fixed` mode, the current workspace used at runtime is always the one set in the currentWorkspace config field.\nIn `dynamic` mode, the current workspace used at runtime is determined by the current directory.\nIf the current directory is within a workspace, that workspace is used.\n"]
#[doc = r""]
#[doc = r" <details><summary>JSON schema</summary>"]
#[doc = r""]
#[doc = r" ```json"]
#[doc = "{"]
#[doc = "  \"description\": \"The mode of the workspace. This can be either `fixed` or `dynamic`.\\nIn `fixed` mode, the current workspace used at runtime is always the one set in the currentWorkspace config field.\\nIn `dynamic` mode, the current workspace used at runtime is determined by the current directory.\\nIf the current directory is within a workspace, that workspace is used.\\n\","]
#[doc = "  \"default\": \"dynamic\","]
#[doc = "  \"type\": \"string\","]
#[doc = "  \"enum\": ["]
#[doc = "    \"fixed\","]
#[doc = "    \"dynamic\""]
#[doc = "  ]"]
#[doc = "}"]
#[doc = r" ```"]
#[doc = r" </details>"]
#[derive(
    :: serde :: Deserialize,
    :: serde :: Serialize,
    Clone,
    Copy,
    Debug,
    Eq,
    Hash,
    Ord,
    PartialEq,
    PartialOrd,
)]
pub enum ConfigWorkspaceMode {
    #[serde(rename = "fixed")]
    Fixed,
    #[serde(rename = "dynamic")]
    Dynamic,
}
impl ::std::convert::From<&Self> for ConfigWorkspaceMode {
    fn from(value: &ConfigWorkspaceMode) -> Self {
        value.clone()
    }
}
impl ::std::fmt::Display for ConfigWorkspaceMode {
    fn fmt(&self, f: &mut ::std::fmt::Formatter<'_>) -> ::std::fmt::Result {
        match *self {
            Self::Fixed => f.write_str("fixed"),
            Self::Dynamic => f.write_str("dynamic"),
        }
    }
}
impl ::std::str::FromStr for ConfigWorkspaceMode {
    type Err = self::error::ConversionError;
    fn from_str(value: &str) -> ::std::result::Result<Self, self::error::ConversionError> {
        match value {
            "fixed" => Ok(Self::Fixed),
            "dynamic" => Ok(Self::Dynamic),
            _ => Err("invalid value".into()),
        }
    }
}
impl ::std::convert::TryFrom<&str> for ConfigWorkspaceMode {
    type Error = self::error::ConversionError;
    fn try_from(value: &str) -> ::std::result::Result<Self, self::error::ConversionError> {
        value.parse()
    }
}
impl ::std::convert::TryFrom<&::std::string::String> for ConfigWorkspaceMode {
    type Error = self::error::ConversionError;
    fn try_from(
        value: &::std::string::String,
    ) -> ::std::result::Result<Self, self::error::ConversionError> {
        value.parse()
    }
}
impl ::std::convert::TryFrom<::std::string::String> for ConfigWorkspaceMode {
    type Error = self::error::ConversionError;
    fn try_from(
        value: ::std::string::String,
    ) -> ::std::result::Result<Self, self::error::ConversionError> {
        value.parse()
    }
}
impl ::std::default::Default for ConfigWorkspaceMode {
    fn default() -> Self {
        ConfigWorkspaceMode::Dynamic
    }
}
#[doc = "Configurations for the interactive UI."]
#[doc = r""]
#[doc = r" <details><summary>JSON schema</summary>"]
#[doc = r""]
#[doc = r" ```json"]
#[doc = "{"]
#[doc = "  \"description\": \"Configurations for the interactive UI.\","]
#[doc = "  \"type\": \"object\","]
#[doc = "  \"required\": ["]
#[doc = "    \"enabled\""]
#[doc = "  ],"]
#[doc = "  \"properties\": {"]
#[doc = "    \"enabled\": {"]
#[doc = "      \"type\": \"boolean\""]
#[doc = "    },"]
#[doc = "    \"notifyOnCompletion\": {"]
#[doc = "      \"description\": \"Whether to send a desktop notification when a command completes.\","]
#[doc = "      \"type\": \"boolean\""]
#[doc = "    },"]
#[doc = "    \"soundOnCompletion\": {"]
#[doc = "      \"description\": \"Whether to play a sound when a command completes.\","]
#[doc = "      \"type\": \"boolean\""]
#[doc = "    }"]
#[doc = "  }"]
#[doc = "}"]
#[doc = r" ```"]
#[doc = r" </details>"]
#[derive(:: serde :: Deserialize, :: serde :: Serialize, Clone, Debug)]
pub struct Interactive {
    pub enabled: bool,
    #[doc = "Whether to send a desktop notification when a command completes."]
    #[serde(
        rename = "notifyOnCompletion",
        default,
        skip_serializing_if = "::std::option::Option::is_none"
    )]
    pub notify_on_completion: ::std::option::Option<bool>,
    #[doc = "Whether to play a sound when a command completes."]
    #[serde(
        rename = "soundOnCompletion",
        default,
        skip_serializing_if = "::std::option::Option::is_none"
    )]
    pub sound_on_completion: ::std::option::Option<bool>,
}
impl ::std::convert::From<&Interactive> for Interactive {
    fn from(value: &Interactive) -> Self {
        value.clone()
    }
}
impl Interactive {
    pub fn builder() -> builder::Interactive {
        Default::default()
    }
}
#[doc = r" Types for composing complex structures."]
pub mod builder {
    #[derive(Clone, Debug)]
    pub struct ColorPalette {
        black: ::std::result::Result<
            ::std::option::Option<::std::string::String>,
            ::std::string::String,
        >,
        body: ::std::result::Result<
            ::std::option::Option<::std::string::String>,
            ::std::string::String,
        >,
        border: ::std::result::Result<
            ::std::option::Option<::std::string::String>,
            ::std::string::String,
        >,
        code_style: ::std::result::Result<
            ::std::option::Option<::std::string::String>,
            ::std::string::String,
        >,
        emphasis: ::std::result::Result<
            ::std::option::Option<::std::string::String>,
            ::std::string::String,
        >,
        error: ::std::result::Result<
            ::std::option::Option<::std::string::String>,
            ::std::string::String,
        >,
        gray: ::std::result::Result<
            ::std::option::Option<::std::string::String>,
            ::std::string::String,
        >,
        info: ::std::result::Result<
            ::std::option::Option<::std::string::String>,
            ::std::string::String,
        >,
        primary: ::std::result::Result<
            ::std::option::Option<::std::string::String>,
            ::std::string::String,
        >,
        secondary: ::std::result::Result<
            ::std::option::Option<::std::string::String>,
            ::std::string::String,
        >,
        success: ::std::result::Result<
            ::std::option::Option<::std::string::String>,
            ::std::string::String,
        >,
        tertiary: ::std::result::Result<
            ::std::option::Option<::std::string::String>,
            ::std::string::String,
        >,
        warning: ::std::result::Result<
            ::std::option::Option<::std::string::String>,
            ::std::string::String,
        >,
        white: ::std::result::Result<
            ::std::option::Option<::std::string::String>,
            ::std::string::String,
        >,
    }
    impl ::std::default::Default for ColorPalette {
        fn default() -> Self {
            Self {
                black: Ok(Default::default()),
                body: Ok(Default::default()),
                border: Ok(Default::default()),
                code_style: Ok(Default::default()),
                emphasis: Ok(Default::default()),
                error: Ok(Default::default()),
                gray: Ok(Default::default()),
                info: Ok(Default::default()),
                primary: Ok(Default::default()),
                secondary: Ok(Default::default()),
                success: Ok(Default::default()),
                tertiary: Ok(Default::default()),
                warning: Ok(Default::default()),
                white: Ok(Default::default()),
            }
        }
    }
    impl ColorPalette {
        pub fn black<T>(mut self, value: T) -> Self
        where
            T: ::std::convert::TryInto<::std::option::Option<::std::string::String>>,
            T::Error: ::std::fmt::Display,
        {
            self.black = value
                .try_into()
                .map_err(|e| format!("error converting supplied value for black: {}", e));
            self
        }
        pub fn body<T>(mut self, value: T) -> Self
        where
            T: ::std::convert::TryInto<::std::option::Option<::std::string::String>>,
            T::Error: ::std::fmt::Display,
        {
            self.body = value
                .try_into()
                .map_err(|e| format!("error converting supplied value for body: {}", e));
            self
        }
        pub fn border<T>(mut self, value: T) -> Self
        where
            T: ::std::convert::TryInto<::std::option::Option<::std::string::String>>,
            T::Error: ::std::fmt::Display,
        {
            self.border = value
                .try_into()
                .map_err(|e| format!("error converting supplied value for border: {}", e));
            self
        }
        pub fn code_style<T>(mut self, value: T) -> Self
        where
            T: ::std::convert::TryInto<::std::option::Option<::std::string::String>>,
            T::Error: ::std::fmt::Display,
        {
            self.code_style = value
                .try_into()
                .map_err(|e| format!("error converting supplied value for code_style: {}", e));
            self
        }
        pub fn emphasis<T>(mut self, value: T) -> Self
        where
            T: ::std::convert::TryInto<::std::option::Option<::std::string::String>>,
            T::Error: ::std::fmt::Display,
        {
            self.emphasis = value
                .try_into()
                .map_err(|e| format!("error converting supplied value for emphasis: {}", e));
            self
        }
        pub fn error<T>(mut self, value: T) -> Self
        where
            T: ::std::convert::TryInto<::std::option::Option<::std::string::String>>,
            T::Error: ::std::fmt::Display,
        {
            self.error = value
                .try_into()
                .map_err(|e| format!("error converting supplied value for error: {}", e));
            self
        }
        pub fn gray<T>(mut self, value: T) -> Self
        where
            T: ::std::convert::TryInto<::std::option::Option<::std::string::String>>,
            T::Error: ::std::fmt::Display,
        {
            self.gray = value
                .try_into()
                .map_err(|e| format!("error converting supplied value for gray: {}", e));
            self
        }
        pub fn info<T>(mut self, value: T) -> Self
        where
            T: ::std::convert::TryInto<::std::option::Option<::std::string::String>>,
            T::Error: ::std::fmt::Display,
        {
            self.info = value
                .try_into()
                .map_err(|e| format!("error converting supplied value for info: {}", e));
            self
        }
        pub fn primary<T>(mut self, value: T) -> Self
        where
            T: ::std::convert::TryInto<::std::option::Option<::std::string::String>>,
            T::Error: ::std::fmt::Display,
        {
            self.primary = value
                .try_into()
                .map_err(|e| format!("error converting supplied value for primary: {}", e));
            self
        }
        pub fn secondary<T>(mut self, value: T) -> Self
        where
            T: ::std::convert::TryInto<::std::option::Option<::std::string::String>>,
            T::Error: ::std::fmt::Display,
        {
            self.secondary = value
                .try_into()
                .map_err(|e| format!("error converting supplied value for secondary: {}", e));
            self
        }
        pub fn success<T>(mut self, value: T) -> Self
        where
            T: ::std::convert::TryInto<::std::option::Option<::std::string::String>>,
            T::Error: ::std::fmt::Display,
        {
            self.success = value
                .try_into()
                .map_err(|e| format!("error converting supplied value for success: {}", e));
            self
        }
        pub fn tertiary<T>(mut self, value: T) -> Self
        where
            T: ::std::convert::TryInto<::std::option::Option<::std::string::String>>,
            T::Error: ::std::fmt::Display,
        {
            self.tertiary = value
                .try_into()
                .map_err(|e| format!("error converting supplied value for tertiary: {}", e));
            self
        }
        pub fn warning<T>(mut self, value: T) -> Self
        where
            T: ::std::convert::TryInto<::std::option::Option<::std::string::String>>,
            T::Error: ::std::fmt::Display,
        {
            self.warning = value
                .try_into()
                .map_err(|e| format!("error converting supplied value for warning: {}", e));
            self
        }
        pub fn white<T>(mut self, value: T) -> Self
        where
            T: ::std::convert::TryInto<::std::option::Option<::std::string::String>>,
            T::Error: ::std::fmt::Display,
        {
            self.white = value
                .try_into()
                .map_err(|e| format!("error converting supplied value for white: {}", e));
            self
        }
    }
    impl ::std::convert::TryFrom<ColorPalette> for super::ColorPalette {
        type Error = super::error::ConversionError;
        fn try_from(
            value: ColorPalette,
        ) -> ::std::result::Result<Self, super::error::ConversionError> {
            Ok(Self {
                black: value.black?,
                body: value.body?,
                border: value.border?,
                code_style: value.code_style?,
                emphasis: value.emphasis?,
                error: value.error?,
                gray: value.gray?,
                info: value.info?,
                primary: value.primary?,
                secondary: value.secondary?,
                success: value.success?,
                tertiary: value.tertiary?,
                warning: value.warning?,
                white: value.white?,
            })
        }
    }
    impl ::std::convert::From<super::ColorPalette> for ColorPalette {
        fn from(value: super::ColorPalette) -> Self {
            Self {
                black: Ok(value.black),
                body: Ok(value.body),
                border: Ok(value.border),
                code_style: Ok(value.code_style),
                emphasis: Ok(value.emphasis),
                error: Ok(value.error),
                gray: Ok(value.gray),
                info: Ok(value.info),
                primary: Ok(value.primary),
                secondary: Ok(value.secondary),
                success: Ok(value.success),
                tertiary: Ok(value.tertiary),
                warning: Ok(value.warning),
                white: Ok(value.white),
            }
        }
    }
    #[derive(Clone, Debug)]
    pub struct Config {
        color_override: ::std::result::Result<
            ::std::option::Option<super::ColorPalette>,
            ::std::string::String,
        >,
        current_namespace: ::std::result::Result<::std::string::String, ::std::string::String>,
        current_vault: ::std::result::Result<
            ::std::option::Option<::std::string::String>,
            ::std::string::String,
        >,
        current_workspace: ::std::result::Result<::std::string::String, ::std::string::String>,
        default_log_mode: ::std::result::Result<::std::string::String, ::std::string::String>,
        default_timeout: ::std::result::Result<::std::string::String, ::std::string::String>,
        interactive:
            ::std::result::Result<::std::option::Option<super::Interactive>, ::std::string::String>,
        templates: ::std::result::Result<
            ::std::collections::HashMap<::std::string::String, ::std::string::String>,
            ::std::string::String,
        >,
        theme: ::std::result::Result<super::ConfigTheme, ::std::string::String>,
        vaults: ::std::result::Result<
            ::std::collections::HashMap<::std::string::String, ::std::string::String>,
            ::std::string::String,
        >,
        workspace_mode: ::std::result::Result<super::ConfigWorkspaceMode, ::std::string::String>,
        workspaces: ::std::result::Result<
            ::std::collections::HashMap<::std::string::String, ::std::string::String>,
            ::std::string::String,
        >,
    }
    impl ::std::default::Default for Config {
        fn default() -> Self {
            Self {
                color_override: Ok(Default::default()),
                current_namespace: Ok(Default::default()),
                current_vault: Ok(Default::default()),
                current_workspace: Err("no value supplied for current_workspace".to_string()),
                default_log_mode: Ok(super::defaults::config_default_log_mode()),
                default_timeout: Ok(super::defaults::config_default_timeout()),
                interactive: Ok(Default::default()),
                templates: Ok(Default::default()),
                theme: Ok(super::defaults::config_theme()),
                vaults: Ok(Default::default()),
                workspace_mode: Ok(super::defaults::config_workspace_mode()),
                workspaces: Err("no value supplied for workspaces".to_string()),
            }
        }
    }
    impl Config {
        pub fn color_override<T>(mut self, value: T) -> Self
        where
            T: ::std::convert::TryInto<::std::option::Option<super::ColorPalette>>,
            T::Error: ::std::fmt::Display,
        {
            self.color_override = value
                .try_into()
                .map_err(|e| format!("error converting supplied value for color_override: {}", e));
            self
        }
        pub fn current_namespace<T>(mut self, value: T) -> Self
        where
            T: ::std::convert::TryInto<::std::string::String>,
            T::Error: ::std::fmt::Display,
        {
            self.current_namespace = value.try_into().map_err(|e| {
                format!(
                    "error converting supplied value for current_namespace: {}",
                    e
                )
            });
            self
        }
        pub fn current_vault<T>(mut self, value: T) -> Self
        where
            T: ::std::convert::TryInto<::std::option::Option<::std::string::String>>,
            T::Error: ::std::fmt::Display,
        {
            self.current_vault = value
                .try_into()
                .map_err(|e| format!("error converting supplied value for current_vault: {}", e));
            self
        }
        pub fn current_workspace<T>(mut self, value: T) -> Self
        where
            T: ::std::convert::TryInto<::std::string::String>,
            T::Error: ::std::fmt::Display,
        {
            self.current_workspace = value.try_into().map_err(|e| {
                format!(
                    "error converting supplied value for current_workspace: {}",
                    e
                )
            });
            self
        }
        pub fn default_log_mode<T>(mut self, value: T) -> Self
        where
            T: ::std::convert::TryInto<::std::string::String>,
            T::Error: ::std::fmt::Display,
        {
            self.default_log_mode = value.try_into().map_err(|e| {
                format!(
                    "error converting supplied value for default_log_mode: {}",
                    e
                )
            });
            self
        }
        pub fn default_timeout<T>(mut self, value: T) -> Self
        where
            T: ::std::convert::TryInto<::std::string::String>,
            T::Error: ::std::fmt::Display,
        {
            self.default_timeout = value
                .try_into()
                .map_err(|e| format!("error converting supplied value for default_timeout: {}", e));
            self
        }
        pub fn interactive<T>(mut self, value: T) -> Self
        where
            T: ::std::convert::TryInto<::std::option::Option<super::Interactive>>,
            T::Error: ::std::fmt::Display,
        {
            self.interactive = value
                .try_into()
                .map_err(|e| format!("error converting supplied value for interactive: {}", e));
            self
        }
        pub fn templates<T>(mut self, value: T) -> Self
        where
            T: ::std::convert::TryInto<
                ::std::collections::HashMap<::std::string::String, ::std::string::String>,
            >,
            T::Error: ::std::fmt::Display,
        {
            self.templates = value
                .try_into()
                .map_err(|e| format!("error converting supplied value for templates: {}", e));
            self
        }
        pub fn theme<T>(mut self, value: T) -> Self
        where
            T: ::std::convert::TryInto<super::ConfigTheme>,
            T::Error: ::std::fmt::Display,
        {
            self.theme = value
                .try_into()
                .map_err(|e| format!("error converting supplied value for theme: {}", e));
            self
        }
        pub fn vaults<T>(mut self, value: T) -> Self
        where
            T: ::std::convert::TryInto<
                ::std::collections::HashMap<::std::string::String, ::std::string::String>,
            >,
            T::Error: ::std::fmt::Display,
        {
            self.vaults = value
                .try_into()
                .map_err(|e| format!("error converting supplied value for vaults: {}", e));
            self
        }
        pub fn workspace_mode<T>(mut self, value: T) -> Self
        where
            T: ::std::convert::TryInto<super::ConfigWorkspaceMode>,
            T::Error: ::std::fmt::Display,
        {
            self.workspace_mode = value
                .try_into()
                .map_err(|e| format!("error converting supplied value for workspace_mode: {}", e));
            self
        }
        pub fn workspaces<T>(mut self, value: T) -> Self
        where
            T: ::std::convert::TryInto<
                ::std::collections::HashMap<::std::string::String, ::std::string::String>,
            >,
            T::Error: ::std::fmt::Display,
        {
            self.workspaces = value
                .try_into()
                .map_err(|e| format!("error converting supplied value for workspaces: {}", e));
            self
        }
    }
    impl ::std::convert::TryFrom<Config> for super::Config {
        type Error = super::error::ConversionError;
        fn try_from(value: Config) -> ::std::result::Result<Self, super::error::ConversionError> {
            Ok(Self {
                color_override: value.color_override?,
                current_namespace: value.current_namespace?,
                current_vault: value.current_vault?,
                current_workspace: value.current_workspace?,
                default_log_mode: value.default_log_mode?,
                default_timeout: value.default_timeout?,
                interactive: value.interactive?,
                templates: value.templates?,
                theme: value.theme?,
                vaults: value.vaults?,
                workspace_mode: value.workspace_mode?,
                workspaces: value.workspaces?,
            })
        }
    }
    impl ::std::convert::From<super::Config> for Config {
        fn from(value: super::Config) -> Self {
            Self {
                color_override: Ok(value.color_override),
                current_namespace: Ok(value.current_namespace),
                current_vault: Ok(value.current_vault),
                current_workspace: Ok(value.current_workspace),
                default_log_mode: Ok(value.default_log_mode),
                default_timeout: Ok(value.default_timeout),
                interactive: Ok(value.interactive),
                templates: Ok(value.templates),
                theme: Ok(value.theme),
                vaults: Ok(value.vaults),
                workspace_mode: Ok(value.workspace_mode),
                workspaces: Ok(value.workspaces),
            }
        }
    }
    #[derive(Clone, Debug)]
    pub struct Interactive {
        enabled: ::std::result::Result<bool, ::std::string::String>,
        notify_on_completion:
            ::std::result::Result<::std::option::Option<bool>, ::std::string::String>,
        sound_on_completion:
            ::std::result::Result<::std::option::Option<bool>, ::std::string::String>,
    }
    impl ::std::default::Default for Interactive {
        fn default() -> Self {
            Self {
                enabled: Err("no value supplied for enabled".to_string()),
                notify_on_completion: Ok(Default::default()),
                sound_on_completion: Ok(Default::default()),
            }
        }
    }
    impl Interactive {
        pub fn enabled<T>(mut self, value: T) -> Self
        where
            T: ::std::convert::TryInto<bool>,
            T::Error: ::std::fmt::Display,
        {
            self.enabled = value
                .try_into()
                .map_err(|e| format!("error converting supplied value for enabled: {}", e));
            self
        }
        pub fn notify_on_completion<T>(mut self, value: T) -> Self
        where
            T: ::std::convert::TryInto<::std::option::Option<bool>>,
            T::Error: ::std::fmt::Display,
        {
            self.notify_on_completion = value.try_into().map_err(|e| {
                format!(
                    "error converting supplied value for notify_on_completion: {}",
                    e
                )
            });
            self
        }
        pub fn sound_on_completion<T>(mut self, value: T) -> Self
        where
            T: ::std::convert::TryInto<::std::option::Option<bool>>,
            T::Error: ::std::fmt::Display,
        {
            self.sound_on_completion = value.try_into().map_err(|e| {
                format!(
                    "error converting supplied value for sound_on_completion: {}",
                    e
                )
            });
            self
        }
    }
    impl ::std::convert::TryFrom<Interactive> for super::Interactive {
        type Error = super::error::ConversionError;
        fn try_from(
            value: Interactive,
        ) -> ::std::result::Result<Self, super::error::ConversionError> {
            Ok(Self {
                enabled: value.enabled?,
                notify_on_completion: value.notify_on_completion?,
                sound_on_completion: value.sound_on_completion?,
            })
        }
    }
    impl ::std::convert::From<super::Interactive> for Interactive {
        fn from(value: super::Interactive) -> Self {
            Self {
                enabled: Ok(value.enabled),
                notify_on_completion: Ok(value.notify_on_completion),
                sound_on_completion: Ok(value.sound_on_completion),
            }
        }
    }
}
#[doc = r" Generation of default values for serde."]
pub mod defaults {
    pub(super) fn config_default_log_mode() -> ::std::string::String {
        "logfmt".to_string()
    }
    pub(super) fn config_default_timeout() -> ::std::string::String {
        "30m".to_string()
    }
    pub(super) fn config_theme() -> super::ConfigTheme {
        super::ConfigTheme::Default
    }
    pub(super) fn config_workspace_mode() -> super::ConfigWorkspaceMode {
        super::ConfigWorkspaceMode::Dynamic
    }
}
