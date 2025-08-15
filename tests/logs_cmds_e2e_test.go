//go:build e2e

package tests_test

import (
	stdCtx "context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/flowexec/flow/tests/utils"
)

var _ = Describe("logs e2e", Ordered, func() {
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

	When("viewing logs (flow logs)", func() {
		It("should display logs in yaml format", func() {
			stdOut := ctx.StdOut()
			Expect(run.Run(ctx.Context, "logs")).To(Succeed())
			out, err := readFileContent(stdOut)
			Expect(err).NotTo(HaveOccurred())
			Expect(out).To(ContainSubstring("logs:"))
		})
	})

	When("viewing last log entry (flow logs --last)", func() {
		It("should display the last log entry", func() {
			// TODO: test that log archiving works
			Skip("e2e test does not include log archiving, so this will not return any logs")
			stdOut := ctx.StdOut()
			Expect(run.Run(ctx.Context, "logs", "--last")).To(Succeed())
			out, err := readFileContent(stdOut)
			Expect(err).NotTo(HaveOccurred())
			Expect(out).To(Or(ContainSubstring("msg="), ContainSubstring("No log entries")))
		})
	})
})
