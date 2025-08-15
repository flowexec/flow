//go:build e2e

package tests_test

import (
	stdCtx "context"
	"os"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/flowexec/flow/tests/utils"
)

var _ = Describe("cache e2e", Ordered, func() {
	var (
		ctx *utils.Context
		run *utils.CommandRunner
	)

	BeforeAll(func() {
		ctx = utils.NewContext(stdCtx.Background(), GinkgoTB())
		run = utils.NewE2ECommandRunner()
	})

	BeforeEach(func() {
		utils.ResetTestContext(ctx, GinkgoTB())
	})

	AfterEach(func() {
		ctx.Finalize()
	})

	When("setting a cache value (flow cache set)", Ordered, func() {
		It("should set the value successfully", func() {
			reader, writer, err := os.Pipe()
			Expect(err).NotTo(HaveOccurred())
			_, err = writer.Write([]byte("test-value\n"))
			Expect(err).ToNot(HaveOccurred())

			ctx.SetIO(reader, ctx.StdOut())
			Expect(run.Run(ctx.Context, "cache", "set", "test-key")).To(Succeed())
			out, err := readFileContent(ctx.StdOut())
			Expect(err).NotTo(HaveOccurred())
			Expect(out).To(ContainSubstring("Key \"test-key\" set in the cache"))
		})

		It("should set the value with direct argument", func() {
			Expect(run.Run(ctx.Context, "cache", "set", "direct-key", "direct-value")).To(Succeed())
			out, err := readFileContent(ctx.StdOut())
			Expect(err).NotTo(HaveOccurred())
			Expect(out).To(ContainSubstring("Key \"direct-key\" set in the cache"))
		})

		It("should handle multiple arguments as single value", func() {
			Expect(run.Run(ctx.Context, "cache", "set", "multi-key", "value1", "value2", "value3")).To(Succeed())
			out, err := readFileContent(ctx.StdOut())
			Expect(err).NotTo(HaveOccurred())
			Expect(out).To(ContainSubstring("Key \"multi-key\" set in the cache"))
		})
	})

	When("getting a cache value (flow cache get)", func() {
		It("should retrieve the value successfully", func() {
			stdOut := ctx.StdOut()
			Expect(run.Run(ctx.Context, "cache", "get", "direct-key")).To(Succeed())
			out, err := readFileContent(stdOut)
			Expect(err).NotTo(HaveOccurred())
			Expect(out).To(ContainSubstring("direct-value"))
		})
	})

	When("listing cache entries (flow cache list)", func() {
		It("should list all cache entries", func() {
			stdOut := ctx.StdOut()
			Expect(run.Run(ctx.Context, "cache", "list")).To(Succeed())
			out, err := readFileContent(stdOut)
			Expect(err).NotTo(HaveOccurred())
			Expect(out).To(ContainSubstring("cache:"))
			Expect(out).To(ContainSubstring("direct-key"))
			Expect(out).To(ContainSubstring("multi-key"))
		})
	})

	When("removing a cache entry (flow cache remove)", func() {
		BeforeEach(func() {
			Expect(run.Run(ctx.Context, "cache", "set", "remove-test-key", "remove-test-value")).To(Succeed())
			utils.ResetTestContext(ctx, GinkgoTB())
		})

		It("should remove the cache entry successfully", func() {
			Expect(run.Run(ctx.Context, "cache", "remove", "remove-test-key")).To(Succeed())
			out, err := readFileContent(ctx.StdOut())
			Expect(err).NotTo(HaveOccurred())
			Expect(out).To(ContainSubstring("Key \"remove-test-key\" removed from the cache"))
		})

		It("should confirm the entry was removed", func() {
			// First remove the entry
			Expect(run.Run(ctx.Context, "cache", "remove", "remove-test-key")).To(Succeed())
			utils.ResetTestContext(ctx, GinkgoTB())
			// Then verify it's gone from the list
			stdOut := ctx.StdOut()
			Expect(run.Run(ctx.Context, "cache", "list")).To(Succeed())
			out, err := readFileContent(stdOut)
			Expect(err).NotTo(HaveOccurred())
			Expect(out).NotTo(ContainSubstring("remove-test-key"))
		})
	})

	When("clearing all cache entries (flow cache clear)", func() {
		It("should clear all cache entries", func() {
			stdOut := ctx.StdOut()
			Expect(run.Run(ctx.Context, "cache", "clear")).To(Succeed())
			out, err := readFileContent(stdOut)
			Expect(err).NotTo(HaveOccurred())
			Expect(out).To(ContainSubstring("Cache cleared"))
		})

		It("should confirm all entries were cleared", func() {
			// First clear the cache
			Expect(run.Run(ctx.Context, "cache", "clear")).To(Succeed())
			utils.ResetTestContext(ctx, GinkgoTB())
			// Then verify the list is empty
			stdOut := ctx.StdOut()
			Expect(run.Run(ctx.Context, "cache", "list")).To(Succeed())
			out, err := readFileContent(stdOut)
			Expect(err).NotTo(HaveOccurred())
			Expect(out).NotTo(ContainSubstring("direct-key"))
			Expect(out).NotTo(ContainSubstring("multi-key"))
		})
	})
})
