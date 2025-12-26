package cli

import "github.com/spf13/cobra"

// AddPreRunHook adds a PreRun hook to a command, preserving any existing PreRun hook.
// If the command already has a PreRun hook, the new hook will run before it.
//
// Example:
//
//	cli.AddPreRunHook(cmd, func(cmd *cobra.Command, args []string) {
//	    log.Println("Starting command:", cmd.Name())
//	})
func AddPreRunHook(cmd *cobra.Command, hook HookFunc) {
	if cmd == nil || hook == nil {
		return
	}

	existingHook := cmd.PreRun
	if existingHook == nil {
		cmd.PreRun = hook
		return
	}

	// Chain the hooks: new hook runs first, then existing hook
	cmd.PreRun = func(c *cobra.Command, args []string) {
		hook(c, args)
		existingHook(c, args)
	}
}

// AddPostRunHook adds a PostRun hook to a command, preserving any existing PostRun hook.
// If the command already has a PostRun hook, the new hook will run after it.
//
// Example:
//
//	cli.AddPostRunHook(cmd, func(cmd *cobra.Command, args []string) {
//	    log.Println("Finished command:", cmd.Name())
//	})
func AddPostRunHook(cmd *cobra.Command, hook HookFunc) {
	if cmd == nil || hook == nil {
		return
	}

	existingHook := cmd.PostRun
	if existingHook == nil {
		cmd.PostRun = hook
		return
	}

	// Chain the hooks: existing hook runs first, then new hook
	cmd.PostRun = func(c *cobra.Command, args []string) {
		existingHook(c, args)
		hook(c, args)
	}
}

// AddPersistentPreRunHook adds a PersistentPreRun hook to a command, preserving any existing hook.
// If the command already has a PersistentPreRun hook, the new hook will run before it.
//
// PersistentPreRun hooks run before all commands in the subtree.
//
// Example:
//
//	cli.AddPersistentPreRunHook(rootCmd, func(cmd *cobra.Command, args []string) {
//	    log.Println("Initializing...")
//	})
func AddPersistentPreRunHook(cmd *cobra.Command, hook HookFunc) {
	if cmd == nil || hook == nil {
		return
	}

	existingHook := cmd.PersistentPreRun
	if existingHook == nil {
		cmd.PersistentPreRun = hook
		return
	}

	// Chain the hooks: new hook runs first, then existing hook
	cmd.PersistentPreRun = func(c *cobra.Command, args []string) {
		hook(c, args)
		existingHook(c, args)
	}
}

// AddPersistentPostRunHook adds a PersistentPostRun hook to a command, preserving any existing hook.
// If the command already has a PersistentPostRun hook, the new hook will run after it.
//
// PersistentPostRun hooks run after all commands in the subtree.
//
// Example:
//
//	cli.AddPersistentPostRunHook(rootCmd, func(cmd *cobra.Command, args []string) {
//	    log.Println("Cleaning up...")
//	})
func AddPersistentPostRunHook(cmd *cobra.Command, hook HookFunc) {
	if cmd == nil || hook == nil {
		return
	}

	existingHook := cmd.PersistentPostRun
	if existingHook == nil {
		cmd.PersistentPostRun = hook
		return
	}

	// Chain the hooks: existing hook runs first, then new hook
	cmd.PersistentPostRun = func(c *cobra.Command, args []string) {
		existingHook(c, args)
		hook(c, args)
	}
}

// ApplyHooksRecursive applies PreRun and PostRun hooks to a command and all its subcommands.
// This is useful for adding cross-cutting concerns like telemetry or logging to all commands.
//
// Pass nil for either preRun or postRun to skip adding that hook type.
//
// Example:
//
//	cli.ApplyHooksRecursive(rootCmd,
//	    func(cmd *cobra.Command, args []string) {
//	        telemetry.Start(cmd.Name())
//	    },
//	    func(cmd *cobra.Command, args []string) {
//	        telemetry.End(cmd.Name())
//	    },
//	)
func ApplyHooksRecursive(cmd *cobra.Command, preRun, postRun HookFunc) {
	if cmd == nil {
		return
	}

	// Add hooks to this command
	if preRun != nil {
		AddPreRunHook(cmd, preRun)
	}
	if postRun != nil {
		AddPostRunHook(cmd, postRun)
	}

	// Recursively apply to all subcommands
	for _, subCmd := range cmd.Commands() {
		ApplyHooksRecursive(subCmd, preRun, postRun)
	}
}

// ApplyPersistentHooksRecursive applies PersistentPreRun and PersistentPostRun hooks
// to a command and all its subcommands.
//
// Pass nil for either preRun or postRun to skip adding that hook type.
//
// Example:
//
//	cli.ApplyPersistentHooksRecursive(rootCmd,
//	    func(cmd *cobra.Command, args []string) {
//	        // Initialize resources
//	    },
//	    func(cmd *cobra.Command, args []string) {
//	        // Clean up resources
//	    },
//	)
func ApplyPersistentHooksRecursive(cmd *cobra.Command, preRun, postRun HookFunc) {
	if cmd == nil {
		return
	}

	// Add hooks to this command
	if preRun != nil {
		AddPersistentPreRunHook(cmd, preRun)
	}
	if postRun != nil {
		AddPersistentPostRunHook(cmd, postRun)
	}

	// Recursively apply to all subcommands
	for _, subCmd := range cmd.Commands() {
		ApplyPersistentHooksRecursive(subCmd, preRun, postRun)
	}
}

// WrapRunFunc wraps a command's Run function with before and after hooks.
// This is useful when you want to wrap the actual command execution logic
// rather than using PreRun/PostRun.
//
// Pass nil for either before or after to skip that hook.
//
// Example:
//
//	cli.WrapRunFunc(cmd,
//	    func(cmd *cobra.Command, args []string) {
//	        log.Println("Before run")
//	    },
//	    func(cmd *cobra.Command, args []string) {
//	        log.Println("After run")
//	    },
//	)
func WrapRunFunc(cmd *cobra.Command, before, after HookFunc) {
	if cmd == nil || cmd.Run == nil {
		return
	}

	existingRun := cmd.Run
	cmd.Run = func(c *cobra.Command, args []string) {
		if before != nil {
			before(c, args)
		}
		existingRun(c, args)
		if after != nil {
			after(c, args)
		}
	}
}

// WrapRunEFunc wraps a command's RunE function with before and after hooks.
// This is similar to WrapRunFunc but for commands that return errors.
//
// Pass nil for either before or after to skip that hook.
//
// Example:
//
//	cli.WrapRunEFunc(cmd,
//	    func(cmd *cobra.Command, args []string) {
//	        log.Println("Before run")
//	    },
//	    func(cmd *cobra.Command, args []string) {
//	        log.Println("After run")
//	    },
//	)
func WrapRunEFunc(cmd *cobra.Command, before, after HookFunc) {
	if cmd == nil || cmd.RunE == nil {
		return
	}

	existingRunE := cmd.RunE
	cmd.RunE = func(c *cobra.Command, args []string) error {
		if before != nil {
			before(c, args)
		}
		err := existingRunE(c, args)
		if after != nil {
			after(c, args)
		}
		return err
	}
}
