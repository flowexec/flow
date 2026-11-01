package git_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/flowexec/flow/v2/internal/services/git"
	"github.com/flowexec/flow/v2/pkg/filesystem"
)

func TestGit(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Git Suite")
}

var _ = Describe("IsManagedClone", func() {
	It("is true only for paths under flow's git-workspaces directory", func() {
		GinkgoT().Setenv(filesystem.FlowCacheDirEnvVar, "/cache/flow")
		Expect(git.IsManagedClone("/cache/flow/git-workspaces/github.com/org/repo")).To(BeTrue())
		Expect(git.IsManagedClone("/cache/flow/git-workspaces")).To(BeFalse())
		Expect(git.IsManagedClone("/cache/flow/git-workspaces-other/repo")).To(BeFalse())
		Expect(git.IsManagedClone("/home/me/repo")).To(BeFalse())
	})
})

var _ = Describe("RepairWorkspaceConfig", func() {
	const upstream = "displayName: Upstream # kept\ntags: [a]\nnotInSchema: true\n"
	const rewritten = "displayName: Upstream\ngitRef: feat\ngitRefType: branch\ngitRemote: https://x/y.git\ntags:\n    - a\n"

	var repo string

	run := func(args ...string) string {
		cmd := exec.Command("git", args...)
		cmd.Dir = repo
		out, err := cmd.CombinedOutput()
		Expect(err).NotTo(HaveOccurred(), string(out))
		return string(out)
	}
	writeConfig := func(content string) {
		Expect(os.WriteFile(filepath.Join(repo, "flow.yaml"), []byte(content), 0600)).To(Succeed())
	}

	BeforeEach(func() {
		repo = GinkgoT().TempDir()
		run("init", "-b", "main")
		run("config", "user.email", "test@test.com")
		run("config", "user.name", "Test")
		writeConfig(upstream)
		run("add", "flow.yaml")
		run("commit", "-m", "initial")
	})

	It("restores a file that differs from HEAD only by flow's rewrite", func() {
		writeConfig(rewritten)
		Expect(git.RepairWorkspaceConfig(repo)).To(BeTrue())
		Expect(run("status", "--porcelain")).To(BeEmpty())
		Expect(os.ReadFile(filepath.Join(repo, "flow.yaml"))).To(BeEquivalentTo(upstream))
	})

	It("leaves a file with other edits alone", func() {
		edited := "displayName: My Own Name\ngitRemote: https://x/y.git\ntags:\n    - a\n"
		writeConfig(edited)
		Expect(git.RepairWorkspaceConfig(repo)).To(BeFalse())
		Expect(os.ReadFile(filepath.Join(repo, "flow.yaml"))).To(BeEquivalentTo(edited))
	})

	It("leaves a reformatted file without git fields alone", func() {
		reformatted := "displayName: Upstream\ntags:\n    - a\n"
		writeConfig(reformatted)
		Expect(git.RepairWorkspaceConfig(repo)).To(BeFalse())
		Expect(os.ReadFile(filepath.Join(repo, "flow.yaml"))).To(BeEquivalentTo(reformatted))
	})

	It("does nothing to a clean clone", func() {
		Expect(git.RepairWorkspaceConfig(repo)).To(BeFalse())
		Expect(run("status", "--porcelain")).To(BeEmpty())
	})

	It("excludes an untracked file flow generated, once", func() {
		run("rm", "-q", "flow.yaml")
		run("commit", "-m", "remove")
		writeConfig(rewritten)

		Expect(git.RepairWorkspaceConfig(repo)).To(BeFalse())
		Expect(git.ExcludeWorkspaceConfig(repo)).To(Succeed())
		Expect(run("status", "--porcelain")).To(BeEmpty())

		exclude, err := os.ReadFile(filepath.Join(repo, ".git", "info", "exclude"))
		Expect(err).NotTo(HaveOccurred())
		Expect(string(exclude)).To(HaveSuffix("\n/flow.yaml\n"))
		Expect(string(exclude)).NotTo(ContainSubstring("/flow.yaml\n\n/flow.yaml"))
	})
})
