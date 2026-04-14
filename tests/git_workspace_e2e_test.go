//go:build e2e

package tests_test

import (
	stdCtx "context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/flowexec/flow/pkg/filesystem"
	"github.com/flowexec/flow/tests/utils"
)

var _ = Describe("git workspace e2e", Ordered, func() {
	var (
		ctx *utils.Context
		run *utils.CommandRunner

		bareRepoDir string // local bare repo path (for cleanup)
		bareRepoURL string // file:// URL for the bare repo
		wsName      string
	)

	BeforeAll(func() {
		ctx = utils.NewContext(stdCtx.Background(), GinkgoTB())
		run = utils.NewE2ECommandRunner()
		wsName = "git-test-ws"

		// Create a local bare git repo with a flow.yaml as a test fixture.
		bareRepoDir = initBareRepo(GinkgoTB())
		bareRepoURL = "file://" + bareRepoDir
	})

	BeforeEach(func() {
		utils.ResetTestContext(ctx, GinkgoTB())
	})

	AfterEach(func() {
		ctx.Finalize()
	})

	AfterAll(func() {
		Expect(os.RemoveAll(bareRepoDir)).To(Succeed())
	})

	When("adding a workspace from a local git URL (flow workspace add)", func() {
		It("clones and registers the workspace", func() {
			stdOut := ctx.StdOut()
			Expect(run.Run(
				ctx.Context, "workspace", "add", wsName, bareRepoURL,
			)).To(Succeed())
			out, err := readFileContent(stdOut)
			Expect(err).NotTo(HaveOccurred())
			Expect(out).To(ContainSubstring(fmt.Sprintf("Workspace '%s' created", wsName)))

			// Verify the workspace was cloned to the cache directory
			clonedPath := ctx.Config.Workspaces[wsName]
			Expect(clonedPath).NotTo(BeEmpty())
			Expect(filesystem.WorkspaceConfigExists(clonedPath)).To(BeTrue())
		})
	})

	When("getting the git workspace (flow workspace get)", func() {
		It("should return the workspace with git metadata", func() {
			stdOut := ctx.StdOut()
			Expect(run.Run(ctx.Context, "workspace", "get", wsName)).To(Succeed())
			out, err := readFileContent(stdOut)
			Expect(err).NotTo(HaveOccurred())
			Expect(out).To(ContainSubstring(wsName))
		})
	})

	When("updating the git workspace (flow workspace update)", func() {
		It("pulls latest changes successfully", func() {
			stdOut := ctx.StdOut()
			Expect(run.Run(ctx.Context, "workspace", "update", wsName)).To(Succeed())
			out, err := readFileContent(stdOut)
			Expect(err).NotTo(HaveOccurred())
			Expect(out).To(ContainSubstring(fmt.Sprintf("Workspace '%s' updated", wsName)))
		})

		It("force updates successfully", func() {
			stdOut := ctx.StdOut()
			Expect(run.Run(
				ctx.Context, "workspace", "update", wsName, "--force",
			)).To(Succeed())
			out, err := readFileContent(stdOut)
			Expect(err).NotTo(HaveOccurred())
			Expect(out).To(ContainSubstring(fmt.Sprintf("Workspace '%s' updated", wsName)))
		})
	})

	When("syncing with --git flag (flow sync --git)", func() {
		It("pulls all git workspaces and syncs cache", func() {
			stdOut := ctx.StdOut()
			Expect(run.Run(ctx.Context, "sync", "--git")).To(Succeed())
			out, err := readFileContent(stdOut)
			Expect(err).NotTo(HaveOccurred())
			Expect(out).To(ContainSubstring("Synced flow cache"))
		})
	})

	// Note: Negative test cases (e.g. updating a non-git workspace, conflicting flags)
	// use logger.Fatalf which calls tb.Fatalf in the test context. Since tb.Fatalf
	// invokes runtime.Goexit, these cannot be caught as errors by the CommandRunner.
})

// initBareRepo creates a local bare git repo with a flow.yaml file,
// suitable for use as a test "remote" without any network calls.
func initBareRepo(tb testing.TB) string {
	// Create a working directory to build the initial commit
	workDir, err := os.MkdirTemp("", "flow-git-work-*")
	Expect(err).NotTo(HaveOccurred())

	// git init
	gitCmd(tb, workDir, "init", "-b", "main")
	gitCmd(tb, workDir, "config", "user.email", "test@test.com")
	gitCmd(tb, workDir, "config", "user.name", "Test")

	// Create a flow.yaml
	flowYAML := `displayName: Git Test Workspace
description: A test workspace from git
tags:
  - test
  - git
`
	Expect(os.WriteFile(
		filepath.Join(workDir, "flow.yaml"), []byte(flowYAML), 0600,
	)).To(Succeed())

	gitCmd(tb, workDir, "add", ".")
	gitCmd(tb, workDir, "commit", "-m", "initial commit")

	// Create a bare clone to act as the "remote"
	bareDir, err := os.MkdirTemp("", "flow-git-bare-*")
	Expect(err).NotTo(HaveOccurred())
	// Remove the dir so git clone --bare can create it
	Expect(os.RemoveAll(bareDir)).To(Succeed())

	cmd := exec.Command("git", "clone", "--bare", workDir, bareDir)
	cmd.Stderr = os.Stderr
	Expect(cmd.Run()).To(Succeed())

	// Clean up the working directory
	Expect(os.RemoveAll(workDir)).To(Succeed())

	return bareDir
}

func gitCmd(tb testing.TB, dir string, args ...string) {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	cmd.Stderr = os.Stderr
	Expect(cmd.Run()).To(Succeed(), "git %v failed", args)
}
