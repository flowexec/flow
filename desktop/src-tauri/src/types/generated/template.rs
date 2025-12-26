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
#[doc = "File source and destination configuration.\nGo templating from form data is supported in all fields.\n"]
#[doc = r""]
#[doc = r" <details><summary>JSON schema</summary>"]
#[doc = r""]
#[doc = r" ```json"]
#[doc = "{"]
#[doc = "  \"description\": \"File source and destination configuration.\\nGo templating from form data is supported in all fields.\\n\","]
#[doc = "  \"type\": \"object\","]
#[doc = "  \"required\": ["]
#[doc = "    \"srcName\""]
#[doc = "  ],"]
#[doc = "  \"properties\": {"]
#[doc = "    \"asTemplate\": {"]
#[doc = "      \"description\": \"If true, the artifact will be copied as a template file. The file will be rendered using Go templating from \\nthe form data. [Expr language functions](https://expr-lang.org/docs/language-definition) are available for use in the template.\\n\","]
#[doc = "      \"default\": false,"]
#[doc = "      \"type\": \"boolean\""]
#[doc = "    },"]
#[doc = "    \"dstDir\": {"]
#[doc = "      \"description\": \"The directory to copy the file to. If not set, the file will be copied to the root of the flow file directory.\\nThe directory will be created if it does not exist.\\n\","]
#[doc = "      \"default\": \"\","]
#[doc = "      \"type\": \"string\""]
#[doc = "    },"]
#[doc = "    \"dstName\": {"]
#[doc = "      \"description\": \"The name of the file to copy to. If not set, the file will be copied with the same name.\","]
#[doc = "      \"default\": \"\","]
#[doc = "      \"type\": \"string\""]
#[doc = "    },"]
#[doc = "    \"if\": {"]
#[doc = "      \"description\": \"An expression that determines whether the the artifact should be copied, using the Expr language syntax. \\nThe expression is evaluated at runtime and must resolve to a boolean value. If the condition is not met, \\nthe artifact will not be copied.\\n\\nThe expression has access to OS/architecture information (os, arch), environment variables (env), form input \\n(form), and context information (name, workspace, directory, etc.).\\n\\nSee the [flow documentation](https://flowexec.io/guide/templating) for more information.\\n\","]
#[doc = "      \"default\": \"\","]
#[doc = "      \"type\": \"string\""]
#[doc = "    },"]
#[doc = "    \"srcDir\": {"]
#[doc = "      \"description\": \"The directory to copy the file from. \\nIf not set, the file will be copied from the directory of the template file.\\n\","]
#[doc = "      \"default\": \"\","]
#[doc = "      \"type\": \"string\""]
#[doc = "    },"]
#[doc = "    \"srcName\": {"]
#[doc = "      \"description\": \"The name of the file to copy.\","]
#[doc = "      \"type\": \"string\""]
#[doc = "    }"]
#[doc = "  }"]
#[doc = "}"]
#[doc = r" ```"]
#[doc = r" </details>"]
#[derive(:: serde :: Deserialize, :: serde :: Serialize, Clone, Debug)]
pub struct Artifact {
    #[doc = "If true, the artifact will be copied as a template file. The file will be rendered using Go templating from \nthe form data. [Expr language functions](https://expr-lang.org/docs/language-definition) are available for use in the template.\n"]
    #[serde(rename = "asTemplate", default)]
    pub as_template: bool,
    #[doc = "The directory to copy the file to. If not set, the file will be copied to the root of the flow file directory.\nThe directory will be created if it does not exist.\n"]
    #[serde(rename = "dstDir", default)]
    pub dst_dir: ::std::string::String,
    #[doc = "The name of the file to copy to. If not set, the file will be copied with the same name."]
    #[serde(rename = "dstName", default)]
    pub dst_name: ::std::string::String,
    #[doc = "An expression that determines whether the the artifact should be copied, using the Expr language syntax. \nThe expression is evaluated at runtime and must resolve to a boolean value. If the condition is not met, \nthe artifact will not be copied.\n\nThe expression has access to OS/architecture information (os, arch), environment variables (env), form input \n(form), and context information (name, workspace, directory, etc.).\n\nSee the [flow documentation](https://flowexec.io/guide/templating) for more information.\n"]
    #[serde(rename = "if", default)]
    pub if_: ::std::string::String,
    #[doc = "The directory to copy the file from. \nIf not set, the file will be copied from the directory of the template file.\n"]
    #[serde(rename = "srcDir", default)]
    pub src_dir: ::std::string::String,
    #[doc = "The name of the file to copy."]
    #[serde(rename = "srcName")]
    pub src_name: ::std::string::String,
}
impl ::std::convert::From<&Artifact> for Artifact {
    fn from(value: &Artifact) -> Self {
        value.clone()
    }
}
impl Artifact {
    pub fn builder() -> builder::Artifact {
        Default::default()
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
#[doc = "A field to be displayed to the user when generating a flow file from a template."]
#[doc = r""]
#[doc = r" <details><summary>JSON schema</summary>"]
#[doc = r""]
#[doc = r" ```json"]
#[doc = "{"]
#[doc = "  \"description\": \"A field to be displayed to the user when generating a flow file from a template.\","]
#[doc = "  \"type\": \"object\","]
#[doc = "  \"required\": ["]
#[doc = "    \"key\","]
#[doc = "    \"prompt\""]
#[doc = "  ],"]
#[doc = "  \"properties\": {"]
#[doc = "    \"default\": {"]
#[doc = "      \"description\": \"The default value to use if a value is not set.\","]
#[doc = "      \"default\": \"\","]
#[doc = "      \"type\": \"string\""]
#[doc = "    },"]
#[doc = "    \"description\": {"]
#[doc = "      \"description\": \"A description of the field.\","]
#[doc = "      \"default\": \"\","]
#[doc = "      \"type\": \"string\""]
#[doc = "    },"]
#[doc = "    \"group\": {"]
#[doc = "      \"description\": \"The group to display the field in. Fields with the same group will be displayed together.\","]
#[doc = "      \"default\": 0,"]
#[doc = "      \"type\": \"integer\""]
#[doc = "    },"]
#[doc = "    \"key\": {"]
#[doc = "      \"description\": \"The key to associate the data with. This is used as the key in the template data map.\","]
#[doc = "      \"type\": \"string\""]
#[doc = "    },"]
#[doc = "    \"prompt\": {"]
#[doc = "      \"description\": \"A prompt to be displayed to the user when collecting an input value.\","]
#[doc = "      \"type\": \"string\""]
#[doc = "    },"]
#[doc = "    \"required\": {"]
#[doc = "      \"description\": \"If true, a value must be set. If false, the default value will be used if a value is not set.\","]
#[doc = "      \"default\": false,"]
#[doc = "      \"type\": \"boolean\""]
#[doc = "    },"]
#[doc = "    \"type\": {"]
#[doc = "      \"description\": \"The type of input field to display.\","]
#[doc = "      \"default\": \"text\","]
#[doc = "      \"type\": \"string\","]
#[doc = "      \"enum\": ["]
#[doc = "        \"text\","]
#[doc = "        \"masked\","]
#[doc = "        \"multiline\","]
#[doc = "        \"confirm\""]
#[doc = "      ]"]
#[doc = "    },"]
#[doc = "    \"validate\": {"]
#[doc = "      \"description\": \"A regular expression to validate the input value against.\","]
#[doc = "      \"default\": \"\","]
#[doc = "      \"type\": \"string\""]
#[doc = "    }"]
#[doc = "  }"]
#[doc = "}"]
#[doc = r" ```"]
#[doc = r" </details>"]
#[derive(:: serde :: Deserialize, :: serde :: Serialize, Clone, Debug)]
pub struct Field {
    #[doc = "The default value to use if a value is not set."]
    #[serde(default)]
    pub default: ::std::string::String,
    #[doc = "A description of the field."]
    #[serde(default)]
    pub description: ::std::string::String,
    #[doc = "The group to display the field in. Fields with the same group will be displayed together."]
    #[serde(default)]
    pub group: i64,
    #[doc = "The key to associate the data with. This is used as the key in the template data map."]
    pub key: ::std::string::String,
    #[doc = "A prompt to be displayed to the user when collecting an input value."]
    pub prompt: ::std::string::String,
    #[doc = "If true, a value must be set. If false, the default value will be used if a value is not set."]
    #[serde(default)]
    pub required: bool,
    #[doc = "The type of input field to display."]
    #[serde(rename = "type", default = "defaults::field_type")]
    pub type_: FieldType,
    #[doc = "A regular expression to validate the input value against."]
    #[serde(default)]
    pub validate: ::std::string::String,
}
impl ::std::convert::From<&Field> for Field {
    fn from(value: &Field) -> Self {
        value.clone()
    }
}
impl Field {
    pub fn builder() -> builder::Field {
        Default::default()
    }
}
#[doc = "The type of input field to display."]
#[doc = r""]
#[doc = r" <details><summary>JSON schema</summary>"]
#[doc = r""]
#[doc = r" ```json"]
#[doc = "{"]
#[doc = "  \"description\": \"The type of input field to display.\","]
#[doc = "  \"default\": \"text\","]
#[doc = "  \"type\": \"string\","]
#[doc = "  \"enum\": ["]
#[doc = "    \"text\","]
#[doc = "    \"masked\","]
#[doc = "    \"multiline\","]
#[doc = "    \"confirm\""]
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
pub enum FieldType {
    #[serde(rename = "text")]
    Text,
    #[serde(rename = "masked")]
    Masked,
    #[serde(rename = "multiline")]
    Multiline,
    #[serde(rename = "confirm")]
    Confirm,
}
impl ::std::convert::From<&Self> for FieldType {
    fn from(value: &FieldType) -> Self {
        value.clone()
    }
}
impl ::std::fmt::Display for FieldType {
    fn fmt(&self, f: &mut ::std::fmt::Formatter<'_>) -> ::std::fmt::Result {
        match *self {
            Self::Text => f.write_str("text"),
            Self::Masked => f.write_str("masked"),
            Self::Multiline => f.write_str("multiline"),
            Self::Confirm => f.write_str("confirm"),
        }
    }
}
impl ::std::str::FromStr for FieldType {
    type Err = self::error::ConversionError;
    fn from_str(value: &str) -> ::std::result::Result<Self, self::error::ConversionError> {
        match value {
            "text" => Ok(Self::Text),
            "masked" => Ok(Self::Masked),
            "multiline" => Ok(Self::Multiline),
            "confirm" => Ok(Self::Confirm),
            _ => Err("invalid value".into()),
        }
    }
}
impl ::std::convert::TryFrom<&str> for FieldType {
    type Error = self::error::ConversionError;
    fn try_from(value: &str) -> ::std::result::Result<Self, self::error::ConversionError> {
        value.parse()
    }
}
impl ::std::convert::TryFrom<&::std::string::String> for FieldType {
    type Error = self::error::ConversionError;
    fn try_from(
        value: &::std::string::String,
    ) -> ::std::result::Result<Self, self::error::ConversionError> {
        value.parse()
    }
}
impl ::std::convert::TryFrom<::std::string::String> for FieldType {
    type Error = self::error::ConversionError;
    fn try_from(
        value: ::std::string::String,
    ) -> ::std::result::Result<Self, self::error::ConversionError> {
        value.parse()
    }
}
impl ::std::default::Default for FieldType {
    fn default() -> Self {
        FieldType::Text
    }
}
#[doc = "Configuration for a flowfile template; templates can be used to generate flow files."]
#[doc = r""]
#[doc = r" <details><summary>JSON schema</summary>"]
#[doc = r""]
#[doc = r" ```json"]
#[doc = "{"]
#[doc = "  \"$id\": \"https://flowexec.io/schemas/template_schema.json\","]
#[doc = "  \"title\": \"Template\","]
#[doc = "  \"description\": \"Configuration for a flowfile template; templates can be used to generate flow files.\","]
#[doc = "  \"type\": \"object\","]
#[doc = "  \"required\": ["]
#[doc = "    \"template\""]
#[doc = "  ],"]
#[doc = "  \"properties\": {"]
#[doc = "    \"artifacts\": {"]
#[doc = "      \"description\": \"A list of artifacts to be copied after generating the flow file.\","]
#[doc = "      \"type\": \"array\","]
#[doc = "      \"items\": {"]
#[doc = "        \"$ref\": \"#/definitions/Artifact\""]
#[doc = "      }"]
#[doc = "    },"]
#[doc = "    \"form\": {"]
#[doc = "      \"description\": \"Form fields to be displayed to the user when generating a flow file from a template. \\nThe form will be rendered first, and the user's input can be used to render the template.\\n\","]
#[doc = "      \"default\": [],"]
#[doc = "      \"type\": \"array\","]
#[doc = "      \"items\": {"]
#[doc = "        \"$ref\": \"#/definitions/Field\""]
#[doc = "      }"]
#[doc = "    },"]
#[doc = "    \"postRun\": {"]
#[doc = "      \"description\": \"A list of exec executables to run after generating the flow file.\","]
#[doc = "      \"type\": \"array\","]
#[doc = "      \"items\": {"]
#[doc = "        \"$ref\": \"#/definitions/TemplateRefConfig\""]
#[doc = "      }"]
#[doc = "    },"]
#[doc = "    \"preRun\": {"]
#[doc = "      \"description\": \"A list of exec executables to run before generating the flow file.\","]
#[doc = "      \"type\": \"array\","]
#[doc = "      \"items\": {"]
#[doc = "        \"$ref\": \"#/definitions/TemplateRefConfig\""]
#[doc = "      }"]
#[doc = "    },"]
#[doc = "    \"template\": {"]
#[doc = "      \"description\": \"The flow file template to generate. The template must be a valid flow file after rendering.\","]
#[doc = "      \"type\": \"string\""]
#[doc = "    }"]
#[doc = "  }"]
#[doc = "}"]
#[doc = r" ```"]
#[doc = r" </details>"]
#[derive(:: serde :: Deserialize, :: serde :: Serialize, Clone, Debug)]
pub struct Template {
    #[doc = "A list of artifacts to be copied after generating the flow file."]
    #[serde(default, skip_serializing_if = "::std::vec::Vec::is_empty")]
    pub artifacts: ::std::vec::Vec<Artifact>,
    #[doc = "Form fields to be displayed to the user when generating a flow file from a template. \nThe form will be rendered first, and the user's input can be used to render the template.\n"]
    #[serde(default, skip_serializing_if = "::std::vec::Vec::is_empty")]
    pub form: ::std::vec::Vec<Field>,
    #[doc = "A list of exec executables to run after generating the flow file."]
    #[serde(
        rename = "postRun",
        default,
        skip_serializing_if = "::std::vec::Vec::is_empty"
    )]
    pub post_run: ::std::vec::Vec<TemplateRefConfig>,
    #[doc = "A list of exec executables to run before generating the flow file."]
    #[serde(
        rename = "preRun",
        default,
        skip_serializing_if = "::std::vec::Vec::is_empty"
    )]
    pub pre_run: ::std::vec::Vec<TemplateRefConfig>,
    #[doc = "The flow file template to generate. The template must be a valid flow file after rendering."]
    pub template: ::std::string::String,
}
impl ::std::convert::From<&Template> for Template {
    fn from(value: &Template) -> Self {
        value.clone()
    }
}
impl Template {
    pub fn builder() -> builder::Template {
        Default::default()
    }
}
#[doc = "Configuration for a template executable."]
#[doc = r""]
#[doc = r" <details><summary>JSON schema</summary>"]
#[doc = r""]
#[doc = r" ```json"]
#[doc = "{"]
#[doc = "  \"description\": \"Configuration for a template executable.\","]
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
#[doc = "      \"description\": \"An expression that determines whether the executable should be run, using the Expr language syntax. \\nThe expression is evaluated at runtime and must resolve to a boolean value. If the condition is not met, \\nthe executable will be skipped.\\n\\nThe expression has access to OS/architecture information (os, arch), environment variables (env), form input \\n(form), and context information (name, workspace, directory, etc.).\\n\\nSee the [flow documentation](https://flowexec.io/guide/templating) for more information.\\n\","]
#[doc = "      \"default\": \"\","]
#[doc = "      \"type\": \"string\""]
#[doc = "    },"]
#[doc = "    \"ref\": {"]
#[doc = "      \"description\": \"A reference to another executable to run in serial.\\nOne of `cmd` or `ref` must be set.\\n\","]
#[doc = "      \"default\": \"\","]
#[doc = "      \"$ref\": \"#/definitions/ExecutableRef\""]
#[doc = "    }"]
#[doc = "  }"]
#[doc = "}"]
#[doc = r" ```"]
#[doc = r" </details>"]
#[derive(:: serde :: Deserialize, :: serde :: Serialize, Clone, Debug)]
pub struct TemplateRefConfig {
    #[doc = "Arguments to pass to the executable."]
    #[serde(default, skip_serializing_if = "::std::vec::Vec::is_empty")]
    pub args: ::std::vec::Vec<::std::string::String>,
    #[doc = "The command to execute.\nOne of `cmd` or `ref` must be set.\n"]
    #[serde(default)]
    pub cmd: ::std::string::String,
    #[doc = "An expression that determines whether the executable should be run, using the Expr language syntax. \nThe expression is evaluated at runtime and must resolve to a boolean value. If the condition is not met, \nthe executable will be skipped.\n\nThe expression has access to OS/architecture information (os, arch), environment variables (env), form input \n(form), and context information (name, workspace, directory, etc.).\n\nSee the [flow documentation](https://flowexec.io/guide/templating) for more information.\n"]
    #[serde(rename = "if", default)]
    pub if_: ::std::string::String,
    #[doc = "A reference to another executable to run in serial.\nOne of `cmd` or `ref` must be set.\n"]
    #[serde(rename = "ref", default = "defaults::template_ref_config_ref")]
    pub ref_: ExecutableRef,
}
impl ::std::convert::From<&TemplateRefConfig> for TemplateRefConfig {
    fn from(value: &TemplateRefConfig) -> Self {
        value.clone()
    }
}
impl ::std::default::Default for TemplateRefConfig {
    fn default() -> Self {
        Self {
            args: Default::default(),
            cmd: Default::default(),
            if_: Default::default(),
            ref_: defaults::template_ref_config_ref(),
        }
    }
}
impl TemplateRefConfig {
    pub fn builder() -> builder::TemplateRefConfig {
        Default::default()
    }
}
#[doc = r" Types for composing complex structures."]
pub mod builder {
    #[derive(Clone, Debug)]
    pub struct Artifact {
        as_template: ::std::result::Result<bool, ::std::string::String>,
        dst_dir: ::std::result::Result<::std::string::String, ::std::string::String>,
        dst_name: ::std::result::Result<::std::string::String, ::std::string::String>,
        if_: ::std::result::Result<::std::string::String, ::std::string::String>,
        src_dir: ::std::result::Result<::std::string::String, ::std::string::String>,
        src_name: ::std::result::Result<::std::string::String, ::std::string::String>,
    }
    impl ::std::default::Default for Artifact {
        fn default() -> Self {
            Self {
                as_template: Ok(Default::default()),
                dst_dir: Ok(Default::default()),
                dst_name: Ok(Default::default()),
                if_: Ok(Default::default()),
                src_dir: Ok(Default::default()),
                src_name: Err("no value supplied for src_name".to_string()),
            }
        }
    }
    impl Artifact {
        pub fn as_template<T>(mut self, value: T) -> Self
        where
            T: ::std::convert::TryInto<bool>,
            T::Error: ::std::fmt::Display,
        {
            self.as_template = value
                .try_into()
                .map_err(|e| format!("error converting supplied value for as_template: {}", e));
            self
        }
        pub fn dst_dir<T>(mut self, value: T) -> Self
        where
            T: ::std::convert::TryInto<::std::string::String>,
            T::Error: ::std::fmt::Display,
        {
            self.dst_dir = value
                .try_into()
                .map_err(|e| format!("error converting supplied value for dst_dir: {}", e));
            self
        }
        pub fn dst_name<T>(mut self, value: T) -> Self
        where
            T: ::std::convert::TryInto<::std::string::String>,
            T::Error: ::std::fmt::Display,
        {
            self.dst_name = value
                .try_into()
                .map_err(|e| format!("error converting supplied value for dst_name: {}", e));
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
        pub fn src_dir<T>(mut self, value: T) -> Self
        where
            T: ::std::convert::TryInto<::std::string::String>,
            T::Error: ::std::fmt::Display,
        {
            self.src_dir = value
                .try_into()
                .map_err(|e| format!("error converting supplied value for src_dir: {}", e));
            self
        }
        pub fn src_name<T>(mut self, value: T) -> Self
        where
            T: ::std::convert::TryInto<::std::string::String>,
            T::Error: ::std::fmt::Display,
        {
            self.src_name = value
                .try_into()
                .map_err(|e| format!("error converting supplied value for src_name: {}", e));
            self
        }
    }
    impl ::std::convert::TryFrom<Artifact> for super::Artifact {
        type Error = super::error::ConversionError;
        fn try_from(value: Artifact) -> ::std::result::Result<Self, super::error::ConversionError> {
            Ok(Self {
                as_template: value.as_template?,
                dst_dir: value.dst_dir?,
                dst_name: value.dst_name?,
                if_: value.if_?,
                src_dir: value.src_dir?,
                src_name: value.src_name?,
            })
        }
    }
    impl ::std::convert::From<super::Artifact> for Artifact {
        fn from(value: super::Artifact) -> Self {
            Self {
                as_template: Ok(value.as_template),
                dst_dir: Ok(value.dst_dir),
                dst_name: Ok(value.dst_name),
                if_: Ok(value.if_),
                src_dir: Ok(value.src_dir),
                src_name: Ok(value.src_name),
            }
        }
    }
    #[derive(Clone, Debug)]
    pub struct Field {
        default: ::std::result::Result<::std::string::String, ::std::string::String>,
        description: ::std::result::Result<::std::string::String, ::std::string::String>,
        group: ::std::result::Result<i64, ::std::string::String>,
        key: ::std::result::Result<::std::string::String, ::std::string::String>,
        prompt: ::std::result::Result<::std::string::String, ::std::string::String>,
        required: ::std::result::Result<bool, ::std::string::String>,
        type_: ::std::result::Result<super::FieldType, ::std::string::String>,
        validate: ::std::result::Result<::std::string::String, ::std::string::String>,
    }
    impl ::std::default::Default for Field {
        fn default() -> Self {
            Self {
                default: Ok(Default::default()),
                description: Ok(Default::default()),
                group: Ok(Default::default()),
                key: Err("no value supplied for key".to_string()),
                prompt: Err("no value supplied for prompt".to_string()),
                required: Ok(Default::default()),
                type_: Ok(super::defaults::field_type()),
                validate: Ok(Default::default()),
            }
        }
    }
    impl Field {
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
        pub fn group<T>(mut self, value: T) -> Self
        where
            T: ::std::convert::TryInto<i64>,
            T::Error: ::std::fmt::Display,
        {
            self.group = value
                .try_into()
                .map_err(|e| format!("error converting supplied value for group: {}", e));
            self
        }
        pub fn key<T>(mut self, value: T) -> Self
        where
            T: ::std::convert::TryInto<::std::string::String>,
            T::Error: ::std::fmt::Display,
        {
            self.key = value
                .try_into()
                .map_err(|e| format!("error converting supplied value for key: {}", e));
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
            T: ::std::convert::TryInto<super::FieldType>,
            T::Error: ::std::fmt::Display,
        {
            self.type_ = value
                .try_into()
                .map_err(|e| format!("error converting supplied value for type_: {}", e));
            self
        }
        pub fn validate<T>(mut self, value: T) -> Self
        where
            T: ::std::convert::TryInto<::std::string::String>,
            T::Error: ::std::fmt::Display,
        {
            self.validate = value
                .try_into()
                .map_err(|e| format!("error converting supplied value for validate: {}", e));
            self
        }
    }
    impl ::std::convert::TryFrom<Field> for super::Field {
        type Error = super::error::ConversionError;
        fn try_from(value: Field) -> ::std::result::Result<Self, super::error::ConversionError> {
            Ok(Self {
                default: value.default?,
                description: value.description?,
                group: value.group?,
                key: value.key?,
                prompt: value.prompt?,
                required: value.required?,
                type_: value.type_?,
                validate: value.validate?,
            })
        }
    }
    impl ::std::convert::From<super::Field> for Field {
        fn from(value: super::Field) -> Self {
            Self {
                default: Ok(value.default),
                description: Ok(value.description),
                group: Ok(value.group),
                key: Ok(value.key),
                prompt: Ok(value.prompt),
                required: Ok(value.required),
                type_: Ok(value.type_),
                validate: Ok(value.validate),
            }
        }
    }
    #[derive(Clone, Debug)]
    pub struct Template {
        artifacts: ::std::result::Result<::std::vec::Vec<super::Artifact>, ::std::string::String>,
        form: ::std::result::Result<::std::vec::Vec<super::Field>, ::std::string::String>,
        post_run:
            ::std::result::Result<::std::vec::Vec<super::TemplateRefConfig>, ::std::string::String>,
        pre_run:
            ::std::result::Result<::std::vec::Vec<super::TemplateRefConfig>, ::std::string::String>,
        template: ::std::result::Result<::std::string::String, ::std::string::String>,
    }
    impl ::std::default::Default for Template {
        fn default() -> Self {
            Self {
                artifacts: Ok(Default::default()),
                form: Ok(Default::default()),
                post_run: Ok(Default::default()),
                pre_run: Ok(Default::default()),
                template: Err("no value supplied for template".to_string()),
            }
        }
    }
    impl Template {
        pub fn artifacts<T>(mut self, value: T) -> Self
        where
            T: ::std::convert::TryInto<::std::vec::Vec<super::Artifact>>,
            T::Error: ::std::fmt::Display,
        {
            self.artifacts = value
                .try_into()
                .map_err(|e| format!("error converting supplied value for artifacts: {}", e));
            self
        }
        pub fn form<T>(mut self, value: T) -> Self
        where
            T: ::std::convert::TryInto<::std::vec::Vec<super::Field>>,
            T::Error: ::std::fmt::Display,
        {
            self.form = value
                .try_into()
                .map_err(|e| format!("error converting supplied value for form: {}", e));
            self
        }
        pub fn post_run<T>(mut self, value: T) -> Self
        where
            T: ::std::convert::TryInto<::std::vec::Vec<super::TemplateRefConfig>>,
            T::Error: ::std::fmt::Display,
        {
            self.post_run = value
                .try_into()
                .map_err(|e| format!("error converting supplied value for post_run: {}", e));
            self
        }
        pub fn pre_run<T>(mut self, value: T) -> Self
        where
            T: ::std::convert::TryInto<::std::vec::Vec<super::TemplateRefConfig>>,
            T::Error: ::std::fmt::Display,
        {
            self.pre_run = value
                .try_into()
                .map_err(|e| format!("error converting supplied value for pre_run: {}", e));
            self
        }
        pub fn template<T>(mut self, value: T) -> Self
        where
            T: ::std::convert::TryInto<::std::string::String>,
            T::Error: ::std::fmt::Display,
        {
            self.template = value
                .try_into()
                .map_err(|e| format!("error converting supplied value for template: {}", e));
            self
        }
    }
    impl ::std::convert::TryFrom<Template> for super::Template {
        type Error = super::error::ConversionError;
        fn try_from(value: Template) -> ::std::result::Result<Self, super::error::ConversionError> {
            Ok(Self {
                artifacts: value.artifacts?,
                form: value.form?,
                post_run: value.post_run?,
                pre_run: value.pre_run?,
                template: value.template?,
            })
        }
    }
    impl ::std::convert::From<super::Template> for Template {
        fn from(value: super::Template) -> Self {
            Self {
                artifacts: Ok(value.artifacts),
                form: Ok(value.form),
                post_run: Ok(value.post_run),
                pre_run: Ok(value.pre_run),
                template: Ok(value.template),
            }
        }
    }
    #[derive(Clone, Debug)]
    pub struct TemplateRefConfig {
        args: ::std::result::Result<::std::vec::Vec<::std::string::String>, ::std::string::String>,
        cmd: ::std::result::Result<::std::string::String, ::std::string::String>,
        if_: ::std::result::Result<::std::string::String, ::std::string::String>,
        ref_: ::std::result::Result<super::ExecutableRef, ::std::string::String>,
    }
    impl ::std::default::Default for TemplateRefConfig {
        fn default() -> Self {
            Self {
                args: Ok(Default::default()),
                cmd: Ok(Default::default()),
                if_: Ok(Default::default()),
                ref_: Ok(super::defaults::template_ref_config_ref()),
            }
        }
    }
    impl TemplateRefConfig {
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
    }
    impl ::std::convert::TryFrom<TemplateRefConfig> for super::TemplateRefConfig {
        type Error = super::error::ConversionError;
        fn try_from(
            value: TemplateRefConfig,
        ) -> ::std::result::Result<Self, super::error::ConversionError> {
            Ok(Self {
                args: value.args?,
                cmd: value.cmd?,
                if_: value.if_?,
                ref_: value.ref_?,
            })
        }
    }
    impl ::std::convert::From<super::TemplateRefConfig> for TemplateRefConfig {
        fn from(value: super::TemplateRefConfig) -> Self {
            Self {
                args: Ok(value.args),
                cmd: Ok(value.cmd),
                if_: Ok(value.if_),
                ref_: Ok(value.ref_),
            }
        }
    }
}
#[doc = r" Generation of default values for serde."]
pub mod defaults {
    pub(super) fn field_type() -> super::FieldType {
        super::FieldType::Text
    }
    pub(super) fn template_ref_config_ref() -> super::ExecutableRef {
        super::ExecutableRef("".to_string())
    }
}
