/**
 * This file was automatically generated from flowfile_schema.json
 * DO NOT MODIFY IT BY HAND
 */

/**
 * Alternate names that can be used to reference the executable in the CLI.
 */
export type CommonAliases = string[];
export type ExecutableArgumentList = ExecutableArgument[];
export type ExecutableParameterList = ExecutableParameter[];
/**
 * A list of executables to run in parallel.
 * Each executable can be a command or a reference to another executable.
 *
 */
export type ExecutableParallelRefConfigList = ExecutableParallelRefConfig[];
/**
 * A list of executables to run in serial.
 * Each executable can be a command or a reference to another executable.
 *
 */
export type ExecutableSerialRefConfigList = ExecutableSerialRefConfig[];
/**
 * A list of tags.
 * Tags can be used with list commands to filter returned data.
 *
 */
export type CommonTags = string[];
export type Verb = unknown;
/**
 * The visibility of the executables to Flow.
 * If not set, the visibility will default to `public`.
 *
 * `public` executables can be executed and listed from anywhere.
 * `private` executables can be executed and listed only within their own workspace.
 * `internal` executables can be executed within their own workspace but are not listed.
 * `hidden` executables cannot be executed or listed.
 *
 */
export type CommonVisibility = 'public' | 'private' | 'internal' | 'hidden';
/**
 * DEPRECATED: Use `imports` instead
 */
export type FromFile = string[];
/**
 * A list of `.sh` files to convert into generated executables in the file's executable group.
 */
export type FromFile1 = string[];

/**
 * Configuration for a group of Flow CLI executables. The file must have the extension `.flow`, `.flow.yaml`, or `.flow.yml`
 * in order to be discovered by the CLI. It's configuration is used to define a group of executables with shared metadata
 * (namespace, tags, etc). A workspace can have multiple flow files located anywhere in the workspace directory
 *
 */
export interface FlowFile {
  /**
   * A description of the executables defined within the flow file. This description will used as a shared description
   * for all executables in the flow file.
   *
   */
  description?: string;
  /**
   * A path to a markdown file that contains the description of the executables defined within the flow file.
   */
  descriptionFile?: string;
  executables?: Executable[];
  fromFile?: FromFile;
  imports?: FromFile1;
  /**
   * The namespace to be given to all executables in the flow file.
   * If not set, the executables in the file will be grouped into the root (*) namespace.
   * Namespaces can be reused across multiple flow files.
   *
   * Namespaces are used to reference executables in the CLI using the format `workspace:namespace/name`.
   *
   */
  namespace?: string;
  /**
   * Tags to be applied to all executables defined within the flow file.
   */
  tags?: string[];
  visibility?: CommonVisibility;
  [k: string]: unknown;
}
/**
 * The executable schema defines the structure of an executable in the Flow CLI.
 * Executables are the building blocks of workflows and are used to define the actions that can be performed in a workspace.
 *
 */
