package filesystem_test

import (
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/flowexec/flow/v2/pkg/filesystem"
)

var _ = Describe("Discovery", func() {
	var tmpDir string

	// mkWorkspace creates dir (relative to tmpDir) and marks it a workspace root.
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

	BeforeEach(func() {
		var err error
		tmpDir, err = os.MkdirTemp("", "flow-discovery-test")
		Expect(err).NotTo(HaveOccurred())
		tmpDir, err = filepath.EvalSymlinks(tmpDir)
		Expect(err).NotTo(HaveOccurred())
		tmpDir = filesystem.NormalizePath(tmpDir)
	})

	AfterEach(func() {
		Expect(os.RemoveAll(tmpDir)).To(Succeed())
	})

	Describe("FindWorkspaceRoot", func() {
		It("finds a flow.yaml in the starting directory", func() {
			root := mkWorkspace("ws")
			found, ok := filesystem.FindWorkspaceRoot(root)
			Expect(ok).To(BeTrue())
			Expect(found).To(Equal(filesystem.NormalizePath(root)))
		})

		It("walks up to an ancestor", func() {
			root := mkWorkspace("ws")
			deep := mkDir("ws/a/b/c")
			found, ok := filesystem.FindWorkspaceRoot(deep)
			Expect(ok).To(BeTrue())
			Expect(found).To(Equal(filesystem.NormalizePath(root)))
		})

		It("prefers the closest root when workspaces nest", func() {
			mkWorkspace("ws")
			nested := mkWorkspace("ws/sub/worktree")
			deep := mkDir("ws/sub/worktree/pkg")
			found, ok := filesystem.FindWorkspaceRoot(deep)
			Expect(ok).To(BeTrue())
			Expect(found).To(Equal(filesystem.NormalizePath(nested)))
		})

		It("returns false when no flow.yaml exists above the directory", func() {
			// tmpDir has no flow.yaml, and neither do the system directories above it.
			_, ok := filesystem.FindWorkspaceRoot(mkDir("plain/nested"))
			Expect(ok).To(BeFalse())
		})

		It("returns false for an empty start directory", func() {
			_, ok := filesystem.FindWorkspaceRoot("")
			Expect(ok).To(BeFalse())
		})

		It("ignores a directory named flow.yaml", func() {
			path := filepath.Join(tmpDir, "notaws")
			Expect(os.MkdirAll(filepath.Join(path, filesystem.WorkspaceConfigFileName), 0750)).To(Succeed())
			_, ok := filesystem.FindWorkspaceRoot(path)
			Expect(ok).To(BeFalse())
		})

		It("walks past candidates the skip function rejects", func() {
			outer := mkWorkspace("ws")
			inner := mkWorkspace("ws/vendor/dep")
			found, ok := filesystem.FindWorkspaceRootExcluding(inner, func(root string) bool {
				return filepath.Base(filepath.Dir(root)) == "vendor"
			})
			Expect(ok).To(BeTrue())
			Expect(found).To(Equal(filesystem.NormalizePath(outer)))
		})
	})

	Describe("IsPathWithin", func() {
		It("matches a directory against itself and its descendants", func() {
			Expect(filesystem.IsPathWithin("/a/ws", "/a/ws")).To(BeTrue())
			Expect(filesystem.IsPathWithin("/a/ws/pkg/x", "/a/ws")).To(BeTrue())
		})

		It("does not match a sibling that shares a name prefix", func() {
			Expect(filesystem.IsPathWithin("/a/wsX", "/a/ws")).To(BeFalse())
			Expect(filesystem.IsPathWithin("/a", "/a/ws")).To(BeFalse())
		})

		It("is false for empty operands", func() {
			Expect(filesystem.IsPathWithin("", "/a")).To(BeFalse())
			Expect(filesystem.IsPathWithin("/a", "")).To(BeFalse())
		})
	})

	Describe("SamePath", func() {
		It("matches a symlinked path to its target", func() {
			target := mkDir("real")
			link := filepath.Join(tmpDir, "link")
			Expect(os.Symlink(target, link)).To(Succeed())
			Expect(filesystem.SamePath(link, target)).To(BeTrue())
		})

		It("does not match distinct directories", func() {
			Expect(filesystem.SamePath(mkDir("a"), mkDir("b"))).To(BeFalse())
		})
	})

	Describe("ReadWorkspaceConfig", func() {
		It("does not create anything when the file is missing", func() {
			path := filepath.Join(tmpDir, "absent")
			_, err := filesystem.ReadWorkspaceConfig("absent", path)
			Expect(err).To(HaveOccurred())
			_, statErr := os.Stat(path)
			Expect(os.IsNotExist(statErr)).To(BeTrue(), "must not create the workspace directory")
		})

		It("reads an existing config and attaches the given name and location", func() {
			root := mkWorkspace("ws")
			cfg, err := filesystem.ReadWorkspaceConfig("custom", root)
			Expect(err).NotTo(HaveOccurred())
			Expect(cfg.AssignedName()).To(Equal("custom"))
			Expect(cfg.Location()).To(Equal(root))
		})

		It("accepts an empty flow.yaml as a valid workspace marker", func() {
			path := filepath.Join(tmpDir, "empty")
			Expect(os.MkdirAll(path, 0750)).To(Succeed())
			Expect(os.WriteFile(filepath.Join(path, filesystem.WorkspaceConfigFileName), nil, 0600)).To(Succeed())
			cfg, err := filesystem.ReadWorkspaceConfig("empty", path)
			Expect(err).NotTo(HaveOccurred())
			Expect(cfg.AssignedName()).To(Equal("empty"))
		})
	})
})
