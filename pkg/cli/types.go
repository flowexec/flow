package cli

import "github.com/spf13/cobra"

// HookFunc is a function that can be used as a PreRun or PostRun hook for a command.
// It receives the command being executed and its arguments.
type HookFunc func(cmd *cobra.Command, args []string)

// RootConfig holds configuration options for building the root command.
type RootConfig struct {
	// Use is the one-line usage message (defaults to "flow")
	Use string

	// Short is the short description shown in help (defaults to Flow's standard description)
	Short string

	// Long is the long description shown in help (defaults to Flow's standard description)
	Long string

	// Version is the version string (defaults to Flow's current version)
	Version string
}

// RootOption is a functional option for configuring the root command.
type RootOption func(*RootConfig)

// WithUse sets the Use field for the root command.
func WithUse(use string) RootOption {
	return func(c *RootConfig) {
		c.Use = use
	}
}

// WithShort sets the Short description for the root command.
func WithShort(short string) RootOption {
	return func(c *RootConfig) {
		c.Short = short
	}
}

// WithLong sets the Long description for the root command.
func WithLong(long string) RootOption {
	return func(c *RootConfig) {
		c.Long = long
	}
}

// WithVersion sets the Version string for the root command.
func WithVersion(version string) RootOption {
	return func(c *RootConfig) {
		c.Version = version
	}
}
