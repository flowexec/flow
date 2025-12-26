package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// WalkCommands traverses the command tree starting from the root and applies
// a function to each command (including the root).
//
// The traversal is depth-first, visiting parent commands before their children.
//
// Example:
//
//	cli.WalkCommands(rootCmd, func(cmd *cobra.Command) {
//	    fmt.Println("Command:", cmd.Name())
//	})
func WalkCommands(rootCmd *cobra.Command, fn func(*cobra.Command)) {
	if rootCmd == nil || fn == nil {
		return
	}

	// Apply function to current command
	fn(rootCmd)

	// Recursively apply to all subcommands
	for _, subCmd := range rootCmd.Commands() {
		WalkCommands(subCmd, fn)
	}
}

// FindCommand finds a command by name in the command tree.
// It performs a breadth-first search starting from the root.
//
// Returns nil if the command is not found.
//
// Example:
//
//	wsCmd := cli.FindCommand(rootCmd, "workspace")
//	if wsCmd != nil {
//	    // Found the workspace command
//	}
func FindCommand(rootCmd *cobra.Command, name string) *cobra.Command {
	if rootCmd == nil || name == "" {
		return nil
	}

	// Check if this is the command we're looking for
	if rootCmd.Name() == name {
		return rootCmd
	}

	// Search in immediate children first (breadth-first)
	for _, cmd := range rootCmd.Commands() {
		if cmd.Name() == name {
			return cmd
		}
	}

	// Search recursively in subcommands
	for _, cmd := range rootCmd.Commands() {
		if found := FindCommand(cmd, name); found != nil {
			return found
		}
	}

	return nil
}

// FindCommandPath finds a command by its full path (e.g., "workspace add").
// The path should be space-separated command names.
//
// Returns nil if the command is not found.
//
// Example:
//
//	addCmd := cli.FindCommandPath(rootCmd, "workspace add")
//	if addCmd != nil {
//	    // Found the workspace add command
//	}
func FindCommandPath(rootCmd *cobra.Command, path string) *cobra.Command {
	if rootCmd == nil || path == "" {
		return nil
	}

	// Use cobra's built-in Find method
	cmd, _, err := rootCmd.Find(splitPath(path))
	if err != nil {
		return nil
	}
	return cmd
}

// ReplaceCommand replaces a command with the given name with a new command.
// The new command will be added to the same parent as the old command.
//
// Returns an error if the old command is not found or if the replacement fails.
//
// Example:
//
//	customExec := &cobra.Command{
//	    Use: "exec",
//	    Run: customExecFunc,
//	}
//	if err := cli.ReplaceCommand(rootCmd, "exec", customExec); err != nil {
//	    log.Fatal(err)
//	}
func ReplaceCommand(rootCmd *cobra.Command, oldName string, newCmd *cobra.Command) error {
	if rootCmd == nil || oldName == "" || newCmd == nil {
		return fmt.Errorf("invalid parameters: rootCmd, oldName, and newCmd must not be nil/empty")
	}

	// Find the command to replace
	oldCmd := FindCommand(rootCmd, oldName)
	if oldCmd == nil {
		return fmt.Errorf("command %q not found", oldName)
	}

	// Get the parent of the old command
	parent := oldCmd.Parent()
	if parent == nil {
		return fmt.Errorf("cannot replace root command")
	}

	// Remove the old command
	parent.RemoveCommand(oldCmd)

	// Add the new command
	parent.AddCommand(newCmd)

	return nil
}

// RemoveCommand removes a command by name from the command tree.
//
// Returns an error if the command is not found or if it's the root command.
//
// Example:
//
//	if err := cli.RemoveCommand(rootCmd, "sync"); err != nil {
//	    log.Fatal(err)
//	}
func RemoveCommand(rootCmd *cobra.Command, name string) error {
	if rootCmd == nil || name == "" {
		return fmt.Errorf("invalid parameters: rootCmd and name must not be nil/empty")
	}

	// Find the command
	cmd := FindCommand(rootCmd, name)
	if cmd == nil {
		return fmt.Errorf("command %q not found", name)
	}

	// Get the parent
	parent := cmd.Parent()
	if parent == nil {
		return fmt.Errorf("cannot remove root command")
	}

	// Remove the command
	parent.RemoveCommand(cmd)

	return nil
}

