package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

// WalkCommands traverses the command tree starting from the root and applies
// a function to each command (including the root).
//
// The traversal is depth-first, visiting parent commands before their children.
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
func FindCommand(rootCmd *cobra.Command, name string) *cobra.Command {
	if rootCmd == nil || name == "" {
		return nil
	}

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
func ReplaceCommand(rootCmd *cobra.Command, oldName string, newCmd *cobra.Command) error {
	if rootCmd == nil || oldName == "" || newCmd == nil {
		return fmt.Errorf("invalid parameters: rootCmd, oldName, and newCmd must not be nil/empty")
	}

	oldCmd := FindCommand(rootCmd, oldName)
	if oldCmd == nil {
		return fmt.Errorf("command %q not found", oldName)
	}

	parent := oldCmd.Parent()
	if parent == nil {
		return fmt.Errorf("cannot replace root command")
	}

	parent.RemoveCommand(oldCmd)
	parent.AddCommand(newCmd)

	return nil
}

// RemoveCommand removes a command by name from the command tree.
//
// Returns an error if the command is not found or if it's the root command.
func RemoveCommand(rootCmd *cobra.Command, name string) error {
	if rootCmd == nil || name == "" {
		return fmt.Errorf("invalid parameters: rootCmd and name must not be nil/empty")
	}

	cmd := FindCommand(rootCmd, name)
	if cmd == nil {
		return fmt.Errorf("command %q not found", name)
	}

	parent := cmd.Parent()
	if parent == nil {
		return fmt.Errorf("cannot remove root command")
	}

	parent.RemoveCommand(cmd)

	return nil
}

// ListCommands returns a list of all command names in the tree.
// The list includes the root command and all subcommands.
func ListCommands(rootCmd *cobra.Command) []string {
	var names []string
	WalkCommands(rootCmd, func(cmd *cobra.Command) {
		names = append(names, cmd.Name())
	})
	return names
}

// GetSubcommands returns a map of subcommand names to commands for a given command.
// This is a convenience wrapper around cobra's Commands() method.
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
