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
#[doc = "A list of tags.\nTags can be used with list commands to filter returned data.\n"]
#[doc = r""]
#[doc = r" <details><summary>JSON schema</summary>"]
#[doc = r""]
#[doc = r" ```json"]
#[doc = "{"]
#[doc = "  \"description\": \"A list of tags.\\nTags can be used with list commands to filter returned data.\\n\","]
#[doc = "  \"type\": \"array\","]
#[doc = "  \"items\": {"]
#[doc = "    \"type\": \"string\""]
#[doc = "  }"]
#[doc = "}"]
#[doc = r" ```"]
#[doc = r" </details>"]
#[derive(:: serde :: Deserialize, :: serde :: Serialize, Clone, Debug)]
#[serde(transparent)]
pub struct CommonTags(pub ::std::vec::Vec<::std::string::String>);
impl ::std::ops::Deref for CommonTags {
    type Target = ::std::vec::Vec<::std::string::String>;
    fn deref(&self) -> &::std::vec::Vec<::std::string::String> {
        &self.0
    }
}
impl ::std::convert::From<CommonTags> for ::std::vec::Vec<::std::string::String> {
    fn from(value: CommonTags) -> Self {
        value.0
    }
}
impl ::std::convert::From<&CommonTags> for CommonTags {
    fn from(value: &CommonTags) -> Self {
        value.clone()
    }
}
impl ::std::convert::From<::std::vec::Vec<::std::string::String>> for CommonTags {
    fn from(value: ::std::vec::Vec<::std::string::String>) -> Self {
        Self(value)
    }
}
#[doc = "`ExecutableFilter`"]
#[doc = r""]
#[doc = r" <details><summary>JSON schema</summary>"]
#[doc = r""]
#[doc = r" ```json"]
#[doc = "{"]
#[doc = "  \"type\": \"object\","]
#[doc = "  \"properties\": {"]
#[doc = "    \"excluded\": {"]
#[doc = "      \"description\": \"A list of directories or file patterns to exclude from the executable search.\\nSupports directory paths (e.g., \\\"node_modules/\\\", \\\"vendor/\\\") and glob patterns for filenames (e.g., \\\"*.js.flow\\\", \\\"*temp*\\\").\\nCommon exclusions like node_modules/, vendor/, third_party/, external/, and *.js.flow are excluded by default.\\n\","]
#[doc = "      \"default\": [],"]
#[doc = "      \"type\": \"array\","]
#[doc = "      \"items\": {"]
#[doc = "        \"type\": \"string\""]
#[doc = "      }"]
#[doc = "    },"]
#[doc = "    \"included\": {"]
#[doc = "      \"description\": \"A list of directories or file patterns to include in the executable search.\\nSupports directory paths (e.g., \\\"src/\\\", \\\"scripts/\\\") and glob patterns for filenames (e.g., \\\"*.test.flow\\\", \\\"example*\\\").\\n\","]
#[doc = "      \"default\": [],"]
#[doc = "      \"type\": \"array\","]
#[doc = "      \"items\": {"]
#[doc = "        \"type\": \"string\""]
#[doc = "      }"]
#[doc = "    }"]
#[doc = "  }"]
#[doc = "}"]
#[doc = r" ```"]
#[doc = r" </details>"]
#[derive(:: serde :: Deserialize, :: serde :: Serialize, Clone, Debug)]
pub struct ExecutableFilter {
    #[doc = "A list of directories or file patterns to exclude from the executable search.\nSupports directory paths (e.g., \"node_modules/\", \"vendor/\") and glob patterns for filenames (e.g., \"*.js.flow\", \"*temp*\").\nCommon exclusions like node_modules/, vendor/, third_party/, external/, and *.js.flow are excluded by default.\n"]
    #[serde(default, skip_serializing_if = "::std::vec::Vec::is_empty")]
    pub excluded: ::std::vec::Vec<::std::string::String>,
    #[doc = "A list of directories or file patterns to include in the executable search.\nSupports directory paths (e.g., \"src/\", \"scripts/\") and glob patterns for filenames (e.g., \"*.test.flow\", \"example*\").\n"]
    #[serde(default, skip_serializing_if = "::std::vec::Vec::is_empty")]
    pub included: ::std::vec::Vec<::std::string::String>,
}
impl ::std::convert::From<&ExecutableFilter> for ExecutableFilter {
    fn from(value: &ExecutableFilter) -> Self {
        value.clone()
    }
}
impl ::std::default::Default for ExecutableFilter {
    fn default() -> Self {
        Self {
            excluded: Default::default(),
            included: Default::default(),
        }
    }
}
impl ExecutableFilter {
    pub fn builder() -> builder::ExecutableFilter {
        Default::default()
    }
}
#[doc = "A map of executable verbs to valid aliases. This allows you to use custom aliases for exec commands in the workspace.\nSetting this will override all of the default flow command aliases. The verbs and its mapped aliases must be valid flow verbs.\n\nIf set to an empty object, verb aliases will be disabled.\n"]
#[doc = r""]
#[doc = r" <details><summary>JSON schema</summary>"]
#[doc = r""]
#[doc = r" ```json"]
#[doc = "{"]
#[doc = "  \"description\": \"A map of executable verbs to valid aliases. This allows you to use custom aliases for exec commands in the workspace.\\nSetting this will override all of the default flow command aliases. The verbs and its mapped aliases must be valid flow verbs.\\n\\nIf set to an empty object, verb aliases will be disabled.\\n\","]
#[doc = "  \"type\": \"object\","]
#[doc = "  \"additionalProperties\": {"]
#[doc = "    \"type\": \"array\","]
#[doc = "    \"items\": {"]
#[doc = "      \"type\": \"string\""]
#[doc = "    }"]
#[doc = "  }"]
#[doc = "}"]
#[doc = r" ```"]
#[doc = r" </details>"]
#[derive(:: serde :: Deserialize, :: serde :: Serialize, Clone, Debug)]
#[serde(transparent)]
pub struct VerbAliases(
    pub ::std::collections::HashMap<::std::string::String, ::std::vec::Vec<::std::string::String>>,
);
impl ::std::ops::Deref for VerbAliases {
    type Target =
        ::std::collections::HashMap<::std::string::String, ::std::vec::Vec<::std::string::String>>;
    fn deref(
        &self,
    ) -> &::std::collections::HashMap<::std::string::String, ::std::vec::Vec<::std::string::String>>
    {
        &self.0
    }
}
impl ::std::convert::From<VerbAliases>
    for ::std::collections::HashMap<::std::string::String, ::std::vec::Vec<::std::string::String>>
{
    fn from(value: VerbAliases) -> Self {
        value.0
    }
}
impl ::std::convert::From<&VerbAliases> for VerbAliases {
    fn from(value: &VerbAliases) -> Self {
        value.clone()
    }
}
impl
    ::std::convert::From<
        ::std::collections::HashMap<::std::string::String, ::std::vec::Vec<::std::string::String>>,
    > for VerbAliases
{
    fn from(
        value: ::std::collections::HashMap<
            ::std::string::String,
            ::std::vec::Vec<::std::string::String>,
        >,
    ) -> Self {
        Self(value)
    }
}
#[doc = "Configuration for a workspace in the Flow CLI.\nThis configuration is used to define the settings for a workspace.\nEvery workspace has a workspace config file named `flow.yaml` in the root of the workspace directory.\n"]
#[doc = r""]
#[doc = r" <details><summary>JSON schema</summary>"]
#[doc = r""]
#[doc = r" ```json"]
#[doc = "{"]
#[doc = "  \"$id\": \"https://flowexec.io/schemas/workspace_schema.json\","]
#[doc = "  \"title\": \"Workspace\","]
#[doc = "  \"description\": \"Configuration for a workspace in the Flow CLI.\\nThis configuration is used to define the settings for a workspace.\\nEvery workspace has a workspace config file named `flow.yaml` in the root of the workspace directory.\\n\","]
#[doc = "  \"type\": \"object\","]
#[doc = "  \"properties\": {"]
#[doc = "    \"description\": {"]
#[doc = "      \"description\": \"A description of the workspace. This description is rendered as markdown in the interactive UI.\","]
#[doc = "      \"default\": \"\","]
#[doc = "      \"type\": \"string\""]
#[doc = "    },"]
#[doc = "    \"descriptionFile\": {"]
#[doc = "      \"description\": \"A path to a markdown file that contains the description of the workspace.\","]
#[doc = "      \"default\": \"\","]
#[doc = "      \"type\": \"string\""]
#[doc = "    },"]
#[doc = "    \"displayName\": {"]
#[doc = "      \"description\": \"The display name of the workspace. This is used in the interactive UI.\","]
#[doc = "      \"default\": \"\","]
#[doc = "      \"type\": \"string\""]
#[doc = "    },"]
#[doc = "    \"envFiles\": {"]
#[doc = "      \"description\": \"A list of environment variable files to load for the workspace. These files should contain key-value pairs of environment variables.\\nBy default, the `.env` file in the workspace root is loaded if it exists.\\n\","]
#[doc = "      \"default\": [],"]
#[doc = "      \"type\": \"array\","]
#[doc = "      \"items\": {"]
#[doc = "        \"type\": \"string\""]
#[doc = "      }"]
#[doc = "    },"]
#[doc = "    \"executables\": {"]
#[doc = "      \"$ref\": \"#/definitions/ExecutableFilter\""]
#[doc = "    },"]
#[doc = "    \"tags\": {"]
#[doc = "      \"default\": [],"]
#[doc = "      \"$ref\": \"#/definitions/CommonTags\""]
#[doc = "    },"]
#[doc = "    \"verbAliases\": {"]
#[doc = "      \"$ref\": \"#/definitions/VerbAliases\""]
#[doc = "    }"]
#[doc = "  }"]
#[doc = "}"]
#[doc = r" ```"]
#[doc = r" </details>"]
#[derive(:: serde :: Deserialize, :: serde :: Serialize, Clone, Debug)]
pub struct Workspace {
    #[doc = "A description of the workspace. This description is rendered as markdown in the interactive UI."]
    #[serde(default)]
    pub description: ::std::string::String,
    #[doc = "A path to a markdown file that contains the description of the workspace."]
    #[serde(rename = "descriptionFile", default)]
    pub description_file: ::std::string::String,
    #[doc = "The display name of the workspace. This is used in the interactive UI."]
    #[serde(rename = "displayName", default)]
    pub display_name: ::std::string::String,
    #[doc = "A list of environment variable files to load for the workspace. These files should contain key-value pairs of environment variables.\nBy default, the `.env` file in the workspace root is loaded if it exists.\n"]
    #[serde(
        rename = "envFiles",
        default,
        skip_serializing_if = "::std::vec::Vec::is_empty"
    )]
    pub env_files: ::std::vec::Vec<::std::string::String>,
    #[serde(default, skip_serializing_if = "::std::option::Option::is_none")]
    pub executables: ::std::option::Option<ExecutableFilter>,
    #[serde(default = "defaults::workspace_tags")]
    pub tags: CommonTags,
    #[serde(
        rename = "verbAliases",
        default,
        skip_serializing_if = "::std::option::Option::is_none"
    )]
    pub verb_aliases: ::std::option::Option<VerbAliases>,
}
impl ::std::convert::From<&Workspace> for Workspace {
    fn from(value: &Workspace) -> Self {
        value.clone()
    }
}
impl ::std::default::Default for Workspace {
    fn default() -> Self {
        Self {
            description: Default::default(),
            description_file: Default::default(),
            display_name: Default::default(),
            env_files: Default::default(),
            executables: Default::default(),
            tags: defaults::workspace_tags(),
            verb_aliases: Default::default(),
        }
    }
}
impl Workspace {
    pub fn builder() -> builder::Workspace {
        Default::default()
    }
}
#[doc = r" Types for composing complex structures."]
pub mod builder {
    #[derive(Clone, Debug)]
    pub struct ExecutableFilter {
        excluded:
            ::std::result::Result<::std::vec::Vec<::std::string::String>, ::std::string::String>,
        included:
            ::std::result::Result<::std::vec::Vec<::std::string::String>, ::std::string::String>,
    }
    impl ::std::default::Default for ExecutableFilter {
        fn default() -> Self {
            Self {
                excluded: Ok(Default::default()),
                included: Ok(Default::default()),
            }
        }
    }
    impl ExecutableFilter {
        pub fn excluded<T>(mut self, value: T) -> Self
        where
            T: ::std::convert::TryInto<::std::vec::Vec<::std::string::String>>,
            T::Error: ::std::fmt::Display,
        {
            self.excluded = value
                .try_into()
                .map_err(|e| format!("error converting supplied value for excluded: {}", e));
            self
        }
        pub fn included<T>(mut self, value: T) -> Self
        where
            T: ::std::convert::TryInto<::std::vec::Vec<::std::string::String>>,
            T::Error: ::std::fmt::Display,
        {
            self.included = value
                .try_into()
                .map_err(|e| format!("error converting supplied value for included: {}", e));
            self
        }
    }
    impl ::std::convert::TryFrom<ExecutableFilter> for super::ExecutableFilter {
        type Error = super::error::ConversionError;
        fn try_from(
            value: ExecutableFilter,
        ) -> ::std::result::Result<Self, super::error::ConversionError> {
            Ok(Self {
                excluded: value.excluded?,
                included: value.included?,
            })
        }
    }
    impl ::std::convert::From<super::ExecutableFilter> for ExecutableFilter {
        fn from(value: super::ExecutableFilter) -> Self {
            Self {
                excluded: Ok(value.excluded),
                included: Ok(value.included),
            }
        }
    }
    #[derive(Clone, Debug)]
    pub struct Workspace {
        description: ::std::result::Result<::std::string::String, ::std::string::String>,
        description_file: ::std::result::Result<::std::string::String, ::std::string::String>,
        display_name: ::std::result::Result<::std::string::String, ::std::string::String>,
        env_files:
            ::std::result::Result<::std::vec::Vec<::std::string::String>, ::std::string::String>,
        executables: ::std::result::Result<
            ::std::option::Option<super::ExecutableFilter>,
            ::std::string::String,
        >,
        tags: ::std::result::Result<super::CommonTags, ::std::string::String>,
        verb_aliases:
            ::std::result::Result<::std::option::Option<super::VerbAliases>, ::std::string::String>,
    }
    impl ::std::default::Default for Workspace {
        fn default() -> Self {
            Self {
                description: Ok(Default::default()),
                description_file: Ok(Default::default()),
                display_name: Ok(Default::default()),
                env_files: Ok(Default::default()),
                executables: Ok(Default::default()),
                tags: Ok(super::defaults::workspace_tags()),
                verb_aliases: Ok(Default::default()),
            }
        }
    }
    impl Workspace {
        pub fn description<T>(mut self, value: T) -> Self
        where
            T: ::std::convert::TryInto<::std::string::String>,
            T::Error: ::std::fmt::Display,
        {
            self.description = value
                .try_into()
                .map_err(|e| format!("error converting supplied value for description: {}", e));
            self
        }
        pub fn description_file<T>(mut self, value: T) -> Self
        where
            T: ::std::convert::TryInto<::std::string::String>,
            T::Error: ::std::fmt::Display,
        {
            self.description_file = value.try_into().map_err(|e| {
                format!(
                    "error converting supplied value for description_file: {}",
                    e
                )
            });
            self
        }
        pub fn display_name<T>(mut self, value: T) -> Self
        where
            T: ::std::convert::TryInto<::std::string::String>,
            T::Error: ::std::fmt::Display,
        {
            self.display_name = value
                .try_into()
                .map_err(|e| format!("error converting supplied value for display_name: {}", e));
            self
        }
        pub fn env_files<T>(mut self, value: T) -> Self
        where
            T: ::std::convert::TryInto<::std::vec::Vec<::std::string::String>>,
            T::Error: ::std::fmt::Display,
        {
            self.env_files = value
                .try_into()
                .map_err(|e| format!("error converting supplied value for env_files: {}", e));
            self
        }
        pub fn executables<T>(mut self, value: T) -> Self
        where
            T: ::std::convert::TryInto<::std::option::Option<super::ExecutableFilter>>,
            T::Error: ::std::fmt::Display,
        {
            self.executables = value
                .try_into()
                .map_err(|e| format!("error converting supplied value for executables: {}", e));
            self
        }
        pub fn tags<T>(mut self, value: T) -> Self
        where
            T: ::std::convert::TryInto<super::CommonTags>,
            T::Error: ::std::fmt::Display,
        {
            self.tags = value
                .try_into()
                .map_err(|e| format!("error converting supplied value for tags: {}", e));
            self
        }
        pub fn verb_aliases<T>(mut self, value: T) -> Self
        where
            T: ::std::convert::TryInto<::std::option::Option<super::VerbAliases>>,
            T::Error: ::std::fmt::Display,
        {
            self.verb_aliases = value
                .try_into()
                .map_err(|e| format!("error converting supplied value for verb_aliases: {}", e));
            self
        }
    }
    impl ::std::convert::TryFrom<Workspace> for super::Workspace {
        type Error = super::error::ConversionError;
        fn try_from(
            value: Workspace,
        ) -> ::std::result::Result<Self, super::error::ConversionError> {
            Ok(Self {
                description: value.description?,
                description_file: value.description_file?,
                display_name: value.display_name?,
                env_files: value.env_files?,
                executables: value.executables?,
                tags: value.tags?,
                verb_aliases: value.verb_aliases?,
            })
        }
    }
    impl ::std::convert::From<super::Workspace> for Workspace {
        fn from(value: super::Workspace) -> Self {
            Self {
                description: Ok(value.description),
                description_file: Ok(value.description_file),
                display_name: Ok(value.display_name),
                env_files: Ok(value.env_files),
                executables: Ok(value.executables),
                tags: Ok(value.tags),
                verb_aliases: Ok(value.verb_aliases),
            }
        }
    }
}
#[doc = r" Generation of default values for serde."]
pub mod defaults {
    pub(super) fn workspace_tags() -> super::CommonTags {
        super::CommonTags(vec![])
    }
}
