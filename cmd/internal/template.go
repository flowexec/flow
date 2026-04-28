package internal

import (
	"fmt"
	"maps"
	"slices"
	"strings"

	"github.com/spf13/cobra"

	errhandler "github.com/flowexec/flow/cmd/internal/errors"
	"github.com/flowexec/flow/cmd/internal/flags"
	"github.com/flowexec/flow/cmd/internal/response"
	"github.com/flowexec/flow/internal/io/executable"
	"github.com/flowexec/flow/internal/runner"
	"github.com/flowexec/flow/internal/runner/exec"
	"github.com/flowexec/flow/internal/templates"
	"github.com/flowexec/flow/pkg/context"
	"github.com/flowexec/flow/pkg/filesystem"
	"github.com/flowexec/flow/pkg/logger"
)

func RegisterTemplateCmd(ctx *context.Context, rootCmd *cobra.Command) {
	templateCmd := &cobra.Command{
		Use:     "template",
		Aliases: []string{"tmpl", "templates"},
		Short:   "Manage flowfile templates.",
		Long:    templateParentLong,
	}
	registerGenerateTemplateCmd(ctx, templateCmd)
	registerAddTemplateCmd(ctx, templateCmd)
	registerRemoveTemplateCmd(ctx, templateCmd)
	registerListTemplateCmd(ctx, templateCmd)
	registerGetTemplateCmd(ctx, templateCmd)
	rootCmd.AddCommand(templateCmd)
}

func registerGenerateTemplateCmd(ctx *context.Context, templateCmd *cobra.Command) {
	generateCmd := &cobra.Command{
		Use:     "generate FLOWFILE_NAME [-w WORKSPACE ] [-d OUTPUT_DIR] [-f FILE | -t TEMPLATE]",
		Aliases: []string{"gen", "scaffold"},
		Short:   "Generate workspace executables and scaffolding from a flowfile template.",
		Long:    templateLong,
		Args:    cobra.MaximumNArgs(1),
		PreRun:  func(cmd *cobra.Command, args []string) { runner.RegisterRunner(exec.NewRunner()) },
		Run:     func(cmd *cobra.Command, args []string) { generateTemplateFunc(ctx, cmd, args) },
	}
	RegisterFlag(ctx, generateCmd, *flags.TemplateOutputPathFlag)
	RegisterFlag(ctx, generateCmd, *flags.TemplateFlag)
	RegisterFlag(ctx, generateCmd, *flags.TemplateFilePathFlag)
	RegisterFlag(ctx, generateCmd, *flags.TemplateWorkspaceFlag)
	RegisterFlag(ctx, generateCmd, *flags.TemplateFieldFlag)
	RegisterFlag(ctx, generateCmd, *flags.OutputFormatFlag)
	MarkFlagMutuallyExclusive(generateCmd, flags.TemplateFlag.Name, flags.TemplateFilePathFlag.Name)
	MarkOneFlagRequired(generateCmd, flags.TemplateFlag.Name, flags.TemplateFilePathFlag.Name)
	MarkFlagFilename(ctx, generateCmd, flags.TemplateFilePathFlag.Name)
	MarkFlagFilename(ctx, generateCmd, flags.TemplateOutputPathFlag.Name)
	templateCmd.AddCommand(generateCmd)
}

func generateTemplateFunc(ctx *context.Context, cmd *cobra.Command, args []string) {
	outputPath := flags.ValueFor[string](cmd, *flags.TemplateOutputPathFlag, false)
	template := flags.ValueFor[string](cmd, *flags.TemplateFlag, false)
	templateFilePath := flags.ValueFor[string](cmd, *flags.TemplateFilePathFlag, false)
	workspaceName := flags.ValueFor[string](cmd, *flags.TemplateWorkspaceFlag, false)
	fieldOverrides := flags.ValueFor[[]string](cmd, *flags.TemplateFieldFlag, false)

	preseeded := make(map[string]string)
	for _, override := range fieldOverrides {
		parts := strings.SplitN(override, "=", 2)
		if len(parts) == 2 {
			preseeded[parts[0]] = parts[1]
		}
	}

	ws := workspaceOrCurrent(ctx, workspaceName)
	if ws == nil {
		errhandler.HandleFatal(ctx, cmd, fmt.Errorf("workspace %s not found", workspaceName))
	}

	tmpl := loadFlowfileTemplate(ctx, template, templateFilePath)
	if tmpl == nil {
		errhandler.HandleFatal(ctx, cmd, fmt.Errorf("unable to load flowfile template"))
	}

	flowFilename := tmpl.Name()
	if len(args) == 1 {
		flowFilename = args[0]
	}
	result, err := templates.ProcessTemplate(ctx, tmpl, ws, flowFilename, outputPath, preseeded)
	if err != nil {
		errhandler.HandleFatal(ctx, cmd, err)
	}

	response.HandleSuccess(ctx, cmd, fmt.Sprintf("Template '%s' rendered successfully", flowFilename), map[string]any{
		"flowfileName": result.FlowfileName,
		"flowfilePath": result.FlowfilePath,
		"outputDir":    result.OutputDir,
		"formValues":   result.FormValues,
		"artifacts":    result.Artifacts,
	})
}

