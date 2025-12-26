package cli_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/spf13/cobra"

	"github.com/flowexec/flow/pkg/cli"
)

var _ = Describe("Hook Injection", func() {
	Describe("AddPreRunHook", func() {
		It("should add PreRun hook to command without existing hook", func() {
			cmd := &cobra.Command{Use: "test"}
			hookCalled := false

			cli.AddPreRunHook(cmd, func(c *cobra.Command, args []string) {
				hookCalled = true
			})

			Expect(cmd.PreRun).NotTo(BeNil())
			cmd.PreRun(cmd, []string{})
			Expect(hookCalled).To(BeTrue())
		})

		It("should chain with existing PreRun hook", func() {
			cmd := &cobra.Command{Use: "test"}
			firstCalled := false
			secondCalled := false
			callOrder := []int{}

			cmd.PreRun = func(c *cobra.Command, args []string) {
				firstCalled = true
				callOrder = append(callOrder, 1)
			}

			cli.AddPreRunHook(cmd, func(c *cobra.Command, args []string) {
				secondCalled = true
				callOrder = append(callOrder, 2)
			})

			cmd.PreRun(cmd, []string{})
			Expect(firstCalled).To(BeTrue())
			Expect(secondCalled).To(BeTrue())
			Expect(callOrder).To(Equal([]int{2, 1}), "New hook should run before existing hook")
		})

		It("should handle nil command gracefully", func() {
			Expect(func() {
				cli.AddPreRunHook(nil, func(c *cobra.Command, args []string) {})
			}).NotTo(Panic())
		})

		It("should handle nil hook gracefully", func() {
			cmd := &cobra.Command{Use: "test"}
			Expect(func() {
				cli.AddPreRunHook(cmd, nil)
			}).NotTo(Panic())
		})
	})

	Describe("AddPostRunHook", func() {
		It("should add PostRun hook to command without existing hook", func() {
			cmd := &cobra.Command{Use: "test"}
			hookCalled := false

			cli.AddPostRunHook(cmd, func(c *cobra.Command, args []string) {
				hookCalled = true
			})

			Expect(cmd.PostRun).NotTo(BeNil())
			cmd.PostRun(cmd, []string{})
			Expect(hookCalled).To(BeTrue())
		})

		It("should chain with existing PostRun hook", func() {
			cmd := &cobra.Command{Use: "test"}
			firstCalled := false
			secondCalled := false
			callOrder := []int{}

			cmd.PostRun = func(c *cobra.Command, args []string) {
				firstCalled = true
				callOrder = append(callOrder, 1)
			}

			cli.AddPostRunHook(cmd, func(c *cobra.Command, args []string) {
				secondCalled = true
				callOrder = append(callOrder, 2)
			})

			cmd.PostRun(cmd, []string{})
			Expect(firstCalled).To(BeTrue())
			Expect(secondCalled).To(BeTrue())
			Expect(callOrder).To(Equal([]int{1, 2}), "New hook should run after existing hook")
		})
	})

	Describe("AddPersistentPreRunHook", func() {
		It("should add PersistentPreRun hook", func() {
			cmd := &cobra.Command{Use: "test"}
			hookCalled := false

			cli.AddPersistentPreRunHook(cmd, func(c *cobra.Command, args []string) {
				hookCalled = true
			})

			Expect(cmd.PersistentPreRun).NotTo(BeNil())
			cmd.PersistentPreRun(cmd, []string{})
			Expect(hookCalled).To(BeTrue())
		})

		It("should chain with existing PersistentPreRun hook", func() {
			cmd := &cobra.Command{Use: "test"}
			callOrder := []int{}

			cmd.PersistentPreRun = func(c *cobra.Command, args []string) {
				callOrder = append(callOrder, 1)
			}

			cli.AddPersistentPreRunHook(cmd, func(c *cobra.Command, args []string) {
				callOrder = append(callOrder, 2)
			})

			cmd.PersistentPreRun(cmd, []string{})
			Expect(callOrder).To(Equal([]int{2, 1}))
		})
	})

	Describe("AddPersistentPostRunHook", func() {
		It("should add PersistentPostRun hook", func() {
			cmd := &cobra.Command{Use: "test"}
			hookCalled := false

			cli.AddPersistentPostRunHook(cmd, func(c *cobra.Command, args []string) {
				hookCalled = true
			})

			Expect(cmd.PersistentPostRun).NotTo(BeNil())
			cmd.PersistentPostRun(cmd, []string{})
			Expect(hookCalled).To(BeTrue())
		})

		It("should chain with existing PersistentPostRun hook", func() {
			cmd := &cobra.Command{Use: "test"}
			callOrder := []int{}

			cmd.PersistentPostRun = func(c *cobra.Command, args []string) {
				callOrder = append(callOrder, 1)
			}

			cli.AddPersistentPostRunHook(cmd, func(c *cobra.Command, args []string) {
				callOrder = append(callOrder, 2)
			})

			cmd.PersistentPostRun(cmd, []string{})
			Expect(callOrder).To(Equal([]int{1, 2}))
		})
	})

	Describe("ApplyHooksRecursive", func() {
		It("should apply hooks to root and all subcommands", func() {
			rootCmd := &cobra.Command{Use: "root"}
			subCmd1 := &cobra.Command{Use: "sub1"}
			subCmd2 := &cobra.Command{Use: "sub2"}
			rootCmd.AddCommand(subCmd1, subCmd2)

			preRunCalls := []string{}
			postRunCalls := []string{}

			cli.ApplyHooksRecursive(rootCmd,
				func(c *cobra.Command, args []string) {
					preRunCalls = append(preRunCalls, c.Use)
				},
				func(c *cobra.Command, args []string) {
					postRunCalls = append(postRunCalls, c.Use)
				},
			)

			// Verify hooks were added
			Expect(rootCmd.PreRun).NotTo(BeNil())
			Expect(rootCmd.PostRun).NotTo(BeNil())
			Expect(subCmd1.PreRun).NotTo(BeNil())
			Expect(subCmd1.PostRun).NotTo(BeNil())
			Expect(subCmd2.PreRun).NotTo(BeNil())
			Expect(subCmd2.PostRun).NotTo(BeNil())

			// Execute hooks
			rootCmd.PreRun(rootCmd, []string{})
			subCmd1.PreRun(subCmd1, []string{})
			subCmd2.PreRun(subCmd2, []string{})

			Expect(preRunCalls).To(ContainElements("root", "sub1", "sub2"))

			rootCmd.PostRun(rootCmd, []string{})
			subCmd1.PostRun(subCmd1, []string{})
			subCmd2.PostRun(subCmd2, []string{})

			Expect(postRunCalls).To(ContainElements("root", "sub1", "sub2"))
		})

		It("should handle nil hooks gracefully", func() {
			cmd := &cobra.Command{Use: "test"}

			Expect(func() {
				cli.ApplyHooksRecursive(cmd, nil, nil)
			}).NotTo(Panic())
		})

		It("should handle nil command gracefully", func() {
			Expect(func() {
				cli.ApplyHooksRecursive(nil, func(c *cobra.Command, args []string) {}, nil)
			}).NotTo(Panic())
		})
	})

	Describe("ApplyPersistentHooksRecursive", func() {
		It("should apply persistent hooks recursively", func() {
			rootCmd := &cobra.Command{Use: "root"}
			subCmd := &cobra.Command{Use: "sub"}
			rootCmd.AddCommand(subCmd)

			hookCalled := false

			cli.ApplyPersistentHooksRecursive(rootCmd,
				func(c *cobra.Command, args []string) {
					hookCalled = true
				},
				nil,
			)

			Expect(rootCmd.PersistentPreRun).NotTo(BeNil())
			Expect(subCmd.PersistentPreRun).NotTo(BeNil())

			rootCmd.PersistentPreRun(rootCmd, []string{})
			Expect(hookCalled).To(BeTrue())
		})
	})

	Describe("WrapRunFunc", func() {
		It("should wrap Run function with before and after hooks", func() {
			callOrder := []string{}
			cmd := &cobra.Command{
				Use: "test",
				Run: func(c *cobra.Command, args []string) {
					callOrder = append(callOrder, "run")
				},
			}

			cli.WrapRunFunc(cmd,
				func(c *cobra.Command, args []string) {
					callOrder = append(callOrder, "before")
				},
				func(c *cobra.Command, args []string) {
					callOrder = append(callOrder, "after")
				},
			)

			cmd.Run(cmd, []string{})
			Expect(callOrder).To(Equal([]string{"before", "run", "after"}))
		})

		It("should handle nil before hook", func() {
			callOrder := []string{}
			cmd := &cobra.Command{
				Use: "test",
				Run: func(c *cobra.Command, args []string) {
					callOrder = append(callOrder, "run")
				},
			}

			cli.WrapRunFunc(cmd, nil, func(c *cobra.Command, args []string) {
				callOrder = append(callOrder, "after")
			})

			cmd.Run(cmd, []string{})
			Expect(callOrder).To(Equal([]string{"run", "after"}))
		})

		It("should handle nil after hook", func() {
			callOrder := []string{}
			cmd := &cobra.Command{
				Use: "test",
				Run: func(c *cobra.Command, args []string) {
					callOrder = append(callOrder, "run")
				},
			}

			cli.WrapRunFunc(cmd, func(c *cobra.Command, args []string) {
				callOrder = append(callOrder, "before")
			}, nil)

			cmd.Run(cmd, []string{})
			Expect(callOrder).To(Equal([]string{"before", "run"}))
		})
	})

	Describe("WrapRunEFunc", func() {
		It("should wrap RunE function with before and after hooks", func() {
			callOrder := []string{}
			cmd := &cobra.Command{
				Use: "test",
				RunE: func(c *cobra.Command, args []string) error {
					callOrder = append(callOrder, "run")
					return nil
				},
			}

			cli.WrapRunEFunc(cmd,
				func(c *cobra.Command, args []string) {
					callOrder = append(callOrder, "before")
				},
				func(c *cobra.Command, args []string) {
					callOrder = append(callOrder, "after")
				},
			)

			err := cmd.RunE(cmd, []string{})
			Expect(err).NotTo(HaveOccurred())
			Expect(callOrder).To(Equal([]string{"before", "run", "after"}))
		})
	})
})
