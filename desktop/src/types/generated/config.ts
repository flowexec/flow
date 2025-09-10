/**
 * This file was automatically generated from config_schema.json
 * DO NOT MODIFY IT BY HAND
 */

/**
 * User Configuration for the Flow CLI.
 * Includes configurations for workspaces, templates, I/O, and other settings for the CLI.
 *
 * It is read from the user's flow config directory:
 * - **MacOS**: `$HOME/Library/Application Support/flow`
 * - **Linux**: `$HOME/.config/flow`
 * - **Windows**: `%APPDATA%\flow`
 *
 * Alternatively, a custom path can be set using the `FLOW_CONFIG_PATH` environment variable.
 *
 */
export interface Config {
  colorOverride?: ColorPalette;
  /**
   * The name of the current namespace.
   *
   * Namespaces are used to reference executables in the CLI using the format `workspace:namespace/name`.
   * If the namespace is not set, only executables defined without a namespace will be discovered.
   *
   */
  currentNamespace?: string;
  /**
   * The name of the current vault. This should match a key in the `vaults` map.
   */
  currentVault?: string;
  /**
   * The name of the current workspace. This should match a key in the `workspaces` or `remoteWorkspaces` map.
   */
  currentWorkspace: string;
  /**
   * The default log mode to use when running executables.
   * This can either be `hidden`, `json`, `logfmt` or `text`
   *
   * `hidden` will not display any logs.
   * `json` will display logs in JSON format.
   * `logfmt` will display logs with a log level, timestamp, and message.
   * `text` will just display the log message.
   *
   */
  defaultLogMode?: string;
  /**
   * The default timeout to use when running executables.
   * This should be a valid duration string.
   *
   */
  defaultTimeout?: string;
  interactive?: Interactive;
  /**
   * A map of flowfile template names to their paths.
   */
  templates?: {
    [k: string]: string;
  };
  /**
   * The theme of the interactive UI.
   */
  theme?: 'default' | 'everforest' | 'dark' | 'light' | 'dracula' | 'tokyo-night';
  /**
   * A map of vault names to their paths. The path should be a valid absolute path to the vault file created by flow.
   */
  vaults?: {
    [k: string]: string;
  };
  /**
   * The mode of the workspace. This can be either `fixed` or `dynamic`.
   * In `fixed` mode, the current workspace used at runtime is always the one set in the currentWorkspace config field.
   * In `dynamic` mode, the current workspace used at runtime is determined by the current directory.
   * If the current directory is within a workspace, that workspace is used.
   *
   */
  workspaceMode?: 'fixed' | 'dynamic';
  /**
   * Map of workspace names to their paths. The path should be a valid absolute path to the workspace directory.
   *
   */
  workspaces: {
    [k: string]: string;
  };
  [k: string]: unknown;
}
/**
 * Override the default color palette for the interactive UI.
 * This can be used to customize the colors of the UI.
 *
 */
export interface ColorPalette {
  black?: string;
  body?: string;
  border?: string;
  /**
   * The style of the code block. For example, `monokai`, `dracula`, `github`, etc.
   * See [chroma styles](https://github.com/alecthomas/chroma/tree/master/styles) for available style names.
   *
   */
  codeStyle?: string;
  emphasis?: string;
  error?: string;
  gray?: string;
  info?: string;
  primary?: string;
  secondary?: string;
  success?: string;
  tertiary?: string;
  warning?: string;
  white?: string;
  [k: string]: unknown;
}
/**
 * Configurations for the interactive UI.
 */
export interface Interactive {
  enabled: boolean;
  /**
   * Whether to send a desktop notification when a command completes.
   */
  notifyOnCompletion?: boolean;
  /**
   * Whether to play a sound when a command completes.
   */
  soundOnCompletion?: boolean;
  [k: string]: unknown;
}
