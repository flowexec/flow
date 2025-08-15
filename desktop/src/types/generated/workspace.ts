/**
 * This file was automatically generated from workspace_schema.json
 * DO NOT MODIFY IT BY HAND
 */

/**
 * A list of tags.
 * Tags can be used with list commands to filter returned data.
 *
 */
export type CommonTags = string[];

/**
 * Configuration for a workspace in the Flow CLI.
 * This configuration is used to define the settings for a workspace.
 * Every workspace has a workspace config file named `flow.yaml` in the root of the workspace directory.
 *
 */
export interface Workspace {
  /**
   * A description of the workspace. This description is rendered as markdown in the interactive UI.
   */
  description?: string;
  /**
   * A path to a markdown file that contains the description of the workspace.
   */
  descriptionFile?: string;
  /**
   * The display name of the workspace. This is used in the interactive UI.
   */
  displayName?: string;
  /**
   * A list of environment variable files to load for the workspace. These files should contain key-value pairs of environment variables.
   * By default, the `.env` file in the workspace root is loaded if it exists.
   *
   */
  envFiles?: string[];
  executables?: ExecutableFilter;
  tags?: CommonTags;
  verbAliases?: VerbAliases;
  [k: string]: unknown;
}
export interface ExecutableFilter {
  /**
   * A list of directories or file patterns to exclude from the executable search.
   * Supports directory paths (e.g., "node_modules/", "vendor/") and glob patterns for filenames (e.g., "*.js.flow", "*temp*").
   * Common exclusions like node_modules/, vendor/, third_party/, external/, and *.js.flow are excluded by default.
   *
   */
  excluded?: string[];
  /**
   * A list of directories or file patterns to include in the executable search.
   * Supports directory paths (e.g., "src/", "scripts/") and glob patterns for filenames (e.g., "*.test.flow", "example*").
   *
   */
  included?: string[];
  [k: string]: unknown;
}
/**
 * A map of executable verbs to valid aliases. This allows you to use custom aliases for exec commands in the workspace.
 * Setting this will override all of the default flow command aliases. The verbs and its mapped aliases must be valid flow verbs.
 *
 * If set to an empty object, verb aliases will be disabled.
 *
 */
export interface VerbAliases {
  [k: string]: string[];
}
