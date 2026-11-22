//go:build e2e

package tests_test

import (
	stdCtx "context"
	"encoding/json"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/flowexec/flow/v2/internal/digest"
	"github.com/flowexec/flow/v2/pkg/store"
	"github.com/flowexec/flow/v2/tests/utils"
)

var _ = Describe("digest e2e", Ordered, func() {
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

	It("should compose run, runs and exec sections as JSON", func() {
		ref := "exec " + ctx.CurrentWorkspace.AssignedName() + "/examples:simple-print"
		Expect(ctx.DataStore.RecordExecution(store.ExecutionRecord{
			ID: "digest-run-1", Ref: ref, StartedAt: time.Now(), Duration: time.Second,
			Status: store.RunFailed, ExitCode: 1, Error: "exit status 1",
		})).To(Succeed())

		stdOut := ctx.StdOut()
		Expect(run.Run(ctx.Context, "digest", "-o", "json",
			"run:digest-run-1", "runs:simple-print status:failed", "exec:exec/examples:simple-print")).To(Succeed())
		out, err := readFileContent(stdOut)
		Expect(err).NotTo(HaveOccurred())

		var doc digest.Document
		Expect(json.Unmarshal([]byte(out), &doc)).To(Succeed())
		Expect(doc.Sections).To(HaveLen(3))
		Expect(doc.Sections[0].Type).To(Equal("run"))
		Expect(doc.Sections[0].Metadata).To(HaveKeyWithValue("status", "failed"))
		Expect(doc.Sections[1].Metadata).To(HaveKeyWithValue("count", BeNumerically("==", 1)))
		Expect(doc.Sections[2].Content).To(ContainSubstring("name: simple-print"))
		// The run was recorded under the alias verb typed (exec); the section still finds it.
		Expect(doc.Sections[2].Content).To(ContainSubstring("| digest-run-1 | failed |"))
		Expect(doc.MaxChars).To(Equal(digest.DefaultMaxChars))
	})

	It("should render text sections", func() {
		stdOut := ctx.StdOut()
		Expect(run.Run(ctx.Context, "digest", "exec:exec examples:simple-print", "--no-history")).To(Succeed())
		out, err := readFileContent(stdOut)
		Expect(err).NotTo(HaveOccurred())
		Expect(out).To(ContainSubstring("## Executable: run "))
		Expect(out).To(ContainSubstring("### Definition"))
		Expect(out).NotTo(ContainSubstring("Recent runs"))
	})

	It("should reject unknown source types", func() {
		ctx.ExpectFailure()
		Expect(run.Run(ctx.Context, "digest", "bogus:x")).NotTo(Succeed())
	})
})
