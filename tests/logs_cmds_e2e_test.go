//go:build e2e

package tests_test

import (
	stdCtx "context"
	"os"
	"path/filepath"
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
			Expect(out).To(ContainSubstring("history: []"))
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

	When("listing running background processes (flow logs --running)", func() {
		It("should display empty runs in yaml format", func() {
			stdOut := ctx.StdOut()
			Expect(run.Run(ctx.Context, "logs", "--running", "-o", "yaml")).To(Succeed())
			out, err := readFileContent(stdOut)
			Expect(err).NotTo(HaveOccurred())
			Expect(out).To(ContainSubstring("runs: []"))
		})

		It("should display active background runs in yaml format", func() {
			bgRun := store.BackgroundRun{
				ID:        "test1234",
				PID:       os.Getpid(), // use our own PID so isProcessAlive returns true
				Ref:       "run default/examples:simple-print",
				StartedAt: time.Now(),
				Status:    store.BackgroundRunning,
			}
			Expect(ctx.DataStore.SaveBackgroundRun(bgRun)).To(Succeed())

			stdOut := ctx.StdOut()
			Expect(run.Run(ctx.Context, "logs", "--running", "-o", "yaml")).To(Succeed())
			out, err := readFileContent(stdOut)
			Expect(err).NotTo(HaveOccurred())
			Expect(out).To(ContainSubstring("id: test1234"))
			Expect(out).To(ContainSubstring("status: running"))
			Expect(out).To(ContainSubstring("default/examples:simple-print"))
		})

		It("should prune stale background runs", func() {
			// Clean up any leftover runs from prior tests in this Ordered suite.
			_ = ctx.DataStore.DeleteBackgroundRun("test1234")

			bgRun := store.BackgroundRun{
				ID:        "stale123",
				PID:       999999999, // non-existent PID
				Ref:       "run default/examples:simple-print",
				StartedAt: time.Now(),
				Status:    store.BackgroundRunning,
			}
			Expect(ctx.DataStore.SaveBackgroundRun(bgRun)).To(Succeed())

			stdOut := ctx.StdOut()
			Expect(run.Run(ctx.Context, "logs", "--running", "-o", "yaml")).To(Succeed())
			out, err := readFileContent(stdOut)
			Expect(err).NotTo(HaveOccurred())
			Expect(out).To(ContainSubstring("runs: []"))

			// Verify the stale run was updated to failed.
			updated, err := ctx.DataStore.GetBackgroundRun("stale123")
			Expect(err).NotTo(HaveOccurred())
			Expect(updated.Status).To(Equal(store.BackgroundFailed))
			Expect(updated.Error).To(ContainSubstring("exited unexpectedly"))
		})
	})

	When("attaching to a background process (flow logs attach)", func() {
		It("should display log content from the archive file", func() {
			// Create a temporary log archive file with known content.
			logDir := GinkgoT().TempDir()
			logFile := filepath.Join(logDir, "test-archive.log")
			Expect(os.WriteFile(logFile, []byte("line 1\nline 2\nline 3\n"), 0600)).To(Succeed())

			bgRun := store.BackgroundRun{
				ID:           "attach12",
				PID:          1,
				Ref:          "run default/examples:simple-print",
				StartedAt:    time.Now(),
				Status:       store.BackgroundCompleted,
				LogArchiveID: logFile,
			}
			Expect(ctx.DataStore.SaveBackgroundRun(bgRun)).To(Succeed())

			stdOut := ctx.StdOut()
			Expect(run.Run(ctx.Context, "logs", "attach", "attach12")).To(Succeed())
			out, err := readFileContent(stdOut)
			Expect(err).NotTo(HaveOccurred())
			Expect(out).To(ContainSubstring("line 1"))
			Expect(out).To(ContainSubstring("line 2"))
			Expect(out).To(ContainSubstring("line 3"))
		})
	})
})
