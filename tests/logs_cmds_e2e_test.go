//go:build e2e

package tests_test

import (
	stdCtx "context"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/flowexec/flow/pkg/store"
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

	When("viewing logs with no history (flow logs)", func() {
		It("should display empty history message", func() {
			stdOut := ctx.StdOut()
			Expect(run.Run(ctx.Context, "logs", "-o", "yaml")).To(Succeed())
			out, err := readFileContent(stdOut)
			Expect(err).NotTo(HaveOccurred())
			Expect(out).To(ContainSubstring("No execution history found."))
		})
	})

	When("viewing logs with history (flow logs)", func() {
		It("should display history in yaml format", func() {
			record := store.ExecutionRecord{
				Ref:       "default/examples:simple-print",
				StartedAt: time.Now(),
				Duration:  100 * time.Millisecond,
			}
			Expect(ctx.DataStore.RecordExecution(record)).To(Succeed())

			stdOut := ctx.StdOut()
			Expect(run.Run(ctx.Context, "logs", "-o", "yaml")).To(Succeed())
			out, err := readFileContent(stdOut)
			Expect(err).NotTo(HaveOccurred())
			Expect(out).To(ContainSubstring("history:"))
			Expect(out).To(ContainSubstring("default/examples:simple-print"))
		})
	})

	When("viewing last log entry (flow logs --last)", func() {
		It("should display the last execution record", func() {
			record := store.ExecutionRecord{
				Ref:       "default/examples:simple-print",
				StartedAt: time.Now(),
				Duration:  200 * time.Millisecond,
			}
			Expect(ctx.DataStore.RecordExecution(record)).To(Succeed())

			stdOut := ctx.StdOut()
			Expect(run.Run(ctx.Context, "logs", "--last")).To(Succeed())
			out, err := readFileContent(stdOut)
			Expect(err).NotTo(HaveOccurred())
			Expect(out).To(ContainSubstring("ref: default/examples:simple-print"))
		})
	})

	When("clearing logs (flow logs clear)", func() {
		It("should clear history for a specific ref", func() {
			record := store.ExecutionRecord{
				Ref:       "default/examples:simple-print",
				StartedAt: time.Now(),
				Duration:  50 * time.Millisecond,
			}
			Expect(ctx.DataStore.RecordExecution(record)).To(Succeed())

			stdOut := ctx.StdOut()
			Expect(run.Run(ctx.Context, "logs", "clear", "default/examples:simple-print")).To(Succeed())
			out, err := readFileContent(stdOut)
			Expect(err).NotTo(HaveOccurred())
			Expect(out).To(ContainSubstring("Cleared history and logs"))
		})
	})
})
