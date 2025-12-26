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
})