export interface Executable {
  aliases?: CommonAliases;
  /**
   * A description of the executable.
   * This description is rendered as markdown in the interactive UI.
   *
   */
  description?: string;
  exec?: ExecutableExecExecutableType;
  launch?: ExecutableLaunchExecutableType;
  /**
   * An optional name for the executable.
   *
   * Name is used to reference the executable in the CLI using the format `workspace/namespace:name`.
   * [Verb group + Name] must be unique within the namespace of the workspace.
   *
   */
  name?: string;
  parallel?: ExecutableParallelExecutableType;
  render?: ExecutableRenderExecutableType;
  request?: ExecutableRequestExecutableType;
  serial?: ExecutableSerialExecutableType;
  tags?: CommonTags;
  /**
   * The maximum amount of time the executable is allowed to run before being terminated.
   * The timeout is specified in Go duration format (e.g. 30s, 5m, 1h).
   *
   */
  timeout?: string;
  /**
   * Keywords that describe the action an executable performs. Executables are configured with a single verb,
   * but core verbs have aliases that can be used interchangeably when referencing executables. This allows users
   * to use the verb that best describes the action they are performing.
   *
   * ### Default Verb Aliases
   *
   * - **Execution Group**: `exec`, `run`, `execute`
   * - **Retrieval Group**: `get`, `fetch`, `retrieve`
   * - **Display Group**: `show`, `view`, `list`
   * - **Configuration Group**: `configure`, `setup`
   * - **Update Group**: `update`, `upgrade`
   *
   * ### Usage Notes
   *
   * 1. [Verb + Name] must be unique within the namespace of the workspace.
   * 2. When referencing an executable, users can use any verb from the default or configured alias group.
   * 3. All other verbs are standalone and self-descriptive.
   *
   * ### Examples
   *
   * - An executable configured with the `exec` verb can also be referenced using "run" or "execute".
   * - An executable configured with `get` can also be called with "list", "show", or "view".
   * - Operations like `backup`, `migrate`, `flush` are standalone verbs without aliases.
   * - Use domain-specific verbs like `deploy`, `scale`, `tunnel` for clear operational intent.
   *
   * By providing minimal aliasing with comprehensive verb coverage, flow enables natural language operations
   * while maintaining simplicity and flexibility for diverse development and operations workflows.
   *
   */
  verb:
    | 'abort'
    | 'activate'
    | 'add'
    | 'analyze'
    | 'apply'
    | 'archive'
    | 'audit'
    | 'backup'
    | 'benchmark'
    | 'build'
    | 'bundle'
    | 'check'
    | 'clean'
    | 'clear'
    | 'commit'
    | 'compile'
    | 'compress'
    | 'configure'
    | 'connect'
    | 'copy'
    | 'create'
    | 'deactivate'
    | 'debug'
    | 'decompress'
    | 'decrypt'
    | 'delete'
    | 'deploy'
    | 'destroy'
    | 'disable'
    | 'disconnect'
    | 'edit'
    | 'enable'
    | 'encrypt'
    | 'erase'
    | 'exec'
    | 'execute'
    | 'export'
    | 'expose'
    | 'fetch'
    | 'fix'
    | 'flush'
    | 'format'
    | 'generate'
    | 'get'
    | 'import'
    | 'index'
    | 'init'
    | 'inspect'
    | 'install'
    | 'join'
    | 'kill'
    | 'launch'
    | 'lint'
    | 'list'
    | 'load'
    | 'lock'
    | 'login'
    | 'logout'
    | 'manage'
    | 'merge'
    | 'migrate'
    | 'modify'
    | 'monitor'
    | 'mount'
    | 'new'
    | 'notify'
    | 'open'
    | 'package'
    | 'partition'
    | 'patch'
    | 'pause'
    | 'ping'
    | 'preload'
    | 'prefetch'
    | 'profile'
    | 'provision'
    | 'publish'
    | 'purge'
    | 'push'
    | 'queue'
    | 'reboot'
    | 'recover'
    | 'refresh'
    | 'release'
    | 'reload'
    | 'remove'
    | 'request'
    | 'reset'
    | 'restart'
    | 'restore'
    | 'retrieve'
    | 'rollback'
    | 'run'
    | 'save'
    | 'scale'
    | 'scan'
    | 'schedule'
    | 'seed'
    | 'send'
    | 'serve'
    | 'set'
    | 'setup'
    | 'show'
    | 'snapshot'
    | 'start'
    | 'stash'
    | 'stop'
    | 'tag'
    | 'teardown'
    | 'terminate'
    | 'test'
    | 'tidy'
    | 'trace'
    | 'transform'
    | 'trigger'
    | 'tunnel'
    | 'undeploy'
    | 'uninstall'
    | 'unmount'
    | 'unset'
    | 'update'
    | 'upgrade'
    | 'validate'
    | 'verify'
    | 'view'
    | 'watch';
  /**
   * A list of aliases for the verb. This allows the executable to be referenced with multiple verbs.
   */
  verbAliases?: Verb[];
  visibility?: CommonVisibility;
  [k: string]: unknown;
}
/**
 * Standard executable type. Runs a command/file in a subprocess.
 */
export interface ExecutableExecExecutableType {
  args?: ExecutableArgumentList;
  /**
   * The command to execute.
   * Only one of `cmd` or `file` must be set.
   *
   */
  cmd?: string;
  /**
   * The directory to execute the command in.
   * If unset, the directory of the flow file will be used.
   * If set to `f:tmp`, a temporary directory will be created for the process.
   * If prefixed with `./`, the path will be relative to the current working directory.
   * If prefixed with `//`, the path will be relative to the workspace root.
   * Environment variables in the path will be expended at runtime.
   *
   */
  dir?: string;
  /**
   * The file to execute.
   * Only one of `cmd` or `file` must be set.
   *
   */
  file?: string;
  /**
   * The log mode to use when running the executable.
   * This can either be `hidden`, `json`, `logfmt` or `text`
   *
   */
  logMode?: string;
  params?: ExecutableParameterList;
  [k: string]: unknown;
}
export interface ExecutableArgument {
  /**
   * The default value to use if the argument is not provided.
   * If the argument is required and no default is provided, the executable will fail.
   *
   */
  default?: string;
  /**
   * The name of the environment variable that will be assigned the value.
   */
  envKey?: string;
  /**
   * The flag to use when setting the argument from the command line.
   * Either `flag` or `pos` must be set, but not both.
   *
   */
  flag?: string;
  /**
   * A path where the argument value will be temporarily written to disk.
   * The file will be created before execution and cleaned up afterwards.
   *
   */
  outputFile?: string;
  /**
   * The position of the argument in the command line ArgumentList. Values start at 1.
   * Either `flag` or `pos` must be set, but not both.
   *
   */
  pos?: number;
  /**
   * If the argument is required, the executable will fail if the argument is not provided.
   * If the argument is not required, the default value will be used if the argument is not provided.
   *
   */
  required?: boolean;
  /**
   * The type of the argument. This is used to determine how to parse the value of the argument.
   */
  type?: 'string' | 'int' | 'float' | 'bool';
  [k: string]: unknown;
}
/**
 * A parameter is a value that can be passed to an executable and all of its sub-executables.
 * Only one of `text`, `secretRef`, `prompt`, or `file` must be set. Specifying more than one will result in an error.
 *
 */
