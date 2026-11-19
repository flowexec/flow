//go:build e2e

package tests_test

import (
	stdCtx "context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	worktreeIO "github.com/flowexec/flow/v2/internal/io/worktree"
	"github.com/flowexec/flow/v2/tests/utils"
)

var _ = Describe("worktree commands e2e", Serial, Ordered, func() {
	var (
		ctx *utils.Context
		run *utils.CommandRunner
	)

	BeforeAll(func() {
		ctx = utils.NewContext(stdCtx.Background(), GinkgoTB())
		run = utils.NewE2ECommandRunner()

		// Turn the workspace directory into a git repository so worktree
		// commands can resolve the repo root from cwd.
		wsDir := ctx.WorkspaceDir()
		gitCmd(GinkgoTB(), wsDir, "init", "-b", "main")
		gitCmd(GinkgoTB(), wsDir, "config", "user.email", "test@test.com")
		gitCmd(GinkgoTB(), wsDir, "config", "user.name", "Test")
		gitCmd(GinkgoTB(), wsDir, "add", ".")
		gitCmd(GinkgoTB(), wsDir, "commit", "-m", "initial")

		GinkgoTB().Chdir(wsDir)
	})

	BeforeEach(func() {
		utils.ResetTestContext(ctx, GinkgoTB())
	})

	AfterEach(func() {
		ctx.Finalize()
	})

	Describe("flow worktree add", func() {
		It("creates a worktree and reports its path", func() {
			stdOut := ctx.StdOut()
			Expect(run.Run(
				ctx.Context, "worktree", "add", "test-branch", "--output", "json",
			)).To(Succeed())
			out, err := readFileContent(stdOut)
			Expect(err).NotTo(HaveOccurred())

			var env map[string]any
			Expect(json.Unmarshal([]byte(out), &env)).To(Succeed())
			result, ok := env["result"].(map[string]any)
			Expect(ok).To(BeTrue(), "expected result envelope, got: %s", out)
			data, ok := result["data"].(map[string]any)
			Expect(ok).To(BeTrue())
			Expect(data["branch"]).To(Equal("test-branch"))
			Expect(data["path"]).NotTo(BeEmpty())

			// Verify the worktree directory was actually created.
			wtPath, ok := data["path"].(string)
			Expect(ok).To(BeTrue())
			Expect(wtPath).To(BeADirectory())
		})

		It("creates a worktree from a base ref", func() {
			stdOut := ctx.StdOut()
			Expect(run.Run(
				ctx.Context, "worktree", "add", "based-branch", "--base", "main", "--output", "json",
			)).To(Succeed())
			out, err := readFileContent(stdOut)
			Expect(err).NotTo(HaveOccurred())
			Expect(out).To(ContainSubstring("based-branch"))
		})

		It("uses an existing branch with --no-create-branch", func() {
			wsDir := ctx.WorkspaceDir()
			gitCmd(GinkgoTB(), wsDir, "branch", "pre-existing")

			stdOut := ctx.StdOut()
			Expect(run.Run(
				ctx.Context, "worktree", "add", "pre-existing", "--no-create-branch", "--output", "json",
			)).To(Succeed())
			out, err := readFileContent(stdOut)
			Expect(err).NotTo(HaveOccurred())
			Expect(out).To(ContainSubstring("pre-existing"))
		})
	})

	Describe("flow worktree list", func() {
		It("returns JSON with the expected fields", func() {
			stdOut := ctx.StdOut()
			Expect(run.Run(
				ctx.Context, "worktree", "list", "--output", "json",
			)).To(Succeed())
			out, err := readFileContent(stdOut)
			Expect(err).NotTo(HaveOccurred())

			var entries []worktreeIO.DisplayInfo
			Expect(json.Unmarshal([]byte(out), &entries)).To(Succeed())
			Expect(entries).NotTo(BeEmpty())

			// The main worktree should be present and marked.
			var mainEntry *worktreeIO.DisplayInfo
			for i, e := range entries {
				if e.IsMain {
					mainEntry = &entries[i]
					break
				}
			}
			Expect(mainEntry).NotTo(BeNil(), "expected a main worktree in the list")
			Expect(mainEntry.Branch).To(Equal("main"))
			Expect(mainEntry.IsCurrent).To(BeTrue())
		})

		It("lists worktrees created by add", func() {
			stdOut := ctx.StdOut()
			Expect(run.Run(
				ctx.Context, "worktree", "list", "--output", "json",
			)).To(Succeed())
			out, err := readFileContent(stdOut)
			Expect(err).NotTo(HaveOccurred())

			var entries []worktreeIO.DisplayInfo
			Expect(json.Unmarshal([]byte(out), &entries)).To(Succeed())

			// Should include the worktrees created in earlier tests.
			branches := make(map[string]bool)
			for _, e := range entries {
				branches[e.Branch] = true
			}
			Expect(branches).To(HaveKey("test-branch"))
		})
	})

	Describe("flow worktree remove", func() {
		It("removes a worktree by branch name", func() {
			stdOut := ctx.StdOut()
			Expect(run.Run(
				ctx.Context, "worktree", "remove", "test-branch", "--yes", "--output", "json",
			)).To(Succeed())
			out, err := readFileContent(stdOut)
			Expect(err).NotTo(HaveOccurred())

			var env map[string]any
			Expect(json.Unmarshal([]byte(out), &env)).To(Succeed())
			result, ok := env["result"].(map[string]any)
			Expect(ok).To(BeTrue())
			Expect(result["message"]).To(ContainSubstring("removed"))
		})

		It("refuses to remove the main worktree", func() {
			ctx.ExpectFailure()
			err := run.Run(ctx.Context, "worktree", "remove", "main", "--yes")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("main"))
		})

		It("fails to remove a dirty worktree without --force", func() {
			// Create a fresh worktree and dirty it.
			Expect(run.Run(
				ctx.Context, "worktree", "add", "dirty-branch", "--output", "json",
			)).To(Succeed())

			// Find its path from the list.
			utils.ResetTestContext(ctx, GinkgoTB())
			stdOut := ctx.StdOut()
			Expect(run.Run(
				ctx.Context, "worktree", "list", "--output", "json",
			)).To(Succeed())
			out, err := readFileContent(stdOut)
			Expect(err).NotTo(HaveOccurred())

			var entries []worktreeIO.DisplayInfo
			Expect(json.Unmarshal([]byte(out), &entries)).To(Succeed())
			var dirtyPath string
			for _, e := range entries {
				if e.Branch == "dirty-branch" {
					dirtyPath = e.Path
					break
				}
			}
			Expect(dirtyPath).NotTo(BeEmpty())

			// Dirty the worktree.
			Expect(os.WriteFile(
				filepath.Join(dirtyPath, "dirty.txt"), []byte("dirty\n"), 0o644,
			)).To(Succeed())

			// Non-force removal should fail.
			utils.ResetTestContext(ctx, GinkgoTB())
			ctx.ExpectFailure()
			err = run.Run(ctx.Context, "worktree", "remove", "dirty-branch", "--yes")
			Expect(err).To(HaveOccurred())
		})

		It("removes a dirty worktree with --force", func() {
			stdOut := ctx.StdOut()
			Expect(run.Run(
				ctx.Context, "worktree", "remove", "dirty-branch", "--yes", "--force", "--output", "json",
			)).To(Succeed())
			out, err := readFileContent(stdOut)
			Expect(err).NotTo(HaveOccurred())
			Expect(out).To(ContainSubstring("removed"))
		})
	})

	Describe("flow worktree prune", func() {
		It("reports nothing to prune on a clean repo", func() {
			stdOut := ctx.StdOut()
			Expect(run.Run(
				ctx.Context, "worktree", "prune", "--dry-run", "--output", "json",
			)).To(Succeed())
			out, err := readFileContent(stdOut)
			Expect(err).NotTo(HaveOccurred())
			Expect(out).To(ContainSubstring("Nothing to prune"))
		})

		It("detects and cleans stale worktree entries", func() {
			// Create a worktree.
			Expect(run.Run(
				ctx.Context, "worktree", "add", "stale-branch", "--output", "json",
			)).To(Succeed())

			// Reset so stdout is fresh for the list command.
			utils.ResetTestContext(ctx, GinkgoTB())
			stdOut := ctx.StdOut()
			Expect(run.Run(
				ctx.Context, "worktree", "list", "--output", "json",
			)).To(Succeed())
			out, err := readFileContent(stdOut)
			Expect(err).NotTo(HaveOccurred())
			var entries []worktreeIO.DisplayInfo
			Expect(json.Unmarshal([]byte(out), &entries)).To(Succeed())
			var stalePath string
			for _, e := range entries {
				if e.Branch == "stale-branch" {
					stalePath = e.Path
					break
				}
			}
			Expect(stalePath).NotTo(BeEmpty())

			// Manually delete the directory to simulate staleness.
			Expect(os.RemoveAll(stalePath)).To(Succeed())

			// Prune should clean up the stale entry.
			utils.ResetTestContext(ctx, GinkgoTB())
			stdOut = ctx.StdOut()
			Expect(run.Run(
				ctx.Context, "worktree", "prune", "--yes", "--output", "json",
			)).To(Succeed())
			out, err = readFileContent(stdOut)
			Expect(err).NotTo(HaveOccurred())
			Expect(out).To(ContainSubstring("Pruned"))
		})

		It("removes merged worktrees with --merged", func() {
			wsDir := ctx.WorkspaceDir()

			// Create a branch, merge it, then leave the worktree around.
			Expect(run.Run(
				ctx.Context, "worktree", "add", "merge-me", "--output", "json",
			)).To(Succeed())

			// Reset so stdout is fresh for the list command.
			utils.ResetTestContext(ctx, GinkgoTB())
			stdOut := ctx.StdOut()
			Expect(run.Run(
				ctx.Context, "worktree", "list", "--output", "json",
			)).To(Succeed())
			out, err := readFileContent(stdOut)
			Expect(err).NotTo(HaveOccurred())
			var entries []worktreeIO.DisplayInfo
			Expect(json.Unmarshal([]byte(out), &entries)).To(Succeed())
			var mergePath string
			for _, e := range entries {
				if e.Branch == "merge-me" {
					mergePath = e.Path
					break
				}
			}
			Expect(mergePath).NotTo(BeEmpty())

			// Make a commit on the branch so the merge is non-trivial.
			Expect(os.WriteFile(
				filepath.Join(mergePath, "merged.txt"), []byte("merged\n"), 0o644,
			)).To(Succeed())
			gitCmd(GinkgoTB(), mergePath, "add", ".")
			gitCmd(GinkgoTB(), mergePath, "commit", "-m", "on merge-me")

			// Merge it into main.
			gitCmd(GinkgoTB(), wsDir, "merge", "--no-ff", "merge-me", "-m", "merge")

			// Prune with --merged should remove it.
			utils.ResetTestContext(ctx, GinkgoTB())
			stdOut = ctx.StdOut()
			Expect(run.Run(
				ctx.Context, "worktree", "prune", "--merged", "--yes", "--output", "json",
			)).To(Succeed())
			out, err = readFileContent(stdOut)
			Expect(err).NotTo(HaveOccurred())
			Expect(out).To(ContainSubstring("Pruned"))

			// Verify the worktree is gone.
			cmd := exec.Command("git", "worktree", "list", "--porcelain")
			cmd.Dir = wsDir
			wtOut, err := cmd.Output()
			Expect(err).NotTo(HaveOccurred())
			Expect(string(wtOut)).NotTo(ContainSubstring("merge-me"))
		})
	})

	Describe("flow worktree add with --dir", func() {
		It("creates worktree in a custom directory", func() {
			customDir := filepath.Join(GinkgoTB().TempDir(), "custom-wt")
			stdOut := ctx.StdOut()
			Expect(run.Run(
				ctx.Context, "worktree", "add", "custom-dir-branch",
				"--dir", customDir, "--output", "json",
			)).To(Succeed())
			out, err := readFileContent(stdOut)
			Expect(err).NotTo(HaveOccurred())

			var env map[string]any
			Expect(json.Unmarshal([]byte(out), &env)).To(Succeed())
			result, ok := env["result"].(map[string]any)
			Expect(ok).To(BeTrue())
			data, ok := result["data"].(map[string]any)
			Expect(ok).To(BeTrue())
			wtPath, ok := data["path"].(string)
			Expect(ok).To(BeTrue())
			Expect(wtPath).To(HavePrefix(customDir))
			Expect(wtPath).To(BeADirectory())

			// Clean up.
			utils.ResetTestContext(ctx, GinkgoTB())
			Expect(run.Run(
				ctx.Context, "worktree", "remove", "custom-dir-branch", "--yes",
			)).To(Succeed())
		})
	})

	Describe("gitignore management", func() {
		It("adds the worktree directory to .gitignore", func() {
			wsDir := ctx.WorkspaceDir()
			gitignorePath := filepath.Join(wsDir, ".gitignore")

			content, err := os.ReadFile(gitignorePath)
			Expect(err).NotTo(HaveOccurred())
			Expect(string(content)).To(ContainSubstring(".worktrees/"))
		})
	})
})