func registerAddTemplateCmd(ctx *context.Context, templateCmd *cobra.Command) {
	addCmd := &cobra.Command{
		Use:     "add NAME DEFINITION_TEMPLATE_PATH",
		Aliases: []string{"register", "new"},
		Short:   "Register a flowfile template by name.",
		Args:    cobra.ExactArgs(2),
		Run:     func(cmd *cobra.Command, args []string) { addTemplateFunc(ctx, cmd, args) },
	}
	RegisterFlag(ctx, addCmd, *flags.OutputFormatFlag)
	templateCmd.AddCommand(addCmd)
}

func addTemplateFunc(ctx *context.Context, cmd *cobra.Command, args []string) {
	name := args[0]
	flowFilePath := args[1]
	loadedTemplates, err := filesystem.LoadFlowFileTemplate(name, flowFilePath)
	if err != nil {
		errhandler.HandleFatal(ctx, cmd, err)
	}
	if err := loadedTemplates.Validate(); err != nil {
		errhandler.HandleFatal(ctx, cmd, err)
	}
	userConfig := ctx.Config
	if userConfig.Templates == nil {
		userConfig.Templates = map[string]string{}
	}
	userConfig.Templates[name] = flowFilePath
	if err := filesystem.WriteConfig(userConfig); err != nil {
		errhandler.HandleFatal(ctx, cmd, err)
	}
	response.HandleSuccess(ctx, cmd, fmt.Sprintf("Template %s set to %s", name, flowFilePath), map[string]any{
		"name": name,
		"path": flowFilePath,
	})
}

func registerRemoveTemplateCmd(ctx *context.Context, templateCmd *cobra.Command) {
	removeCmd := &cobra.Command{
		Use:     "remove NAME",
		Aliases: []string{"delete", "rm", "unregister"},
		Short:   "Unregister a flowfile template by name.",
		Args:    cobra.ExactArgs(1),
		ValidArgsFunction: func(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
			return slices.Sorted(maps.Keys(ctx.Config.Templates)), cobra.ShellCompDirectiveNoFileComp
		},
		Run: func(cmd *cobra.Command, args []string) { removeTemplateFunc(ctx, cmd, args) },
	}
	RegisterFlag(ctx, removeCmd, *flags.OutputFormatFlag)
	templateCmd.AddCommand(removeCmd)
}

func removeTemplateFunc(ctx *context.Context, cmd *cobra.Command, args []string) {
	name := args[0]

	userConfig := ctx.Config
	previousPath, found := userConfig.Templates[name]
	if !found {
		errhandler.HandleFatal(ctx, cmd, fmt.Errorf("template %s not found", name))
	}

	delete(userConfig.Templates, name)
	if err := filesystem.WriteConfig(userConfig); err != nil {
		errhandler.HandleFatal(ctx, cmd, err)
	}

	response.HandleSuccess(ctx, cmd, fmt.Sprintf("Template %s removed", name), map[string]any{
		"name":         name,
		"previousPath": previousPath,
	})
}