export interface ExecutableParameter {
  /**
   * A path to a file containing environment variables to be passed to the executable.
   * The file should contain one variable per line in the format `KEY=VALUE`.
   *
   */
  envFile?: string;
  /**
   * The name of the environment variable that will be assigned the value.
   *
   * When specified with `envFile`, only the environment variable with this name will be set.
   *
   */
  envKey?: string;
  /**
   * A path where the parameter value will be temporarily written to disk.
   * The file will be created before execution and cleaned up afterwards.
   *
   */
  outputFile?: string;
  /**
   * A prompt to be displayed to the user when collecting an input value.
   */
  prompt?: string;
  /**
   * A reference to a secret to be passed to the executable.
   */
  secretRef?: string;
  /**
   * A static value to be passed to the executable.
   */
  text?: string;
  [k: string]: unknown;
}
/**
 * Launches an application or opens a URI.
 */
export interface ExecutableLaunchExecutableType {
  /**
   * The application to launch the URI with.
   */
  app?: string;
  args?: ExecutableArgumentList;
  params?: ExecutableParameterList;
  /**
   * The URI to launch. This can be a file path or a web URL.
   */
  uri: string;
  [k: string]: unknown;
}
export interface ExecutableParallelExecutableType {
  args?: ExecutableArgumentList;
  /**
   * The directory to execute the command in.
   * If unset, the directory of the flow file will be used.
   * If set to `f:tmp`, a temporary directory will be created for the process.
   * If prefixed with `./`, the path will be relative to the current working directory.
   * If prefixed with `//`, the path will be relative to the workspace root.
   * Environment variables in the path will be expended at runtime.
   *
   */
  dir?: string;
  execs: ExecutableParallelRefConfigList;
  /**
   * End the parallel execution as soon as an exec exits with a non-zero status. This is the default behavior.
   * When set to false, all execs will be run regardless of the exit status of parallel execs.
   *
   */
  failFast?: boolean;
  /**
   * The maximum number of threads to use when executing the parallel executables.
   */
  maxThreads?: number;
  params?: ExecutableParameterList;
  [k: string]: unknown;
}
/**
 * Configuration for a parallel executable.
 */
export interface ExecutableParallelRefConfig {
  /**
   * Arguments to pass to the executable.
   */
  args?: string[];
  /**
   * The command to execute.
   * One of `cmd` or `ref` must be set.
   *
   */
  cmd?: string;
  /**
   * An expression that determines whether the executable should run, using the Expr language syntax.
   * The expression is evaluated at runtime and must resolve to a boolean value.
   *
   * The expression has access to OS/architecture information (os, arch), environment variables (env), stored data
   * (store), and context information (ctx) like workspace and paths.
   *
   * For example, `os == "darwin"` will only run on macOS, `len(store["feature"]) > 0` will run if a value exists
   * in the store, and `env["CI"] == "true"` will run in CI environments.
   * See the [Expr documentation](https://expr-lang.org/docs/language-definition) for more information.
   *
   */
  if?: string;
  /**
   * A reference to another executable to run in serial.
   * One of `cmd` or `ref` must be set.
   *
   */
  ref?: string;
  /**
   * The number of times to retry the executable if it fails.
   */
  retries?: number;
  [k: string]: unknown;
}
/**
 * Renders a markdown template file with data.
 */
export interface ExecutableRenderExecutableType {
  args?: ExecutableArgumentList;
  /**
   * The directory to execute the command in.
   * If unset, the directory of the flow file will be used.
   * If set to `f:tmp`, a temporary directory will be created for the process.
   * If prefixed with `./`, the path will be relative to the current working directory.
   * If prefixed with `//`, the path will be relative to the workspace root.
   * Environment variables in the path will be expended at runtime.
   *
   */
  dir?: string;
  params?: ExecutableParameterList;
  /**
   * The path to the JSON or YAML file containing the template data.
   */
  templateDataFile?: string;
  /**
   * The path to the markdown template file to render.
   */
  templateFile: string;
  [k: string]: unknown;
}
/**
 * Makes an HTTP request.
 */
