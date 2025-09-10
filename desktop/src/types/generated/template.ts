/**
 * This file was automatically generated from template_schema.json
 * DO NOT MODIFY IT BY HAND
 */

/**
 * Configuration for a flowfile template; templates can be used to generate flow files.
 */
export interface Template {
  /**
   * A list of artifacts to be copied after generating the flow file.
   */
  artifacts?: Artifact[];
  /**
   * Form fields to be displayed to the user when generating a flow file from a template.
   * The form will be rendered first, and the user's input can be used to render the template.
   * For example, a form field with the key `name` can be used in the template as `{{.name}}`.
   *
   */
  form?: Field[];
  /**
   * A list of exec executables to run after generating the flow file.
   */
  postRun?: TemplateRefConfig[];
  /**
   * A list of exec executables to run before generating the flow file.
   */
  preRun?: TemplateRefConfig[];
  /**
   * The flow file template to generate. The template must be a valid flow file after rendering.
   */
  template: string;
  [k: string]: unknown;
}
/**
 * File source and destination configuration.
 * Go templating from form data is supported in all fields.
 *
 */
export interface Artifact {
  /**
   * If true, the artifact will be copied as a template file. The file will be rendered using Go templating from
   * the form data. [Expr language functions](https://expr-lang.org/docs/language-definition) are available for use in the template.
   *
   */
  asTemplate?: boolean;
  /**
   * The directory to copy the file to. If not set, the file will be copied to the root of the flow file directory.
   * The directory will be created if it does not exist.
   *
   */
  dstDir?: string;
  /**
   * The name of the file to copy to. If not set, the file will be copied with the same name.
   */
  dstName?: string;
  /**
   * An expression that determines whether the the artifact should be copied, using the Expr language syntax.
   * The expression is evaluated at runtime and must resolve to a boolean value. If the condition is not met,
   * the artifact will not be copied.
   *
   * The expression has access to OS/architecture information (os, arch), environment variables (env), form input
   * (form), and context information (name, workspace, directory, etc.).
   *
   * See the [flow documentation](https://flowexec.io/guide/templating) for more information.
   *
   */
  if?: string;
  /**
   * The directory to copy the file from.
   * If not set, the file will be copied from the directory of the template file.
   *
   */
  srcDir?: string;
  /**
   * The name of the file to copy.
   */
  srcName: string;
  [k: string]: unknown;
}
/**
 * A field to be displayed to the user when generating a flow file from a template.
 */
export interface Field {
  /**
   * The default value to use if a value is not set.
   */
  default?: string;
  /**
   * A description of the field.
   */
  description?: string;
  /**
   * The group to display the field in. Fields with the same group will be displayed together.
   */
  group?: number;
  /**
   * The key to associate the data with. This is used as the key in the template data map.
   */
  key: string;
  /**
   * A prompt to be displayed to the user when collecting an input value.
   */
  prompt: string;
  /**
   * If true, a value must be set. If false, the default value will be used if a value is not set.
   */
  required?: boolean;
  /**
   * The type of input field to display.
   */
  type?: 'text' | 'masked' | 'multiline' | 'confirm';
  /**
   * A regular expression to validate the input value against.
   */
  validate?: string;
  [k: string]: unknown;
}
/**
 * Configuration for a template executable.
 */
export interface TemplateRefConfig {
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
   * An expression that determines whether the executable should be run, using the Expr language syntax.
   * The expression is evaluated at runtime and must resolve to a boolean value. If the condition is not met,
   * the executable will be skipped.
   *
   * The expression has access to OS/architecture information (os, arch), environment variables (env), form input
   * (form), and context information (name, workspace, directory, etc.).
   *
   * See the [flow documentation](https://flowexec.io/guide/templating) for more information.
   *
   */
  if?: string;
  /**
   * A reference to another executable to run in serial.
   * One of `cmd` or `ref` must be set.
   *
   */
  ref?: string;
  [k: string]: unknown;
}