func registerListTemplateCmd(ctx *context.Context, templateCmd *cobra.Command) {
	listCmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List registered flowfile templates.",
		Args:    cobra.NoArgs,
		PreRun:  func(cmd *cobra.Command, args []string) { StartTUI(ctx, cmd) },
		PostRun: func(cmd *cobra.Command, args []string) { WaitForTUI(ctx, cmd) },
		Run:     func(cmd *cobra.Command, args []string) { listTemplateFunc(ctx, cmd, args) },
	}
	RegisterFlag(ctx, listCmd, *flags.OutputFormatFlag)
	templateCmd.AddCommand(listCmd)
}

func listTemplateFunc(ctx *context.Context, cmd *cobra.Command, _ []string) {
	// TODO: include unregistered templates within the current ws;
	// add --annotation filter flags (mirroring browse / workspace list)
	tmpls, err := filesystem.LoadFlowFileTemplates(ctx.Config.Templates)
	if err != nil {
		errhandler.HandleFatal(ctx, cmd, err)
	}

	outputFormat := flags.ValueFor[string](cmd, *flags.OutputFormatFlag, false)
	if TUIEnabled(ctx, cmd) {
		view := executable.NewTemplateListView(
			ctx, tmpls,
			func(name string) error {
				tmpl := tmpls.Find(name)
				if tmpl == nil {
					return fmt.Errorf("template %s not found", name)
				}
				ws := ctx.CurrentWorkspace
				// TODO: support specifying a path/name
				if _, err := templates.ProcessTemplate(ctx, tmpl, ws, tmpl.Name(), "//", nil); err != nil {
					return err
				}
				logger.Log().PlainTextSuccess("Template rendered successfully")
				return nil
			},
		)
		SetView(ctx, cmd, view)
	} else {
		executable.PrintTemplateList(outputFormat, tmpls)
	}
}

func registerGetTemplateCmd(ctx *context.Context, getCmd *cobra.Command) {
	templateCmd := &cobra.Command{
		Use:     "get",
		Aliases: []string{"show", "view", "info"},
		Short:   "Get a flowfile template's details. Either it's registered name or file path can be used.",
		PreRun:  func(cmd *cobra.Command, args []string) { StartTUI(ctx, cmd) },
		PostRun: func(cmd *cobra.Command, args []string) { WaitForTUI(ctx, cmd) },
		Run:     func(cmd *cobra.Command, args []string) { getTemplateFunc(ctx, cmd, args) },
	}
	RegisterFlag(ctx, templateCmd, *flags.TemplateFlag)
	RegisterFlag(ctx, templateCmd, *flags.TemplateFilePathFlag)
	MarkOneFlagRequired(templateCmd, flags.TemplateFlag.Name, flags.TemplateFilePathFlag.Name)
	RegisterFlag(ctx, templateCmd, *flags.OutputFormatFlag)
	getCmd.AddCommand(templateCmd)
}

func getTemplateFunc(ctx *context.Context, cmd *cobra.Command, _ []string) {
	template := flags.ValueFor[string](cmd, *flags.TemplateFlag, false)
	templateFilePath := flags.ValueFor[string](cmd, *flags.TemplateFilePathFlag, false)

	tmpl := loadFlowfileTemplate(ctx, template, templateFilePath)
	if tmpl == nil {
		errhandler.HandleFatal(ctx, cmd, fmt.Errorf("unable to load flowfile template"))
	}

	outputFormat := flags.ValueFor[string](cmd, *flags.OutputFormatFlag, false)
	if TUIEnabled(ctx, cmd) {
		runFunc := func(ref string) error { return runByRef(ctx, cmd, ref) }
		view := executable.NewTemplateView(ctx, tmpl, runFunc)
		SetView(ctx, cmd, view)
	} else {
		executable.PrintTemplate(outputFormat, tmpl)
	}
}

const templateParentLong = `Manage flowfile templates. A template is a reusable flowfile scaffold that can generate
executables, directory structures, and configuration files via 'template generate'.

Templates are registered by name for easy reuse. Use 'template add' to register a
template, 'template generate' to scaffold from one, and 'template list' to see what's available.`

var templateLong = `Add rendered executables from a flowfile template to a workspace.

The WORKSPACE_NAME is the name of the workspace to initialize the flowfile template in.
The FLOWFILE_NAME is the name to give the flowfile (if applicable) when rendering its template.

One of -f or -t must be provided and must point to a valid flowfile template.
The -d flag can be used to specify an output directory within the workspace to create
the flowfile and its artifacts in.`
