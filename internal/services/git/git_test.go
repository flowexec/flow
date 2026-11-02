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

	It("sees through a symlinked cache directory", func() {
		base := GinkgoT().TempDir()
		clone := filepath.Join(base, "real", "git-workspaces", "host", "repo")
		Expect(os.MkdirAll(clone, 0750)).To(Succeed())
		if err := os.Symlink(filepath.Join(base, "real"), filepath.Join(base, "link")); err != nil {
			Skip("symlinks are unavailable: " + err.Error())
		}
		GinkgoT().Setenv(filesystem.FlowCacheDirEnvVar, filepath.Join(base, "link"))
		Expect(git.IsManagedClone(clone)).To(BeTrue())
	})
})

var _ = Describe("DetectSource", func() {
	var repo string

	run := func(args ...string) {
		cmd := exec.Command("git", args...)
		cmd.Dir = repo
		out, err := cmd.CombinedOutput()
		Expect(err).NotTo(HaveOccurred(), string(out))
	}

	BeforeEach(func() {
		repo = GinkgoT().TempDir()
		run("init", "-b", "feature")
		run("config", "user.email", "test@test.com")
		run("config", "user.name", "Test")
		run("remote", "add", "origin", "https://example.com/org/repo.git")
		run("commit", "--allow-empty", "-m", "initial")
	})

	expectSource := func(wantRef, wantRefType string) {
		remote, ref, refType, err := git.DetectSource(repo)
		Expect(err).NotTo(HaveOccurred())
		Expect(remote).To(Equal("https://example.com/org/repo.git"))
		Expect(ref).To(Equal(wantRef))
		Expect(refType).To(Equal(wantRefType))
	}

	It("reports the checked out branch", func() {
		expectSource("feature", "branch")
	})

	It("reports the tag at a detached HEAD", func() {
		run("tag", "v1.0.0")
		run("checkout", "--detach", "v1.0.0")
		expectSource("v1.0.0", "tag")
	})

	It("reports no ref when several tags point at a detached HEAD", func() {
		run("tag", "a-other")
		run("tag", "v1.0.0")
		run("checkout", "--detach", "v1.0.0")
		expectSource("", "")
	})

	It("reports no ref at a detached HEAD without a tag", func() {
		run("checkout", "--detach")
		expectSource("", "")
	})

	It("recognizes its clone by the configured origin", func() {
		Expect(git.IsCloneOf(repo, "https://example.com/org/repo.git")).To(BeTrue())
		Expect(git.IsCloneOf(repo, "https://example.com/org/other.git")).To(BeFalse())
	})

	It("fails without an origin remote", func() {
		run("remote", "remove", "origin")
		_, _, _, err := git.DetectSource(repo)
		Expect(err).To(HaveOccurred())
	})
})

var _ = Describe("IsRepoRoot", func() {
	It("is true only at the top level of a repository", func() {
		repo := GinkgoT().TempDir()
		cmd := exec.Command("git", "init")
		cmd.Dir = repo
		Expect(cmd.Run()).To(Succeed())
		sub := filepath.Join(repo, "cache", "clone")
		Expect(os.MkdirAll(sub, 0750)).To(Succeed())

		Expect(git.IsRepoRoot(repo)).To(BeTrue())
		Expect(git.IsRepoRoot(sub)).To(BeFalse())
		Expect(git.IsRepoRoot(filepath.Join(repo, "missing"))).To(BeFalse())
		Expect(git.IsRepoRoot(GinkgoT().TempDir())).To(BeFalse())
		Expect(git.ResetPull(sub, "", "")).To(MatchError(ContainSubstring("refusing to reset")))
	})
})

var _ = Describe("RepairWorkspaceConfig", func() {
	const upstream = "displayName: Upstream # kept\ntags: [a]\nnotInSchema: true\n"
	const rewritten = "displayName: Upstream\ngitRef: feat\ngitRefType: branch\n" +
		"gitRemote: https://x/y.git\ntags:\n    - a\n"

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
		// The specs compare bytes, which a Windows checkout would otherwise convert to CRLF.
		run("config", "core.autocrlf", "false")
		writeConfig(upstream)
		run("add", "flow.yaml")
		run("commit", "-m", "initial")
	})

	repair := func() (bool, error) { return git.RepairWorkspaceConfig(repo, "ws") }

	It("restores a file that is exactly flow's rewrite of HEAD", func() {
		writeConfig(rewritten)
		Expect(repair()).To(BeTrue())
		Expect(run("status", "--porcelain")).To(BeEmpty())
		Expect(os.ReadFile(filepath.Join(repo, "flow.yaml"))).To(BeEquivalentTo(upstream))
	})

	DescribeTable("leaves other changes alone",
		func(content string) {
			writeConfig(content)
			Expect(repair()).To(BeFalse())
			Expect(os.ReadFile(filepath.Join(repo, "flow.yaml"))).To(BeEquivalentTo(content))
		},
		Entry("an edited field", "displayName: My Own Name\ngitRemote: https://x/y.git\ntags:\n    - a\n"),
		Entry("a comment added to the rewrite", "# my note\n"+rewritten),
		Entry("an unknown key added to the rewrite", rewritten+"futureSetting: 42\n"),
		Entry("a reformat without git fields", "displayName: Upstream\ntags:\n    - a\n"),
	)

	It("does nothing to a clean clone", func() {
		Expect(repair()).To(BeFalse())
		Expect(run("status", "--porcelain")).To(BeEmpty())
	})

	Context("when flow.yaml is untracked", func() {
		BeforeEach(func() {
			run("rm", "-q", "flow.yaml")
			run("commit", "-m", "remove")
		})

		DescribeTable("removes a file flow generated",
			func(content string) {
				writeConfig(content)
				Expect(repair()).To(BeFalse())
				Expect(filepath.Join(repo, "flow.yaml")).NotTo(BeAnExistingFile())
			},
			Entry("the current form", "displayName: ws\n"),
			Entry("the form with git fields", "displayName: ws\ngitRemote: https://x/y.git\n"),
		)

		It("keeps a generated file the user edited", func() {
			edited := "displayName: ws\nenvFiles:\n    - .env.local\n"
			writeConfig(edited)
			Expect(repair()).To(BeFalse())
			Expect(os.ReadFile(filepath.Join(repo, "flow.yaml"))).To(BeEquivalentTo(edited))
		})
	})
})
