//go:build e2e

package tests_test

import (
	stdCtx "context"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/flowexec/flow/v2/pkg/filesystem"
	"github.com/flowexec/flow/v2/tests/utils"
	"github.com/flowexec/flow/v2/types/config"
)

// These specs build their context through context.NewContext from a real working directory,
// which is the only way to exercise workspace discovery. That changes the process working
// directory, hence Serial.
var _ = Describe("workspace discovery e2e", Serial, func() {
	var root string

	dynamicCtx := func(dir string, registered map[string]string) *utils.DiscoveryContext {
		return utils.NewDiscoveryContext(
			stdCtx.Background(), GinkgoTB(), dir, registered, config.ConfigWorkspaceModeDynamic,
		)
	}

	BeforeEach(func() {
		root = filesystem.NormalizePath(GinkgoTB().TempDir())
	})

	It("runs in a freshly cloned repo with no workspaces registered at all", func() {
		clone := utils.WriteWorkspace(GinkgoTB(), filepath.Join(root, "cloned-repo"), "greet")
		nested := filepath.Join(clone, "internal", "pkg")
		Expect(os.MkdirAll(nested, 0750)).To(Succeed())

		ctx := dynamicCtx(nested, nil)

		Expect(ctx.CurrentWorkspaceName()).To(Equal("cloned-repo"))
		Expect(ctx.WorkspaceIsRegistered()).To(BeFalse())
		Expect(ctx.CurrentWorkspace.Location()).To(Equal(clone))

		list, err := ctx.ExecutableCache.GetExecutableList()
		Expect(err).NotTo(HaveOccurred())
		Expect(list.FilterByWorkspace("cloned-repo")).To(HaveLen(1))
	})

	It("prefers a worktree nested inside a registered workspace", func() {
		parent := utils.WriteWorkspace(GinkgoTB(), filepath.Join(root, "repo"), "build")
		worktree := utils.WriteWorkspace(GinkgoTB(), filepath.Join(parent, ".worktrees", "feature"), "deploy")

		ctx := dynamicCtx(worktree, map[string]string{"repo": parent})

		Expect(ctx.CurrentWorkspaceName()).To(Equal("feature"))
		Expect(ctx.WorkspaceIsRegistered()).To(BeFalse())

		list, err := ctx.ExecutableCache.GetExecutableList()
		Expect(err).NotTo(HaveOccurred())
		Expect(list.FilterByWorkspace("feature")).To(HaveLen(1))
	})

	It("does not index a nested worktree's executables into its parent workspace", func() {
		parent := utils.WriteWorkspace(GinkgoTB(), filepath.Join(root, "repo"), "build")
		utils.WriteWorkspace(GinkgoTB(), filepath.Join(parent, ".worktrees", "feature"), "deploy")

		ctx := dynamicCtx(parent, map[string]string{"repo": parent})
		Expect(ctx.CurrentWorkspaceName()).To(Equal("repo"))
		Expect(ctx.ExecutableCache.Update()).To(Succeed())

		list, err := ctx.ExecutableCache.GetExecutableList()
		Expect(err).NotTo(HaveOccurred())
		execs := list.FilterByWorkspace("repo")
		Expect(execs).To(HaveLen(1))
		Expect(execs[0].Name).To(Equal("build"))
	})

	It("resolves a worktree that lives outside every registered workspace", func() {
		registered := utils.WriteWorkspace(GinkgoTB(), filepath.Join(root, "repo"), "build")
		worktree := utils.WriteWorkspace(GinkgoTB(), filepath.Join(root, "worktrees", "repo-feature"), "deploy")

		ctx := dynamicCtx(worktree, map[string]string{"repo": registered})
		Expect(ctx.CurrentWorkspaceName()).To(Equal("repo-feature"))
		Expect(ctx.WorkspaceIsRegistered()).To(BeFalse())
	})

	It("keeps the registered name when the directory basename differs", func() {
		dir := utils.WriteWorkspace(GinkgoTB(), filepath.Join(root, "dir-name"), "build")

		ctx := dynamicCtx(dir, map[string]string{"registered-name": dir})
		Expect(ctx.CurrentWorkspaceName()).To(Equal("registered-name"))
		Expect(ctx.WorkspaceIsRegistered()).To(BeTrue())
	})

	It("never writes a discovered workspace to the user config or the shared cache", func() {
		clone := utils.WriteWorkspace(GinkgoTB(), filepath.Join(root, "throwaway"), "greet")
		registered := utils.WriteWorkspace(GinkgoTB(), filepath.Join(root, "kept"), "build")

		ctx := dynamicCtx(clone, map[string]string{"kept": registered})
		Expect(ctx.CurrentWorkspaceName()).To(Equal("throwaway"))

		// Exercise the paths that populate and persist the caches.
		_, err := ctx.ExecutableCache.GetExecutableList()
		Expect(err).NotTo(HaveOccurred())
		Expect(ctx.ExecutableCache.Update()).To(Succeed())
		Expect(ctx.WorkspacesCache.Update()).To(Succeed())

		persisted, err := filesystem.LoadConfig()
		Expect(err).NotTo(HaveOccurred())
		Expect(persisted.Workspaces).To(HaveKey("kept"))
		Expect(persisted.Workspaces).NotTo(HaveKey("throwaway"))

		raw, err := ctx.DataStore.GetCacheEntry("workspaces")
		Expect(err).NotTo(HaveOccurred())
		Expect(string(raw)).NotTo(ContainSubstring("throwaway"))
	})

	Context("overrides", func() {
		It("honors FLOW_WORKSPACE pointing at an unregistered path", func() {
			other := utils.WriteWorkspace(GinkgoTB(), filepath.Join(root, "other"), "deploy")
			here := utils.WriteWorkspace(GinkgoTB(), filepath.Join(root, "here"), "build")
			GinkgoTB().Setenv(filesystem.WorkspaceOverrideEnv, other)

			ctx := dynamicCtx(here, nil)
			Expect(ctx.CurrentWorkspaceName()).To(Equal("other"))
			Expect(ctx.WorkspaceResolution.Source).To(Equal(filesystem.SourceOverride))
		})

		It("is the only thing that moves the workspace in fixed mode", func() {
			pinned := utils.WriteWorkspace(GinkgoTB(), filepath.Join(root, "pinned"), "build")
			elsewhere := utils.WriteWorkspace(GinkgoTB(), filepath.Join(root, "elsewhere"), "deploy")

			fixed := utils.NewDiscoveryContext(
				stdCtx.Background(), GinkgoTB(), elsewhere,
				map[string]string{"pinned": pinned}, config.ConfigWorkspaceModeFixed,
			)
			Expect(fixed.CurrentWorkspaceName()).To(Equal("pinned"),
				"fixed mode must ignore the working directory")
		})
	})
})
