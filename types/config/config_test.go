package config_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/flowexec/flow/v2/types/config"
)

func TestConfig(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Config Suite")
}

var _ = Describe("Workspace path lookups", func() {
	Describe("WorkspaceForPath", func() {
		It("returns the workspace containing the path", func() {
			cfg := &config.Config{Workspaces: map[string]string{"a": "/src/a", "b": "/src/b"}}
			name, found := cfg.WorkspaceForPath("/src/a/pkg/deep")
			Expect(found).To(BeTrue())
			Expect(name).To(Equal("a"))
		})

		It("prefers the longest match when workspaces nest", func() {
			// Map iteration order used to decide this, so the nested workspace won only by luck.
			cfg := &config.Config{Workspaces: map[string]string{
				"outer": "/src/repo",
				"inner": "/src/repo/sub/worktree",
			}}
			for range 20 {
				name, found := cfg.WorkspaceForPath("/src/repo/sub/worktree/cmd")
				Expect(found).To(BeTrue())
				Expect(name).To(Equal("inner"))
			}
		})

		It("does not match a sibling sharing a name prefix", func() {
			cfg := &config.Config{Workspaces: map[string]string{"ws": "/src/ws"}}
			_, found := cfg.WorkspaceForPath("/src/wsX")
			Expect(found).To(BeFalse())
		})

		It("matches the workspace root itself", func() {
			cfg := &config.Config{Workspaces: map[string]string{"ws": "/src/ws"}}
			name, found := cfg.WorkspaceForPath("/src/ws")
			Expect(found).To(BeTrue())
			Expect(name).To(Equal("ws"))
		})

		It("returns false for an empty path or an empty config", func() {
			cfg := &config.Config{Workspaces: map[string]string{"ws": "/src/ws"}}
			_, found := cfg.WorkspaceForPath("")
			Expect(found).To(BeFalse())
			_, found = (&config.Config{}).WorkspaceForPath("/src/ws")
			Expect(found).To(BeFalse())
		})
	})

	Describe("NameForWorkspacePath", func() {
		It("matches an exact root only", func() {
			cfg := &config.Config{Workspaces: map[string]string{"ws": "/src/ws"}}
			name, found := cfg.NameForWorkspacePath("/src/ws")
			Expect(found).To(BeTrue())
			Expect(name).To(Equal("ws"))

			_, found = cfg.NameForWorkspacePath("/src/ws/sub")
			Expect(found).To(BeFalse())
		})

		It("tolerates unclean paths", func() {
			cfg := &config.Config{Workspaces: map[string]string{"ws": "/src/ws/"}}
			_, found := cfg.NameForWorkspacePath("/src/ws/./")
			Expect(found).To(BeTrue())
		})
	})

	Describe("CurrentWorkspaceName", func() {
		It("falls back to the configured workspace in fixed mode", func() {
			cfg := &config.Config{
				Workspaces:       map[string]string{"pinned": "/src/pinned"},
				CurrentWorkspace: "pinned",
				WorkspaceMode:    config.ConfigWorkspaceModeFixed,
			}
			name, err := cfg.CurrentWorkspaceName()
			Expect(err).NotTo(HaveOccurred())
			Expect(name).To(Equal("pinned"))
		})

		It("errors when nothing is configured", func() {
			cfg := &config.Config{WorkspaceMode: config.ConfigWorkspaceModeFixed}
			_, err := cfg.CurrentWorkspaceName()
			Expect(err).To(MatchError(ContainSubstring("current workspace not found")))
		})
	})

	Describe("SetDefaults", func() {
		It("picks the first workspace by sort order, not map order", func() {
			for range 20 {
				cfg := &config.Config{Workspaces: map[string]string{
					"zulu": "/z", "alpha": "/a", "mike": "/m",
				}}
				cfg.SetDefaults()
				Expect(cfg.CurrentWorkspace).To(Equal("alpha"))
			}
		})
	})
})