export interface ExecutableRequestExecutableType {
  args?: ExecutableArgumentList;
  /**
   * The body of the request.
   */
  body?: string;
  /**
   * A map of headers to include in the request.
   */
  headers?: {
    [k: string]: string;
  };
  /**
   * If set to true, the response will be logged as program output.
   */
  logResponse?: boolean;
  /**
   * The HTTP method to use when making the request.
   */
  method?: 'GET' | 'POST' | 'PUT' | 'PATCH' | 'DELETE';
  params?: ExecutableParameterList;
  responseFile?: ExecutableRequestResponseFile;
  /**
   * The timeout for the request in Go duration format (e.g. 30s, 5m, 1h).
   */
  timeout?: string;
  /**
   * [Expr](https://expr-lang.org/docs/language-definition) expression used to transform the response before
   * saving it to a file or outputting it.
   *
   * The following variables are available in the expression:
   *   - `status`: The response status string.
   *   - `code`: The response status code.
   *   - `body`: The response body.
   *   - `headers`: The response headers.
   *
   * For example, to capitalize a JSON body field's value, you can use `upper(fromJSON(body)["field"])`.
   *
   */
  transformResponse?: string;
  /**
   * The URL to make the request to.
   */
  url: string;
  /**
   * A list of valid status codes. If the response status code is not in this list, the executable will fail.
   * If not set, the response status code will not be checked.
   *
   */
  validStatusCodes?: number[];
  [k: string]: unknown;
}
/**
 * Configuration for saving the response of a request to a file.
 */
export interface ExecutableRequestResponseFile {
  /**
   * The directory to execute the command in.
   * If unset, the directory of the flow file will be used.
   * If set to `f:tmp`, a temporary directory will be created for the process.
   * If prefixed with `./`, the path will be relative to the current working directory.
   * If prefixed with `//`, the path will be relative to the workspace root.
   * Environment variables in the path will be expended at runtime.
   *
   */
  dir?: string;
  /**
   * The name of the file to save the response to.
   */
  filename: string;
  /**
   * The format to save the response as.
   */
  saveAs?: 'raw' | 'json' | 'indented-json' | 'yaml' | 'yml';
  [k: string]: unknown;
}
/**
 * Executes a list of executables in serial.
 */
export interface ExecutableSerialExecutableType {
  args?: ExecutableArgumentList;
  /**
   * The directory to execute the command in.
   * If unset, the directory of the flow file will be used.
   * If set to `f:tmp`, a temporary directory will be created for the process.
   * If prefixed with `./`, the path will be relative to the current working directory.
   * If prefixed with `//`, the path will be relative to the workspace root.
   * Environment variables in the path will be expended at runtime.
   *
   */
  dir?: string;
  execs: ExecutableSerialRefConfigList;
  /**
   * End the serial execution as soon as an exec exits with a non-zero status. This is the default behavior.
   * When set to false, all execs will be run regardless of the exit status of the previous exec.
   *
   */
  failFast?: boolean;
  params?: ExecutableParameterList;
  [k: string]: unknown;
}
/**
 * Configuration for a serial executable.
 */
export interface ExecutableSerialRefConfig {
  /**
   * Arguments to pass to the executable.
   */
  args?: string[];
  /**
   * The command to execute.
   * One of `cmd` or `ref` must be set.
   *
   */
  cmd?: string;
  /**
   * An expression that determines whether the executable should run, using the Expr language syntax.
   * The expression is evaluated at runtime and must resolve to a boolean value.
   *
   * The expression has access to OS/architecture information (os, arch), environment variables (env), stored data
   * (store), and context information (ctx) like workspace and paths.
   *
   * For example, `os == "darwin"` will only run on macOS, `len(store["feature"]) > 0` will run if a value exists
   * in the store, and `env["CI"] == "true"` will run in CI environments.
   * See the [Expr documentation](https://expr-lang.org/docs/language-definition) for more information.
   *
   */
  if?: string;
  /**
   * A reference to another executable to run in serial.
   * One of `cmd` or `ref` must be set.
   *
   */
  ref?: string;
  /**
   * The number of times to retry the executable if it fails.
   */
  retries?: number;
  /**
   * If set to true, the user will be prompted to review the output of the executable before continuing.
   */
  reviewRequired?: boolean;
  [k: string]: unknown;
}
