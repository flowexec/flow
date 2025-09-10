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
#[doc = "Alternate names that can be used to reference the executable in the CLI."]
#[doc = r""]
#[doc = r" <details><summary>JSON schema</summary>"]
#[doc = r""]
#[doc = r" ```json"]
#[doc = "{"]
#[doc = "  \"description\": \"Alternate names that can be used to reference the executable in the CLI.\","]
#[doc = "  \"type\": \"array\","]
#[doc = "  \"items\": {"]
#[doc = "    \"type\": \"string\""]
#[doc = "  }"]
#[doc = "}"]
#[doc = r" ```"]
#[doc = r" </details>"]
#[derive(:: serde :: Deserialize, :: serde :: Serialize, Clone, Debug)]
#[serde(transparent)]
pub struct CommonAliases(pub ::std::vec::Vec<::std::string::String>);
impl ::std::ops::Deref for CommonAliases {
    type Target = ::std::vec::Vec<::std::string::String>;
    fn deref(&self) -> &::std::vec::Vec<::std::string::String> {
        &self.0
    }
}
impl ::std::convert::From<CommonAliases> for ::std::vec::Vec<::std::string::String> {
    fn from(value: CommonAliases) -> Self {
        value.0
    }
}
impl ::std::convert::From<&CommonAliases> for CommonAliases {
    fn from(value: &CommonAliases) -> Self {
        value.clone()
    }
}
impl ::std::convert::From<::std::vec::Vec<::std::string::String>> for CommonAliases {
    fn from(value: ::std::vec::Vec<::std::string::String>) -> Self {
        Self(value)
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
#[doc = "The visibility of the executables to Flow.\nIf not set, the visibility will default to `public`.\n\n`public` executables can be executed and listed from anywhere.\n`private` executables can be executed and listed only within their own workspace.\n`internal` executables can be executed within their own workspace but are not listed.\n`hidden` executables cannot be executed or listed.\n"]
#[doc = r""]
#[doc = r" <details><summary>JSON schema</summary>"]
#[doc = r""]
#[doc = r" ```json"]
#[doc = "{"]
#[doc = "  \"description\": \"The visibility of the executables to Flow.\\nIf not set, the visibility will default to `public`.\\n\\n`public` executables can be executed and listed from anywhere.\\n`private` executables can be executed and listed only within their own workspace.\\n`internal` executables can be executed within their own workspace but are not listed.\\n`hidden` executables cannot be executed or listed.\\n\","]
#[doc = "  \"default\": \"public\","]
#[doc = "  \"type\": \"string\","]
#[doc = "  \"enum\": ["]
#[doc = "    \"public\","]
#[doc = "    \"private\","]
#[doc = "    \"internal\","]
#[doc = "    \"hidden\""]
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
pub enum CommonVisibility {
    #[serde(rename = "public")]
    Public,
    #[serde(rename = "private")]
    Private,
    #[serde(rename = "internal")]
    Internal,
    #[serde(rename = "hidden")]
    Hidden,
}
impl ::std::convert::From<&Self> for CommonVisibility {
    fn from(value: &CommonVisibility) -> Self {
        value.clone()
    }
}
impl ::std::fmt::Display for CommonVisibility {
    fn fmt(&self, f: &mut ::std::fmt::Formatter<'_>) -> ::std::fmt::Result {
        match *self {
            Self::Public => f.write_str("public"),
            Self::Private => f.write_str("private"),
            Self::Internal => f.write_str("internal"),
            Self::Hidden => f.write_str("hidden"),
        }
    }
}
impl ::std::str::FromStr for CommonVisibility {
    type Err = self::error::ConversionError;
    fn from_str(value: &str) -> ::std::result::Result<Self, self::error::ConversionError> {
        match value {
            "public" => Ok(Self::Public),
            "private" => Ok(Self::Private),
            "internal" => Ok(Self::Internal),
            "hidden" => Ok(Self::Hidden),
            _ => Err("invalid value".into()),
        }
    }
}
impl ::std::convert::TryFrom<&str> for CommonVisibility {
    type Error = self::error::ConversionError;
    fn try_from(value: &str) -> ::std::result::Result<Self, self::error::ConversionError> {
        value.parse()
    }
}
impl ::std::convert::TryFrom<&::std::string::String> for CommonVisibility {
    type Error = self::error::ConversionError;
    fn try_from(
        value: &::std::string::String,
    ) -> ::std::result::Result<Self, self::error::ConversionError> {
        value.parse()
    }
}
impl ::std::convert::TryFrom<::std::string::String> for CommonVisibility {
    type Error = self::error::ConversionError;
    fn try_from(
        value: ::std::string::String,
    ) -> ::std::result::Result<Self, self::error::ConversionError> {
        value.parse()
    }
}
impl ::std::default::Default for CommonVisibility {
    fn default() -> Self {
        CommonVisibility::Public
    }
}
#[doc = "The executable schema defines the structure of an executable in the Flow CLI.\nExecutables are the building blocks of workflows and are used to define the actions that can be performed in a workspace.\n"]
#[doc = r""]
#[doc = r" <details><summary>JSON schema</summary>"]
#[doc = r""]
#[doc = r" ```json"]
#[doc = "{"]
#[doc = "  \"title\": \"Executable\","]
#[doc = "  \"description\": \"The executable schema defines the structure of an executable in the Flow CLI.\\nExecutables are the building blocks of workflows and are used to define the actions that can be performed in a workspace.\\n\","]
#[doc = "  \"type\": \"object\","]
#[doc = "  \"required\": ["]
#[doc = "    \"verb\""]
#[doc = "  ],"]
#[doc = "  \"properties\": {"]
#[doc = "    \"aliases\": {"]
#[doc = "      \"default\": [],"]
#[doc = "      \"$ref\": \"#/definitions/CommonAliases\""]
#[doc = "    },"]
#[doc = "    \"description\": {"]
#[doc = "      \"description\": \"A description of the executable.\\nThis description is rendered as markdown in the interactive UI.\\n\","]
#[doc = "      \"default\": \"\","]
#[doc = "      \"type\": \"string\""]
#[doc = "    },"]
#[doc = "    \"exec\": {"]
#[doc = "      \"$ref\": \"#/definitions/ExecutableExecExecutableType\""]
#[doc = "    },"]
#[doc = "    \"launch\": {"]
#[doc = "      \"$ref\": \"#/definitions/ExecutableLaunchExecutableType\""]
#[doc = "    },"]
#[doc = "    \"name\": {"]
#[doc = "      \"description\": \"An optional name for the executable.\\n\\nName is used to reference the executable in the CLI using the format `workspace/namespace:name`.\\n[Verb group + Name] must be unique within the namespace of the workspace.\\n\","]
#[doc = "      \"default\": \"\","]
#[doc = "      \"type\": \"string\""]
#[doc = "    },"]
#[doc = "    \"parallel\": {"]
#[doc = "      \"$ref\": \"#/definitions/ExecutableParallelExecutableType\""]
#[doc = "    },"]
#[doc = "    \"render\": {"]
#[doc = "      \"$ref\": \"#/definitions/ExecutableRenderExecutableType\""]
#[doc = "    },"]
#[doc = "    \"request\": {"]
#[doc = "      \"$ref\": \"#/definitions/ExecutableRequestExecutableType\""]
#[doc = "    },"]
#[doc = "    \"serial\": {"]
#[doc = "      \"$ref\": \"#/definitions/ExecutableSerialExecutableType\""]
#[doc = "    },"]
#[doc = "    \"tags\": {"]
#[doc = "      \"default\": [],"]
#[doc = "      \"$ref\": \"#/definitions/CommonTags\""]
#[doc = "    },"]
#[doc = "    \"timeout\": {"]
#[doc = "      \"description\": \"The maximum amount of time the executable is allowed to run before being terminated.\\nThe timeout is specified in Go duration format (e.g. 30s, 5m, 1h).\\n\","]
#[doc = "      \"type\": \"string\""]
#[doc = "    },"]
#[doc = "    \"verb\": {"]
#[doc = "      \"default\": \"exec\","]
#[doc = "      \"$ref\": \"#/definitions/ExecutableVerb\""]
#[doc = "    },"]
#[doc = "    \"verbAliases\": {"]
#[doc = "      \"description\": \"A list of aliases for the verb. This allows the executable to be referenced with multiple verbs.\","]
#[doc = "      \"default\": [],"]
#[doc = "      \"type\": \"array\","]
#[doc = "      \"items\": {"]
#[doc = "        \"$ref\": \"#/definitions/Verb\""]
#[doc = "      }"]
#[doc = "    },"]
#[doc = "    \"visibility\": {"]
#[doc = "      \"$ref\": \"#/definitions/CommonVisibility\""]
#[doc = "    }"]
#[doc = "  }"]
#[doc = "}"]
#[doc = r" ```"]
#[doc = r" </details>"]
#[derive(:: serde :: Deserialize, :: serde :: Serialize, Clone, Debug)]
pub struct Executable {
    #[serde(default = "defaults::executable_aliases")]
    pub aliases: CommonAliases,
    #[doc = "A description of the executable.\nThis description is rendered as markdown in the interactive UI.\n"]
    #[serde(default)]
    pub description: ::std::string::String,
    #[serde(default, skip_serializing_if = "::std::option::Option::is_none")]
    pub exec: ::std::option::Option<ExecutableExecExecutableType>,
    #[serde(default, skip_serializing_if = "::std::option::Option::is_none")]
    pub launch: ::std::option::Option<ExecutableLaunchExecutableType>,
    #[doc = "An optional name for the executable.\n\nName is used to reference the executable in the CLI using the format `workspace/namespace:name`.\n[Verb group + Name] must be unique within the namespace of the workspace.\n"]
    #[serde(default)]
    pub name: ::std::string::String,
    #[serde(default, skip_serializing_if = "::std::option::Option::is_none")]
    pub parallel: ::std::option::Option<ExecutableParallelExecutableType>,
    #[serde(default, skip_serializing_if = "::std::option::Option::is_none")]
    pub render: ::std::option::Option<ExecutableRenderExecutableType>,
    #[serde(default, skip_serializing_if = "::std::option::Option::is_none")]
    pub request: ::std::option::Option<ExecutableRequestExecutableType>,
    #[serde(default, skip_serializing_if = "::std::option::Option::is_none")]
    pub serial: ::std::option::Option<ExecutableSerialExecutableType>,
    #[serde(default = "defaults::executable_tags")]
    pub tags: CommonTags,
    #[doc = "The maximum amount of time the executable is allowed to run before being terminated.\nThe timeout is specified in Go duration format (e.g. 30s, 5m, 1h).\n"]
    #[serde(default, skip_serializing_if = "::std::option::Option::is_none")]
    pub timeout: ::std::option::Option<::std::string::String>,
    pub verb: ExecutableVerb,
    #[doc = "A list of aliases for the verb. This allows the executable to be referenced with multiple verbs."]
    #[serde(
        rename = "verbAliases",
        default,
        skip_serializing_if = "::std::vec::Vec::is_empty"
    )]
    pub verb_aliases: ::std::vec::Vec<Verb>,
    #[serde(default, skip_serializing_if = "::std::option::Option::is_none")]
    pub visibility: ::std::option::Option<CommonVisibility>,
}
impl ::std::convert::From<&Executable> for Executable {
    fn from(value: &Executable) -> Self {
        value.clone()
    }
}
impl Executable {
    pub fn builder() -> builder::Executable {
        Default::default()
    }
}
#[doc = "`ExecutableArgument`"]
#[doc = r""]
#[doc = r" <details><summary>JSON schema</summary>"]
#[doc = r""]
#[doc = r" ```json"]
#[doc = "{"]
#[doc = "  \"type\": \"object\","]
#[doc = "  \"properties\": {"]
#[doc = "    \"default\": {"]
#[doc = "      \"description\": \"The default value to use if the argument is not provided.\\nIf the argument is required and no default is provided, the executable will fail.\\n\","]
#[doc = "      \"default\": \"\","]
#[doc = "      \"type\": \"string\""]
#[doc = "    },"]
#[doc = "    \"envKey\": {"]
#[doc = "      \"description\": \"The name of the environment variable that will be assigned the value.\","]
#[doc = "      \"default\": \"\","]
#[doc = "      \"type\": \"string\""]
#[doc = "    },"]
#[doc = "    \"flag\": {"]
#[doc = "      \"description\": \"The flag to use when setting the argument from the command line.\\nEither `flag` or `pos` must be set, but not both.\\n\","]
#[doc = "      \"default\": \"\","]
#[doc = "      \"type\": \"string\""]
#[doc = "    },"]
#[doc = "    \"outputFile\": {"]
#[doc = "      \"description\": \"A path where the argument value will be temporarily written to disk.\\nThe file will be created before execution and cleaned up afterwards.\\n\","]
#[doc = "      \"default\": \"\","]
#[doc = "      \"type\": \"string\""]
#[doc = "    },"]
#[doc = "    \"pos\": {"]
#[doc = "      \"description\": \"The position of the argument in the command line ArgumentList. Values start at 1.\\nEither `flag` or `pos` must be set, but not both.\\n\","]
#[doc = "      \"type\": \"integer\""]
#[doc = "    },"]
#[doc = "    \"required\": {"]
#[doc = "      \"description\": \"If the argument is required, the executable will fail if the argument is not provided.\\nIf the argument is not required, the default value will be used if the argument is not provided.\\n\","]
#[doc = "      \"default\": false,"]
#[doc = "      \"type\": \"boolean\""]
#[doc = "    },"]
#[doc = "    \"type\": {"]
#[doc = "      \"description\": \"The type of the argument. This is used to determine how to parse the value of the argument.\","]
#[doc = "      \"default\": \"string\","]
#[doc = "      \"type\": \"string\","]
#[doc = "      \"enum\": ["]
#[doc = "        \"string\","]
#[doc = "        \"int\","]
#[doc = "        \"float\","]
#[doc = "        \"bool\""]
#[doc = "      ]"]
#[doc = "    }"]
#[doc = "  }"]
#[doc = "}"]
#[doc = r" ```"]
#[doc = r" </details>"]
#[derive(:: serde :: Deserialize, :: serde :: Serialize, Clone, Debug)]
pub struct ExecutableArgument {
    #[doc = "The default value to use if the argument is not provided.\nIf the argument is required and no default is provided, the executable will fail.\n"]
    #[serde(default)]
    pub default: ::std::string::String,
    #[doc = "The name of the environment variable that will be assigned the value."]
    #[serde(rename = "envKey", default)]
    pub env_key: ::std::string::String,
    #[doc = "The flag to use when setting the argument from the command line.\nEither `flag` or `pos` must be set, but not both.\n"]
    #[serde(default)]
    pub flag: ::std::string::String,
    #[doc = "A path where the argument value will be temporarily written to disk.\nThe file will be created before execution and cleaned up afterwards.\n"]
    #[serde(rename = "outputFile", default)]
    pub output_file: ::std::string::String,
    #[doc = "The position of the argument in the command line ArgumentList. Values start at 1.\nEither `flag` or `pos` must be set, but not both.\n"]
    #[serde(default, skip_serializing_if = "::std::option::Option::is_none")]
    pub pos: ::std::option::Option<i64>,
    #[doc = "If the argument is required, the executable will fail if the argument is not provided.\nIf the argument is not required, the default value will be used if the argument is not provided.\n"]
    #[serde(default)]
    pub required: bool,
    #[doc = "The type of the argument. This is used to determine how to parse the value of the argument."]
    #[serde(rename = "type", default = "defaults::executable_argument_type")]
    pub type_: ExecutableArgumentType,
}
impl ::std::convert::From<&ExecutableArgument> for ExecutableArgument {
    fn from(value: &ExecutableArgument) -> Self {
        value.clone()
    }
}
impl ::std::default::Default for ExecutableArgument {
    fn default() -> Self {
        Self {
            default: Default::default(),
            env_key: Default::default(),
            flag: Default::default(),
            output_file: Default::default(),
            pos: Default::default(),
            required: Default::default(),
            type_: defaults::executable_argument_type(),
        }
    }
}
impl ExecutableArgument {
    pub fn builder() -> builder::ExecutableArgument {
        Default::default()
    }
}
#[doc = "`ExecutableArgumentList`"]
#[doc = r""]
#[doc = r" <details><summary>JSON schema</summary>"]
#[doc = r""]
#[doc = r" ```json"]
#[doc = "{"]
#[doc = "  \"type\": \"array\","]
#[doc = "  \"items\": {"]
#[doc = "    \"$ref\": \"#/definitions/ExecutableArgument\""]
#[doc = "  }"]
#[doc = "}"]
#[doc = r" ```"]
#[doc = r" </details>"]
#[derive(:: serde :: Deserialize, :: serde :: Serialize, Clone, Debug)]
#[serde(transparent)]
pub struct ExecutableArgumentList(pub ::std::vec::Vec<ExecutableArgument>);
impl ::std::ops::Deref for ExecutableArgumentList {
    type Target = ::std::vec::Vec<ExecutableArgument>;
    fn deref(&self) -> &::std::vec::Vec<ExecutableArgument> {
        &self.0
    }
}
impl ::std::convert::From<ExecutableArgumentList> for ::std::vec::Vec<ExecutableArgument> {
    fn from(value: ExecutableArgumentList) -> Self {
        value.0
    }
}
impl ::std::convert::From<&ExecutableArgumentList> for ExecutableArgumentList {
    fn from(value: &ExecutableArgumentList) -> Self {
        value.clone()
    }
}
impl ::std::convert::From<::std::vec::Vec<ExecutableArgument>> for ExecutableArgumentList {
    fn from(value: ::std::vec::Vec<ExecutableArgument>) -> Self {
        Self(value)
    }
}
#[doc = "The type of the argument. This is used to determine how to parse the value of the argument."]
#[doc = r""]
#[doc = r" <details><summary>JSON schema</summary>"]
#[doc = r""]
#[doc = r" ```json"]
#[doc = "{"]
#[doc = "  \"description\": \"The type of the argument. This is used to determine how to parse the value of the argument.\","]
#[doc = "  \"default\": \"string\","]
#[doc = "  \"type\": \"string\","]
#[doc = "  \"enum\": ["]
#[doc = "    \"string\","]
#[doc = "    \"int\","]
#[doc = "    \"float\","]
#[doc = "    \"bool\""]
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
pub enum ExecutableArgumentType {
    #[serde(rename = "string")]
    String,
    #[serde(rename = "int")]
    Int,
    #[serde(rename = "float")]
    Float,
    #[serde(rename = "bool")]
    Bool,
}
impl ::std::convert::From<&Self> for ExecutableArgumentType {
    fn from(value: &ExecutableArgumentType) -> Self {
        value.clone()
    }
}
impl ::std::fmt::Display for ExecutableArgumentType {
    fn fmt(&self, f: &mut ::std::fmt::Formatter<'_>) -> ::std::fmt::Result {
        match *self {
            Self::String => f.write_str("string"),
            Self::Int => f.write_str("int"),
            Self::Float => f.write_str("float"),
            Self::Bool => f.write_str("bool"),
        }
    }
}
impl ::std::str::FromStr for ExecutableArgumentType {
    type Err = self::error::ConversionError;
    fn from_str(value: &str) -> ::std::result::Result<Self, self::error::ConversionError> {
        match value {
            "string" => Ok(Self::String),
            "int" => Ok(Self::Int),
            "float" => Ok(Self::Float),
            "bool" => Ok(Self::Bool),
            _ => Err("invalid value".into()),
        }
    }
}
impl ::std::convert::TryFrom<&str> for ExecutableArgumentType {
    type Error = self::error::ConversionError;
    fn try_from(value: &str) -> ::std::result::Result<Self, self::error::ConversionError> {
        value.parse()
    }
}
impl ::std::convert::TryFrom<&::std::string::String> for ExecutableArgumentType {
    type Error = self::error::ConversionError;
    fn try_from(
        value: &::std::string::String,
    ) -> ::std::result::Result<Self, self::error::ConversionError> {
        value.parse()
    }
}
impl ::std::convert::TryFrom<::std::string::String> for ExecutableArgumentType {
    type Error = self::error::ConversionError;
    fn try_from(
        value: ::std::string::String,
    ) -> ::std::result::Result<Self, self::error::ConversionError> {
        value.parse()
    }
}
impl ::std::default::Default for ExecutableArgumentType {
    fn default() -> Self {
        ExecutableArgumentType::String
    }
}
#[doc = "The directory to execute the command in.\nIf unset, the directory of the flow file will be used.\nIf set to `f:tmp`, a temporary directory will be created for the process.\nIf prefixed with `./`, the path will be relative to the current working directory.\nIf prefixed with `//`, the path will be relative to the workspace root.\nEnvironment variables in the path will be expended at runtime.\n"]
#[doc = r""]
#[doc = r" <details><summary>JSON schema</summary>"]
#[doc = r""]
#[doc = r" ```json"]
#[doc = "{"]
#[doc = "  \"description\": \"The directory to execute the command in.\\nIf unset, the directory of the flow file will be used.\\nIf set to `f:tmp`, a temporary directory will be created for the process.\\nIf prefixed with `./`, the path will be relative to the current working directory.\\nIf prefixed with `//`, the path will be relative to the workspace root.\\nEnvironment variables in the path will be expended at runtime.\\n\","]
#[doc = "  \"default\": \"\","]
#[doc = "  \"type\": \"string\""]
#[doc = "}"]
#[doc = r" ```"]
#[doc = r" </details>"]
#[derive(
    :: serde :: Deserialize,
    :: serde :: Serialize,
    Clone,
    Debug,
    Eq,
    Hash,
    Ord,
    PartialEq,
    PartialOrd,
)]
#[serde(transparent)]
pub struct ExecutableDirectory(pub ::std::string::String);
impl ::std::ops::Deref for ExecutableDirectory {
    type Target = ::std::string::String;
    fn deref(&self) -> &::std::string::String {
        &self.0
    }
}
impl ::std::convert::From<ExecutableDirectory> for ::std::string::String {
    fn from(value: ExecutableDirectory) -> Self {
        value.0
    }
}
impl ::std::convert::From<&ExecutableDirectory> for ExecutableDirectory {
    fn from(value: &ExecutableDirectory) -> Self {
        value.clone()
    }
}
impl ::std::convert::From<::std::string::String> for ExecutableDirectory {
    fn from(value: ::std::string::String) -> Self {
        Self(value)
    }
}
impl ::std::str::FromStr for ExecutableDirectory {
    type Err = ::std::convert::Infallible;
    fn from_str(value: &str) -> ::std::result::Result<Self, Self::Err> {
        Ok(Self(value.to_string()))
    }
}
impl ::std::fmt::Display for ExecutableDirectory {
    fn fmt(&self, f: &mut ::std::fmt::Formatter<'_>) -> ::std::fmt::Result {
        self.0.fmt(f)
    }
}
#[doc = "Standard executable type. Runs a command/file in a subprocess."]
#[doc = r""]
#[doc = r" <details><summary>JSON schema</summary>"]
#[doc = r""]
#[doc = r" ```json"]
#[doc = "{"]
#[doc = "  \"description\": \"Standard executable type. Runs a command/file in a subprocess.\","]
#[doc = "  \"type\": \"object\","]
#[doc = "  \"properties\": {"]
#[doc = "    \"args\": {"]
#[doc = "      \"$ref\": \"#/definitions/ExecutableArgumentList\""]
#[doc = "    },"]
#[doc = "    \"cmd\": {"]
#[doc = "      \"description\": \"The command to execute.\\nOnly one of `cmd` or `file` must be set.\\n\","]
#[doc = "      \"default\": \"\","]
#[doc = "      \"type\": \"string\""]
#[doc = "    },"]
#[doc = "    \"dir\": {"]
#[doc = "      \"default\": \"\","]
#[doc = "      \"$ref\": \"#/definitions/ExecutableDirectory\""]
#[doc = "    },"]
#[doc = "    \"file\": {"]
#[doc = "      \"description\": \"The file to execute.\\nOnly one of `cmd` or `file` must be set.\\n\","]
#[doc = "      \"default\": \"\","]
#[doc = "      \"type\": \"string\""]
#[doc = "    },"]
#[doc = "    \"logMode\": {"]
#[doc = "      \"description\": \"The log mode to use when running the executable.\\nThis can either be `hidden`, `json`, `logfmt` or `text`\\n\","]
#[doc = "      \"default\": \"logfmt\","]
#[doc = "      \"type\": \"string\""]
#[doc = "    },"]
#[doc = "    \"params\": {"]
#[doc = "      \"$ref\": \"#/definitions/ExecutableParameterList\""]
#[doc = "    }"]
#[doc = "  }"]
#[doc = "}"]
#[doc = r" ```"]
#[doc = r" </details>"]
#[derive(:: serde :: Deserialize, :: serde :: Serialize, Clone, Debug)]
pub struct ExecutableExecExecutableType {
    #[serde(default, skip_serializing_if = "::std::option::Option::is_none")]
    pub args: ::std::option::Option<ExecutableArgumentList>,
    #[doc = "The command to execute.\nOnly one of `cmd` or `file` must be set.\n"]
    #[serde(default)]
    pub cmd: ::std::string::String,
    #[serde(default = "defaults::executable_exec_executable_type_dir")]
    pub dir: ExecutableDirectory,
    #[doc = "The file to execute.\nOnly one of `cmd` or `file` must be set.\n"]
    #[serde(default)]
    pub file: ::std::string::String,
    #[doc = "The log mode to use when running the executable.\nThis can either be `hidden`, `json`, `logfmt` or `text`\n"]
    #[serde(
        rename = "logMode",
        default = "defaults::executable_exec_executable_type_log_mode"
    )]
    pub log_mode: ::std::string::String,
    #[serde(default, skip_serializing_if = "::std::option::Option::is_none")]
    pub params: ::std::option::Option<ExecutableParameterList>,
}
impl ::std::convert::From<&ExecutableExecExecutableType> for ExecutableExecExecutableType {
    fn from(value: &ExecutableExecExecutableType) -> Self {
        value.clone()
    }
}
impl ::std::default::Default for ExecutableExecExecutableType {
    fn default() -> Self {
        Self {
            args: Default::default(),
            cmd: Default::default(),
            dir: defaults::executable_exec_executable_type_dir(),
            file: Default::default(),
            log_mode: defaults::executable_exec_executable_type_log_mode(),
            params: Default::default(),
        }
    }
}
impl ExecutableExecExecutableType {
    pub fn builder() -> builder::ExecutableExecExecutableType {
        Default::default()
    }
}
#[doc = "Launches an application or opens a URI."]
#[doc = r""]
#[doc = r" <details><summary>JSON schema</summary>"]
#[doc = r""]
#[doc = r" ```json"]
#[doc = "{"]
#[doc = "  \"description\": \"Launches an application or opens a URI.\","]
#[doc = "  \"type\": \"object\","]
#[doc = "  \"required\": ["]
#[doc = "    \"uri\""]
#[doc = "  ],"]
#[doc = "  \"properties\": {"]
#[doc = "    \"app\": {"]
#[doc = "      \"description\": \"The application to launch the URI with.\","]
#[doc = "      \"default\": \"\","]
#[doc = "      \"type\": \"string\""]
#[doc = "    },"]
#[doc = "    \"args\": {"]
#[doc = "      \"$ref\": \"#/definitions/ExecutableArgumentList\""]
#[doc = "    },"]
#[doc = "    \"params\": {"]
#[doc = "      \"$ref\": \"#/definitions/ExecutableParameterList\""]
#[doc = "    },"]
#[doc = "    \"uri\": {"]
#[doc = "      \"description\": \"The URI to launch. This can be a file path or a web URL.\","]
#[doc = "      \"default\": \"\","]
#[doc = "      \"type\": \"string\""]
#[doc = "    }"]
#[doc = "  }"]
#[doc = "}"]
#[doc = r" ```"]
#[doc = r" </details>"]
#[derive(:: serde :: Deserialize, :: serde :: Serialize, Clone, Debug)]
pub struct ExecutableLaunchExecutableType {
    #[doc = "The application to launch the URI with."]
    #[serde(default)]
    pub app: ::std::string::String,
    #[serde(default, skip_serializing_if = "::std::option::Option::is_none")]
    pub args: ::std::option::Option<ExecutableArgumentList>,
    #[serde(default, skip_serializing_if = "::std::option::Option::is_none")]
    pub params: ::std::option::Option<ExecutableParameterList>,
    #[doc = "The URI to launch. This can be a file path or a web URL."]
    pub uri: ::std::string::String,
}
impl ::std::convert::From<&ExecutableLaunchExecutableType> for ExecutableLaunchExecutableType {
    fn from(value: &ExecutableLaunchExecutableType) -> Self {
        value.clone()
    }
}
impl ExecutableLaunchExecutableType {
    pub fn builder() -> builder::ExecutableLaunchExecutableType {
        Default::default()
    }
}
#[doc = "`ExecutableParallelExecutableType`"]
#[doc = r""]
#[doc = r" <details><summary>JSON schema</summary>"]
#[doc = r""]
#[doc = r" ```json"]
#[doc = "{"]
#[doc = "  \"type\": \"object\","]
#[doc = "  \"required\": ["]
#[doc = "    \"execs\""]
#[doc = "  ],"]
#[doc = "  \"properties\": {"]
#[doc = "    \"args\": {"]
#[doc = "      \"$ref\": \"#/definitions/ExecutableArgumentList\""]
#[doc = "    },"]
#[doc = "    \"dir\": {"]
#[doc = "      \"default\": \"\","]
#[doc = "      \"$ref\": \"#/definitions/ExecutableDirectory\""]
#[doc = "    },"]
#[doc = "    \"execs\": {"]
#[doc = "      \"description\": \"A list of executables to run in parallel.\\nEach executable can be a command or a reference to another executable.\\n\","]
#[doc = "      \"$ref\": \"#/definitions/ExecutableParallelRefConfigList\""]
#[doc = "    },"]
#[doc = "    \"failFast\": {"]
#[doc = "      \"description\": \"End the parallel execution as soon as an exec exits with a non-zero status. This is the default behavior.\\nWhen set to false, all execs will be run regardless of the exit status of parallel execs.\\n\","]
#[doc = "      \"type\": \"boolean\""]
#[doc = "    },"]
#[doc = "    \"maxThreads\": {"]
#[doc = "      \"description\": \"The maximum number of threads to use when executing the parallel executables.\","]
#[doc = "      \"default\": 5,"]
#[doc = "      \"type\": \"integer\""]
#[doc = "    },"]
#[doc = "    \"params\": {"]
#[doc = "      \"$ref\": \"#/definitions/ExecutableParameterList\""]
#[doc = "    }"]
#[doc = "  }"]
#[doc = "}"]
#[doc = r" ```"]
#[doc = r" </details>"]
#[derive(:: serde :: Deserialize, :: serde :: Serialize, Clone, Debug)]
pub struct ExecutableParallelExecutableType {
    #[serde(default, skip_serializing_if = "::std::option::Option::is_none")]
    pub args: ::std::option::Option<ExecutableArgumentList>,
    #[serde(default = "defaults::executable_parallel_executable_type_dir")]
    pub dir: ExecutableDirectory,
    #[doc = "A list of executables to run in parallel.\nEach executable can be a command or a reference to another executable.\n"]
    pub execs: ExecutableParallelRefConfigList,
    #[doc = "End the parallel execution as soon as an exec exits with a non-zero status. This is the default behavior.\nWhen set to false, all execs will be run regardless of the exit status of parallel execs.\n"]
    #[serde(
        rename = "failFast",
        default,
        skip_serializing_if = "::std::option::Option::is_none"
    )]
    pub fail_fast: ::std::option::Option<bool>,
    #[doc = "The maximum number of threads to use when executing the parallel executables."]
    #[serde(rename = "maxThreads", default = "defaults::default_u64::<i64, 5>")]
    pub max_threads: i64,
    #[serde(default, skip_serializing_if = "::std::option::Option::is_none")]
    pub params: ::std::option::Option<ExecutableParameterList>,
}
impl ::std::convert::From<&ExecutableParallelExecutableType> for ExecutableParallelExecutableType {
    fn from(value: &ExecutableParallelExecutableType) -> Self {
        value.clone()
    }
}
impl ExecutableParallelExecutableType {
    pub fn builder() -> builder::ExecutableParallelExecutableType {
        Default::default()
    }
}
#[doc = "Configuration for a parallel executable."]
#[doc = r""]
#[doc = r" <details><summary>JSON schema</summary>"]
#[doc = r""]
#[doc = r" ```json"]
#[doc = "{"]
#[doc = "  \"description\": \"Configuration for a parallel executable.\","]
#[doc = "  \"type\": \"object\","]
#[doc = "  \"properties\": {"]
#[doc = "    \"args\": {"]
#[doc = "      \"description\": \"Arguments to pass to the executable.\","]
#[doc = "      \"default\": [],"]
#[doc = "      \"type\": \"array\","]
#[doc = "      \"items\": {"]
#[doc = "        \"type\": \"string\""]
#[doc = "      }"]
#[doc = "    },"]
#[doc = "    \"cmd\": {"]
#[doc = "      \"description\": \"The command to execute.\\nOne of `cmd` or `ref` must be set.\\n\","]
#[doc = "      \"default\": \"\","]
#[doc = "      \"type\": \"string\""]
#[doc = "    },"]
#[doc = "    \"if\": {"]
#[doc = "      \"description\": \"An expression that determines whether the executable should run, using the Expr language syntax.\\nThe expression is evaluated at runtime and must resolve to a boolean value.\\n\\nThe expression has access to OS/architecture information (os, arch), environment variables (env), stored data\\n(store), and context information (ctx) like workspace and paths.\\n\\nFor example, `os == \\\"darwin\\\"` will only run on macOS, `len(store[\\\"feature\\\"]) > 0` will run if a value exists\\nin the store, and `env[\\\"CI\\\"] == \\\"true\\\"` will run in CI environments.\\nSee the [Expr documentation](https://expr-lang.org/docs/language-definition) for more information.\\n\","]
#[doc = "      \"default\": \"\","]
#[doc = "      \"type\": \"string\""]
#[doc = "    },"]
#[doc = "    \"ref\": {"]
#[doc = "      \"description\": \"A reference to another executable to run in serial.\\nOne of `cmd` or `ref` must be set.\\n\","]
#[doc = "      \"default\": \"\","]
#[doc = "      \"$ref\": \"#/definitions/ExecutableRef\""]
#[doc = "    },"]
#[doc = "    \"retries\": {"]
#[doc = "      \"description\": \"The number of times to retry the executable if it fails.\","]
#[doc = "      \"default\": 0,"]
#[doc = "      \"type\": \"integer\""]
#[doc = "    }"]
#[doc = "  }"]
#[doc = "}"]
#[doc = r" ```"]
#[doc = r" </details>"]
#[derive(:: serde :: Deserialize, :: serde :: Serialize, Clone, Debug)]
pub struct ExecutableParallelRefConfig {
    #[doc = "Arguments to pass to the executable."]
    #[serde(default, skip_serializing_if = "::std::vec::Vec::is_empty")]
    pub args: ::std::vec::Vec<::std::string::String>,
    #[doc = "The command to execute.\nOne of `cmd` or `ref` must be set.\n"]
    #[serde(default)]
    pub cmd: ::std::string::String,
    #[doc = "An expression that determines whether the executable should run, using the Expr language syntax.\nThe expression is evaluated at runtime and must resolve to a boolean value.\n\nThe expression has access to OS/architecture information (os, arch), environment variables (env), stored data\n(store), and context information (ctx) like workspace and paths.\n\nFor example, `os == \"darwin\"` will only run on macOS, `len(store[\"feature\"]) > 0` will run if a value exists\nin the store, and `env[\"CI\"] == \"true\"` will run in CI environments.\nSee the [Expr documentation](https://expr-lang.org/docs/language-definition) for more information.\n"]
    #[serde(rename = "if", default)]
    pub if_: ::std::string::String,
    #[doc = "A reference to another executable to run in serial.\nOne of `cmd` or `ref` must be set.\n"]
    #[serde(
        rename = "ref",
        default = "defaults::executable_parallel_ref_config_ref"
    )]
    pub ref_: ExecutableRef,
    #[doc = "The number of times to retry the executable if it fails."]
    #[serde(default)]
    pub retries: i64,
}
impl ::std::convert::From<&ExecutableParallelRefConfig> for ExecutableParallelRefConfig {
    fn from(value: &ExecutableParallelRefConfig) -> Self {
        value.clone()
    }
}
impl ::std::default::Default for ExecutableParallelRefConfig {
    fn default() -> Self {
        Self {
            args: Default::default(),
            cmd: Default::default(),
            if_: Default::default(),
            ref_: defaults::executable_parallel_ref_config_ref(),
            retries: Default::default(),
        }
    }
}
impl ExecutableParallelRefConfig {
    pub fn builder() -> builder::ExecutableParallelRefConfig {
        Default::default()
    }
}
#[doc = "A list of executables to run in parallel. The executables can be defined by it's exec `cmd` or `ref`.\n"]
#[doc = r""]
#[doc = r" <details><summary>JSON schema</summary>"]
#[doc = r""]
#[doc = r" ```json"]
#[doc = "{"]
#[doc = "  \"description\": \"A list of executables to run in parallel. The executables can be defined by it's exec `cmd` or `ref`.\\n\","]
#[doc = "  \"type\": \"array\","]
#[doc = "  \"items\": {"]
#[doc = "    \"$ref\": \"#/definitions/ExecutableParallelRefConfig\""]
#[doc = "  }"]
#[doc = "}"]
#[doc = r" ```"]
#[doc = r" </details>"]
#[derive(:: serde :: Deserialize, :: serde :: Serialize, Clone, Debug)]
#[serde(transparent)]
pub struct ExecutableParallelRefConfigList(pub ::std::vec::Vec<ExecutableParallelRefConfig>);
impl ::std::ops::Deref for ExecutableParallelRefConfigList {
    type Target = ::std::vec::Vec<ExecutableParallelRefConfig>;
    fn deref(&self) -> &::std::vec::Vec<ExecutableParallelRefConfig> {
        &self.0
    }
}
impl ::std::convert::From<ExecutableParallelRefConfigList>
    for ::std::vec::Vec<ExecutableParallelRefConfig>
{
    fn from(value: ExecutableParallelRefConfigList) -> Self {
        value.0
    }
}
impl ::std::convert::From<&ExecutableParallelRefConfigList> for ExecutableParallelRefConfigList {
    fn from(value: &ExecutableParallelRefConfigList) -> Self {
        value.clone()
    }
}
impl ::std::convert::From<::std::vec::Vec<ExecutableParallelRefConfig>>
    for ExecutableParallelRefConfigList
{
    fn from(value: ::std::vec::Vec<ExecutableParallelRefConfig>) -> Self {
        Self(value)
    }
}
#[doc = "A parameter is a value that can be passed to an executable and all of its sub-executables.\nOnly one of `text`, `secretRef`, `prompt`, or `file` must be set. Specifying more than one will result in an error.\n"]
#[doc = r""]
#[doc = r" <details><summary>JSON schema</summary>"]
#[doc = r""]
#[doc = r" ```json"]
#[doc = "{"]
#[doc = "  \"description\": \"A parameter is a value that can be passed to an executable and all of its sub-executables.\\nOnly one of `text`, `secretRef`, `prompt`, or `file` must be set. Specifying more than one will result in an error.\\n\","]
#[doc = "  \"type\": \"object\","]
#[doc = "  \"properties\": {"]
#[doc = "    \"envFile\": {"]
#[doc = "      \"description\": \"A path to a file containing environment variables to be passed to the executable.\\nThe file should contain one variable per line in the format `KEY=VALUE`.\\n\","]
#[doc = "      \"default\": \"\","]
#[doc = "      \"type\": \"string\""]
#[doc = "    },"]
#[doc = "    \"envKey\": {"]
#[doc = "      \"description\": \"The name of the environment variable that will be assigned the value.\\n\\nWhen specified with `envFile`, only the environment variable with this name will be set.\\n\","]
#[doc = "      \"default\": \"\","]
#[doc = "      \"type\": \"string\""]
#[doc = "    },"]
#[doc = "    \"outputFile\": {"]
#[doc = "      \"description\": \"A path where the parameter value will be temporarily written to disk.\\nThe file will be created before execution and cleaned up afterwards.\\n\","]
#[doc = "      \"default\": \"\","]
#[doc = "      \"type\": \"string\""]
#[doc = "    },"]
#[doc = "    \"prompt\": {"]
#[doc = "      \"description\": \"A prompt to be displayed to the user when collecting an input value.\","]
#[doc = "      \"default\": \"\","]
#[doc = "      \"type\": \"string\""]
#[doc = "    },"]
#[doc = "    \"secretRef\": {"]
#[doc = "      \"description\": \"A reference to a secret to be passed to the executable.\","]
#[doc = "      \"default\": \"\","]
#[doc = "      \"type\": \"string\""]
#[doc = "    },"]
#[doc = "    \"text\": {"]
#[doc = "      \"description\": \"A static value to be passed to the executable.\","]
#[doc = "      \"default\": \"\","]
#[doc = "      \"type\": \"string\""]
#[doc = "    }"]
#[doc = "  }"]
#[doc = "}"]
#[doc = r" ```"]
#[doc = r" </details>"]
#[derive(:: serde :: Deserialize, :: serde :: Serialize, Clone, Debug)]
pub struct ExecutableParameter {
    #[doc = "A path to a file containing environment variables to be passed to the executable.\nThe file should contain one variable per line in the format `KEY=VALUE`.\n"]
    #[serde(rename = "envFile", default)]
    pub env_file: ::std::string::String,
    #[doc = "The name of the environment variable that will be assigned the value.\n\nWhen specified with `envFile`, only the environment variable with this name will be set.\n"]
    #[serde(rename = "envKey", default)]
    pub env_key: ::std::string::String,
    #[doc = "A path where the parameter value will be temporarily written to disk.\nThe file will be created before execution and cleaned up afterwards.\n"]
    #[serde(rename = "outputFile", default)]
    pub output_file: ::std::string::String,
    #[doc = "A prompt to be displayed to the user when collecting an input value."]
    #[serde(default)]
    pub prompt: ::std::string::String,
    #[doc = "A reference to a secret to be passed to the executable."]
    #[serde(rename = "secretRef", default)]
    pub secret_ref: ::std::string::String,
    #[doc = "A static value to be passed to the executable."]
    #[serde(default)]
    pub text: ::std::string::String,
}
impl ::std::convert::From<&ExecutableParameter> for ExecutableParameter {
    fn from(value: &ExecutableParameter) -> Self {
        value.clone()
    }
}
impl ::std::default::Default for ExecutableParameter {
    fn default() -> Self {
        Self {
            env_file: Default::default(),
            env_key: Default::default(),
            output_file: Default::default(),
            prompt: Default::default(),
            secret_ref: Default::default(),
            text: Default::default(),
        }
    }
}
impl ExecutableParameter {
    pub fn builder() -> builder::ExecutableParameter {
        Default::default()
    }
}
#[doc = "`ExecutableParameterList`"]
#[doc = r""]
#[doc = r" <details><summary>JSON schema</summary>"]
#[doc = r""]
#[doc = r" ```json"]
#[doc = "{"]
#[doc = "  \"type\": \"array\","]
#[doc = "  \"items\": {"]
#[doc = "    \"$ref\": \"#/definitions/ExecutableParameter\""]
#[doc = "  }"]
#[doc = "}"]
#[doc = r" ```"]
#[doc = r" </details>"]
#[derive(:: serde :: Deserialize, :: serde :: Serialize, Clone, Debug)]
#[serde(transparent)]
pub struct ExecutableParameterList(pub ::std::vec::Vec<ExecutableParameter>);
impl ::std::ops::Deref for ExecutableParameterList {
    type Target = ::std::vec::Vec<ExecutableParameter>;
    fn deref(&self) -> &::std::vec::Vec<ExecutableParameter> {
        &self.0
    }
}
impl ::std::convert::From<ExecutableParameterList> for ::std::vec::Vec<ExecutableParameter> {
    fn from(value: ExecutableParameterList) -> Self {
        value.0
    }
}
impl ::std::convert::From<&ExecutableParameterList> for ExecutableParameterList {
    fn from(value: &ExecutableParameterList) -> Self {
        value.clone()
    }
}
impl ::std::convert::From<::std::vec::Vec<ExecutableParameter>> for ExecutableParameterList {
    fn from(value: ::std::vec::Vec<ExecutableParameter>) -> Self {
        Self(value)
    }
}
#[doc = "A reference to an executable.\nThe format is `<verb> <workspace>/<namespace>:<executable name>`.\nFor example, `exec ws/ns:my-workflow`.\n\n- If the workspace is not specified, the current workspace will be used.\n- If the namespace is not specified, the current namespace will be used.\n- Excluding the name will reference the executable with a matching verb but an unspecified name and namespace (e.g. `exec ws` or simply `exec`).\n"]
#[doc = r""]
#[doc = r" <details><summary>JSON schema</summary>"]
#[doc = r""]
#[doc = r" ```json"]
#[doc = "{"]
#[doc = "  \"description\": \"A reference to an executable.\\nThe format is `<verb> <workspace>/<namespace>:<executable name>`.\\nFor example, `exec ws/ns:my-workflow`.\\n\\n- If the workspace is not specified, the current workspace will be used.\\n- If the namespace is not specified, the current namespace will be used.\\n- Excluding the name will reference the executable with a matching verb but an unspecified name and namespace (e.g. `exec ws` or simply `exec`).\\n\","]
#[doc = "  \"type\": \"string\""]
#[doc = "}"]
#[doc = r" ```"]
#[doc = r" </details>"]
#[derive(
    :: serde :: Deserialize,
    :: serde :: Serialize,
    Clone,
    Debug,
    Eq,
    Hash,
    Ord,
    PartialEq,
    PartialOrd,
)]
#[serde(transparent)]
pub struct ExecutableRef(pub ::std::string::String);
impl ::std::ops::Deref for ExecutableRef {
    type Target = ::std::string::String;
    fn deref(&self) -> &::std::string::String {
        &self.0
    }
}
impl ::std::convert::From<ExecutableRef> for ::std::string::String {
    fn from(value: ExecutableRef) -> Self {
        value.0
    }
}
impl ::std::convert::From<&ExecutableRef> for ExecutableRef {
    fn from(value: &ExecutableRef) -> Self {
        value.clone()
    }
}
impl ::std::convert::From<::std::string::String> for ExecutableRef {
    fn from(value: ::std::string::String) -> Self {
        Self(value)
    }
}
impl ::std::str::FromStr for ExecutableRef {
    type Err = ::std::convert::Infallible;
    fn from_str(value: &str) -> ::std::result::Result<Self, Self::Err> {
        Ok(Self(value.to_string()))
    }
}
impl ::std::fmt::Display for ExecutableRef {
    fn fmt(&self, f: &mut ::std::fmt::Formatter<'_>) -> ::std::fmt::Result {
        self.0.fmt(f)
    }
}
#[doc = "Renders a markdown template file with data."]
#[doc = r""]
#[doc = r" <details><summary>JSON schema</summary>"]
#[doc = r""]
#[doc = r" ```json"]
#[doc = "{"]
#[doc = "  \"description\": \"Renders a markdown template file with data.\","]
#[doc = "  \"type\": \"object\","]
#[doc = "  \"required\": ["]
#[doc = "    \"templateFile\""]
#[doc = "  ],"]
#[doc = "  \"properties\": {"]
#[doc = "    \"args\": {"]
#[doc = "      \"$ref\": \"#/definitions/ExecutableArgumentList\""]
#[doc = "    },"]
#[doc = "    \"dir\": {"]
#[doc = "      \"default\": \"\","]
#[doc = "      \"$ref\": \"#/definitions/ExecutableDirectory\""]
#[doc = "    },"]
#[doc = "    \"params\": {"]
#[doc = "      \"$ref\": \"#/definitions/ExecutableParameterList\""]
#[doc = "    },"]
#[doc = "    \"templateDataFile\": {"]
#[doc = "      \"description\": \"The path to the JSON or YAML file containing the template data.\","]
#[doc = "      \"default\": \"\","]
#[doc = "      \"type\": \"string\""]
#[doc = "    },"]
#[doc = "    \"templateFile\": {"]
#[doc = "      \"description\": \"The path to the markdown template file to render.\","]
#[doc = "      \"default\": \"\","]
#[doc = "      \"type\": \"string\""]
#[doc = "    }"]
#[doc = "  }"]
#[doc = "}"]
#[doc = r" ```"]
#[doc = r" </details>"]
#[derive(:: serde :: Deserialize, :: serde :: Serialize, Clone, Debug)]
pub struct ExecutableRenderExecutableType {
    #[serde(default, skip_serializing_if = "::std::option::Option::is_none")]
    pub args: ::std::option::Option<ExecutableArgumentList>,
    #[serde(default = "defaults::executable_render_executable_type_dir")]
    pub dir: ExecutableDirectory,
    #[serde(default, skip_serializing_if = "::std::option::Option::is_none")]
    pub params: ::std::option::Option<ExecutableParameterList>,
    #[doc = "The path to the JSON or YAML file containing the template data."]
    #[serde(rename = "templateDataFile", default)]
    pub template_data_file: ::std::string::String,
    #[doc = "The path to the markdown template file to render."]
    #[serde(rename = "templateFile")]
    pub template_file: ::std::string::String,
}
impl ::std::convert::From<&ExecutableRenderExecutableType> for ExecutableRenderExecutableType {
    fn from(value: &ExecutableRenderExecutableType) -> Self {
        value.clone()
    }
}
impl ExecutableRenderExecutableType {
    pub fn builder() -> builder::ExecutableRenderExecutableType {
        Default::default()
    }
}
#[doc = "Makes an HTTP request."]
#[doc = r""]
#[doc = r" <details><summary>JSON schema</summary>"]
#[doc = r""]
#[doc = r" ```json"]
#[doc = "{"]
#[doc = "  \"description\": \"Makes an HTTP request.\","]
#[doc = "  \"type\": \"object\","]
#[doc = "  \"required\": ["]
#[doc = "    \"url\""]
#[doc = "  ],"]
#[doc = "  \"properties\": {"]
#[doc = "    \"args\": {"]
#[doc = "      \"$ref\": \"#/definitions/ExecutableArgumentList\""]
#[doc = "    },"]
#[doc = "    \"body\": {"]
#[doc = "      \"description\": \"The body of the request.\","]
#[doc = "      \"default\": \"\","]
#[doc = "      \"type\": \"string\""]
#[doc = "    },"]
#[doc = "    \"headers\": {"]
#[doc = "      \"description\": \"A map of headers to include in the request.\","]
#[doc = "      \"default\": {},"]
#[doc = "      \"type\": \"object\","]
#[doc = "      \"additionalProperties\": {"]
#[doc = "        \"type\": \"string\""]
#[doc = "      }"]
#[doc = "    },"]
#[doc = "    \"logResponse\": {"]
#[doc = "      \"description\": \"If set to true, the response will be logged as program output.\","]
#[doc = "      \"default\": false,"]
#[doc = "      \"type\": \"boolean\""]
#[doc = "    },"]
#[doc = "    \"method\": {"]
#[doc = "      \"description\": \"The HTTP method to use when making the request.\","]
#[doc = "      \"default\": \"GET\","]
#[doc = "      \"type\": \"string\","]
#[doc = "      \"enum\": ["]
#[doc = "        \"GET\","]
#[doc = "        \"POST\","]
#[doc = "        \"PUT\","]
#[doc = "        \"PATCH\","]
#[doc = "        \"DELETE\""]
#[doc = "      ]"]
#[doc = "    },"]
#[doc = "    \"params\": {"]
#[doc = "      \"$ref\": \"#/definitions/ExecutableParameterList\""]
#[doc = "    },"]
#[doc = "    \"responseFile\": {"]
#[doc = "      \"$ref\": \"#/definitions/ExecutableRequestResponseFile\""]
#[doc = "    },"]
#[doc = "    \"timeout\": {"]
#[doc = "      \"description\": \"The timeout for the request in Go duration format (e.g. 30s, 5m, 1h).\","]
#[doc = "      \"default\": \"30m0s\","]
#[doc = "      \"type\": \"string\""]
#[doc = "    },"]
#[doc = "    \"transformResponse\": {"]
#[doc = "      \"description\": \"[Expr](https://expr-lang.org/docs/language-definition) expression used to transform the response before\\nsaving it to a file or outputting it.\\n\\nThe following variables are available in the expression:\\n  - `status`: The response status string.\\n  - `code`: The response status code.\\n  - `body`: The response body.\\n  - `headers`: The response headers.\\n\\nFor example, to capitalize a JSON body field's value, you can use `upper(fromJSON(body)[\\\"field\\\"])`.\\n\","]
#[doc = "      \"default\": \"\","]
#[doc = "      \"type\": \"string\""]
#[doc = "    },"]
#[doc = "    \"url\": {"]
#[doc = "      \"description\": \"The URL to make the request to.\","]
#[doc = "      \"default\": \"\","]
#[doc = "      \"type\": \"string\""]
#[doc = "    },"]
#[doc = "    \"validStatusCodes\": {"]
#[doc = "      \"description\": \"A list of valid status codes. If the response status code is not in this list, the executable will fail.\\nIf not set, the response status code will not be checked.\\n\","]
#[doc = "      \"default\": [],"]
#[doc = "      \"type\": \"array\","]
#[doc = "      \"items\": {"]
#[doc = "        \"type\": \"integer\""]
#[doc = "      }"]
#[doc = "    }"]
#[doc = "  }"]
#[doc = "}"]
#[doc = r" ```"]
#[doc = r" </details>"]
#[derive(:: serde :: Deserialize, :: serde :: Serialize, Clone, Debug)]
pub struct ExecutableRequestExecutableType {
    #[serde(default, skip_serializing_if = "::std::option::Option::is_none")]
    pub args: ::std::option::Option<ExecutableArgumentList>,
    #[doc = "The body of the request."]
    #[serde(default)]
    pub body: ::std::string::String,
    #[doc = "A map of headers to include in the request."]
    #[serde(
        default,
        skip_serializing_if = ":: std :: collections :: HashMap::is_empty"
    )]
    pub headers: ::std::collections::HashMap<::std::string::String, ::std::string::String>,
    #[doc = "If set to true, the response will be logged as program output."]
    #[serde(rename = "logResponse", default)]
    pub log_response: bool,
    #[doc = "The HTTP method to use when making the request."]
    #[serde(default = "defaults::executable_request_executable_type_method")]
    pub method: ExecutableRequestExecutableTypeMethod,
    #[serde(default, skip_serializing_if = "::std::option::Option::is_none")]
    pub params: ::std::option::Option<ExecutableParameterList>,
    #[serde(
        rename = "responseFile",
        default,
        skip_serializing_if = "::std::option::Option::is_none"
    )]
    pub response_file: ::std::option::Option<ExecutableRequestResponseFile>,
    #[doc = "The timeout for the request in Go duration format (e.g. 30s, 5m, 1h)."]
    #[serde(default = "defaults::executable_request_executable_type_timeout")]
    pub timeout: ::std::string::String,
    #[doc = "[Expr](https://expr-lang.org/docs/language-definition) expression used to transform the response before\nsaving it to a file or outputting it.\n\nThe following variables are available in the expression:\n  - `status`: The response status string.\n  - `code`: The response status code.\n  - `body`: The response body.\n  - `headers`: The response headers.\n\nFor example, to capitalize a JSON body field's value, you can use `upper(fromJSON(body)[\"field\"])`.\n"]
    #[serde(rename = "transformResponse", default)]
    pub transform_response: ::std::string::String,
    #[doc = "The URL to make the request to."]
    pub url: ::std::string::String,
    #[doc = "A list of valid status codes. If the response status code is not in this list, the executable will fail.\nIf not set, the response status code will not be checked.\n"]
    #[serde(
        rename = "validStatusCodes",
        default,
        skip_serializing_if = "::std::vec::Vec::is_empty"
    )]
    pub valid_status_codes: ::std::vec::Vec<i64>,
}
impl ::std::convert::From<&ExecutableRequestExecutableType> for ExecutableRequestExecutableType {
    fn from(value: &ExecutableRequestExecutableType) -> Self {
        value.clone()
    }
}
impl ExecutableRequestExecutableType {
    pub fn builder() -> builder::ExecutableRequestExecutableType {
        Default::default()
    }
}
#[doc = "The HTTP method to use when making the request."]
#[doc = r""]
#[doc = r" <details><summary>JSON schema</summary>"]
#[doc = r""]
#[doc = r" ```json"]
#[doc = "{"]
#[doc = "  \"description\": \"The HTTP method to use when making the request.\","]
#[doc = "  \"default\": \"GET\","]
#[doc = "  \"type\": \"string\","]
#[doc = "  \"enum\": ["]
#[doc = "    \"GET\","]
#[doc = "    \"POST\","]
#[doc = "    \"PUT\","]
#[doc = "    \"PATCH\","]
#[doc = "    \"DELETE\""]
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
pub enum ExecutableRequestExecutableTypeMethod {
    #[serde(rename = "GET")]
    Get,
    #[serde(rename = "POST")]
    Post,
    #[serde(rename = "PUT")]
    Put,
    #[serde(rename = "PATCH")]
    Patch,
    #[serde(rename = "DELETE")]
    Delete,
}
impl ::std::convert::From<&Self> for ExecutableRequestExecutableTypeMethod {
    fn from(value: &ExecutableRequestExecutableTypeMethod) -> Self {
        value.clone()
    }
}
impl ::std::fmt::Display for ExecutableRequestExecutableTypeMethod {
    fn fmt(&self, f: &mut ::std::fmt::Formatter<'_>) -> ::std::fmt::Result {
        match *self {
            Self::Get => f.write_str("GET"),
            Self::Post => f.write_str("POST"),
            Self::Put => f.write_str("PUT"),
            Self::Patch => f.write_str("PATCH"),
            Self::Delete => f.write_str("DELETE"),
        }
    }
}
impl ::std::str::FromStr for ExecutableRequestExecutableTypeMethod {
    type Err = self::error::ConversionError;
    fn from_str(value: &str) -> ::std::result::Result<Self, self::error::ConversionError> {
        match value {
            "GET" => Ok(Self::Get),
            "POST" => Ok(Self::Post),
            "PUT" => Ok(Self::Put),
            "PATCH" => Ok(Self::Patch),
            "DELETE" => Ok(Self::Delete),
            _ => Err("invalid value".into()),
        }
    }
}
impl ::std::convert::TryFrom<&str> for ExecutableRequestExecutableTypeMethod {
    type Error = self::error::ConversionError;
    fn try_from(value: &str) -> ::std::result::Result<Self, self::error::ConversionError> {
        value.parse()
    }
}
impl ::std::convert::TryFrom<&::std::string::String> for ExecutableRequestExecutableTypeMethod {
    type Error = self::error::ConversionError;
    fn try_from(
        value: &::std::string::String,
    ) -> ::std::result::Result<Self, self::error::ConversionError> {
        value.parse()
    }
}
impl ::std::convert::TryFrom<::std::string::String> for ExecutableRequestExecutableTypeMethod {
    type Error = self::error::ConversionError;
    fn try_from(
        value: ::std::string::String,
    ) -> ::std::result::Result<Self, self::error::ConversionError> {
        value.parse()
    }
}
impl ::std::default::Default for ExecutableRequestExecutableTypeMethod {
    fn default() -> Self {
        ExecutableRequestExecutableTypeMethod::Get
    }
}
#[doc = "Configuration for saving the response of a request to a file."]
#[doc = r""]
#[doc = r" <details><summary>JSON schema</summary>"]
#[doc = r""]
#[doc = r" ```json"]
#[doc = "{"]
#[doc = "  \"description\": \"Configuration for saving the response of a request to a file.\","]
#[doc = "  \"type\": \"object\","]
#[doc = "  \"required\": ["]
#[doc = "    \"filename\""]
#[doc = "  ],"]
#[doc = "  \"properties\": {"]
#[doc = "    \"dir\": {"]
#[doc = "      \"default\": \"\","]
#[doc = "      \"$ref\": \"#/definitions/ExecutableDirectory\""]
#[doc = "    },"]
#[doc = "    \"filename\": {"]
#[doc = "      \"description\": \"The name of the file to save the response to.\","]
#[doc = "      \"default\": \"\","]
#[doc = "      \"type\": \"string\""]
#[doc = "    },"]
#[doc = "    \"saveAs\": {"]
#[doc = "      \"description\": \"The format to save the response as.\","]
#[doc = "      \"default\": \"raw\","]
#[doc = "      \"type\": \"string\","]
#[doc = "      \"enum\": ["]
#[doc = "        \"raw\","]
#[doc = "        \"json\","]
#[doc = "        \"indented-json\","]
#[doc = "        \"yaml\","]
#[doc = "        \"yml\""]
#[doc = "      ]"]
#[doc = "    }"]
#[doc = "  }"]
#[doc = "}"]
#[doc = r" ```"]
#[doc = r" </details>"]
#[derive(:: serde :: Deserialize, :: serde :: Serialize, Clone, Debug)]
pub struct ExecutableRequestResponseFile {
    #[serde(default = "defaults::executable_request_response_file_dir")]
    pub dir: ExecutableDirectory,
    #[doc = "The name of the file to save the response to."]
    pub filename: ::std::string::String,
    #[doc = "The format to save the response as."]
    #[serde(
        rename = "saveAs",
        default = "defaults::executable_request_response_file_save_as"
    )]
    pub save_as: ExecutableRequestResponseFileSaveAs,
}
impl ::std::convert::From<&ExecutableRequestResponseFile> for ExecutableRequestResponseFile {
    fn from(value: &ExecutableRequestResponseFile) -> Self {
        value.clone()
    }
}
impl ExecutableRequestResponseFile {
    pub fn builder() -> builder::ExecutableRequestResponseFile {
        Default::default()
    }
}
#[doc = "The format to save the response as."]
#[doc = r""]
#[doc = r" <details><summary>JSON schema</summary>"]
#[doc = r""]
#[doc = r" ```json"]
#[doc = "{"]
#[doc = "  \"description\": \"The format to save the response as.\","]
#[doc = "  \"default\": \"raw\","]
#[doc = "  \"type\": \"string\","]
#[doc = "  \"enum\": ["]
#[doc = "    \"raw\","]
#[doc = "    \"json\","]
#[doc = "    \"indented-json\","]
#[doc = "    \"yaml\","]
#[doc = "    \"yml\""]
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
pub enum ExecutableRequestResponseFileSaveAs {
    #[serde(rename = "raw")]
    Raw,
    #[serde(rename = "json")]
    Json,
    #[serde(rename = "indented-json")]
    IndentedJson,
    #[serde(rename = "yaml")]
    Yaml,
    #[serde(rename = "yml")]
    Yml,
}
impl ::std::convert::From<&Self> for ExecutableRequestResponseFileSaveAs {
    fn from(value: &ExecutableRequestResponseFileSaveAs) -> Self {
        value.clone()
    }
}
impl ::std::fmt::Display for ExecutableRequestResponseFileSaveAs {
    fn fmt(&self, f: &mut ::std::fmt::Formatter<'_>) -> ::std::fmt::Result {
        match *self {
            Self::Raw => f.write_str("raw"),
            Self::Json => f.write_str("json"),
            Self::IndentedJson => f.write_str("indented-json"),
            Self::Yaml => f.write_str("yaml"),
            Self::Yml => f.write_str("yml"),
        }
    }
}
impl ::std::str::FromStr for ExecutableRequestResponseFileSaveAs {
    type Err = self::error::ConversionError;
    fn from_str(value: &str) -> ::std::result::Result<Self, self::error::ConversionError> {
        match value {
            "raw" => Ok(Self::Raw),
            "json" => Ok(Self::Json),
            "indented-json" => Ok(Self::IndentedJson),
            "yaml" => Ok(Self::Yaml),
            "yml" => Ok(Self::Yml),
            _ => Err("invalid value".into()),
        }
    }
}
impl ::std::convert::TryFrom<&str> for ExecutableRequestResponseFileSaveAs {
    type Error = self::error::ConversionError;
    fn try_from(value: &str) -> ::std::result::Result<Self, self::error::ConversionError> {
        value.parse()
    }
}
impl ::std::convert::TryFrom<&::std::string::String> for ExecutableRequestResponseFileSaveAs {
    type Error = self::error::ConversionError;
    fn try_from(
        value: &::std::string::String,
    ) -> ::std::result::Result<Self, self::error::ConversionError> {
        value.parse()
    }
}
impl ::std::convert::TryFrom<::std::string::String> for ExecutableRequestResponseFileSaveAs {
    type Error = self::error::ConversionError;
    fn try_from(
        value: ::std::string::String,
    ) -> ::std::result::Result<Self, self::error::ConversionError> {
        value.parse()
    }
}
impl ::std::default::Default for ExecutableRequestResponseFileSaveAs {
    fn default() -> Self {
        ExecutableRequestResponseFileSaveAs::Raw
    }
}
#[doc = "Executes a list of executables in serial."]
#[doc = r""]
#[doc = r" <details><summary>JSON schema</summary>"]
#[doc = r""]
#[doc = r" ```json"]
#[doc = "{"]
#[doc = "  \"description\": \"Executes a list of executables in serial.\","]
#[doc = "  \"type\": \"object\","]
#[doc = "  \"required\": ["]
#[doc = "    \"execs\""]
#[doc = "  ],"]
#[doc = "  \"properties\": {"]
#[doc = "    \"args\": {"]
#[doc = "      \"$ref\": \"#/definitions/ExecutableArgumentList\""]
#[doc = "    },"]
#[doc = "    \"dir\": {"]
#[doc = "      \"default\": \"\","]
#[doc = "      \"$ref\": \"#/definitions/ExecutableDirectory\""]
#[doc = "    },"]
#[doc = "    \"execs\": {"]
#[doc = "      \"description\": \"A list of executables to run in serial.\\nEach executable can be a command or a reference to another executable.\\n\","]
#[doc = "      \"$ref\": \"#/definitions/ExecutableSerialRefConfigList\""]
#[doc = "    },"]
#[doc = "    \"failFast\": {"]
#[doc = "      \"description\": \"End the serial execution as soon as an exec exits with a non-zero status. This is the default behavior.\\nWhen set to false, all execs will be run regardless of the exit status of the previous exec.\\n\","]
#[doc = "      \"type\": \"boolean\""]
#[doc = "    },"]
#[doc = "    \"params\": {"]
#[doc = "      \"$ref\": \"#/definitions/ExecutableParameterList\""]
#[doc = "    }"]
#[doc = "  }"]
#[doc = "}"]
#[doc = r" ```"]
#[doc = r" </details>"]
#[derive(:: serde :: Deserialize, :: serde :: Serialize, Clone, Debug)]
pub struct ExecutableSerialExecutableType {
    #[serde(default, skip_serializing_if = "::std::option::Option::is_none")]
    pub args: ::std::option::Option<ExecutableArgumentList>,
    #[serde(default = "defaults::executable_serial_executable_type_dir")]
    pub dir: ExecutableDirectory,
    #[doc = "A list of executables to run in serial.\nEach executable can be a command or a reference to another executable.\n"]
    pub execs: ExecutableSerialRefConfigList,
    #[doc = "End the serial execution as soon as an exec exits with a non-zero status. This is the default behavior.\nWhen set to false, all execs will be run regardless of the exit status of the previous exec.\n"]
    #[serde(
        rename = "failFast",
        default,
        skip_serializing_if = "::std::option::Option::is_none"
    )]
    pub fail_fast: ::std::option::Option<bool>,
    #[serde(default, skip_serializing_if = "::std::option::Option::is_none")]
    pub params: ::std::option::Option<ExecutableParameterList>,
}
impl ::std::convert::From<&ExecutableSerialExecutableType> for ExecutableSerialExecutableType {
    fn from(value: &ExecutableSerialExecutableType) -> Self {
        value.clone()
    }
}
impl ExecutableSerialExecutableType {
    pub fn builder() -> builder::ExecutableSerialExecutableType {
        Default::default()
    }
}
#[doc = "Configuration for a serial executable."]
#[doc = r""]
#[doc = r" <details><summary>JSON schema</summary>"]
#[doc = r""]
#[doc = r" ```json"]
#[doc = "{"]
#[doc = "  \"description\": \"Configuration for a serial executable.\","]
#[doc = "  \"type\": \"object\","]
#[doc = "  \"properties\": {"]
#[doc = "    \"args\": {"]
#[doc = "      \"description\": \"Arguments to pass to the executable.\","]
#[doc = "      \"default\": [],"]
#[doc = "      \"type\": \"array\","]
#[doc = "      \"items\": {"]
#[doc = "        \"type\": \"string\""]
#[doc = "      }"]
#[doc = "    },"]
#[doc = "    \"cmd\": {"]
#[doc = "      \"description\": \"The command to execute.\\nOne of `cmd` or `ref` must be set.\\n\","]
#[doc = "      \"default\": \"\","]
#[doc = "      \"type\": \"string\""]
#[doc = "    },"]
#[doc = "    \"if\": {"]
#[doc = "      \"description\": \"An expression that determines whether the executable should run, using the Expr language syntax.\\nThe expression is evaluated at runtime and must resolve to a boolean value.\\n\\nThe expression has access to OS/architecture information (os, arch), environment variables (env), stored data\\n(store), and context information (ctx) like workspace and paths.\\n\\nFor example, `os == \\\"darwin\\\"` will only run on macOS, `len(store[\\\"feature\\\"]) > 0` will run if a value exists\\nin the store, and `env[\\\"CI\\\"] == \\\"true\\\"` will run in CI environments.\\nSee the [Expr documentation](https://expr-lang.org/docs/language-definition) for more information.\\n\","]
#[doc = "      \"default\": \"\","]
#[doc = "      \"type\": \"string\""]
#[doc = "    },"]
#[doc = "    \"ref\": {"]
#[doc = "      \"description\": \"A reference to another executable to run in serial.\\nOne of `cmd` or `ref` must be set.\\n\","]
#[doc = "      \"default\": \"\","]
#[doc = "      \"$ref\": \"#/definitions/ExecutableRef\""]
#[doc = "    },"]
#[doc = "    \"retries\": {"]
#[doc = "      \"description\": \"The number of times to retry the executable if it fails.\","]
#[doc = "      \"default\": 0,"]
#[doc = "      \"type\": \"integer\""]
#[doc = "    },"]
#[doc = "    \"reviewRequired\": {"]
#[doc = "      \"description\": \"If set to true, the user will be prompted to review the output of the executable before continuing.\","]
#[doc = "      \"default\": false,"]
#[doc = "      \"type\": \"boolean\""]
#[doc = "    }"]
#[doc = "  }"]
#[doc = "}"]
#[doc = r" ```"]
#[doc = r" </details>"]
#[derive(:: serde :: Deserialize, :: serde :: Serialize, Clone, Debug)]
pub struct ExecutableSerialRefConfig {
    #[doc = "Arguments to pass to the executable."]
    #[serde(default, skip_serializing_if = "::std::vec::Vec::is_empty")]
    pub args: ::std::vec::Vec<::std::string::String>,
    #[doc = "The command to execute.\nOne of `cmd` or `ref` must be set.\n"]
    #[serde(default)]
    pub cmd: ::std::string::String,
    #[doc = "An expression that determines whether the executable should run, using the Expr language syntax.\nThe expression is evaluated at runtime and must resolve to a boolean value.\n\nThe expression has access to OS/architecture information (os, arch), environment variables (env), stored data\n(store), and context information (ctx) like workspace and paths.\n\nFor example, `os == \"darwin\"` will only run on macOS, `len(store[\"feature\"]) > 0` will run if a value exists\nin the store, and `env[\"CI\"] == \"true\"` will run in CI environments.\nSee the [Expr documentation](https://expr-lang.org/docs/language-definition) for more information.\n"]
    #[serde(rename = "if", default)]
    pub if_: ::std::string::String,
    #[doc = "A reference to another executable to run in serial.\nOne of `cmd` or `ref` must be set.\n"]
    #[serde(rename = "ref", default = "defaults::executable_serial_ref_config_ref")]
    pub ref_: ExecutableRef,
    #[doc = "The number of times to retry the executable if it fails."]
    #[serde(default)]
    pub retries: i64,
    #[doc = "If set to true, the user will be prompted to review the output of the executable before continuing."]
    #[serde(rename = "reviewRequired", default)]
    pub review_required: bool,
}
impl ::std::convert::From<&ExecutableSerialRefConfig> for ExecutableSerialRefConfig {
    fn from(value: &ExecutableSerialRefConfig) -> Self {
        value.clone()
    }
}
impl ::std::default::Default for ExecutableSerialRefConfig {
    fn default() -> Self {
        Self {
            args: Default::default(),
            cmd: Default::default(),
            if_: Default::default(),
            ref_: defaults::executable_serial_ref_config_ref(),
            retries: Default::default(),
            review_required: Default::default(),
        }
    }
}
impl ExecutableSerialRefConfig {
    pub fn builder() -> builder::ExecutableSerialRefConfig {
        Default::default()
    }
}
#[doc = "A list of executables to run in serial. The executables can be defined by it's exec `cmd` or `ref`.\n"]
#[doc = r""]
#[doc = r" <details><summary>JSON schema</summary>"]
#[doc = r""]
#[doc = r" ```json"]
#[doc = "{"]
#[doc = "  \"description\": \"A list of executables to run in serial. The executables can be defined by it's exec `cmd` or `ref`.\\n\","]
#[doc = "  \"type\": \"array\","]
#[doc = "  \"items\": {"]
#[doc = "    \"$ref\": \"#/definitions/ExecutableSerialRefConfig\""]
#[doc = "  }"]
#[doc = "}"]
#[doc = r" ```"]
#[doc = r" </details>"]
#[derive(:: serde :: Deserialize, :: serde :: Serialize, Clone, Debug)]
#[serde(transparent)]
pub struct ExecutableSerialRefConfigList(pub ::std::vec::Vec<ExecutableSerialRefConfig>);
impl ::std::ops::Deref for ExecutableSerialRefConfigList {
    type Target = ::std::vec::Vec<ExecutableSerialRefConfig>;
    fn deref(&self) -> &::std::vec::Vec<ExecutableSerialRefConfig> {
        &self.0
    }
}
impl ::std::convert::From<ExecutableSerialRefConfigList>
    for ::std::vec::Vec<ExecutableSerialRefConfig>
{
    fn from(value: ExecutableSerialRefConfigList) -> Self {
        value.0
    }
}
impl ::std::convert::From<&ExecutableSerialRefConfigList> for ExecutableSerialRefConfigList {
    fn from(value: &ExecutableSerialRefConfigList) -> Self {
        value.clone()
    }
}
impl ::std::convert::From<::std::vec::Vec<ExecutableSerialRefConfig>>
    for ExecutableSerialRefConfigList
{
    fn from(value: ::std::vec::Vec<ExecutableSerialRefConfig>) -> Self {
        Self(value)
    }
}
#[doc = "Keywords that describe the action an executable performs. Executables are configured with a single verb,\nbut core verbs have aliases that can be used interchangeably when referencing executables. This allows users \nto use the verb that best describes the action they are performing.\n\n### Default Verb Aliases\n\n- **Execution Group**: `exec`, `run`, `execute`\n- **Retrieval Group**: `get`, `fetch`, `retrieve`\n- **Display Group**: `show`, `view`, `list`\n- **Configuration Group**: `configure`, `setup`\n- **Update Group**: `update`, `upgrade`\n\n### Usage Notes\n\n1. [Verb + Name] must be unique within the namespace of the workspace.\n2. When referencing an executable, users can use any verb from the default or configured alias group.\n3. All other verbs are standalone and self-descriptive.\n\n### Examples\n\n- An executable configured with the `exec` verb can also be referenced using \"run\" or \"execute\".\n- An executable configured with `get` can also be called with \"list\", \"show\", or \"view\".\n- Operations like `backup`, `migrate`, `flush` are standalone verbs without aliases.\n- Use domain-specific verbs like `deploy`, `scale`, `tunnel` for clear operational intent.\n\nBy providing minimal aliasing with comprehensive verb coverage, flow enables natural language operations\nwhile maintaining simplicity and flexibility for diverse development and operations workflows.\n"]
#[doc = r""]
#[doc = r" <details><summary>JSON schema</summary>"]
#[doc = r""]
#[doc = r" ```json"]
#[doc = "{"]
#[doc = "  \"description\": \"Keywords that describe the action an executable performs. Executables are configured with a single verb,\\nbut core verbs have aliases that can be used interchangeably when referencing executables. This allows users \\nto use the verb that best describes the action they are performing.\\n\\n### Default Verb Aliases\\n\\n- **Execution Group**: `exec`, `run`, `execute`\\n- **Retrieval Group**: `get`, `fetch`, `retrieve`\\n- **Display Group**: `show`, `view`, `list`\\n- **Configuration Group**: `configure`, `setup`\\n- **Update Group**: `update`, `upgrade`\\n\\n### Usage Notes\\n\\n1. [Verb + Name] must be unique within the namespace of the workspace.\\n2. When referencing an executable, users can use any verb from the default or configured alias group.\\n3. All other verbs are standalone and self-descriptive.\\n\\n### Examples\\n\\n- An executable configured with the `exec` verb can also be referenced using \\\"run\\\" or \\\"execute\\\".\\n- An executable configured with `get` can also be called with \\\"list\\\", \\\"show\\\", or \\\"view\\\".\\n- Operations like `backup`, `migrate`, `flush` are standalone verbs without aliases.\\n- Use domain-specific verbs like `deploy`, `scale`, `tunnel` for clear operational intent.\\n\\nBy providing minimal aliasing with comprehensive verb coverage, flow enables natural language operations\\nwhile maintaining simplicity and flexibility for diverse development and operations workflows.\\n\","]
#[doc = "  \"default\": \"exec\","]
#[doc = "  \"type\": \"string\","]
#[doc = "  \"enum\": ["]
#[doc = "    \"abort\","]
#[doc = "    \"activate\","]
#[doc = "    \"add\","]
#[doc = "    \"analyze\","]
#[doc = "    \"apply\","]
#[doc = "    \"archive\","]
#[doc = "    \"audit\","]
#[doc = "    \"backup\","]
#[doc = "    \"benchmark\","]
#[doc = "    \"build\","]
#[doc = "    \"bundle\","]
#[doc = "    \"check\","]
#[doc = "    \"clean\","]
#[doc = "    \"clear\","]
#[doc = "    \"commit\","]
#[doc = "    \"compile\","]
#[doc = "    \"compress\","]
#[doc = "    \"configure\","]
#[doc = "    \"connect\","]
#[doc = "    \"copy\","]
#[doc = "    \"create\","]
#[doc = "    \"deactivate\","]
#[doc = "    \"debug\","]
#[doc = "    \"decompress\","]
#[doc = "    \"decrypt\","]
#[doc = "    \"delete\","]
#[doc = "    \"deploy\","]
#[doc = "    \"destroy\","]
#[doc = "    \"disable\","]
#[doc = "    \"disconnect\","]
#[doc = "    \"edit\","]
#[doc = "    \"enable\","]
#[doc = "    \"encrypt\","]
#[doc = "    \"erase\","]
#[doc = "    \"exec\","]
#[doc = "    \"execute\","]
#[doc = "    \"export\","]
#[doc = "    \"expose\","]
#[doc = "    \"fetch\","]
#[doc = "    \"fix\","]
#[doc = "    \"flush\","]
#[doc = "    \"format\","]
#[doc = "    \"generate\","]
#[doc = "    \"get\","]
#[doc = "    \"import\","]
#[doc = "    \"index\","]
#[doc = "    \"init\","]
#[doc = "    \"inspect\","]
#[doc = "    \"install\","]
#[doc = "    \"join\","]
#[doc = "    \"kill\","]
#[doc = "    \"launch\","]
#[doc = "    \"lint\","]
#[doc = "    \"list\","]
#[doc = "    \"load\","]
#[doc = "    \"lock\","]
#[doc = "    \"login\","]
#[doc = "    \"logout\","]
#[doc = "    \"manage\","]
#[doc = "    \"merge\","]
#[doc = "    \"migrate\","]
#[doc = "    \"modify\","]
#[doc = "    \"monitor\","]
#[doc = "    \"mount\","]
#[doc = "    \"new\","]
#[doc = "    \"notify\","]
#[doc = "    \"open\","]
#[doc = "    \"package\","]
#[doc = "    \"partition\","]
#[doc = "    \"patch\","]
#[doc = "    \"pause\","]
#[doc = "    \"ping\","]
#[doc = "    \"preload\","]
#[doc = "    \"prefetch\","]
#[doc = "    \"profile\","]
#[doc = "    \"provision\","]
#[doc = "    \"publish\","]
#[doc = "    \"purge\","]
#[doc = "    \"push\","]
#[doc = "    \"queue\","]
#[doc = "    \"reboot\","]
#[doc = "    \"recover\","]
#[doc = "    \"refresh\","]
#[doc = "    \"release\","]
#[doc = "    \"reload\","]
#[doc = "    \"remove\","]
#[doc = "    \"request\","]
#[doc = "    \"reset\","]
#[doc = "    \"restart\","]
#[doc = "    \"restore\","]
#[doc = "    \"retrieve\","]
#[doc = "    \"rollback\","]
#[doc = "    \"run\","]
#[doc = "    \"save\","]
#[doc = "    \"scale\","]
#[doc = "    \"scan\","]
#[doc = "    \"schedule\","]
#[doc = "    \"seed\","]
#[doc = "    \"send\","]
#[doc = "    \"serve\","]
#[doc = "    \"set\","]
#[doc = "    \"setup\","]
#[doc = "    \"show\","]
#[doc = "    \"snapshot\","]
#[doc = "    \"start\","]
#[doc = "    \"stash\","]
#[doc = "    \"stop\","]
#[doc = "    \"tag\","]
#[doc = "    \"teardown\","]
#[doc = "    \"terminate\","]
#[doc = "    \"test\","]
#[doc = "    \"tidy\","]
#[doc = "    \"trace\","]
#[doc = "    \"transform\","]
#[doc = "    \"trigger\","]
#[doc = "    \"tunnel\","]
#[doc = "    \"undeploy\","]
#[doc = "    \"uninstall\","]
#[doc = "    \"unmount\","]
#[doc = "    \"unset\","]
#[doc = "    \"update\","]
#[doc = "    \"upgrade\","]
#[doc = "    \"validate\","]
#[doc = "    \"verify\","]
#[doc = "    \"view\","]
#[doc = "    \"watch\""]
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
pub enum ExecutableVerb {
    #[serde(rename = "abort")]
    Abort,
    #[serde(rename = "activate")]
    Activate,
    #[serde(rename = "add")]
    Add,
    #[serde(rename = "analyze")]
    Analyze,
    #[serde(rename = "apply")]
    Apply,
    #[serde(rename = "archive")]
    Archive,
    #[serde(rename = "audit")]
    Audit,
    #[serde(rename = "backup")]
    Backup,
    #[serde(rename = "benchmark")]
    Benchmark,
    #[serde(rename = "build")]
    Build,
    #[serde(rename = "bundle")]
    Bundle,
    #[serde(rename = "check")]
    Check,
    #[serde(rename = "clean")]
    Clean,
    #[serde(rename = "clear")]
    Clear,
    #[serde(rename = "commit")]
    Commit,
    #[serde(rename = "compile")]
    Compile,
    #[serde(rename = "compress")]
    Compress,
    #[serde(rename = "configure")]
    Configure,
    #[serde(rename = "connect")]
    Connect,
    #[serde(rename = "copy")]
    Copy,
    #[serde(rename = "create")]
    Create,
    #[serde(rename = "deactivate")]
    Deactivate,
    #[serde(rename = "debug")]
    Debug,
    #[serde(rename = "decompress")]
    Decompress,
    #[serde(rename = "decrypt")]
    Decrypt,
    #[serde(rename = "delete")]
    Delete,
    #[serde(rename = "deploy")]
    Deploy,
    #[serde(rename = "destroy")]
    Destroy,
    #[serde(rename = "disable")]
    Disable,
    #[serde(rename = "disconnect")]
    Disconnect,
    #[serde(rename = "edit")]
    Edit,
    #[serde(rename = "enable")]
    Enable,
    #[serde(rename = "encrypt")]
    Encrypt,
    #[serde(rename = "erase")]
    Erase,
    #[serde(rename = "exec")]
    Exec,
    #[serde(rename = "execute")]
    Execute,
    #[serde(rename = "export")]
    Export,
    #[serde(rename = "expose")]
    Expose,
    #[serde(rename = "fetch")]
    Fetch,
    #[serde(rename = "fix")]
    Fix,
    #[serde(rename = "flush")]
    Flush,
    #[serde(rename = "format")]
    Format,
    #[serde(rename = "generate")]
    Generate,
    #[serde(rename = "get")]
    Get,
    #[serde(rename = "import")]
    Import,
    #[serde(rename = "index")]
    Index,
    #[serde(rename = "init")]
    Init,
    #[serde(rename = "inspect")]
    Inspect,
    #[serde(rename = "install")]
    Install,
    #[serde(rename = "join")]
    Join,
    #[serde(rename = "kill")]
    Kill,
    #[serde(rename = "launch")]
    Launch,
    #[serde(rename = "lint")]
    Lint,
    #[serde(rename = "list")]
    List,
    #[serde(rename = "load")]
    Load,
    #[serde(rename = "lock")]
    Lock,
    #[serde(rename = "login")]
    Login,
    #[serde(rename = "logout")]
    Logout,
    #[serde(rename = "manage")]
    Manage,
    #[serde(rename = "merge")]
    Merge,
    #[serde(rename = "migrate")]
    Migrate,
    #[serde(rename = "modify")]
    Modify,
    #[serde(rename = "monitor")]
    Monitor,
    #[serde(rename = "mount")]
    Mount,
    #[serde(rename = "new")]
    New,
    #[serde(rename = "notify")]
    Notify,
    #[serde(rename = "open")]
    Open,
    #[serde(rename = "package")]
    Package,
    #[serde(rename = "partition")]
    Partition,
    #[serde(rename = "patch")]
    Patch,
    #[serde(rename = "pause")]
    Pause,
    #[serde(rename = "ping")]
    Ping,
    #[serde(rename = "preload")]
    Preload,
    #[serde(rename = "prefetch")]
    Prefetch,
    #[serde(rename = "profile")]
    Profile,
    #[serde(rename = "provision")]
    Provision,
    #[serde(rename = "publish")]
    Publish,
    #[serde(rename = "purge")]
    Purge,
    #[serde(rename = "push")]
    Push,
    #[serde(rename = "queue")]
    Queue,
    #[serde(rename = "reboot")]
    Reboot,
    #[serde(rename = "recover")]
    Recover,
    #[serde(rename = "refresh")]
    Refresh,
    #[serde(rename = "release")]
    Release,
    #[serde(rename = "reload")]
    Reload,
    #[serde(rename = "remove")]
    Remove,
    #[serde(rename = "request")]
    Request,
    #[serde(rename = "reset")]
    Reset,
    #[serde(rename = "restart")]
    Restart,
    #[serde(rename = "restore")]
    Restore,
    #[serde(rename = "retrieve")]
    Retrieve,
    #[serde(rename = "rollback")]
    Rollback,
    #[serde(rename = "run")]
    Run,
    #[serde(rename = "save")]
    Save,
    #[serde(rename = "scale")]
    Scale,
    #[serde(rename = "scan")]
    Scan,
    #[serde(rename = "schedule")]
    Schedule,
    #[serde(rename = "seed")]
    Seed,
    #[serde(rename = "send")]
    Send,
    #[serde(rename = "serve")]
    Serve,
    #[serde(rename = "set")]
    Set,
    #[serde(rename = "setup")]
    Setup,
    #[serde(rename = "show")]
    Show,
    #[serde(rename = "snapshot")]
    Snapshot,
    #[serde(rename = "start")]
    Start,
    #[serde(rename = "stash")]
    Stash,
    #[serde(rename = "stop")]
    Stop,
    #[serde(rename = "tag")]
    Tag,
    #[serde(rename = "teardown")]
    Teardown,
    #[serde(rename = "terminate")]
    Terminate,
    #[serde(rename = "test")]
    Test,
    #[serde(rename = "tidy")]
    Tidy,
    #[serde(rename = "trace")]
    Trace,
    #[serde(rename = "transform")]
    Transform,
    #[serde(rename = "trigger")]
    Trigger,
    #[serde(rename = "tunnel")]
    Tunnel,
    #[serde(rename = "undeploy")]
    Undeploy,
    #[serde(rename = "uninstall")]
    Uninstall,
    #[serde(rename = "unmount")]
    Unmount,
    #[serde(rename = "unset")]
    Unset,
    #[serde(rename = "update")]
    Update,
    #[serde(rename = "upgrade")]
    Upgrade,
    #[serde(rename = "validate")]
    Validate,
    #[serde(rename = "verify")]
    Verify,
    #[serde(rename = "view")]
    View,
    #[serde(rename = "watch")]
    Watch,
}
impl ::std::convert::From<&Self> for ExecutableVerb {
    fn from(value: &ExecutableVerb) -> Self {
        value.clone()
    }
}
impl ::std::fmt::Display for ExecutableVerb {
    fn fmt(&self, f: &mut ::std::fmt::Formatter<'_>) -> ::std::fmt::Result {
        match *self {
            Self::Abort => f.write_str("abort"),
            Self::Activate => f.write_str("activate"),
            Self::Add => f.write_str("add"),
            Self::Analyze => f.write_str("analyze"),
            Self::Apply => f.write_str("apply"),
            Self::Archive => f.write_str("archive"),
            Self::Audit => f.write_str("audit"),
            Self::Backup => f.write_str("backup"),
            Self::Benchmark => f.write_str("benchmark"),
            Self::Build => f.write_str("build"),
            Self::Bundle => f.write_str("bundle"),
            Self::Check => f.write_str("check"),
            Self::Clean => f.write_str("clean"),
            Self::Clear => f.write_str("clear"),
            Self::Commit => f.write_str("commit"),
            Self::Compile => f.write_str("compile"),
            Self::Compress => f.write_str("compress"),
            Self::Configure => f.write_str("configure"),
            Self::Connect => f.write_str("connect"),
            Self::Copy => f.write_str("copy"),
            Self::Create => f.write_str("create"),
            Self::Deactivate => f.write_str("deactivate"),
            Self::Debug => f.write_str("debug"),
            Self::Decompress => f.write_str("decompress"),
            Self::Decrypt => f.write_str("decrypt"),
            Self::Delete => f.write_str("delete"),
            Self::Deploy => f.write_str("deploy"),
            Self::Destroy => f.write_str("destroy"),
            Self::Disable => f.write_str("disable"),
            Self::Disconnect => f.write_str("disconnect"),
            Self::Edit => f.write_str("edit"),
            Self::Enable => f.write_str("enable"),
            Self::Encrypt => f.write_str("encrypt"),
            Self::Erase => f.write_str("erase"),
            Self::Exec => f.write_str("exec"),
            Self::Execute => f.write_str("execute"),
            Self::Export => f.write_str("export"),
            Self::Expose => f.write_str("expose"),
            Self::Fetch => f.write_str("fetch"),
            Self::Fix => f.write_str("fix"),
            Self::Flush => f.write_str("flush"),
            Self::Format => f.write_str("format"),
            Self::Generate => f.write_str("generate"),
            Self::Get => f.write_str("get"),
            Self::Import => f.write_str("import"),
            Self::Index => f.write_str("index"),
            Self::Init => f.write_str("init"),
            Self::Inspect => f.write_str("inspect"),
            Self::Install => f.write_str("install"),
            Self::Join => f.write_str("join"),
            Self::Kill => f.write_str("kill"),
            Self::Launch => f.write_str("launch"),
            Self::Lint => f.write_str("lint"),
            Self::List => f.write_str("list"),
            Self::Load => f.write_str("load"),
            Self::Lock => f.write_str("lock"),
            Self::Login => f.write_str("login"),
            Self::Logout => f.write_str("logout"),
            Self::Manage => f.write_str("manage"),
            Self::Merge => f.write_str("merge"),
            Self::Migrate => f.write_str("migrate"),
            Self::Modify => f.write_str("modify"),
            Self::Monitor => f.write_str("monitor"),
            Self::Mount => f.write_str("mount"),
            Self::New => f.write_str("new"),
            Self::Notify => f.write_str("notify"),
            Self::Open => f.write_str("open"),
            Self::Package => f.write_str("package"),
            Self::Partition => f.write_str("partition"),
            Self::Patch => f.write_str("patch"),
            Self::Pause => f.write_str("pause"),
            Self::Ping => f.write_str("ping"),
            Self::Preload => f.write_str("preload"),
            Self::Prefetch => f.write_str("prefetch"),
            Self::Profile => f.write_str("profile"),
            Self::Provision => f.write_str("provision"),
            Self::Publish => f.write_str("publish"),
            Self::Purge => f.write_str("purge"),
            Self::Push => f.write_str("push"),
            Self::Queue => f.write_str("queue"),
            Self::Reboot => f.write_str("reboot"),
            Self::Recover => f.write_str("recover"),
            Self::Refresh => f.write_str("refresh"),
            Self::Release => f.write_str("release"),
            Self::Reload => f.write_str("reload"),
            Self::Remove => f.write_str("remove"),
            Self::Request => f.write_str("request"),
            Self::Reset => f.write_str("reset"),
            Self::Restart => f.write_str("restart"),
            Self::Restore => f.write_str("restore"),
            Self::Retrieve => f.write_str("retrieve"),
            Self::Rollback => f.write_str("rollback"),
            Self::Run => f.write_str("run"),
            Self::Save => f.write_str("save"),
            Self::Scale => f.write_str("scale"),
            Self::Scan => f.write_str("scan"),
            Self::Schedule => f.write_str("schedule"),
            Self::Seed => f.write_str("seed"),
            Self::Send => f.write_str("send"),
            Self::Serve => f.write_str("serve"),
            Self::Set => f.write_str("set"),
            Self::Setup => f.write_str("setup"),
            Self::Show => f.write_str("show"),
            Self::Snapshot => f.write_str("snapshot"),
            Self::Start => f.write_str("start"),
            Self::Stash => f.write_str("stash"),
            Self::Stop => f.write_str("stop"),
            Self::Tag => f.write_str("tag"),
            Self::Teardown => f.write_str("teardown"),
            Self::Terminate => f.write_str("terminate"),
            Self::Test => f.write_str("test"),
            Self::Tidy => f.write_str("tidy"),
            Self::Trace => f.write_str("trace"),
            Self::Transform => f.write_str("transform"),
            Self::Trigger => f.write_str("trigger"),
            Self::Tunnel => f.write_str("tunnel"),
            Self::Undeploy => f.write_str("undeploy"),
            Self::Uninstall => f.write_str("uninstall"),
            Self::Unmount => f.write_str("unmount"),
            Self::Unset => f.write_str("unset"),
            Self::Update => f.write_str("update"),
            Self::Upgrade => f.write_str("upgrade"),
            Self::Validate => f.write_str("validate"),
            Self::Verify => f.write_str("verify"),
            Self::View => f.write_str("view"),
            Self::Watch => f.write_str("watch"),
        }
    }
}
impl ::std::str::FromStr for ExecutableVerb {
    type Err = self::error::ConversionError;
    fn from_str(value: &str) -> ::std::result::Result<Self, self::error::ConversionError> {
        match value {
            "abort" => Ok(Self::Abort),
            "activate" => Ok(Self::Activate),
            "add" => Ok(Self::Add),
            "analyze" => Ok(Self::Analyze),
            "apply" => Ok(Self::Apply),
            "archive" => Ok(Self::Archive),
            "audit" => Ok(Self::Audit),
            "backup" => Ok(Self::Backup),
            "benchmark" => Ok(Self::Benchmark),
            "build" => Ok(Self::Build),
            "bundle" => Ok(Self::Bundle),
            "check" => Ok(Self::Check),
            "clean" => Ok(Self::Clean),
            "clear" => Ok(Self::Clear),
            "commit" => Ok(Self::Commit),
            "compile" => Ok(Self::Compile),
            "compress" => Ok(Self::Compress),
            "configure" => Ok(Self::Configure),
            "connect" => Ok(Self::Connect),
            "copy" => Ok(Self::Copy),
            "create" => Ok(Self::Create),
            "deactivate" => Ok(Self::Deactivate),
            "debug" => Ok(Self::Debug),
            "decompress" => Ok(Self::Decompress),
            "decrypt" => Ok(Self::Decrypt),
            "delete" => Ok(Self::Delete),
            "deploy" => Ok(Self::Deploy),
            "destroy" => Ok(Self::Destroy),
            "disable" => Ok(Self::Disable),
            "disconnect" => Ok(Self::Disconnect),
            "edit" => Ok(Self::Edit),
            "enable" => Ok(Self::Enable),
            "encrypt" => Ok(Self::Encrypt),
            "erase" => Ok(Self::Erase),
            "exec" => Ok(Self::Exec),
            "execute" => Ok(Self::Execute),
            "export" => Ok(Self::Export),
            "expose" => Ok(Self::Expose),
            "fetch" => Ok(Self::Fetch),
            "fix" => Ok(Self::Fix),
            "flush" => Ok(Self::Flush),
            "format" => Ok(Self::Format),
            "generate" => Ok(Self::Generate),
            "get" => Ok(Self::Get),
            "import" => Ok(Self::Import),
            "index" => Ok(Self::Index),
            "init" => Ok(Self::Init),
            "inspect" => Ok(Self::Inspect),
            "install" => Ok(Self::Install),
            "join" => Ok(Self::Join),
            "kill" => Ok(Self::Kill),
            "launch" => Ok(Self::Launch),
            "lint" => Ok(Self::Lint),
            "list" => Ok(Self::List),
            "load" => Ok(Self::Load),
            "lock" => Ok(Self::Lock),
            "login" => Ok(Self::Login),
            "logout" => Ok(Self::Logout),
            "manage" => Ok(Self::Manage),
            "merge" => Ok(Self::Merge),
            "migrate" => Ok(Self::Migrate),
            "modify" => Ok(Self::Modify),
            "monitor" => Ok(Self::Monitor),
            "mount" => Ok(Self::Mount),
            "new" => Ok(Self::New),
            "notify" => Ok(Self::Notify),
            "open" => Ok(Self::Open),
            "package" => Ok(Self::Package),
            "partition" => Ok(Self::Partition),
            "patch" => Ok(Self::Patch),
            "pause" => Ok(Self::Pause),
            "ping" => Ok(Self::Ping),
            "preload" => Ok(Self::Preload),
            "prefetch" => Ok(Self::Prefetch),
            "profile" => Ok(Self::Profile),
            "provision" => Ok(Self::Provision),
            "publish" => Ok(Self::Publish),
            "purge" => Ok(Self::Purge),
            "push" => Ok(Self::Push),
            "queue" => Ok(Self::Queue),
            "reboot" => Ok(Self::Reboot),
            "recover" => Ok(Self::Recover),
            "refresh" => Ok(Self::Refresh),
            "release" => Ok(Self::Release),
            "reload" => Ok(Self::Reload),
            "remove" => Ok(Self::Remove),
            "request" => Ok(Self::Request),
            "reset" => Ok(Self::Reset),
            "restart" => Ok(Self::Restart),
            "restore" => Ok(Self::Restore),
            "retrieve" => Ok(Self::Retrieve),
            "rollback" => Ok(Self::Rollback),
            "run" => Ok(Self::Run),
            "save" => Ok(Self::Save),
            "scale" => Ok(Self::Scale),
            "scan" => Ok(Self::Scan),
            "schedule" => Ok(Self::Schedule),
            "seed" => Ok(Self::Seed),
            "send" => Ok(Self::Send),
            "serve" => Ok(Self::Serve),
            "set" => Ok(Self::Set),
            "setup" => Ok(Self::Setup),
            "show" => Ok(Self::Show),
            "snapshot" => Ok(Self::Snapshot),
            "start" => Ok(Self::Start),
            "stash" => Ok(Self::Stash),
            "stop" => Ok(Self::Stop),
            "tag" => Ok(Self::Tag),
            "teardown" => Ok(Self::Teardown),
            "terminate" => Ok(Self::Terminate),
            "test" => Ok(Self::Test),
            "tidy" => Ok(Self::Tidy),
            "trace" => Ok(Self::Trace),
            "transform" => Ok(Self::Transform),
            "trigger" => Ok(Self::Trigger),
            "tunnel" => Ok(Self::Tunnel),
            "undeploy" => Ok(Self::Undeploy),
            "uninstall" => Ok(Self::Uninstall),
            "unmount" => Ok(Self::Unmount),
            "unset" => Ok(Self::Unset),
            "update" => Ok(Self::Update),
            "upgrade" => Ok(Self::Upgrade),
            "validate" => Ok(Self::Validate),
            "verify" => Ok(Self::Verify),
            "view" => Ok(Self::View),
            "watch" => Ok(Self::Watch),
            _ => Err("invalid value".into()),
        }
    }
}
impl ::std::convert::TryFrom<&str> for ExecutableVerb {
    type Error = self::error::ConversionError;
    fn try_from(value: &str) -> ::std::result::Result<Self, self::error::ConversionError> {
        value.parse()
    }
}
impl ::std::convert::TryFrom<&::std::string::String> for ExecutableVerb {
    type Error = self::error::ConversionError;
    fn try_from(
        value: &::std::string::String,
    ) -> ::std::result::Result<Self, self::error::ConversionError> {
        value.parse()
    }
}
impl ::std::convert::TryFrom<::std::string::String> for ExecutableVerb {
    type Error = self::error::ConversionError;
    fn try_from(
        value: ::std::string::String,
    ) -> ::std::result::Result<Self, self::error::ConversionError> {
        value.parse()
    }
}
impl ::std::default::Default for ExecutableVerb {
    fn default() -> Self {
        ExecutableVerb::Exec
    }
}
#[doc = "Configuration for a group of Flow CLI executables. The file must have the extension `.flow`, `.flow.yaml`, or `.flow.yml` \nin order to be discovered by the CLI. It's configuration is used to define a group of executables with shared metadata \n(namespace, tags, etc). A workspace can have multiple flow files located anywhere in the workspace directory\n"]
#[doc = r""]
#[doc = r" <details><summary>JSON schema</summary>"]
#[doc = r""]
#[doc = r" ```json"]
#[doc = "{"]
#[doc = "  \"$id\": \"https://flowexec.io/schemas/flowfile_schema.json\","]
#[doc = "  \"title\": \"FlowFile\","]
#[doc = "  \"description\": \"Configuration for a group of Flow CLI executables. The file must have the extension `.flow`, `.flow.yaml`, or `.flow.yml` \\nin order to be discovered by the CLI. It's configuration is used to define a group of executables with shared metadata \\n(namespace, tags, etc). A workspace can have multiple flow files located anywhere in the workspace directory\\n\","]
#[doc = "  \"type\": \"object\","]
#[doc = "  \"properties\": {"]
#[doc = "    \"description\": {"]
#[doc = "      \"description\": \"A description of the executables defined within the flow file. This description will used as a shared description\\nfor all executables in the flow file.\\n\","]
#[doc = "      \"default\": \"\","]
#[doc = "      \"type\": \"string\""]
#[doc = "    },"]
#[doc = "    \"descriptionFile\": {"]
#[doc = "      \"description\": \"A path to a markdown file that contains the description of the executables defined within the flow file.\","]
#[doc = "      \"default\": \"\","]
#[doc = "      \"type\": \"string\""]
#[doc = "    },"]
#[doc = "    \"executables\": {"]
#[doc = "      \"default\": [],"]
#[doc = "      \"type\": \"array\","]
#[doc = "      \"items\": {"]
#[doc = "        \"$ref\": \"#/definitions/Executable\""]
#[doc = "      }"]
#[doc = "    },"]
#[doc = "    \"fromFile\": {"]
#[doc = "      \"description\": \"DEPRECATED: Use `imports` instead\","]
#[doc = "      \"default\": [],"]
#[doc = "      \"$ref\": \"#/definitions/FromFile\""]
#[doc = "    },"]
#[doc = "    \"imports\": {"]
#[doc = "      \"default\": [],"]
#[doc = "      \"$ref\": \"#/definitions/FromFile\""]
#[doc = "    },"]
#[doc = "    \"namespace\": {"]
#[doc = "      \"description\": \"The namespace to be given to all executables in the flow file.\\nIf not set, the executables in the file will be grouped into the root (*) namespace. \\nNamespaces can be reused across multiple flow files.\\n\\nNamespaces are used to reference executables in the CLI using the format `workspace:namespace/name`.\\n\","]
#[doc = "      \"default\": \"\","]
#[doc = "      \"type\": \"string\""]
#[doc = "    },"]
#[doc = "    \"tags\": {"]
#[doc = "      \"description\": \"Tags to be applied to all executables defined within the flow file.\","]
#[doc = "      \"default\": [],"]
#[doc = "      \"type\": \"array\","]
#[doc = "      \"items\": {"]
#[doc = "        \"type\": \"string\""]
#[doc = "      }"]
#[doc = "    },"]
#[doc = "    \"visibility\": {"]
#[doc = "      \"$ref\": \"#/definitions/CommonVisibility\""]
#[doc = "    }"]
#[doc = "  }"]
#[doc = "}"]
#[doc = r" ```"]
#[doc = r" </details>"]
#[derive(:: serde :: Deserialize, :: serde :: Serialize, Clone, Debug)]
pub struct FlowFile {
    #[doc = "A description of the executables defined within the flow file. This description will used as a shared description\nfor all executables in the flow file.\n"]
    #[serde(default)]
    pub description: ::std::string::String,
    #[doc = "A path to a markdown file that contains the description of the executables defined within the flow file."]
    #[serde(rename = "descriptionFile", default)]
    pub description_file: ::std::string::String,
    #[serde(default, skip_serializing_if = "::std::vec::Vec::is_empty")]
    pub executables: ::std::vec::Vec<Executable>,
    #[doc = "DEPRECATED: Use `imports` instead"]
    #[serde(rename = "fromFile", default = "defaults::flow_file_from_file")]
    pub from_file: FromFile,
    #[serde(default = "defaults::flow_file_imports")]
    pub imports: FromFile,
    #[doc = "The namespace to be given to all executables in the flow file.\nIf not set, the executables in the file will be grouped into the root (*) namespace. \nNamespaces can be reused across multiple flow files.\n\nNamespaces are used to reference executables in the CLI using the format `workspace:namespace/name`.\n"]
    #[serde(default)]
    pub namespace: ::std::string::String,
    #[doc = "Tags to be applied to all executables defined within the flow file."]
    #[serde(default, skip_serializing_if = "::std::vec::Vec::is_empty")]
    pub tags: ::std::vec::Vec<::std::string::String>,
    #[serde(default, skip_serializing_if = "::std::option::Option::is_none")]
    pub visibility: ::std::option::Option<CommonVisibility>,
}
impl ::std::convert::From<&FlowFile> for FlowFile {
    fn from(value: &FlowFile) -> Self {
        value.clone()
    }
}
impl ::std::default::Default for FlowFile {
    fn default() -> Self {
        Self {
            description: Default::default(),
            description_file: Default::default(),
            executables: Default::default(),
            from_file: defaults::flow_file_from_file(),
            imports: defaults::flow_file_imports(),
            namespace: Default::default(),
            tags: Default::default(),
            visibility: Default::default(),
        }
    }
}
impl FlowFile {
    pub fn builder() -> builder::FlowFile {
        Default::default()
    }
}
#[doc = "A list of `.sh` files to convert into generated executables in the file's executable group."]
#[doc = r""]
#[doc = r" <details><summary>JSON schema</summary>"]
#[doc = r""]
#[doc = r" ```json"]
#[doc = "{"]
#[doc = "  \"description\": \"A list of `.sh` files to convert into generated executables in the file's executable group.\","]
#[doc = "  \"default\": [],"]
#[doc = "  \"type\": \"array\","]
#[doc = "  \"items\": {"]
#[doc = "    \"type\": \"string\""]
#[doc = "  }"]
#[doc = "}"]
#[doc = r" ```"]
#[doc = r" </details>"]
#[derive(:: serde :: Deserialize, :: serde :: Serialize, Clone, Debug)]
#[serde(transparent)]
pub struct FromFile(pub ::std::vec::Vec<::std::string::String>);
impl ::std::ops::Deref for FromFile {
    type Target = ::std::vec::Vec<::std::string::String>;
    fn deref(&self) -> &::std::vec::Vec<::std::string::String> {
        &self.0
    }
}
impl ::std::convert::From<FromFile> for ::std::vec::Vec<::std::string::String> {
    fn from(value: FromFile) -> Self {
        value.0
    }
}
impl ::std::convert::From<&FromFile> for FromFile {
    fn from(value: &FromFile) -> Self {
        value.clone()
    }
}
impl ::std::convert::From<::std::vec::Vec<::std::string::String>> for FromFile {
    fn from(value: ::std::vec::Vec<::std::string::String>) -> Self {
        Self(value)
    }
}
#[doc = "`Ref`"]
#[doc = r""]
#[doc = r" <details><summary>JSON schema</summary>"]
#[doc = r""]
#[doc = r" ```json"]
#[doc = "{}"]
#[doc = r" ```"]
#[doc = r" </details>"]
#[derive(:: serde :: Deserialize, :: serde :: Serialize, Clone, Debug)]
#[serde(transparent)]
pub struct Ref(pub ::serde_json::Value);
impl ::std::ops::Deref for Ref {
    type Target = ::serde_json::Value;
    fn deref(&self) -> &::serde_json::Value {
        &self.0
    }
}
impl ::std::convert::From<Ref> for ::serde_json::Value {
    fn from(value: Ref) -> Self {
        value.0
    }
}
impl ::std::convert::From<&Ref> for Ref {
    fn from(value: &Ref) -> Self {
        value.clone()
    }
}
impl ::std::convert::From<::serde_json::Value> for Ref {
    fn from(value: ::serde_json::Value) -> Self {
        Self(value)
    }
}
#[doc = "`Verb`"]
#[doc = r""]
#[doc = r" <details><summary>JSON schema</summary>"]
#[doc = r""]
#[doc = r" ```json"]
#[doc = "{}"]
#[doc = r" ```"]
#[doc = r" </details>"]
#[derive(:: serde :: Deserialize, :: serde :: Serialize, Clone, Debug)]
#[serde(transparent)]
pub struct Verb(pub ::serde_json::Value);
impl ::std::ops::Deref for Verb {
    type Target = ::serde_json::Value;
    fn deref(&self) -> &::serde_json::Value {
        &self.0
    }
}
impl ::std::convert::From<Verb> for ::serde_json::Value {
    fn from(value: Verb) -> Self {
        value.0
    }
}
impl ::std::convert::From<&Verb> for Verb {
    fn from(value: &Verb) -> Self {
        value.clone()
    }
}
impl ::std::convert::From<::serde_json::Value> for Verb {
    fn from(value: ::serde_json::Value) -> Self {
        Self(value)
    }
}
#[doc = r" Types for composing complex structures."]
pub mod builder {
    #[derive(Clone, Debug)]
    pub struct Executable {
        aliases: ::std::result::Result<super::CommonAliases, ::std::string::String>,
        description: ::std::result::Result<::std::string::String, ::std::string::String>,
        exec: ::std::result::Result<
            ::std::option::Option<super::ExecutableExecExecutableType>,
            ::std::string::String,
        >,
        launch: ::std::result::Result<
            ::std::option::Option<super::ExecutableLaunchExecutableType>,
            ::std::string::String,
        >,
        name: ::std::result::Result<::std::string::String, ::std::string::String>,
        parallel: ::std::result::Result<
            ::std::option::Option<super::ExecutableParallelExecutableType>,
            ::std::string::String,
        >,
        render: ::std::result::Result<
            ::std::option::Option<super::ExecutableRenderExecutableType>,
            ::std::string::String,
        >,
        request: ::std::result::Result<
            ::std::option::Option<super::ExecutableRequestExecutableType>,
            ::std::string::String,
        >,
        serial: ::std::result::Result<
            ::std::option::Option<super::ExecutableSerialExecutableType>,
            ::std::string::String,
        >,
        tags: ::std::result::Result<super::CommonTags, ::std::string::String>,
        timeout: ::std::result::Result<
            ::std::option::Option<::std::string::String>,
            ::std::string::String,
        >,
        verb: ::std::result::Result<super::ExecutableVerb, ::std::string::String>,
        verb_aliases: ::std::result::Result<::std::vec::Vec<super::Verb>, ::std::string::String>,
        visibility: ::std::result::Result<
            ::std::option::Option<super::CommonVisibility>,
            ::std::string::String,
        >,
    }
    impl ::std::default::Default for Executable {
        fn default() -> Self {
            Self {
                aliases: Ok(super::defaults::executable_aliases()),
                description: Ok(Default::default()),
                exec: Ok(Default::default()),
                launch: Ok(Default::default()),
                name: Ok(Default::default()),
                parallel: Ok(Default::default()),
                render: Ok(Default::default()),
                request: Ok(Default::default()),
                serial: Ok(Default::default()),
                tags: Ok(super::defaults::executable_tags()),
                timeout: Ok(Default::default()),
                verb: Err("no value supplied for verb".to_string()),
                verb_aliases: Ok(Default::default()),
                visibility: Ok(Default::default()),
            }
        }
    }
    impl Executable {
        pub fn aliases<T>(mut self, value: T) -> Self
        where
            T: ::std::convert::TryInto<super::CommonAliases>,
            T::Error: ::std::fmt::Display,
        {
            self.aliases = value
                .try_into()
                .map_err(|e| format!("error converting supplied value for aliases: {}", e));
            self
        }
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
        pub fn exec<T>(mut self, value: T) -> Self
        where
            T: ::std::convert::TryInto<::std::option::Option<super::ExecutableExecExecutableType>>,
            T::Error: ::std::fmt::Display,
        {
            self.exec = value
                .try_into()
                .map_err(|e| format!("error converting supplied value for exec: {}", e));
            self
        }
        pub fn launch<T>(mut self, value: T) -> Self
        where
            T: ::std::convert::TryInto<
                ::std::option::Option<super::ExecutableLaunchExecutableType>,
            >,
            T::Error: ::std::fmt::Display,
        {
            self.launch = value
                .try_into()
                .map_err(|e| format!("error converting supplied value for launch: {}", e));
            self
        }
        pub fn name<T>(mut self, value: T) -> Self
        where
            T: ::std::convert::TryInto<::std::string::String>,
            T::Error: ::std::fmt::Display,
        {
            self.name = value
                .try_into()
                .map_err(|e| format!("error converting supplied value for name: {}", e));
            self
        }
        pub fn parallel<T>(mut self, value: T) -> Self
        where
            T: ::std::convert::TryInto<
                ::std::option::Option<super::ExecutableParallelExecutableType>,
            >,
            T::Error: ::std::fmt::Display,
        {
            self.parallel = value
                .try_into()
                .map_err(|e| format!("error converting supplied value for parallel: {}", e));
            self
        }
        pub fn render<T>(mut self, value: T) -> Self
        where
            T: ::std::convert::TryInto<
                ::std::option::Option<super::ExecutableRenderExecutableType>,
            >,
            T::Error: ::std::fmt::Display,
        {
            self.render = value
                .try_into()
                .map_err(|e| format!("error converting supplied value for render: {}", e));
            self
        }
        pub fn request<T>(mut self, value: T) -> Self
        where
            T: ::std::convert::TryInto<
                ::std::option::Option<super::ExecutableRequestExecutableType>,
            >,
            T::Error: ::std::fmt::Display,
        {
            self.request = value
                .try_into()
                .map_err(|e| format!("error converting supplied value for request: {}", e));
            self
        }
        pub fn serial<T>(mut self, value: T) -> Self
        where
            T: ::std::convert::TryInto<
                ::std::option::Option<super::ExecutableSerialExecutableType>,
            >,
            T::Error: ::std::fmt::Display,
        {
            self.serial = value
                .try_into()
                .map_err(|e| format!("error converting supplied value for serial: {}", e));
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
        pub fn timeout<T>(mut self, value: T) -> Self
        where
            T: ::std::convert::TryInto<::std::option::Option<::std::string::String>>,
            T::Error: ::std::fmt::Display,
        {
            self.timeout = value
                .try_into()
                .map_err(|e| format!("error converting supplied value for timeout: {}", e));
            self
        }
        pub fn verb<T>(mut self, value: T) -> Self
        where
            T: ::std::convert::TryInto<super::ExecutableVerb>,
            T::Error: ::std::fmt::Display,
        {
            self.verb = value
                .try_into()
                .map_err(|e| format!("error converting supplied value for verb: {}", e));
            self
        }
        pub fn verb_aliases<T>(mut self, value: T) -> Self
        where
            T: ::std::convert::TryInto<::std::vec::Vec<super::Verb>>,
            T::Error: ::std::fmt::Display,
        {
            self.verb_aliases = value
                .try_into()
                .map_err(|e| format!("error converting supplied value for verb_aliases: {}", e));
            self
        }
        pub fn visibility<T>(mut self, value: T) -> Self
        where
            T: ::std::convert::TryInto<::std::option::Option<super::CommonVisibility>>,
            T::Error: ::std::fmt::Display,
        {
            self.visibility = value
                .try_into()
                .map_err(|e| format!("error converting supplied value for visibility: {}", e));
            self
        }
    }
    impl ::std::convert::TryFrom<Executable> for super::Executable {
        type Error = super::error::ConversionError;
        fn try_from(
            value: Executable,
        ) -> ::std::result::Result<Self, super::error::ConversionError> {
            Ok(Self {
                aliases: value.aliases?,
                description: value.description?,
                exec: value.exec?,
                launch: value.launch?,
                name: value.name?,
                parallel: value.parallel?,
                render: value.render?,
                request: value.request?,
                serial: value.serial?,
                tags: value.tags?,
                timeout: value.timeout?,
                verb: value.verb?,
                verb_aliases: value.verb_aliases?,
                visibility: value.visibility?,
            })
        }
    }
    impl ::std::convert::From<super::Executable> for Executable {
        fn from(value: super::Executable) -> Self {
            Self {
                aliases: Ok(value.aliases),
                description: Ok(value.description),
                exec: Ok(value.exec),
                launch: Ok(value.launch),
                name: Ok(value.name),
                parallel: Ok(value.parallel),
                render: Ok(value.render),
                request: Ok(value.request),
                serial: Ok(value.serial),
                tags: Ok(value.tags),
                timeout: Ok(value.timeout),
                verb: Ok(value.verb),
                verb_aliases: Ok(value.verb_aliases),
                visibility: Ok(value.visibility),
            }
        }
    }
    #[derive(Clone, Debug)]
    pub struct ExecutableArgument {
        default: ::std::result::Result<::std::string::String, ::std::string::String>,
        env_key: ::std::result::Result<::std::string::String, ::std::string::String>,
        flag: ::std::result::Result<::std::string::String, ::std::string::String>,
        output_file: ::std::result::Result<::std::string::String, ::std::string::String>,
        pos: ::std::result::Result<::std::option::Option<i64>, ::std::string::String>,
        required: ::std::result::Result<bool, ::std::string::String>,
        type_: ::std::result::Result<super::ExecutableArgumentType, ::std::string::String>,
    }
    impl ::std::default::Default for ExecutableArgument {
        fn default() -> Self {
            Self {
                default: Ok(Default::default()),
                env_key: Ok(Default::default()),
                flag: Ok(Default::default()),
                output_file: Ok(Default::default()),
                pos: Ok(Default::default()),
                required: Ok(Default::default()),
                type_: Ok(super::defaults::executable_argument_type()),
            }
        }
    }
    impl ExecutableArgument {
        pub fn default<T>(mut self, value: T) -> Self
        where
            T: ::std::convert::TryInto<::std::string::String>,
            T::Error: ::std::fmt::Display,
        {
            self.default = value
                .try_into()
                .map_err(|e| format!("error converting supplied value for default: {}", e));
            self
        }
        pub fn env_key<T>(mut self, value: T) -> Self
        where
            T: ::std::convert::TryInto<::std::string::String>,
            T::Error: ::std::fmt::Display,
        {
            self.env_key = value
                .try_into()
                .map_err(|e| format!("error converting supplied value for env_key: {}", e));
            self
        }
        pub fn flag<T>(mut self, value: T) -> Self
        where
            T: ::std::convert::TryInto<::std::string::String>,
            T::Error: ::std::fmt::Display,
        {
            self.flag = value
                .try_into()
                .map_err(|e| format!("error converting supplied value for flag: {}", e));
            self
        }
        pub fn output_file<T>(mut self, value: T) -> Self
        where
            T: ::std::convert::TryInto<::std::string::String>,
            T::Error: ::std::fmt::Display,
        {
            self.output_file = value
                .try_into()
                .map_err(|e| format!("error converting supplied value for output_file: {}", e));
            self
        }
        pub fn pos<T>(mut self, value: T) -> Self
        where
            T: ::std::convert::TryInto<::std::option::Option<i64>>,
            T::Error: ::std::fmt::Display,
        {
            self.pos = value
                .try_into()
                .map_err(|e| format!("error converting supplied value for pos: {}", e));
            self
        }
        pub fn required<T>(mut self, value: T) -> Self
        where
            T: ::std::convert::TryInto<bool>,
            T::Error: ::std::fmt::Display,
        {
            self.required = value
                .try_into()
                .map_err(|e| format!("error converting supplied value for required: {}", e));
            self
        }
        pub fn type_<T>(mut self, value: T) -> Self
        where
            T: ::std::convert::TryInto<super::ExecutableArgumentType>,
            T::Error: ::std::fmt::Display,
        {
            self.type_ = value
                .try_into()
                .map_err(|e| format!("error converting supplied value for type_: {}", e));
            self
        }
    }
    impl ::std::convert::TryFrom<ExecutableArgument> for super::ExecutableArgument {
        type Error = super::error::ConversionError;
        fn try_from(
            value: ExecutableArgument,
        ) -> ::std::result::Result<Self, super::error::ConversionError> {
            Ok(Self {
                default: value.default?,
                env_key: value.env_key?,
                flag: value.flag?,
                output_file: value.output_file?,
                pos: value.pos?,
                required: value.required?,
                type_: value.type_?,
            })
        }
    }
    impl ::std::convert::From<super::ExecutableArgument> for ExecutableArgument {
        fn from(value: super::ExecutableArgument) -> Self {
            Self {
                default: Ok(value.default),
                env_key: Ok(value.env_key),
                flag: Ok(value.flag),
                output_file: Ok(value.output_file),
                pos: Ok(value.pos),
                required: Ok(value.required),
                type_: Ok(value.type_),
            }
        }
    }
    #[derive(Clone, Debug)]
    pub struct ExecutableExecExecutableType {
        args: ::std::result::Result<
            ::std::option::Option<super::ExecutableArgumentList>,
            ::std::string::String,
        >,
        cmd: ::std::result::Result<::std::string::String, ::std::string::String>,
        dir: ::std::result::Result<super::ExecutableDirectory, ::std::string::String>,
        file: ::std::result::Result<::std::string::String, ::std::string::String>,
        log_mode: ::std::result::Result<::std::string::String, ::std::string::String>,
        params: ::std::result::Result<
            ::std::option::Option<super::ExecutableParameterList>,
            ::std::string::String,
        >,
    }
    impl ::std::default::Default for ExecutableExecExecutableType {
        fn default() -> Self {
            Self {
                args: Ok(Default::default()),
                cmd: Ok(Default::default()),
                dir: Ok(super::defaults::executable_exec_executable_type_dir()),
                file: Ok(Default::default()),
                log_mode: Ok(super::defaults::executable_exec_executable_type_log_mode()),
                params: Ok(Default::default()),
            }
        }
    }
    impl ExecutableExecExecutableType {
        pub fn args<T>(mut self, value: T) -> Self
        where
            T: ::std::convert::TryInto<::std::option::Option<super::ExecutableArgumentList>>,
            T::Error: ::std::fmt::Display,
        {
            self.args = value
                .try_into()
                .map_err(|e| format!("error converting supplied value for args: {}", e));
            self
        }
        pub fn cmd<T>(mut self, value: T) -> Self
        where
            T: ::std::convert::TryInto<::std::string::String>,
            T::Error: ::std::fmt::Display,
        {
            self.cmd = value
                .try_into()
                .map_err(|e| format!("error converting supplied value for cmd: {}", e));
            self
        }
        pub fn dir<T>(mut self, value: T) -> Self
        where
            T: ::std::convert::TryInto<super::ExecutableDirectory>,
            T::Error: ::std::fmt::Display,
        {
            self.dir = value
                .try_into()
                .map_err(|e| format!("error converting supplied value for dir: {}", e));
            self
        }
        pub fn file<T>(mut self, value: T) -> Self
        where
            T: ::std::convert::TryInto<::std::string::String>,
            T::Error: ::std::fmt::Display,
        {
            self.file = value
                .try_into()
                .map_err(|e| format!("error converting supplied value for file: {}", e));
            self
        }
        pub fn log_mode<T>(mut self, value: T) -> Self
        where
            T: ::std::convert::TryInto<::std::string::String>,
            T::Error: ::std::fmt::Display,
        {
            self.log_mode = value
                .try_into()
                .map_err(|e| format!("error converting supplied value for log_mode: {}", e));
            self
        }
        pub fn params<T>(mut self, value: T) -> Self
        where
            T: ::std::convert::TryInto<::std::option::Option<super::ExecutableParameterList>>,
            T::Error: ::std::fmt::Display,
        {
            self.params = value
                .try_into()
                .map_err(|e| format!("error converting supplied value for params: {}", e));
            self
        }
    }
    impl ::std::convert::TryFrom<ExecutableExecExecutableType> for super::ExecutableExecExecutableType {
        type Error = super::error::ConversionError;
        fn try_from(
            value: ExecutableExecExecutableType,
        ) -> ::std::result::Result<Self, super::error::ConversionError> {
            Ok(Self {
                args: value.args?,
                cmd: value.cmd?,
                dir: value.dir?,
                file: value.file?,
                log_mode: value.log_mode?,
                params: value.params?,
            })
        }
    }
    impl ::std::convert::From<super::ExecutableExecExecutableType> for ExecutableExecExecutableType {
        fn from(value: super::ExecutableExecExecutableType) -> Self {
            Self {
                args: Ok(value.args),
                cmd: Ok(value.cmd),
                dir: Ok(value.dir),
                file: Ok(value.file),
                log_mode: Ok(value.log_mode),
                params: Ok(value.params),
            }
        }
    }
    #[derive(Clone, Debug)]
    pub struct ExecutableLaunchExecutableType {
        app: ::std::result::Result<::std::string::String, ::std::string::String>,
        args: ::std::result::Result<
            ::std::option::Option<super::ExecutableArgumentList>,
            ::std::string::String,
        >,
        params: ::std::result::Result<
            ::std::option::Option<super::ExecutableParameterList>,
            ::std::string::String,
        >,
        uri: ::std::result::Result<::std::string::String, ::std::string::String>,
    }
    impl ::std::default::Default for ExecutableLaunchExecutableType {
        fn default() -> Self {
            Self {
                app: Ok(Default::default()),
                args: Ok(Default::default()),
                params: Ok(Default::default()),
                uri: Err("no value supplied for uri".to_string()),
            }
        }
    }
    impl ExecutableLaunchExecutableType {
        pub fn app<T>(mut self, value: T) -> Self
        where
            T: ::std::convert::TryInto<::std::string::String>,
            T::Error: ::std::fmt::Display,
        {
            self.app = value
                .try_into()
                .map_err(|e| format!("error converting supplied value for app: {}", e));
            self
        }
        pub fn args<T>(mut self, value: T) -> Self
        where
            T: ::std::convert::TryInto<::std::option::Option<super::ExecutableArgumentList>>,
            T::Error: ::std::fmt::Display,
        {
            self.args = value
                .try_into()
                .map_err(|e| format!("error converting supplied value for args: {}", e));
            self
        }
        pub fn params<T>(mut self, value: T) -> Self
        where
            T: ::std::convert::TryInto<::std::option::Option<super::ExecutableParameterList>>,
            T::Error: ::std::fmt::Display,
        {
            self.params = value
                .try_into()
                .map_err(|e| format!("error converting supplied value for params: {}", e));
            self
        }
        pub fn uri<T>(mut self, value: T) -> Self
        where
            T: ::std::convert::TryInto<::std::string::String>,
            T::Error: ::std::fmt::Display,
        {
            self.uri = value
                .try_into()
                .map_err(|e| format!("error converting supplied value for uri: {}", e));
            self
        }
    }
    impl ::std::convert::TryFrom<ExecutableLaunchExecutableType>
        for super::ExecutableLaunchExecutableType
    {
        type Error = super::error::ConversionError;
        fn try_from(
            value: ExecutableLaunchExecutableType,
        ) -> ::std::result::Result<Self, super::error::ConversionError> {
            Ok(Self {
                app: value.app?,
                args: value.args?,
                params: value.params?,
                uri: value.uri?,
            })
        }
    }
    impl ::std::convert::From<super::ExecutableLaunchExecutableType>
        for ExecutableLaunchExecutableType
    {
        fn from(value: super::ExecutableLaunchExecutableType) -> Self {
            Self {
                app: Ok(value.app),
                args: Ok(value.args),
                params: Ok(value.params),
                uri: Ok(value.uri),
            }
        }
    }
    #[derive(Clone, Debug)]
    pub struct ExecutableParallelExecutableType {
        args: ::std::result::Result<
            ::std::option::Option<super::ExecutableArgumentList>,
            ::std::string::String,
        >,
        dir: ::std::result::Result<super::ExecutableDirectory, ::std::string::String>,
        execs: ::std::result::Result<super::ExecutableParallelRefConfigList, ::std::string::String>,
        fail_fast: ::std::result::Result<::std::option::Option<bool>, ::std::string::String>,
        max_threads: ::std::result::Result<i64, ::std::string::String>,
        params: ::std::result::Result<
            ::std::option::Option<super::ExecutableParameterList>,
            ::std::string::String,
        >,
    }
    impl ::std::default::Default for ExecutableParallelExecutableType {
        fn default() -> Self {
            Self {
                args: Ok(Default::default()),
                dir: Ok(super::defaults::executable_parallel_executable_type_dir()),
                execs: Err("no value supplied for execs".to_string()),
                fail_fast: Ok(Default::default()),
                max_threads: Ok(super::defaults::default_u64::<i64, 5>()),
                params: Ok(Default::default()),
            }
        }
    }
    impl ExecutableParallelExecutableType {
        pub fn args<T>(mut self, value: T) -> Self
        where
            T: ::std::convert::TryInto<::std::option::Option<super::ExecutableArgumentList>>,
            T::Error: ::std::fmt::Display,
        {
            self.args = value
                .try_into()
                .map_err(|e| format!("error converting supplied value for args: {}", e));
            self
        }
        pub fn dir<T>(mut self, value: T) -> Self
        where
            T: ::std::convert::TryInto<super::ExecutableDirectory>,
            T::Error: ::std::fmt::Display,
        {
            self.dir = value
                .try_into()
                .map_err(|e| format!("error converting supplied value for dir: {}", e));
            self
        }
        pub fn execs<T>(mut self, value: T) -> Self
        where
            T: ::std::convert::TryInto<super::ExecutableParallelRefConfigList>,
            T::Error: ::std::fmt::Display,
        {
            self.execs = value
                .try_into()
                .map_err(|e| format!("error converting supplied value for execs: {}", e));
            self
        }
        pub fn fail_fast<T>(mut self, value: T) -> Self
        where
            T: ::std::convert::TryInto<::std::option::Option<bool>>,
            T::Error: ::std::fmt::Display,
        {
            self.fail_fast = value
                .try_into()
                .map_err(|e| format!("error converting supplied value for fail_fast: {}", e));
            self
        }
        pub fn max_threads<T>(mut self, value: T) -> Self
        where
            T: ::std::convert::TryInto<i64>,
            T::Error: ::std::fmt::Display,
        {
            self.max_threads = value
                .try_into()
                .map_err(|e| format!("error converting supplied value for max_threads: {}", e));
            self
        }
        pub fn params<T>(mut self, value: T) -> Self
        where
            T: ::std::convert::TryInto<::std::option::Option<super::ExecutableParameterList>>,
            T::Error: ::std::fmt::Display,
        {
            self.params = value
                .try_into()
                .map_err(|e| format!("error converting supplied value for params: {}", e));
            self
        }
    }
    impl ::std::convert::TryFrom<ExecutableParallelExecutableType>
        for super::ExecutableParallelExecutableType
    {
        type Error = super::error::ConversionError;
        fn try_from(
            value: ExecutableParallelExecutableType,
        ) -> ::std::result::Result<Self, super::error::ConversionError> {
            Ok(Self {
                args: value.args?,
                dir: value.dir?,
                execs: value.execs?,
                fail_fast: value.fail_fast?,
                max_threads: value.max_threads?,
                params: value.params?,
            })
        }
    }
    impl ::std::convert::From<super::ExecutableParallelExecutableType>
        for ExecutableParallelExecutableType
    {
        fn from(value: super::ExecutableParallelExecutableType) -> Self {
            Self {
                args: Ok(value.args),
                dir: Ok(value.dir),
                execs: Ok(value.execs),
                fail_fast: Ok(value.fail_fast),
                max_threads: Ok(value.max_threads),
                params: Ok(value.params),
            }
        }
    }
    #[derive(Clone, Debug)]
    pub struct ExecutableParallelRefConfig {
        args: ::std::result::Result<::std::vec::Vec<::std::string::String>, ::std::string::String>,
        cmd: ::std::result::Result<::std::string::String, ::std::string::String>,
        if_: ::std::result::Result<::std::string::String, ::std::string::String>,
        ref_: ::std::result::Result<super::ExecutableRef, ::std::string::String>,
        retries: ::std::result::Result<i64, ::std::string::String>,
    }
    impl ::std::default::Default for ExecutableParallelRefConfig {
        fn default() -> Self {
            Self {
                args: Ok(Default::default()),
                cmd: Ok(Default::default()),
                if_: Ok(Default::default()),
                ref_: Ok(super::defaults::executable_parallel_ref_config_ref()),
                retries: Ok(Default::default()),
            }
        }
    }
    impl ExecutableParallelRefConfig {
        pub fn args<T>(mut self, value: T) -> Self
        where
            T: ::std::convert::TryInto<::std::vec::Vec<::std::string::String>>,
            T::Error: ::std::fmt::Display,
        {
            self.args = value
                .try_into()
                .map_err(|e| format!("error converting supplied value for args: {}", e));
            self
        }
        pub fn cmd<T>(mut self, value: T) -> Self
        where
            T: ::std::convert::TryInto<::std::string::String>,
            T::Error: ::std::fmt::Display,
        {
            self.cmd = value
                .try_into()
                .map_err(|e| format!("error converting supplied value for cmd: {}", e));
            self
        }
        pub fn if_<T>(mut self, value: T) -> Self
        where
            T: ::std::convert::TryInto<::std::string::String>,
            T::Error: ::std::fmt::Display,
        {
            self.if_ = value
                .try_into()
                .map_err(|e| format!("error converting supplied value for if_: {}", e));
            self
        }
        pub fn ref_<T>(mut self, value: T) -> Self
        where
            T: ::std::convert::TryInto<super::ExecutableRef>,
            T::Error: ::std::fmt::Display,
        {
            self.ref_ = value
                .try_into()
                .map_err(|e| format!("error converting supplied value for ref_: {}", e));
            self
        }
        pub fn retries<T>(mut self, value: T) -> Self
        where
            T: ::std::convert::TryInto<i64>,
            T::Error: ::std::fmt::Display,
        {
            self.retries = value
                .try_into()
                .map_err(|e| format!("error converting supplied value for retries: {}", e));
            self
        }
    }
    impl ::std::convert::TryFrom<ExecutableParallelRefConfig> for super::ExecutableParallelRefConfig {
        type Error = super::error::ConversionError;
        fn try_from(
            value: ExecutableParallelRefConfig,
        ) -> ::std::result::Result<Self, super::error::ConversionError> {
            Ok(Self {
                args: value.args?,
                cmd: value.cmd?,
                if_: value.if_?,
                ref_: value.ref_?,
                retries: value.retries?,
            })
        }
    }
    impl ::std::convert::From<super::ExecutableParallelRefConfig> for ExecutableParallelRefConfig {
        fn from(value: super::ExecutableParallelRefConfig) -> Self {
            Self {
                args: Ok(value.args),
                cmd: Ok(value.cmd),
                if_: Ok(value.if_),
                ref_: Ok(value.ref_),
                retries: Ok(value.retries),
            }
        }
    }
    #[derive(Clone, Debug)]
    pub struct ExecutableParameter {
        env_file: ::std::result::Result<::std::string::String, ::std::string::String>,
        env_key: ::std::result::Result<::std::string::String, ::std::string::String>,
        output_file: ::std::result::Result<::std::string::String, ::std::string::String>,
        prompt: ::std::result::Result<::std::string::String, ::std::string::String>,
        secret_ref: ::std::result::Result<::std::string::String, ::std::string::String>,
        text: ::std::result::Result<::std::string::String, ::std::string::String>,
    }
    impl ::std::default::Default for ExecutableParameter {
        fn default() -> Self {
            Self {
                env_file: Ok(Default::default()),
                env_key: Ok(Default::default()),
                output_file: Ok(Default::default()),
                prompt: Ok(Default::default()),
                secret_ref: Ok(Default::default()),
                text: Ok(Default::default()),
            }
        }
    }
    impl ExecutableParameter {
        pub fn env_file<T>(mut self, value: T) -> Self
        where
            T: ::std::convert::TryInto<::std::string::String>,
            T::Error: ::std::fmt::Display,
        {
            self.env_file = value
                .try_into()
                .map_err(|e| format!("error converting supplied value for env_file: {}", e));
            self
        }
        pub fn env_key<T>(mut self, value: T) -> Self
        where
            T: ::std::convert::TryInto<::std::string::String>,
            T::Error: ::std::fmt::Display,
        {
            self.env_key = value
                .try_into()
                .map_err(|e| format!("error converting supplied value for env_key: {}", e));
            self
        }
        pub fn output_file<T>(mut self, value: T) -> Self
        where
            T: ::std::convert::TryInto<::std::string::String>,
            T::Error: ::std::fmt::Display,
        {
            self.output_file = value
                .try_into()
                .map_err(|e| format!("error converting supplied value for output_file: {}", e));
            self
        }
        pub fn prompt<T>(mut self, value: T) -> Self
        where
            T: ::std::convert::TryInto<::std::string::String>,
            T::Error: ::std::fmt::Display,
        {
            self.prompt = value
                .try_into()
                .map_err(|e| format!("error converting supplied value for prompt: {}", e));
            self
        }
        pub fn secret_ref<T>(mut self, value: T) -> Self
        where
            T: ::std::convert::TryInto<::std::string::String>,
            T::Error: ::std::fmt::Display,
        {
            self.secret_ref = value
                .try_into()
                .map_err(|e| format!("error converting supplied value for secret_ref: {}", e));
            self
        }
        pub fn text<T>(mut self, value: T) -> Self
        where
            T: ::std::convert::TryInto<::std::string::String>,
            T::Error: ::std::fmt::Display,
        {
            self.text = value
                .try_into()
                .map_err(|e| format!("error converting supplied value for text: {}", e));
            self
        }
    }
    impl ::std::convert::TryFrom<ExecutableParameter> for super::ExecutableParameter {
        type Error = super::error::ConversionError;
        fn try_from(
            value: ExecutableParameter,
        ) -> ::std::result::Result<Self, super::error::ConversionError> {
            Ok(Self {
                env_file: value.env_file?,
                env_key: value.env_key?,
                output_file: value.output_file?,
                prompt: value.prompt?,
                secret_ref: value.secret_ref?,
                text: value.text?,
            })
        }
    }
    impl ::std::convert::From<super::ExecutableParameter> for ExecutableParameter {
        fn from(value: super::ExecutableParameter) -> Self {
            Self {
                env_file: Ok(value.env_file),
                env_key: Ok(value.env_key),
                output_file: Ok(value.output_file),
                prompt: Ok(value.prompt),
                secret_ref: Ok(value.secret_ref),
                text: Ok(value.text),
            }
        }
    }
    #[derive(Clone, Debug)]
    pub struct ExecutableRenderExecutableType {
        args: ::std::result::Result<
            ::std::option::Option<super::ExecutableArgumentList>,
            ::std::string::String,
        >,
        dir: ::std::result::Result<super::ExecutableDirectory, ::std::string::String>,
        params: ::std::result::Result<
            ::std::option::Option<super::ExecutableParameterList>,
            ::std::string::String,
        >,
        template_data_file: ::std::result::Result<::std::string::String, ::std::string::String>,
        template_file: ::std::result::Result<::std::string::String, ::std::string::String>,
    }
    impl ::std::default::Default for ExecutableRenderExecutableType {
        fn default() -> Self {
            Self {
                args: Ok(Default::default()),
                dir: Ok(super::defaults::executable_render_executable_type_dir()),
                params: Ok(Default::default()),
                template_data_file: Ok(Default::default()),
                template_file: Err("no value supplied for template_file".to_string()),
            }
        }
    }
    impl ExecutableRenderExecutableType {
        pub fn args<T>(mut self, value: T) -> Self
        where
            T: ::std::convert::TryInto<::std::option::Option<super::ExecutableArgumentList>>,
            T::Error: ::std::fmt::Display,
        {
            self.args = value
                .try_into()
                .map_err(|e| format!("error converting supplied value for args: {}", e));
            self
        }
        pub fn dir<T>(mut self, value: T) -> Self
        where
            T: ::std::convert::TryInto<super::ExecutableDirectory>,
            T::Error: ::std::fmt::Display,
        {
            self.dir = value
                .try_into()
                .map_err(|e| format!("error converting supplied value for dir: {}", e));
            self
        }
        pub fn params<T>(mut self, value: T) -> Self
        where
            T: ::std::convert::TryInto<::std::option::Option<super::ExecutableParameterList>>,
            T::Error: ::std::fmt::Display,
        {
            self.params = value
                .try_into()
                .map_err(|e| format!("error converting supplied value for params: {}", e));
            self
        }
        pub fn template_data_file<T>(mut self, value: T) -> Self
        where
            T: ::std::convert::TryInto<::std::string::String>,
            T::Error: ::std::fmt::Display,
        {
            self.template_data_file = value.try_into().map_err(|e| {
                format!(
                    "error converting supplied value for template_data_file: {}",
                    e
                )
            });
            self
        }
        pub fn template_file<T>(mut self, value: T) -> Self
        where
            T: ::std::convert::TryInto<::std::string::String>,
            T::Error: ::std::fmt::Display,
        {
            self.template_file = value
                .try_into()
                .map_err(|e| format!("error converting supplied value for template_file: {}", e));
            self
        }
    }
    impl ::std::convert::TryFrom<ExecutableRenderExecutableType>
        for super::ExecutableRenderExecutableType
    {
        type Error = super::error::ConversionError;
        fn try_from(
            value: ExecutableRenderExecutableType,
        ) -> ::std::result::Result<Self, super::error::ConversionError> {
            Ok(Self {
                args: value.args?,
                dir: value.dir?,
                params: value.params?,
                template_data_file: value.template_data_file?,
                template_file: value.template_file?,
            })
        }
    }
    impl ::std::convert::From<super::ExecutableRenderExecutableType>
        for ExecutableRenderExecutableType
    {
        fn from(value: super::ExecutableRenderExecutableType) -> Self {
            Self {
                args: Ok(value.args),
                dir: Ok(value.dir),
                params: Ok(value.params),
                template_data_file: Ok(value.template_data_file),
                template_file: Ok(value.template_file),
            }
        }
    }
    #[derive(Clone, Debug)]
    pub struct ExecutableRequestExecutableType {
        args: ::std::result::Result<
            ::std::option::Option<super::ExecutableArgumentList>,
            ::std::string::String,
        >,
        body: ::std::result::Result<::std::string::String, ::std::string::String>,
        headers: ::std::result::Result<
            ::std::collections::HashMap<::std::string::String, ::std::string::String>,
            ::std::string::String,
        >,
        log_response: ::std::result::Result<bool, ::std::string::String>,
        method: ::std::result::Result<
            super::ExecutableRequestExecutableTypeMethod,
            ::std::string::String,
        >,
        params: ::std::result::Result<
            ::std::option::Option<super::ExecutableParameterList>,
            ::std::string::String,
        >,
        response_file: ::std::result::Result<
            ::std::option::Option<super::ExecutableRequestResponseFile>,
            ::std::string::String,
        >,
        timeout: ::std::result::Result<::std::string::String, ::std::string::String>,
        transform_response: ::std::result::Result<::std::string::String, ::std::string::String>,
        url: ::std::result::Result<::std::string::String, ::std::string::String>,
        valid_status_codes: ::std::result::Result<::std::vec::Vec<i64>, ::std::string::String>,
    }
    impl ::std::default::Default for ExecutableRequestExecutableType {
        fn default() -> Self {
            Self {
                args: Ok(Default::default()),
                body: Ok(Default::default()),
                headers: Ok(Default::default()),
                log_response: Ok(Default::default()),
                method: Ok(super::defaults::executable_request_executable_type_method()),
                params: Ok(Default::default()),
                response_file: Ok(Default::default()),
                timeout: Ok(super::defaults::executable_request_executable_type_timeout()),
                transform_response: Ok(Default::default()),
                url: Err("no value supplied for url".to_string()),
                valid_status_codes: Ok(Default::default()),
            }
        }
    }
    impl ExecutableRequestExecutableType {
        pub fn args<T>(mut self, value: T) -> Self
        where
            T: ::std::convert::TryInto<::std::option::Option<super::ExecutableArgumentList>>,
            T::Error: ::std::fmt::Display,
        {
            self.args = value
                .try_into()
                .map_err(|e| format!("error converting supplied value for args: {}", e));
            self
        }
        pub fn body<T>(mut self, value: T) -> Self
        where
            T: ::std::convert::TryInto<::std::string::String>,
            T::Error: ::std::fmt::Display,
        {
            self.body = value
                .try_into()
                .map_err(|e| format!("error converting supplied value for body: {}", e));
            self
        }
        pub fn headers<T>(mut self, value: T) -> Self
        where
            T: ::std::convert::TryInto<
                ::std::collections::HashMap<::std::string::String, ::std::string::String>,
            >,
            T::Error: ::std::fmt::Display,
        {
            self.headers = value
                .try_into()
                .map_err(|e| format!("error converting supplied value for headers: {}", e));
            self
        }
        pub fn log_response<T>(mut self, value: T) -> Self
        where
            T: ::std::convert::TryInto<bool>,
            T::Error: ::std::fmt::Display,
        {
            self.log_response = value
                .try_into()
                .map_err(|e| format!("error converting supplied value for log_response: {}", e));
            self
        }
        pub fn method<T>(mut self, value: T) -> Self
        where
            T: ::std::convert::TryInto<super::ExecutableRequestExecutableTypeMethod>,
            T::Error: ::std::fmt::Display,
        {
            self.method = value
                .try_into()
                .map_err(|e| format!("error converting supplied value for method: {}", e));
            self
        }
        pub fn params<T>(mut self, value: T) -> Self
        where
            T: ::std::convert::TryInto<::std::option::Option<super::ExecutableParameterList>>,
            T::Error: ::std::fmt::Display,
        {
            self.params = value
                .try_into()
                .map_err(|e| format!("error converting supplied value for params: {}", e));
            self
        }
        pub fn response_file<T>(mut self, value: T) -> Self
        where
            T: ::std::convert::TryInto<::std::option::Option<super::ExecutableRequestResponseFile>>,
            T::Error: ::std::fmt::Display,
        {
            self.response_file = value
                .try_into()
                .map_err(|e| format!("error converting supplied value for response_file: {}", e));
            self
        }
        pub fn timeout<T>(mut self, value: T) -> Self
        where
            T: ::std::convert::TryInto<::std::string::String>,
            T::Error: ::std::fmt::Display,
        {
            self.timeout = value
                .try_into()
                .map_err(|e| format!("error converting supplied value for timeout: {}", e));
            self
        }
        pub fn transform_response<T>(mut self, value: T) -> Self
        where
            T: ::std::convert::TryInto<::std::string::String>,
            T::Error: ::std::fmt::Display,
        {
            self.transform_response = value.try_into().map_err(|e| {
                format!(
                    "error converting supplied value for transform_response: {}",
                    e
                )
            });
            self
        }
        pub fn url<T>(mut self, value: T) -> Self
        where
            T: ::std::convert::TryInto<::std::string::String>,
            T::Error: ::std::fmt::Display,
        {
            self.url = value
                .try_into()
                .map_err(|e| format!("error converting supplied value for url: {}", e));
            self
        }
        pub fn valid_status_codes<T>(mut self, value: T) -> Self
        where
            T: ::std::convert::TryInto<::std::vec::Vec<i64>>,
            T::Error: ::std::fmt::Display,
        {
            self.valid_status_codes = value.try_into().map_err(|e| {
                format!(
                    "error converting supplied value for valid_status_codes: {}",
                    e
                )
            });
            self
        }
    }
    impl ::std::convert::TryFrom<ExecutableRequestExecutableType>
        for super::ExecutableRequestExecutableType
    {
        type Error = super::error::ConversionError;
        fn try_from(
            value: ExecutableRequestExecutableType,
        ) -> ::std::result::Result<Self, super::error::ConversionError> {
            Ok(Self {
                args: value.args?,
                body: value.body?,
                headers: value.headers?,
                log_response: value.log_response?,
                method: value.method?,
                params: value.params?,
                response_file: value.response_file?,
                timeout: value.timeout?,
                transform_response: value.transform_response?,
                url: value.url?,
                valid_status_codes: value.valid_status_codes?,
            })
        }
    }
    impl ::std::convert::From<super::ExecutableRequestExecutableType>
        for ExecutableRequestExecutableType
    {
        fn from(value: super::ExecutableRequestExecutableType) -> Self {
            Self {
                args: Ok(value.args),
                body: Ok(value.body),
                headers: Ok(value.headers),
                log_response: Ok(value.log_response),
                method: Ok(value.method),
                params: Ok(value.params),
                response_file: Ok(value.response_file),
                timeout: Ok(value.timeout),
                transform_response: Ok(value.transform_response),
                url: Ok(value.url),
                valid_status_codes: Ok(value.valid_status_codes),
            }
        }
    }
    #[derive(Clone, Debug)]
    pub struct ExecutableRequestResponseFile {
        dir: ::std::result::Result<super::ExecutableDirectory, ::std::string::String>,
        filename: ::std::result::Result<::std::string::String, ::std::string::String>,
        save_as: ::std::result::Result<
            super::ExecutableRequestResponseFileSaveAs,
            ::std::string::String,
        >,
    }
    impl ::std::default::Default for ExecutableRequestResponseFile {
        fn default() -> Self {
            Self {
                dir: Ok(super::defaults::executable_request_response_file_dir()),
                filename: Err("no value supplied for filename".to_string()),
                save_as: Ok(super::defaults::executable_request_response_file_save_as()),
            }
        }
    }
    impl ExecutableRequestResponseFile {
        pub fn dir<T>(mut self, value: T) -> Self
        where
            T: ::std::convert::TryInto<super::ExecutableDirectory>,
            T::Error: ::std::fmt::Display,
        {
            self.dir = value
                .try_into()
                .map_err(|e| format!("error converting supplied value for dir: {}", e));
            self
        }
        pub fn filename<T>(mut self, value: T) -> Self
        where
            T: ::std::convert::TryInto<::std::string::String>,
            T::Error: ::std::fmt::Display,
        {
            self.filename = value
                .try_into()
                .map_err(|e| format!("error converting supplied value for filename: {}", e));
            self
        }
        pub fn save_as<T>(mut self, value: T) -> Self
        where
            T: ::std::convert::TryInto<super::ExecutableRequestResponseFileSaveAs>,
            T::Error: ::std::fmt::Display,
        {
            self.save_as = value
                .try_into()
                .map_err(|e| format!("error converting supplied value for save_as: {}", e));
            self
        }
    }
    impl ::std::convert::TryFrom<ExecutableRequestResponseFile>
        for super::ExecutableRequestResponseFile
    {
        type Error = super::error::ConversionError;
        fn try_from(
            value: ExecutableRequestResponseFile,
        ) -> ::std::result::Result<Self, super::error::ConversionError> {
            Ok(Self {
                dir: value.dir?,
                filename: value.filename?,
                save_as: value.save_as?,
            })
        }
    }
    impl ::std::convert::From<super::ExecutableRequestResponseFile> for ExecutableRequestResponseFile {
        fn from(value: super::ExecutableRequestResponseFile) -> Self {
            Self {
                dir: Ok(value.dir),
                filename: Ok(value.filename),
                save_as: Ok(value.save_as),
            }
        }
    }
    #[derive(Clone, Debug)]
    pub struct ExecutableSerialExecutableType {
        args: ::std::result::Result<
            ::std::option::Option<super::ExecutableArgumentList>,
            ::std::string::String,
        >,
        dir: ::std::result::Result<super::ExecutableDirectory, ::std::string::String>,
        execs: ::std::result::Result<super::ExecutableSerialRefConfigList, ::std::string::String>,
        fail_fast: ::std::result::Result<::std::option::Option<bool>, ::std::string::String>,
        params: ::std::result::Result<
            ::std::option::Option<super::ExecutableParameterList>,
            ::std::string::String,
        >,
    }
    impl ::std::default::Default for ExecutableSerialExecutableType {
        fn default() -> Self {
            Self {
                args: Ok(Default::default()),
                dir: Ok(super::defaults::executable_serial_executable_type_dir()),
                execs: Err("no value supplied for execs".to_string()),
                fail_fast: Ok(Default::default()),
                params: Ok(Default::default()),
            }
        }
    }
    impl ExecutableSerialExecutableType {
        pub fn args<T>(mut self, value: T) -> Self
        where
            T: ::std::convert::TryInto<::std::option::Option<super::ExecutableArgumentList>>,
            T::Error: ::std::fmt::Display,
        {
            self.args = value
                .try_into()
                .map_err(|e| format!("error converting supplied value for args: {}", e));
            self
        }
        pub fn dir<T>(mut self, value: T) -> Self
        where
            T: ::std::convert::TryInto<super::ExecutableDirectory>,
            T::Error: ::std::fmt::Display,
        {
            self.dir = value
                .try_into()
                .map_err(|e| format!("error converting supplied value for dir: {}", e));
            self
        }
        pub fn execs<T>(mut self, value: T) -> Self
        where
            T: ::std::convert::TryInto<super::ExecutableSerialRefConfigList>,
            T::Error: ::std::fmt::Display,
        {
            self.execs = value
                .try_into()
                .map_err(|e| format!("error converting supplied value for execs: {}", e));
            self
        }
        pub fn fail_fast<T>(mut self, value: T) -> Self
        where
            T: ::std::convert::TryInto<::std::option::Option<bool>>,
            T::Error: ::std::fmt::Display,
        {
            self.fail_fast = value
                .try_into()
                .map_err(|e| format!("error converting supplied value for fail_fast: {}", e));
            self
        }
        pub fn params<T>(mut self, value: T) -> Self
        where
            T: ::std::convert::TryInto<::std::option::Option<super::ExecutableParameterList>>,
            T::Error: ::std::fmt::Display,
        {
            self.params = value
                .try_into()
                .map_err(|e| format!("error converting supplied value for params: {}", e));
            self
        }
    }
    impl ::std::convert::TryFrom<ExecutableSerialExecutableType>
        for super::ExecutableSerialExecutableType
    {
        type Error = super::error::ConversionError;
        fn try_from(
            value: ExecutableSerialExecutableType,
        ) -> ::std::result::Result<Self, super::error::ConversionError> {
            Ok(Self {
                args: value.args?,
                dir: value.dir?,
                execs: value.execs?,
                fail_fast: value.fail_fast?,
                params: value.params?,
            })
        }
    }
    impl ::std::convert::From<super::ExecutableSerialExecutableType>
        for ExecutableSerialExecutableType
    {
        fn from(value: super::ExecutableSerialExecutableType) -> Self {
            Self {
                args: Ok(value.args),
                dir: Ok(value.dir),
                execs: Ok(value.execs),
                fail_fast: Ok(value.fail_fast),
                params: Ok(value.params),
            }
        }
    }
    #[derive(Clone, Debug)]
    pub struct ExecutableSerialRefConfig {
        args: ::std::result::Result<::std::vec::Vec<::std::string::String>, ::std::string::String>,
        cmd: ::std::result::Result<::std::string::String, ::std::string::String>,
        if_: ::std::result::Result<::std::string::String, ::std::string::String>,
        ref_: ::std::result::Result<super::ExecutableRef, ::std::string::String>,
        retries: ::std::result::Result<i64, ::std::string::String>,
        review_required: ::std::result::Result<bool, ::std::string::String>,
    }
    impl ::std::default::Default for ExecutableSerialRefConfig {
        fn default() -> Self {
            Self {
                args: Ok(Default::default()),
                cmd: Ok(Default::default()),
                if_: Ok(Default::default()),
                ref_: Ok(super::defaults::executable_serial_ref_config_ref()),
                retries: Ok(Default::default()),
                review_required: Ok(Default::default()),
            }
        }
    }
    impl ExecutableSerialRefConfig {
        pub fn args<T>(mut self, value: T) -> Self
        where
            T: ::std::convert::TryInto<::std::vec::Vec<::std::string::String>>,
            T::Error: ::std::fmt::Display,
        {
            self.args = value
                .try_into()
                .map_err(|e| format!("error converting supplied value for args: {}", e));
            self
        }
        pub fn cmd<T>(mut self, value: T) -> Self
        where
            T: ::std::convert::TryInto<::std::string::String>,
            T::Error: ::std::fmt::Display,
        {
            self.cmd = value
                .try_into()
                .map_err(|e| format!("error converting supplied value for cmd: {}", e));
            self
        }
        pub fn if_<T>(mut self, value: T) -> Self
        where
            T: ::std::convert::TryInto<::std::string::String>,
            T::Error: ::std::fmt::Display,
        {
            self.if_ = value
                .try_into()
                .map_err(|e| format!("error converting supplied value for if_: {}", e));
            self
        }
        pub fn ref_<T>(mut self, value: T) -> Self
        where
            T: ::std::convert::TryInto<super::ExecutableRef>,
            T::Error: ::std::fmt::Display,
        {
            self.ref_ = value
                .try_into()
                .map_err(|e| format!("error converting supplied value for ref_: {}", e));
            self
        }
        pub fn retries<T>(mut self, value: T) -> Self
        where
            T: ::std::convert::TryInto<i64>,
            T::Error: ::std::fmt::Display,
        {
            self.retries = value
                .try_into()
                .map_err(|e| format!("error converting supplied value for retries: {}", e));
            self
        }
        pub fn review_required<T>(mut self, value: T) -> Self
        where
            T: ::std::convert::TryInto<bool>,
            T::Error: ::std::fmt::Display,
        {
            self.review_required = value
                .try_into()
                .map_err(|e| format!("error converting supplied value for review_required: {}", e));
            self
        }
    }
    impl ::std::convert::TryFrom<ExecutableSerialRefConfig> for super::ExecutableSerialRefConfig {
        type Error = super::error::ConversionError;
        fn try_from(
            value: ExecutableSerialRefConfig,
        ) -> ::std::result::Result<Self, super::error::ConversionError> {
            Ok(Self {
                args: value.args?,
                cmd: value.cmd?,
                if_: value.if_?,
                ref_: value.ref_?,
                retries: value.retries?,
                review_required: value.review_required?,
            })
        }
    }
    impl ::std::convert::From<super::ExecutableSerialRefConfig> for ExecutableSerialRefConfig {
        fn from(value: super::ExecutableSerialRefConfig) -> Self {
            Self {
                args: Ok(value.args),
                cmd: Ok(value.cmd),
                if_: Ok(value.if_),
                ref_: Ok(value.ref_),
                retries: Ok(value.retries),
                review_required: Ok(value.review_required),
            }
        }
    }
    #[derive(Clone, Debug)]
    pub struct FlowFile {
        description: ::std::result::Result<::std::string::String, ::std::string::String>,
        description_file: ::std::result::Result<::std::string::String, ::std::string::String>,
        executables:
            ::std::result::Result<::std::vec::Vec<super::Executable>, ::std::string::String>,
        from_file: ::std::result::Result<super::FromFile, ::std::string::String>,
        imports: ::std::result::Result<super::FromFile, ::std::string::String>,
        namespace: ::std::result::Result<::std::string::String, ::std::string::String>,
        tags: ::std::result::Result<::std::vec::Vec<::std::string::String>, ::std::string::String>,
        visibility: ::std::result::Result<
            ::std::option::Option<super::CommonVisibility>,
            ::std::string::String,
        >,
    }
    impl ::std::default::Default for FlowFile {
        fn default() -> Self {
            Self {
                description: Ok(Default::default()),
                description_file: Ok(Default::default()),
                executables: Ok(Default::default()),
                from_file: Ok(super::defaults::flow_file_from_file()),
                imports: Ok(super::defaults::flow_file_imports()),
                namespace: Ok(Default::default()),
                tags: Ok(Default::default()),
                visibility: Ok(Default::default()),
            }
        }
    }
    impl FlowFile {
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
        pub fn executables<T>(mut self, value: T) -> Self
        where
            T: ::std::convert::TryInto<::std::vec::Vec<super::Executable>>,
            T::Error: ::std::fmt::Display,
        {
            self.executables = value
                .try_into()
                .map_err(|e| format!("error converting supplied value for executables: {}", e));
            self
        }
        pub fn from_file<T>(mut self, value: T) -> Self
        where
            T: ::std::convert::TryInto<super::FromFile>,
            T::Error: ::std::fmt::Display,
        {
            self.from_file = value
                .try_into()
                .map_err(|e| format!("error converting supplied value for from_file: {}", e));
            self
        }
        pub fn imports<T>(mut self, value: T) -> Self
        where
            T: ::std::convert::TryInto<super::FromFile>,
            T::Error: ::std::fmt::Display,
        {
            self.imports = value
                .try_into()
                .map_err(|e| format!("error converting supplied value for imports: {}", e));
            self
        }
        pub fn namespace<T>(mut self, value: T) -> Self
        where
            T: ::std::convert::TryInto<::std::string::String>,
            T::Error: ::std::fmt::Display,
        {
            self.namespace = value
                .try_into()
                .map_err(|e| format!("error converting supplied value for namespace: {}", e));
            self
        }
        pub fn tags<T>(mut self, value: T) -> Self
        where
            T: ::std::convert::TryInto<::std::vec::Vec<::std::string::String>>,
            T::Error: ::std::fmt::Display,
        {
            self.tags = value
                .try_into()
                .map_err(|e| format!("error converting supplied value for tags: {}", e));
            self
        }
        pub fn visibility<T>(mut self, value: T) -> Self
        where
            T: ::std::convert::TryInto<::std::option::Option<super::CommonVisibility>>,
            T::Error: ::std::fmt::Display,
        {
            self.visibility = value
                .try_into()
                .map_err(|e| format!("error converting supplied value for visibility: {}", e));
            self
        }
    }
    impl ::std::convert::TryFrom<FlowFile> for super::FlowFile {
        type Error = super::error::ConversionError;
        fn try_from(value: FlowFile) -> ::std::result::Result<Self, super::error::ConversionError> {
            Ok(Self {
                description: value.description?,
                description_file: value.description_file?,
                executables: value.executables?,
                from_file: value.from_file?,
                imports: value.imports?,
                namespace: value.namespace?,
                tags: value.tags?,
                visibility: value.visibility?,
            })
        }
    }
    impl ::std::convert::From<super::FlowFile> for FlowFile {
        fn from(value: super::FlowFile) -> Self {
            Self {
                description: Ok(value.description),
                description_file: Ok(value.description_file),
                executables: Ok(value.executables),
                from_file: Ok(value.from_file),
                imports: Ok(value.imports),
                namespace: Ok(value.namespace),
                tags: Ok(value.tags),
                visibility: Ok(value.visibility),
            }
        }
    }
}
#[doc = r" Generation of default values for serde."]
pub mod defaults {
    pub(super) fn default_u64<T, const V: u64>() -> T
    where
        T: ::std::convert::TryFrom<u64>,
        <T as ::std::convert::TryFrom<u64>>::Error: ::std::fmt::Debug,
    {
        T::try_from(V).unwrap()
    }
    pub(super) fn executable_aliases() -> super::CommonAliases {
        super::CommonAliases(vec![])
    }
    pub(super) fn executable_tags() -> super::CommonTags {
        super::CommonTags(vec![])
    }
    pub(super) fn executable_argument_type() -> super::ExecutableArgumentType {
        super::ExecutableArgumentType::String
    }
    pub(super) fn executable_exec_executable_type_dir() -> super::ExecutableDirectory {
        super::ExecutableDirectory("".to_string())
    }
    pub(super) fn executable_exec_executable_type_log_mode() -> ::std::string::String {
        "logfmt".to_string()
    }
    pub(super) fn executable_parallel_executable_type_dir() -> super::ExecutableDirectory {
        super::ExecutableDirectory("".to_string())
    }
    pub(super) fn executable_parallel_ref_config_ref() -> super::ExecutableRef {
        super::ExecutableRef("".to_string())
    }
    pub(super) fn executable_render_executable_type_dir() -> super::ExecutableDirectory {
        super::ExecutableDirectory("".to_string())
    }
    pub(super) fn executable_request_executable_type_method(
    ) -> super::ExecutableRequestExecutableTypeMethod {
        super::ExecutableRequestExecutableTypeMethod::Get
    }
    pub(super) fn executable_request_executable_type_timeout() -> ::std::string::String {
        "30m0s".to_string()
    }
    pub(super) fn executable_request_response_file_dir() -> super::ExecutableDirectory {
        super::ExecutableDirectory("".to_string())
    }
    pub(super) fn executable_request_response_file_save_as(
    ) -> super::ExecutableRequestResponseFileSaveAs {
        super::ExecutableRequestResponseFileSaveAs::Raw
    }
    pub(super) fn executable_serial_executable_type_dir() -> super::ExecutableDirectory {
        super::ExecutableDirectory("".to_string())
    }
    pub(super) fn executable_serial_ref_config_ref() -> super::ExecutableRef {
        super::ExecutableRef("".to_string())
    }
    pub(super) fn flow_file_from_file() -> super::FromFile {
        super::FromFile(vec![])
    }
    pub(super) fn flow_file_imports() -> super::FromFile {
        super::FromFile(vec![])
    }
}
