package internal

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/flowexec/flow/cmd/internal/flags"
	"github.com/flowexec/flow/internal/context"
	"github.com/flowexec/flow/internal/filesystem"
	"github.com/flowexec/flow/internal/io/executable"
	"github.com/flowexec/flow/internal/logger"
	"github.com/flowexec/flow/internal/runner"
	"github.com/flowexec/flow/internal/runner/exec"
	"github.com/flowexec/flow/internal/templates"
)

func RegisterTemplateCmd(ctx *context.Context, rootCmd *cobra.Command) {
	templateCmd := &cobra.Command{
		Use:     "template",
		Aliases: []string{"tmpl", "templates"},
		Short:   "Manage flowfile templates.",
	}
	registerGenerateTemplateCmd(ctx, templateCmd)
	registerAddTemplateCmd(ctx, templateCmd)
	registerListTemplateCmd(ctx, templateCmd)
	registerGetTemplateCmd(ctx, templateCmd)
	rootCmd.AddCommand(templateCmd)
}

func registerGenerateTemplateCmd(ctx *context.Context, templateCmd *cobra.Command) {
	generateCmd := &cobra.Command{
		Use:     "generate FLOWFILE_NAME [-w WORKSPACE ] [-o OUTPUT_DIR] [-f FILE | -t TEMPLATE]",
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

	ws := workspaceOrCurrent(ctx, workspaceName)
	if ws == nil {
		logger.Log().Fatalf("workspace %s not found", workspaceName)
	}

	tmpl := loadFlowfileTemplate(ctx, template, templateFilePath)
	if tmpl == nil {
		logger.Log().Fatalf("unable to load flowfile template")
	}

	flowFilename := tmpl.Name()
	if len(args) == 1 {
		flowFilename = args[0]
	}
	if err := templates.ProcessTemplate(ctx, tmpl, ws, flowFilename, outputPath); err != nil {
		logger.Log().FatalErr(err)
	}

	logger.Log().PlainTextSuccess(fmt.Sprintf("Template '%s' rendered successfully", flowFilename))
}

func registerAddTemplateCmd(ctx *context.Context, templateCmd *cobra.Command) {
	addCmd := &cobra.Command{
		Use:     "add NAME DEFINITION_TEMPLATE_PATH",
		Aliases: []string{"register", "new"},
		Short:   "Register a flowfile template by name.",
		Args:    cobra.ExactArgs(2),
		Run:     func(cmd *cobra.Command, args []string) { addTemplateFunc(ctx, cmd, args) },
	}
	templateCmd.AddCommand(addCmd)
}

func addTemplateFunc(ctx *context.Context, _ *cobra.Command, args []string) {
	name := args[0]
	flowFilePath := args[1]
	loadedTemplates, err := filesystem.LoadFlowFileTemplate(name, flowFilePath)
	if err != nil {
		logger.Log().FatalErr(err)
	}
	if err := loadedTemplates.Validate(); err != nil {
		logger.Log().FatalErr(err)
	}
	userConfig := ctx.Config
	if userConfig.Templates == nil {
		userConfig.Templates = map[string]string{}
	}
	userConfig.Templates[name] = flowFilePath
	if err := filesystem.WriteConfig(userConfig); err != nil {
		logger.Log().FatalErr(err)
	}
	logger.Log().PlainTextSuccess(fmt.Sprintf("Template %s set to %s", name, flowFilePath))
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
	// TODO: include unregistered templates within the current ws
	tmpls, err := filesystem.LoadFlowFileTemplates(ctx.Config.Templates)
	if err != nil {
		logger.Log().FatalErr(err)
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
				if err := templates.ProcessTemplate(ctx, tmpl, ws, tmpl.Name(), "//"); err != nil {
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
		logger.Log().Fatalf("unable to load flowfile template")
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

var templateLong = `Add rendered executables from a flowfile template to a workspace.

The WORKSPACE_NAME is the name of the workspace to initialize the flowfile template in.
The FLOWFILE_NAME is the name to give the flowfile (if applicable) when rendering its template.

One one of -f or -t must be provided and must point to a valid flowfile template.
The -o flag can be used to specify an output path within the workspace to create the flowfile and its artifacts in.`
