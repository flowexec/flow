package internal

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/flowexec/flow/cmd/internal/flags"
	"github.com/flowexec/flow/internal/io"
	execIO "github.com/flowexec/flow/internal/io/executable"
	"github.com/flowexec/flow/internal/io/library"
	"github.com/flowexec/flow/pkg/context"
	flowErrors "github.com/flowexec/flow/pkg/errors"
	"github.com/flowexec/flow/pkg/logger"
	"github.com/flowexec/flow/types/common"
	"github.com/flowexec/flow/types/executable"
)

func RegisterBrowseCmd(ctx *context.Context, rootCmd *cobra.Command) {
	browseCmd := &cobra.Command{
		Use:     "browse [EXECUTABLE-REFERENCE]",
		Short:   "Discover and explore available executables.",
		Aliases: []string{"ls", "library"},
		Long: "Browse executables across workspaces.\n\n" +
			"  flow browse                # Interactive multi-pane executable browser\n" +
			"  flow browse --list         # Simple list view of executables\n" +
			"  flow browse VERB [ID]      # Detailed view of specific executable\n\n" +
			fmt.Sprintf(
				"See %s for more information on executable verbs and "+
					"%s for more information on executable references.",
				io.TypesDocsURL("flowfile", "executableverb"),
				io.TypesDocsURL("flowfile", "executableref"),
			),
		Args: cobra.MaximumNArgs(2),
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			verbStr := cmd.CalledAs()
			verb := executable.Verb(verbStr)
			execList, err := ctx.ExecutableCache.GetExecutableList()
			if err != nil {
				return nil, cobra.ShellCompDirectiveError
			}
			execIDs := make([]string, 0, len(execList))
			for _, e := range execList {
				if e.Verb.Equals(verb) {
					execIDs = append(execIDs, e.ID())
				}
			}
			return execIDs, cobra.ShellCompDirectiveNoFileComp
		},
		PreRun:  func(cmd *cobra.Command, args []string) { StartTUI(ctx, cmd) },
		PostRun: func(cmd *cobra.Command, args []string) { WaitForTUI(ctx, cmd) },
		Run:     func(cmd *cobra.Command, args []string) { browseFunc(ctx, cmd, args) },
	}
	RegisterFlag(ctx, browseCmd, *flags.ListFlag)
	RegisterFlag(ctx, browseCmd, *flags.OutputFormatFlag)
	RegisterFlag(ctx, browseCmd, *flags.FilterWorkspaceFlag)
	RegisterFlag(ctx, browseCmd, *flags.FilterNamespaceFlag)
	RegisterFlag(ctx, browseCmd, *flags.FilterVerbFlag)
	RegisterFlag(ctx, browseCmd, *flags.FilterTagFlag)
	RegisterFlag(ctx, browseCmd, *flags.FilterExecSubstringFlag)
	RegisterFlag(ctx, browseCmd, *flags.AllNamespacesFlag)
	RegisterFlag(ctx, browseCmd, *flags.VisibilityFlag)
	rootCmd.AddCommand(browseCmd)
}

func browseFunc(ctx *context.Context, cmd *cobra.Command, args []string) {
	if len(args) >= 1 {
		viewExecutable(ctx, cmd, args)
		return
	}

	listMode := flags.ValueFor[bool](cmd, *flags.ListFlag, false)
	if listMode || !TUIEnabled(ctx, cmd) {
		listExecutables(ctx, cmd, args)
		return
	}

	executableLibrary(ctx, cmd, args)
}

func executableLibrary(ctx *context.Context, cmd *cobra.Command, _ []string) {
	if !TUIEnabled(ctx, cmd) {
		logger.Log().FatalErr(errors.New("interactive discovery requires an interactive terminal"))
	}

	wsFilter := flags.ValueFor[string](cmd, *flags.FilterWorkspaceFlag, false)
	switch wsFilter {
	case ".":
		wsFilter = ctx.Config.CurrentWorkspace
	case executable.WildcardWorkspace:
		wsFilter = ""
	}

	nsFilter := flags.ValueFor[string](cmd, *flags.FilterNamespaceFlag, false)
	allNs := flags.ValueFor[bool](cmd, *flags.AllNamespacesFlag, false)
	switch {
	case allNs && nsFilter != "":
		logger.Log().PlainTextWarn("cannot use both --all and --namespace flags, ignoring --namespace")
		fallthrough
	case allNs:
		nsFilter = executable.WildcardNamespace
	case nsFilter == ".":
		nsFilter = ctx.Config.CurrentNamespace
	}

	verbFilter := flags.ValueFor[string](cmd, *flags.FilterVerbFlag, false)
	tagsFilter := flags.ValueFor[[]string](cmd, *flags.FilterTagFlag, false)
	subStr := flags.ValueFor[string](cmd, *flags.FilterExecSubstringFlag, false)

	visStr := flags.ValueFor[string](cmd, *flags.VisibilityFlag, false)
	visibilityFilter := common.VisibilityPrivate
	if visStr != "" {
		visibilityFilter = common.Visibility(visStr)
		if !isValidVisibility(visibilityFilter) {
			logger.Log().Fatalf("invalid visibility: %s (valid: public, private, internal, hidden)", visStr)
		}
	}

	allExecs, err := ctx.ExecutableCache.GetExecutableList()
	if err != nil {
		logger.Log().FatalErr(err)
	}
	allWs, err := ctx.WorkspacesCache.GetWorkspaceConfigList()
	if err != nil {
		logger.Log().FatalErr(err)
	}

	runFunc := func(ref string) error { return runByRef(ctx, cmd, ref) }
	libraryModel := library.NewLibraryView(
		ctx, allWs, allExecs,
		library.Filter{
			Workspace:  wsFilter,
			Namespace:  nsFilter,
			Verb:       executable.Verb(verbFilter),
			Tags:       tagsFilter,
			Substring:  subStr,
			Visibility: visibilityFilter,
		},
		logger.Theme(ctx.Config.Theme.String()),
		runFunc,
	)
	SetView(ctx, cmd, libraryModel)
}

