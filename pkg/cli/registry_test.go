package cli_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/spf13/cobra"

	"github.com/flowexec/flow/pkg/cli"
)

var _ = Describe("Command Registry", func() {
	Describe("WalkCommands", func() {
		It("should traverse all commands", func() {
			rootCmd := &cobra.Command{Use: "root"}
			sub1 := &cobra.Command{Use: "sub1"}
			sub2 := &cobra.Command{Use: "sub2"}
			subsub := &cobra.Command{Use: "subsub"}
			sub1.AddCommand(subsub)
			rootCmd.AddCommand(sub1, sub2)

			visited := []string{}
			cli.WalkCommands(rootCmd, func(cmd *cobra.Command) {
				visited = append(visited, cmd.Use)
			})

			Expect(visited).To(Equal([]string{"root", "sub1", "subsub", "sub2"}))
		})

		It("should handle nil command gracefully", func() {
			Expect(func() {
				cli.WalkCommands(nil, func(cmd *cobra.Command) {})
			}).NotTo(Panic())
		})

		It("should handle nil function gracefully", func() {
			cmd := &cobra.Command{Use: "test"}
			Expect(func() {
				cli.WalkCommands(cmd, nil)
			}).NotTo(Panic())
		})
	})

	Describe("FindCommand", func() {
		var rootCmd *cobra.Command

		BeforeEach(func() {
			rootCmd = &cobra.Command{Use: "root"}
			sub1 := &cobra.Command{Use: "sub1"}
			sub2 := &cobra.Command{Use: "sub2"}
			subsub := &cobra.Command{Use: "subsub"}
			sub1.AddCommand(subsub)
			rootCmd.AddCommand(sub1, sub2)
		})

		It("should find root command", func() {
			cmd := cli.FindCommand(rootCmd, "root")
			Expect(cmd).NotTo(BeNil())
			Expect(cmd.Use).To(Equal("root"))
		})

		It("should find immediate child", func() {
			cmd := cli.FindCommand(rootCmd, "sub1")
			Expect(cmd).NotTo(BeNil())
			Expect(cmd.Use).To(Equal("sub1"))
		})

		It("should find nested command", func() {
			cmd := cli.FindCommand(rootCmd, "subsub")
			Expect(cmd).NotTo(BeNil())
			Expect(cmd.Use).To(Equal("subsub"))
		})

		It("should return nil for non-existent command", func() {
			cmd := cli.FindCommand(rootCmd, "nonexistent")
			Expect(cmd).To(BeNil())
		})

		It("should handle nil root command", func() {
			cmd := cli.FindCommand(nil, "test")
			Expect(cmd).To(BeNil())
		})

		It("should handle empty name", func() {
			cmd := cli.FindCommand(rootCmd, "")
			Expect(cmd).To(BeNil())
		})
	})

	Describe("FindCommandPath", func() {
		var rootCmd *cobra.Command

		BeforeEach(func() {
			rootCmd = &cobra.Command{Use: "root"}
			sub1 := &cobra.Command{Use: "sub1"}
			subsub := &cobra.Command{Use: "subsub"}
			sub1.AddCommand(subsub)
			rootCmd.AddCommand(sub1)
		})

		It("should find command by single name", func() {
			cmd := cli.FindCommandPath(rootCmd, "sub1")
			Expect(cmd).NotTo(BeNil())
			Expect(cmd.Use).To(Equal("sub1"))
		})

		It("should find command by path", func() {
			cmd := cli.FindCommandPath(rootCmd, "sub1 subsub")
			Expect(cmd).NotTo(BeNil())
			Expect(cmd.Use).To(Equal("subsub"))
		})

		It("should return nil for invalid path", func() {
			cmd := cli.FindCommandPath(rootCmd, "nonexistent path")
			Expect(cmd).To(BeNil())
		})
	})

	Describe("ReplaceCommand", func() {
		var rootCmd *cobra.Command

		BeforeEach(func() {
			rootCmd = &cobra.Command{Use: "root"}
			oldCmd := &cobra.Command{Use: "old"}
			rootCmd.AddCommand(oldCmd)
		})

		It("should replace existing command", func() {
			newCmd := &cobra.Command{
				Use:   "old",
				Short: "New command",
			}

			err := cli.ReplaceCommand(rootCmd, "old", newCmd)
			Expect(err).NotTo(HaveOccurred())

			found := cli.FindCommand(rootCmd, "old")
			Expect(found).NotTo(BeNil())
			Expect(found.Short).To(Equal("New command"))
		})

		It("should return error for non-existent command", func() {
			newCmd := &cobra.Command{Use: "new"}
			err := cli.ReplaceCommand(rootCmd, "nonexistent", newCmd)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("not found"))
		})

		It("should return error for nil root", func() {
			newCmd := &cobra.Command{Use: "new"}
			err := cli.ReplaceCommand(nil, "old", newCmd)
			Expect(err).To(HaveOccurred())
		})

		It("should return error for nil new command", func() {
			err := cli.ReplaceCommand(rootCmd, "old", nil)
			Expect(err).To(HaveOccurred())
		})

		It("should return error when trying to replace root", func() {
			newCmd := &cobra.Command{Use: "root"}
			err := cli.ReplaceCommand(rootCmd, "root", newCmd)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("cannot replace root"))
		})
	})

	Describe("RemoveCommand", func() {
		var rootCmd *cobra.Command

		BeforeEach(func() {
			rootCmd = &cobra.Command{Use: "root"}
			cmd1 := &cobra.Command{Use: "cmd1"}
			cmd2 := &cobra.Command{Use: "cmd2"}
			rootCmd.AddCommand(cmd1, cmd2)
		})

		It("should remove existing command", func() {
			err := cli.RemoveCommand(rootCmd, "cmd1")
			Expect(err).NotTo(HaveOccurred())

			found := cli.FindCommand(rootCmd, "cmd1")
			Expect(found).To(BeNil())

			// cmd2 should still exist
			found = cli.FindCommand(rootCmd, "cmd2")
			Expect(found).NotTo(BeNil())
		})

		It("should return error for non-existent command", func() {
			err := cli.RemoveCommand(rootCmd, "nonexistent")
			Expect(err).To(HaveOccurred())
		})

		It("should return error when trying to remove root", func() {
			err := cli.RemoveCommand(rootCmd, "root")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("cannot remove root"))
		})
	})

	Describe("ListCommands", func() {
		It("should list all command names", func() {
			rootCmd := &cobra.Command{Use: "root"}
			sub1 := &cobra.Command{Use: "sub1"}
			sub2 := &cobra.Command{Use: "sub2"}
			rootCmd.AddCommand(sub1, sub2)

			names := cli.ListCommands(rootCmd)
			Expect(names).To(ConsistOf("root", "sub1", "sub2"))
		})

		It("should handle empty tree", func() {
			rootCmd := &cobra.Command{Use: "root"}
			names := cli.ListCommands(rootCmd)
			Expect(names).To(Equal([]string{"root"}))
		})
	})

	Describe("GetSubcommands", func() {
		It("should return map of subcommands", func() {
			rootCmd := &cobra.Command{Use: "root"}
			sub1 := &cobra.Command{Use: "sub1"}
			sub2 := &cobra.Command{Use: "sub2"}
			rootCmd.AddCommand(sub1, sub2)

			subCmds := cli.GetSubcommands(rootCmd)
			Expect(subCmds).To(HaveLen(2))
			Expect(subCmds).To(HaveKey("sub1"))
			Expect(subCmds).To(HaveKey("sub2"))
			Expect(subCmds["sub1"]).To(Equal(sub1))
			Expect(subCmds["sub2"]).To(Equal(sub2))
		})

		It("should return empty map for command with no subcommands", func() {
			cmd := &cobra.Command{Use: "test"}
			subCmds := cli.GetSubcommands(cmd)
			Expect(subCmds).To(BeEmpty())
		})

		It("should return nil for nil command", func() {
			subCmds := cli.GetSubcommands(nil)
			Expect(subCmds).To(BeNil())
		})
	})
})
