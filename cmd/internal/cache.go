package internal

import (
	"fmt"
	"strings"

	"github.com/flowexec/tuikit/views"
	"github.com/spf13/cobra"

	errhandler "github.com/flowexec/flow/v2/cmd/internal/errors"
	"github.com/flowexec/flow/v2/cmd/internal/flags"
	"github.com/flowexec/flow/v2/cmd/internal/response"
	cacheIO "github.com/flowexec/flow/v2/internal/io/cache"
	"github.com/flowexec/flow/v2/pkg/context"
	"github.com/flowexec/flow/v2/pkg/logger"
	"github.com/flowexec/flow/v2/pkg/store"
)

func RegisterCacheCmd(ctx *context.Context, rootCmd *cobra.Command) {
	subCmd := &cobra.Command{
		Use:     "cache",
		Aliases: []string{"store", "data"}, // TODO: deprecate "store" alias
		Short:   "Manage temporary key-value data.",
		Long: "Manage temporary key-value data. " +
			"Values set outside executables runs persist globally, while values set within executables persist only for " +
			"that execution scope.",
		Args: cobra.NoArgs,
	}
	registerCacheSetCmd(ctx, subCmd)
	registerCacheGetCmd(ctx, subCmd)
	registerCacheListCmd(ctx, subCmd)
	registerCacheRemoveCmd(ctx, subCmd)
	registerCacheClearCmd(ctx, subCmd)
	rootCmd.AddCommand(subCmd)
}

func registerCacheSetCmd(ctx *context.Context, rootCmd *cobra.Command) {
	subCmd := &cobra.Command{
		Use:     "set KEY [VALUE]",
		Aliases: []string{"put", "add"},
		Short:   "Set cached data by key.",
		Long:    dataStoreDescription + "This will overwrite any existing value for the key.",
		Args:    cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			cacheSetFunc(ctx, cmd, args)
		},
	}
	RegisterFlag(ctx, subCmd, *flags.GlobalCacheFlag)
	RegisterFlag(ctx, subCmd, *flags.OutputFormatFlag)
	rootCmd.AddCommand(subCmd)
}

func cacheSetFunc(ctx *context.Context, cmd *cobra.Command, args []string) {
	key := args[0]

	var value string
	switch {
	case len(args) == 1:
		form, err := views.NewForm(
			logger.Theme(ctx.Config.Theme.String()),
			ctx.StdIn(),
			ctx.StdOut(),
			&views.FormField{
				Key:   "value",
				Type:  views.PromptTypeMultiline,
				Title: "Enter the value to store",
			})
		if err != nil {
			errhandler.HandleFatal(ctx, cmd, err)
		}
		if err = form.Run(ctx); err != nil {
			errhandler.HandleFatal(ctx, cmd, err)
		}
		value = form.FindByKey("value").Value()
	case len(args) == 2:
		value = args[1]
	default:
		logger.Log().PlainTextWarn(fmt.Sprintf("merging multiple (%d) arguments into a single value", len(args)-1))
		value = strings.Join(args[1:], " ")
	}

	bucketName := store.EnvironmentBucket()
	global := flags.ValueFor[bool](cmd, *flags.GlobalCacheFlag, false)
	if global {
		bucketName = store.RootBucket
	}
	if err := ctx.DataStore.SetProcessVar(bucketName, key, value); err != nil {
		errhandler.HandleFatal(ctx, cmd, err)
	}
	response.HandleSuccess(ctx, cmd, fmt.Sprintf("Key %q set in the cache", key), map[string]any{
		"key": key,
	})
}

func registerCacheGetCmd(ctx *context.Context, rootCmd *cobra.Command) {
	subCmd := &cobra.Command{
		Use:     "get KEY",
		Aliases: []string{"view", "show"},
		Short:   "Get cached data by key.",
		Long:    dataStoreDescription + "This will retrieve the value for the given key.",
		Args:    cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			cacheGetFunc(ctx, cmd, args)
		},
	}
	RegisterFlag(ctx, subCmd, *flags.GlobalCacheFlag)
	RegisterFlag(ctx, subCmd, *flags.OutputFormatFlag)
	rootCmd.AddCommand(subCmd)
}

func cacheGetFunc(ctx *context.Context, cmd *cobra.Command, args []string) {
	key := args[0]

	bucketName := store.EnvironmentBucket()
	global := flags.ValueFor[bool](cmd, *flags.GlobalCacheFlag, false)
	if global {
		bucketName = store.RootBucket
	}
	value, err := ctx.DataStore.GetProcessVar(bucketName, key)
	if err != nil {
		errhandler.HandleFatal(ctx, cmd, err)
	}
	response.HandleSuccess(ctx, cmd, value, map[string]any{
		"key":   key,
		"value": value,
	})
}