func listExecutables(ctx *context.Context, cmd *cobra.Command, _ []string) {
	wsFilter := flags.ValueFor[string](cmd, *flags.FilterWorkspaceFlag, false)
	if wsFilter == "." {
		wsFilter = ctx.Config.CurrentWorkspace
	}

	nsFilter := flags.ValueFor[string](cmd, *flags.FilterNamespaceFlag, false)
	allNs := flags.ValueFor[bool](cmd, *flags.AllNamespacesFlag, false)
	switch {
	case allNs && nsFilter != "":
		logger.Log().PlainTextWarn("cannot use both --all and --namespace flags, ignoring --namespace")
		fallthrough
	case allNs:
		nsFilter = executable.WildcardNamespace
	case nsFilter == ".":
		nsFilter = ctx.Config.CurrentNamespace
	}

	verbFilter := flags.ValueFor[string](cmd, *flags.FilterVerbFlag, false)
	tagsFilter := flags.ValueFor[[]string](cmd, *flags.FilterTagFlag, false)
	outputFormat := flags.ValueFor[string](cmd, *flags.OutputFormatFlag, false)
	substr := flags.ValueFor[string](cmd, *flags.FilterExecSubstringFlag, false)

	visStr := flags.ValueFor[string](cmd, *flags.VisibilityFlag, false)
	visibilityFilter := common.VisibilityPrivate
	if visStr != "" {
		visibilityFilter = common.Visibility(visStr)
		if !isValidVisibility(visibilityFilter) {
			logger.Log().Fatalf("invalid visibility: %s (valid: public, private, internal, hidden)", visStr)
		}
	}

	allExecs, err := ctx.ExecutableCache.GetExecutableList()
	if err != nil {
		logger.Log().FatalErr(err)
	}
	filteredExec := allExecs
	filteredExec = filteredExec.
		FilterByWorkspaceWithVisibility(wsFilter, visibilityFilter).
		FilterByNamespace(nsFilter).
		FilterByVerb(executable.Verb(verbFilter)).
		FilterByTags(tagsFilter).
		FilterBySubstring(substr)

	if TUIEnabled(ctx, cmd) {
		runFunc := func(ref string) error { return runByRef(ctx, cmd, ref) }
		view := execIO.NewExecutableListView(ctx, filteredExec, runFunc)
		SetView(ctx, cmd, view)
	} else {
		execIO.PrintExecutableList(outputFormat, filteredExec)
	}
}

func viewExecutable(ctx *context.Context, cmd *cobra.Command, args []string) {
	verbStr := args[0]
	verb := executable.Verb(verbStr)
	if err := verb.Validate(); err != nil {
		logger.Log().FatalErr(err)
	}

	var execID string
	if len(args) > 1 {
		id := args[1]
		ws, ns, name := executable.MustParseExecutableID(id)
		if ws == executable.WildcardWorkspace {
			ws = ctx.CurrentWorkspace.AssignedName()
		}
		if ns == executable.WildcardNamespace && ctx.Config.CurrentNamespace != "" {
			ns = ctx.Config.CurrentNamespace
		}
		execID = executable.NewExecutableID(ws, ns, name)
	}
	ref := executable.NewRef(execID, verb)

	exec, err := ctx.ExecutableCache.GetExecutableByRef(ref)
	if err != nil && errors.Is(flowErrors.NewExecutableNotFoundError(ref.String()), err) {
		logger.Log().Debugf("Executable %s not found in cache, syncing cache", ref)
		if err := ctx.ExecutableCache.Update(); err != nil {
			logger.Log().FatalErr(err)
		}
		exec, err = ctx.ExecutableCache.GetExecutableByRef(ref)
	}
	if err != nil {
		logger.Log().FatalErr(err)
	} else if exec == nil {
		logger.Log().Fatalf("executable %s not found", ref)
	}

	outputFormat := flags.ValueFor[string](cmd, *flags.OutputFormatFlag, false)
	if TUIEnabled(ctx, cmd) {
		runFunc := func(ref string) error { return runByRef(ctx, cmd, ref) }
		view := execIO.NewExecutableView(ctx, exec, runFunc)
		SetView(ctx, cmd, view)
	} else {
		execIO.PrintExecutable(outputFormat, exec)
	}
}

func isValidVisibility(v common.Visibility) bool {
	switch v {
	case common.VisibilityPublic, common.VisibilityPrivate, common.VisibilityInternal, common.VisibilityHidden:
		return true
	default:
		return false
	}
}
