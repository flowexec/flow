//go:build e2e

package tests_test

import (
	stdCtx "context"
	"fmt"
	"os"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/flowexec/flow/tests/utils"
)

var _ = Describe("workspace e2e", Ordered, func() {
	var (
		ctx *utils.Context
		run *utils.CommandRunner

		wsName, wsPath, origWsName string
	)

	BeforeAll(func() {
		ctx = utils.NewContext(stdCtx.Background(), GinkgoTB())
		run = utils.NewE2ECommandRunner()
		tmp, err := os.MkdirTemp("", "flow-test-*")
		Expect(err).NotTo(HaveOccurred())
		origWsName = ctx.Config.CurrentWorkspace
		wsName = "test-workspace"
		wsPath = tmp
	})

	BeforeEach(func() {
		utils.ResetTestContext(ctx, GinkgoTB())
	})

	AfterEach(func() {
		ctx.Finalize()
	})

	AfterAll(func() {
		Expect(os.RemoveAll(wsPath)).To(Succeed())
	})

	When("creating a new workspace (flow workspace add)", func() {
		It("creates successfully", func() {
			stdOut := ctx.StdOut()
			Expect(run.Run(ctx.Context, "workspace", "add", wsName, wsPath)).To(Succeed())
			out, err := readFileContent(stdOut)
			Expect(err).NotTo(HaveOccurred())
			Expect(out).To(ContainSubstring(fmt.Sprintf("Workspace '%s' created", wsName)))
		})
	})

	When("setting a workspace (flow workspace switch)", func() {
		It("sets successfully", func() {
			Expect(run.Run(ctx.Context, "workspace", "switch", wsName)).To(Succeed())
			out, err := readFileContent(ctx.StdOut())
			Expect(err).NotTo(HaveOccurred())
			Expect(out).To(ContainSubstring(fmt.Sprintf("Workspace set to %s", wsName)))
		})
	})

	When("getting a workspace (flow workspace get)", func() {
		It("should returns the workspace", func() {
			stdOut := ctx.StdOut()
			Expect(run.Run(ctx.Context, "workspace", "get", wsName)).To(Succeed())
			out, err := readFileContent(stdOut)
			Expect(err).NotTo(HaveOccurred())
			Expect(out).To(ContainSubstring(wsName))
		})
	})

	When("listing workspaces (flow workspace list)", func() {
		It("should return the list of workspaces", func() {
			stdOut := ctx.StdOut()
			Expect(run.Run(ctx.Context, "workspace", "list")).To(Succeed())
			out, err := readFileContent(stdOut)
			Expect(err).NotTo(HaveOccurred())
			Expect(out).To(ContainSubstring(wsName))
		})
	})

	When("deleting a workspace (flow workspace remove)", func() {
		It("should remove the workspace from the user config", func() {
			reader, writer, err := os.Pipe()
			Expect(err).NotTo(HaveOccurred())
			_, err = writer.Write([]byte("yes\n"))
			Expect(err).ToNot(HaveOccurred())

			ctx.SetIO(reader, ctx.StdOut())
			Expect(run.Run(ctx.Context, "workspace", "remove", origWsName)).To(Succeed())
			out, err := readFileContent(ctx.StdOut())
			Expect(err).NotTo(HaveOccurred())
			Expect(out).To(ContainSubstring(fmt.Sprintf("Workspace '%s' deleted", origWsName)))
		})
	})
})