func registerCacheListCmd(ctx *context.Context, rootCmd *cobra.Command) {
	subCmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List all keys in the store.",
		Long:    dataStoreDescription + "This will list all keys currently stored in the data store.",
		Args:    cobra.NoArgs,
		PreRun:  func(cmd *cobra.Command, args []string) { StartTUI(ctx, cmd) },
		PostRun: func(cmd *cobra.Command, args []string) { WaitForTUI(ctx, cmd) },
		Run: func(cmd *cobra.Command, args []string) {
			cacheListFunc(ctx, cmd, args)
		},
	}
	RegisterFlag(ctx, subCmd, *flags.OutputFormatFlag)
	rootCmd.AddCommand(subCmd)
}

func cacheListFunc(ctx *context.Context, cmd *cobra.Command, _ []string) {
	data, err := ctx.DataStore.GetAllProcessVars(store.EnvironmentBucket())
	if err != nil {
		errhandler.HandleFatal(ctx, cmd, err)
	}
	outputFormat := flags.ValueFor[string](cmd, *flags.OutputFormatFlag, false)
	if TUIEnabled(ctx, cmd) {
		view := cacheIO.NewCacheListView(ctx.TUIContainer(), data)
		SetView(ctx, cmd, view)
	} else {
		cacheIO.PrintCache(data, outputFormat)
	}
}

func registerCacheRemoveCmd(ctx *context.Context, rootCmd *cobra.Command) {
	subCmd := &cobra.Command{
		Use:     "remove KEY",
		Aliases: []string{"rm", "delete"},
		Short:   "Remove a key from the cached data store.",
		Long:    dataStoreDescription + "This will remove the specified key and its value from the data store.",
		Args:    cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			cacheRemoveFunc(ctx, cmd, args)
		},
	}
	RegisterFlag(ctx, subCmd, *flags.GlobalCacheFlag)
	RegisterFlag(ctx, subCmd, *flags.OutputFormatFlag)
	rootCmd.AddCommand(subCmd)
}

func cacheRemoveFunc(ctx *context.Context, cmd *cobra.Command, args []string) {
	key := args[0]

	bucketName := store.EnvironmentBucket()
	global := flags.ValueFor[bool](cmd, *flags.GlobalCacheFlag, false)
	if global {
		bucketName = store.RootBucket
	}
	if err := ctx.DataStore.DeleteProcessVar(bucketName, key); err != nil {
		errhandler.HandleFatal(ctx, cmd, err)
	}
	response.HandleSuccess(ctx, cmd, fmt.Sprintf("Key %q removed from the cache", key), map[string]any{
		"key": key,
	})
}

func registerCacheClearCmd(ctx *context.Context, rootCmd *cobra.Command) {
	subCmd := &cobra.Command{
		Use:     "clear",
		Aliases: []string{"reset", "flush"},
		Short:   "Clear cache data. Use --all to remove data across all scopes.",
		Long:    dataStoreDescription + "This will remove all keys and values from the data store.",
		Args:    cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			cacheClearFunc(ctx, cmd, args)
		},
	}
	RegisterFlag(ctx, subCmd, *flags.StoreAllFlag)
	RegisterFlag(ctx, subCmd, *flags.OutputFormatFlag)
	rootCmd.AddCommand(subCmd)
}

func cacheClearFunc(ctx *context.Context, cmd *cobra.Command, _ []string) {
	full := flags.ValueFor[bool](cmd, *flags.StoreAllFlag, false)
	if full {
		if err := store.DestroyStore(); err != nil {
			errhandler.HandleFatal(ctx, cmd, err)
		}
		response.HandleSuccess(ctx, cmd, "Cache cleared", nil)
		return
	}
	if err := ctx.DataStore.DeleteProcessBucket(store.EnvironmentBucket()); err != nil {
		errhandler.HandleFatal(ctx, cmd, err)
	}
	response.HandleSuccess(ctx, cmd, "Cache cleared", nil)
}

var dataStoreDescription = "The data store is a key-value store that can be used to persist data across executions. " +
	"Values that are set outside of an executable will persist across all executions until they are cleared. " +
	"When set within an executable, the data will only persist across serial or parallel sub-executables but all " +
	"values will be cleared when the parent executable completes. " +
	"Use the --global flag to force use of the global cache scope, even when called from within an executable.\n\n"
