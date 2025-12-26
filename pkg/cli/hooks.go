package cli

import "github.com/spf13/cobra"

// AddPreRunHook adds a PreRun hook to a command, preserving any existing PreRun hook.
// If the command already has a PreRun hook, the new hook will run before it.
func AddPreRunHook(cmd *cobra.Command, hook HookFunc) {
	if cmd == nil || hook == nil {
		return
	}

	existingHook := cmd.PreRun
	if existingHook == nil {
		cmd.PreRun = hook
		return
	}

	cmd.PreRun = func(c *cobra.Command, args []string) {
		hook(c, args)
		existingHook(c, args)
	}
}

// AddPostRunHook adds a PostRun hook to a command, preserving any existing PostRun hook.
// If the command already has a PostRun hook, the new hook will run after it.
func AddPostRunHook(cmd *cobra.Command, hook HookFunc) {
	if cmd == nil || hook == nil {
		return
	}

	existingHook := cmd.PostRun
	if existingHook == nil {
		cmd.PostRun = hook
		return
	}

	cmd.PostRun = func(c *cobra.Command, args []string) {
		existingHook(c, args)
		hook(c, args)
	}
}

// AddPersistentPreRunHook adds a PersistentPreRun hook to a command, preserving any existing hook.
// If the command already has a PersistentPreRun hook, the new hook will run before it.
//
// PersistentPreRun hooks run before all commands in the subtree.
func AddPersistentPreRunHook(cmd *cobra.Command, hook HookFunc) {
	if cmd == nil || hook == nil {
		return
	}

	existingHook := cmd.PersistentPreRun
	if existingHook == nil {
		cmd.PersistentPreRun = hook
		return
	}

	cmd.PersistentPreRun = func(c *cobra.Command, args []string) {
		hook(c, args)
		existingHook(c, args)
	}
}

// AddPersistentPostRunHook adds a PersistentPostRun hook to a command, preserving any existing hook.
// If the command already has a PersistentPostRun hook, the new hook will run after it.
//
// PersistentPostRun hooks run after all commands in the subtree.
func AddPersistentPostRunHook(cmd *cobra.Command, hook HookFunc) {
	if cmd == nil || hook == nil {
		return
	}

	existingHook := cmd.PersistentPostRun
	if existingHook == nil {
		cmd.PersistentPostRun = hook
		return
	}

	cmd.PersistentPostRun = func(c *cobra.Command, args []string) {
		existingHook(c, args)
		hook(c, args)
	}
}
