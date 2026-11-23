package internal

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	errhandler "github.com/flowexec/flow/v2/cmd/internal/errors"
	"github.com/flowexec/flow/v2/cmd/internal/flags"
	"github.com/flowexec/flow/v2/internal/digest"
	"github.com/flowexec/flow/v2/pkg/context"
	"github.com/flowexec/flow/v2/pkg/filesystem"
)

func RegisterDigestCmd(ctx *context.Context, rootCmd *cobra.Command) {
	subCmd := &cobra.Command{
		Use:   "digest SOURCE [SOURCE...]",
		Short: "Gather runs, executables and files into one document.",
		Long: "Resolve each source against run history, the executable cache and the workspace, and print " +
			"one labeled section per source. Sources are type:value pointers:\n\n" +
			"  run:<id>         one run's metadata and log tail; a unique ID prefix (6+ characters) works\n" +
			"  runs:<filter>    a table of matching runs; filter terms are ref:<text>, status:<completed|failed|" +
			"cancelled|running> (passed works too), ran:<age> (or since:) and limit:<n>; bare words match the " +
			"ref or an ad-hoc run's label\n" +
			"  exec:<verb>/<id> an executable's definition, script source and last 5 runs\n" +
			"  file:<path>      a file's contents, relative to the workspace root and confined to it",
		Example: digestExamples,
		Args:    cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			digestFunc(ctx, cmd, args)
		},
	}
	RegisterFlag(ctx, subCmd, *flags.OutputFormatFlag)
	RegisterFlag(ctx, subCmd, *flags.DigestMaxCharsFlag)
	RegisterFlag(ctx, subCmd, *flags.DigestMaxLinesFlag)
	RegisterFlag(ctx, subCmd, *flags.DigestNoHistoryFlag)
	rootCmd.AddCommand(subCmd)
}

func digestFunc(ctx *context.Context, cmd *cobra.Command, args []string) {
	format := strings.ToLower(flags.ValueFor[string](cmd, *flags.OutputFormatFlag, false))
	switch format {
	case "", "text", "json", "yaml", "yml":
	default:
		errhandler.HandleUsage(ctx, cmd, "unsupported output format %q (use text, json or yaml)", format)
	}
	opts := digest.Options{
		MaxChars:  flags.ValueFor[int](cmd, *flags.DigestMaxCharsFlag, false),
		MaxLines:  flags.ValueFor[int](cmd, *flags.DigestMaxLinesFlag, false),
		NoHistory: flags.ValueFor[bool](cmd, *flags.DigestNoHistoryFlag, false),
	}
	if opts.MaxChars <= 0 || opts.MaxLines <= 0 {
		errhandler.HandleUsage(ctx, cmd, "--max-chars and --max-lines must be positive")
	}

	deps := digest.Deps{
		Store:       ctx.DataStore,
		LogsDir:     filesystem.LogsDir(),
		Executables: ctx.ExecutableCache,
	}
	if ctx.CurrentWorkspace != nil {
		deps.WorkspaceRoot = ctx.CurrentWorkspace.Location()
		deps.CurrentWorkspace = ctx.CurrentWorkspace.AssignedName()
	}
	if ctx.Config != nil {
		deps.CurrentNamespace = ctx.Config.CurrentNamespace
	}

	doc, err := digest.Resolve(ctx, args, opts, deps)
	if err != nil {
		errhandler.HandleFatal(ctx, cmd, err)
	}

	var out string
	switch format {
	case "json":
		data, err := json.MarshalIndent(doc, "", "  ")
		if err != nil {
			errhandler.HandleFatal(ctx, cmd, err)
		}
		out = string(data) + "\n"
	case "yaml", "yml":
		data, err := yaml.Marshal(doc)
		if err != nil {
			errhandler.HandleFatal(ctx, cmd, err)
		}
		out = string(data)
	default:
		out = doc.Text()
	}
	_, _ = fmt.Fprint(ctx.StdOut(), out)
}

const digestExamples = `
  flow digest run:3f2a9c1e-...                     # one run's metadata and log tail
  flow digest runs:"test status:failed ran:7d"     # failed test runs from the last week
  flow digest exec:test/unit                       # definition, script and recent runs
  flow digest file:./flow.yaml exec:build/binary   # several sources in one document
  flow digest runs:status:failed -o json           # structured sections for scripts and agents
`
