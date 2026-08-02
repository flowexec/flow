package filesystem_test

import (
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/flowexec/flow/v2/pkg/filesystem"
	"github.com/flowexec/flow/v2/types/config"
)

var _ = Describe("ResolveWorkspace", func() {
	var tmpDir string

	mkWorkspace := func(rel string) string {
		path := filepath.Join(tmpDir, rel)
		Expect(os.MkdirAll(path, 0750)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(path, filesystem.WorkspaceConfigFileName), []byte("{}\n"), 0600)).To(Succeed())
		return path
	}
	mkDir := func(rel string) string {
		path := filepath.Join(tmpDir, rel)
		Expect(os.MkdirAll(path, 0750)).To(Succeed())
		return path
	}
	dynamic := func(workspaces map[string]string, current string) *config.Config {
		return &config.Config{
			Workspaces:       workspaces,
			CurrentWorkspace: current,
			WorkspaceMode:    config.ConfigWorkspaceModeDynamic,
		}
	}

	BeforeEach(func() {
		var err error
		tmpDir, err = os.MkdirTemp("", "flow-resolve-test")
		Expect(err).NotTo(HaveOccurred())
		tmpDir, err = filepath.EvalSymlinks(tmpDir)
		Expect(err).NotTo(HaveOccurred())
		// Resolution reports normalized paths; normalize the fixture root so the two forms of a
		// macOS temp dir (/private/var/... and /var/...) don't make every comparison fail.
		tmpDir = filesystem.NormalizePath(tmpDir)
		Expect(os.Unsetenv(filesystem.WorkspaceOverrideEnv)).To(Succeed())
	})

	AfterEach(func() {
		Expect(os.RemoveAll(tmpDir)).To(Succeed())
		Expect(os.Unsetenv(filesystem.WorkspaceOverrideEnv)).To(Succeed())
	})

	Context("discovery", func() {
		It("resolves an unregistered root by walking up, named for its directory", func() {
			root := mkWorkspace("cloned-repo")
			res, err := filesystem.ResolveWorkspace(dynamic(nil, ""), filesystem.ResolveOptions{
				Dir: mkDir("cloned-repo/internal/pkg"),
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(res).NotTo(BeNil())
			Expect(res.Name).To(Equal("cloned-repo"))
			Expect(res.Path).To(Equal(root))
			Expect(res.Registered).To(BeFalse())
			Expect(res.Source).To(Equal(filesystem.SourceDiscovered))
		})

		It("uses the registered name when the discovered root is registered", func() {
			root := mkWorkspace("checkout")
			cfg := dynamic(map[string]string{"myproject": root}, "myproject")
			res, err := filesystem.ResolveWorkspace(cfg, filesystem.ResolveOptions{Dir: mkDir("checkout/sub")})
			Expect(err).NotTo(HaveOccurred())
			Expect(res.Name).To(Equal("myproject"), "registered name wins over the directory basename")
			Expect(res.Registered).To(BeTrue())
			Expect(res.Source).To(Equal(filesystem.SourceRegistered))
		})

		It("prefers a nested worktree over its registered parent workspace", func() {
			parent := mkWorkspace("repo")
			worktree := mkWorkspace("repo/.worktrees/feature")
			cfg := dynamic(map[string]string{"repo": parent}, "repo")
			res, err := filesystem.ResolveWorkspace(cfg, filesystem.ResolveOptions{
				Dir: mkDir("repo/.worktrees/feature/cmd"),
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(res.Name).To(Equal("feature"))
			Expect(res.Path).To(Equal(worktree))
			Expect(res.Registered).To(BeFalse())
		})

		It("resolves a worktree that lives outside every registered workspace", func() {
			registered := mkWorkspace("repo")
			worktree := mkWorkspace("elsewhere/repo-feature")
			cfg := dynamic(map[string]string{"repo": registered}, "repo")
			res, err := filesystem.ResolveWorkspace(cfg, filesystem.ResolveOptions{Dir: worktree})
			Expect(err).NotTo(HaveOccurred())
			Expect(res.Name).To(Equal("repo-feature"))
			Expect(res.Registered).To(BeFalse())
		})

		It("walks past a vendored copy inside a registered workspace", func() {
			parent := mkWorkspace("repo")
			mkWorkspace("repo/vendor/dep")
			cfg := dynamic(map[string]string{"repo": parent}, "repo")
			res, err := filesystem.ResolveWorkspace(cfg, filesystem.ResolveOptions{Dir: mkDir("repo/vendor/dep")})
			Expect(err).NotTo(HaveOccurred())
			Expect(res.Name).To(Equal("repo"))
			Expect(res.Path).To(Equal(parent))
		})

		It("still discovers a standalone clone under a directory named external", func() {
			// The boundary only applies inside a registered workspace; an unrelated path that
			// happens to contain "external" is a legitimate root.
			root := mkWorkspace("external/standalone")
			res, err := filesystem.ResolveWorkspace(dynamic(nil, ""), filesystem.ResolveOptions{Dir: root})
			Expect(err).NotTo(HaveOccurred())
			Expect(res.Name).To(Equal("standalone"))
		})

		It("falls back to a registered workspace containing the directory", func() {
			// Registered, but its flow.yaml is gone, so there is nothing to walk up to.
			root := mkDir("bare")
			cfg := dynamic(map[string]string{"bare": root}, "bare")
			res, err := filesystem.ResolveWorkspace(cfg, filesystem.ResolveOptions{Dir: mkDir("bare/sub")})
			Expect(err).NotTo(HaveOccurred())
			Expect(res.Name).To(Equal("bare"))
			Expect(res.Source).To(Equal(filesystem.SourcePrefix))
		})

		It("falls back to the configured current workspace outside any workspace", func() {
			root := mkWorkspace("home-ws")
			cfg := dynamic(map[string]string{"home-ws": root}, "home-ws")
			res, err := filesystem.ResolveWorkspace(cfg, filesystem.ResolveOptions{Dir: mkDir("unrelated")})
			Expect(err).NotTo(HaveOccurred())
			Expect(res.Name).To(Equal("home-ws"))
			Expect(res.Source).To(Equal(filesystem.SourceCurrent))
		})

		It("returns nil when nothing resolves", func() {
			res, err := filesystem.ResolveWorkspace(dynamic(nil, ""), filesystem.ResolveOptions{
				Dir: mkDir("nowhere"),
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(res).To(BeNil())
		})
	})

	Context("fixed mode", func() {
		It("ignores discovery and uses the configured workspace", func() {
			pinned := mkWorkspace("pinned")
			mkWorkspace("other")
			cfg := &config.Config{
				Workspaces:       map[string]string{"pinned": pinned},
				CurrentWorkspace: "pinned",
				WorkspaceMode:    config.ConfigWorkspaceModeFixed,
			}
			res, err := filesystem.ResolveWorkspace(cfg, filesystem.ResolveOptions{Dir: filepath.Join(tmpDir, "other")})
			Expect(err).NotTo(HaveOccurred())
			Expect(res.Name).To(Equal("pinned"))
			Expect(res.Source).To(Equal(filesystem.SourceCurrent))
		})

		It("still honors an explicit override", func() {
			pinned := mkWorkspace("pinned")
			other := mkWorkspace("other")
			cfg := &config.Config{
				Workspaces:       map[string]string{"pinned": pinned},
				CurrentWorkspace: "pinned",
				WorkspaceMode:    config.ConfigWorkspaceModeFixed,
			}
			res, err := filesystem.ResolveWorkspace(cfg, filesystem.ResolveOptions{Override: other})
			Expect(err).NotTo(HaveOccurred())
			Expect(res.Name).To(Equal("other"))
			Expect(res.Source).To(Equal(filesystem.SourceOverride))
		})
	})

	Context("override", func() {
		It("accepts a registered workspace name", func() {
			root := mkWorkspace("ws")
			cfg := dynamic(map[string]string{"named": root}, "named")
			res, err := filesystem.ResolveWorkspace(cfg, filesystem.ResolveOptions{
				Override: "named", Dir: mkDir("unrelated"),
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(res.Name).To(Equal("named"))
			Expect(res.Registered).To(BeTrue())
		})

		It("accepts an absolute path to an unregistered workspace", func() {
			root := mkWorkspace("adhoc")
			res, err := filesystem.ResolveWorkspace(dynamic(nil, ""), filesystem.ResolveOptions{Override: root})
			Expect(err).NotTo(HaveOccurred())
			Expect(res.Name).To(Equal("adhoc"))
			Expect(res.Registered).To(BeFalse())
			Expect(res.Source).To(Equal(filesystem.SourceOverride))
		})

		It("maps a path that is registered back to its registered name", func() {
			root := mkWorkspace("dir-name")
			cfg := dynamic(map[string]string{"registered-name": root}, "registered-name")
			res, err := filesystem.ResolveWorkspace(cfg, filesystem.ResolveOptions{Override: root})
			Expect(err).NotTo(HaveOccurred())
			Expect(res.Name).To(Equal("registered-name"))
			Expect(res.Registered).To(BeTrue())
		})

		It("reads the override from the environment when no value is passed", func() {
			root := mkWorkspace("env-ws")
			Expect(os.Setenv(filesystem.WorkspaceOverrideEnv, root)).To(Succeed())
			res, err := filesystem.ResolveWorkspace(dynamic(nil, ""), filesystem.ResolveOptions{
				Dir: mkDir("unrelated"),
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(res.Name).To(Equal("env-ws"))
		})

		It("errors on an unknown name", func() {
			_, err := filesystem.ResolveWorkspace(dynamic(nil, ""), filesystem.ResolveOptions{Override: "nope"})
			Expect(err).To(MatchError(ContainSubstring("unknown workspace")))
		})

		It("errors on a path with no flow.yaml", func() {
			_, err := filesystem.ResolveWorkspace(dynamic(nil, ""), filesystem.ResolveOptions{
				Override: mkDir("plain"),
			})
			Expect(err).To(MatchError(ContainSubstring("no flow.yaml found")))
		})
	})
})