// ListCommands returns a list of all command names in the tree.
// The list includes the root command and all subcommands.
//
// Example:
//
//	names := cli.ListCommands(rootCmd)
//	fmt.Println("Available commands:", names)
func ListCommands(rootCmd *cobra.Command) []string {
	var names []string
	WalkCommands(rootCmd, func(cmd *cobra.Command) {
		names = append(names, cmd.Name())
	})
	return names
}

// GetSubcommands returns a map of subcommand names to commands for a given command.
// This is a convenience wrapper around cobra's Commands() method.
//
// Example:
//
//	wsCmd := cli.FindCommand(rootCmd, "workspace")
//	subCmds := cli.GetSubcommands(wsCmd)
//	for name, cmd := range subCmds {
//	    fmt.Printf("Subcommand: %s - %s\n", name, cmd.Short)
//	}
func GetSubcommands(cmd *cobra.Command) map[string]*cobra.Command {
	if cmd == nil {
		return nil
	}

	subCmds := make(map[string]*cobra.Command)
	for _, subCmd := range cmd.Commands() {
		subCmds[subCmd.Name()] = subCmd
	}
	return subCmds
}

// CloneCommand creates a shallow copy of a command.
// This is useful when you want to modify a command without affecting the original.
//
// Note: This creates a shallow copy - the Run functions and other function fields
// will reference the same functions as the original.
//
// Example:
//
//	origCmd := cli.FindCommand(rootCmd, "exec")
//	clonedCmd := cli.CloneCommand(origCmd)
//	clonedCmd.Short = "Custom exec command"
func CloneCommand(cmd *cobra.Command) *cobra.Command {
	if cmd == nil {
		return nil
	}

	clone := &cobra.Command{
		Use:                cmd.Use,
		Aliases:            append([]string(nil), cmd.Aliases...),
		Short:              cmd.Short,
		Long:               cmd.Long,
		Example:            cmd.Example,
		ValidArgs:          append([]string(nil), cmd.ValidArgs...),
		Args:               cmd.Args,
		ArgAliases:         append([]string(nil), cmd.ArgAliases...),
		BashCompletionFunction: cmd.BashCompletionFunction,
		Deprecated:         cmd.Deprecated,
		Hidden:             cmd.Hidden,
		Annotations:        copyMap(cmd.Annotations),
		Version:            cmd.Version,
		PersistentPreRun:   cmd.PersistentPreRun,
		PersistentPreRunE:  cmd.PersistentPreRunE,
		PreRun:             cmd.PreRun,
		PreRunE:            cmd.PreRunE,
		Run:                cmd.Run,
		RunE:               cmd.RunE,
		PostRun:            cmd.PostRun,
		PostRunE:           cmd.PostRunE,
		PersistentPostRun:  cmd.PersistentPostRun,
		PersistentPostRunE: cmd.PersistentPostRunE,
		SilenceErrors:      cmd.SilenceErrors,
		SilenceUsage:       cmd.SilenceUsage,
		DisableFlagParsing: cmd.DisableFlagParsing,
		DisableAutoGenTag:  cmd.DisableAutoGenTag,
		DisableFlagsInUseLine: cmd.DisableFlagsInUseLine,
		DisableSuggestions: cmd.DisableSuggestions,
		SuggestionsMinimumDistance: cmd.SuggestionsMinimumDistance,
		TraverseChildren:   cmd.TraverseChildren,
		FParseErrWhitelist: cmd.FParseErrWhitelist,
	}

	// Copy flags
	cmd.Flags().VisitAll(func(f *pflag.Flag) {
		clone.Flags().AddFlag(f)
	})
	cmd.PersistentFlags().VisitAll(func(f *pflag.Flag) {
		clone.PersistentFlags().AddFlag(f)
	})

	return clone
}

// Helper function to split a command path into individual command names
func splitPath(path string) []string {
	var parts []string
	current := ""
	for _, ch := range path {
		if ch == ' ' {
			if current != "" {
				parts = append(parts, current)
				current = ""
			}
		} else {
			current += string(ch)
		}
	}
	if current != "" {
		parts = append(parts, current)
	}
	return parts
}

// Helper function to copy a map
func copyMap(m map[string]string) map[string]string {
	if m == nil {
		return nil
	}
	copy := make(map[string]string, len(m))
	for k, v := range m {
		copy[k] = v
	}
	return copy
}
