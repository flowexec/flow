/*
Copyright © 2024 Jahvon Dockery <jahvondockery@gmail.com>
*/

package cmd

import (
	"github.com/spf13/cobra"

	"github.com/jahvon/pilotcli/internal/cmd/version"
)

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the application version.",
	Long:  `Print the application version.`,
	Run: func(cmd *cobra.Command, args []string) {
		version.Print()
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
