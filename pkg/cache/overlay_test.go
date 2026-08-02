package cache_test

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/flowexec/tuikit/io/mocks"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/mock/gomock"

	"github.com/flowexec/flow/v2/pkg/cache"
	cacheMocks "github.com/flowexec/flow/v2/pkg/cache/mocks"
	"github.com/flowexec/flow/v2/pkg/filesystem"
	"github.com/flowexec/flow/v2/pkg/logger"
	"github.com/flowexec/flow/v2/pkg/store"
	"github.com/flowexec/flow/v2/types/common"
	"github.com/flowexec/flow/v2/types/executable"
	"github.com/flowexec/flow/v2/types/workspace"
)

var _ = Describe("Local workspace overlay", func() {
	var (
		ds                    store.DataStore
		baseExecCache         cache.ExecutableCache
		baseWsCache           cache.WorkspaceCache
		registeredWs, localWs *workspace.Workspace
		tmpDir                string
	)

	// writeWorkspace creates a workspace rooted at tmpDir/rel holding one executable named
	// execName, and returns its loaded config.
	writeWorkspace := func(name, rel, execName string) *workspace.Workspace {
		path := filepath.Join(tmpDir, rel)
		Expect(filesystem.InitWorkspaceConfig(name, path)).To(Succeed())
		wsCfg, err := filesystem.LoadWorkspaceConfig(name, path)
		Expect(err).NotTo(HaveOccurred())

		v := executable.FlowFileVisibility(common.VisibilityPrivate)
		flowFile := &executable.FlowFile{
			Namespace:   "ns",
			Visibility:  &v,
			Executables: executable.ExecutableList{{Verb: "run", Name: execName, Exec: &executable.ExecExecutableType{}}},
		}
		flowFile.SetContext(name, path, filepath.Join(path, "test"+executable.FlowFileExt))
		Expect(filesystem.WriteFlowFile(flowFile.ConfigPath(), flowFile)).To(Succeed())
		return wsCfg
	}

	BeforeEach(func() {
		mockLogger := mocks.NewMockLogger(gomock.NewController(GinkgoT()))
		logger.Init(logger.InitOptions{Logger: mockLogger, TestingTB: GinkgoTB()})
		mockLogger.EXPECT().Debug(gomock.Any(), gomock.Any()).AnyTimes()
		mockLogger.EXPECT().Debugf(gomock.Any()).AnyTimes()
		mockLogger.EXPECT().Debug(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
		mockLogger.EXPECT().Warn(gomock.Any(), gomock.Any()).AnyTimes()

		var err error
		tmpDir, err = os.MkdirTemp("", "flow-overlay-test")
		Expect(err).NotTo(HaveOccurred())
		Expect(os.Setenv(filesystem.FlowCacheDirEnvVar, tmpDir)).To(Succeed())

		ds, err = store.NewDataStore(filepath.Join(tmpDir, "test.db"))
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(func() { _ = ds.Close() })

		registeredWs = writeWorkspace("registered", "registered", "build")
		localWs = writeWorkspace("worktree", "worktree", "deploy")

		mockWsCache := cacheMocks.NewMockWorkspaceCache(gomock.NewController(GinkgoT()))
		mockWsCache.EXPECT().GetLatestData().Return(&cache.WorkspaceCacheData{
			Workspaces:         map[string]*workspace.Workspace{"registered": registeredWs},
			WorkspaceLocations: map[string]string{"registered": registeredWs.Location()},
		}, nil).AnyTimes()
		mockWsCache.EXPECT().GetData().Return(&cache.WorkspaceCacheData{
			Workspaces:         map[string]*workspace.Workspace{"registered": registeredWs},
			WorkspaceLocations: map[string]string{"registered": registeredWs.Location()},
		}).AnyTimes()
		mockWsCache.EXPECT().GetWorkspaceConfigList().
			Return(workspace.WorkspaceList{registeredWs}, nil).AnyTimes()
		mockWsCache.EXPECT().Update().Return(nil).AnyTimes()
		baseWsCache = mockWsCache

		baseExecCache = cache.NewExecutableCache(baseWsCache, ds)
		Expect(baseExecCache.Update()).To(Succeed())
	})

	AfterEach(func() {
		Expect(os.RemoveAll(tmpDir)).To(Succeed())
		Expect(os.Unsetenv(filesystem.FlowCacheDirEnvVar)).To(Succeed())
	})

	Describe("executables", func() {
		It("resolves an executable from the discovered workspace", func() {
			overlay := cache.NewLocalExecutableCache(baseExecCache, localWs)
			exec, err := overlay.GetExecutableByRef(executable.NewRef("worktree/ns:deploy", "run"))
			Expect(err).NotTo(HaveOccurred())
			Expect(exec.Name).To(Equal("deploy"))
		})

		It("resolves a verb alias from the discovered workspace", func() {
			overlay := cache.NewLocalExecutableCache(baseExecCache, localWs)
			// "exec" is a related verb of "run", so the alias map has to be populated too.
			exec, err := overlay.GetExecutableByRef(executable.NewRef("worktree/ns:deploy", "exec"))
			Expect(err).NotTo(HaveOccurred())
			Expect(exec.Name).To(Equal("deploy"))
		})

		It("falls through to the persisted cache for other workspaces", func() {
			overlay := cache.NewLocalExecutableCache(baseExecCache, localWs)
			exec, err := overlay.GetExecutableByRef(executable.NewRef("registered/ns:build", "run"))
			Expect(err).NotTo(HaveOccurred())
			Expect(exec.Name).To(Equal("build"))
		})

		It("returns not-found for a ref in neither", func() {
			overlay := cache.NewLocalExecutableCache(baseExecCache, localWs)
			_, err := overlay.GetExecutableByRef(executable.NewRef("worktree/ns:absent", "run"))
			Expect(err).To(HaveOccurred())
		})

		It("lists executables from both the overlay and the persisted cache", func() {
			overlay := cache.NewLocalExecutableCache(baseExecCache, localWs)
			list, err := overlay.GetExecutableList()
			Expect(err).NotTo(HaveOccurred())
			Expect(list.FilterByWorkspace("worktree")).To(HaveLen(1))
			Expect(list.FilterByWorkspace("registered")).To(HaveLen(1))
		})

		It("shadows a persisted workspace of the same name", func() {
			// A clone whose directory name matches an existing workspace resolves to itself,
			// so the persisted entries for that name must not leak through.
			collide := writeWorkspace("registered", "clone-of-registered", "deploy")
			overlay := cache.NewLocalExecutableCache(baseExecCache, collide)

			list, err := overlay.GetExecutableList()
			Expect(err).NotTo(HaveOccurred())
			execs := list.FilterByWorkspace("registered")
			Expect(execs).To(HaveLen(1))
			Expect(execs[0].Name).To(Equal("deploy"))
			Expect(execs[0].WorkspacePath()).To(Equal(collide.Location()))
		})
	})

	Describe("persistence", func() {
		It("never writes the discovered workspace into the data store", func() {
			overlay := cache.NewLocalExecutableCache(baseExecCache, localWs)
			wsOverlay := cache.NewLocalWorkspaceCache(baseWsCache, localWs)

			// Exercise every path that could plausibly persist: reads, then an explicit Update.
			_, err := overlay.GetExecutableList()
			Expect(err).NotTo(HaveOccurred())
			_, err = wsOverlay.GetLatestData()
			Expect(err).NotTo(HaveOccurred())
			Expect(overlay.Update()).To(Succeed())
			Expect(wsOverlay.Update()).To(Succeed())

			raw, err := ds.GetCacheEntry("executables")
			Expect(err).NotTo(HaveOccurred())
			var persisted cache.ExecutableCacheData
			Expect(json.Unmarshal(raw, &persisted)).To(Succeed())
			for ref := range persisted.ExecutableMap {
				Expect(ref.String()).NotTo(ContainSubstring("worktree/"))
			}
			for _, info := range persisted.ConfigMap {
				Expect(info.WorkspaceName).NotTo(Equal("worktree"))
			}
		})
	})

	Describe("workspaces", func() {
		It("includes the discovered workspace in the config list", func() {
			overlay := cache.NewLocalWorkspaceCache(baseWsCache, localWs)
			list, err := overlay.GetWorkspaceConfigList()
			Expect(err).NotTo(HaveOccurred())
			Expect(list.FindByName("worktree")).NotTo(BeNil())
			Expect(list.FindByName("registered")).NotTo(BeNil())
		})

		It("does not mutate the base cache data", func() {
			overlay := cache.NewLocalWorkspaceCache(baseWsCache, localWs)
			_, err := overlay.GetLatestData()
			Expect(err).NotTo(HaveOccurred())

			base, err := baseWsCache.GetLatestData()
			Expect(err).NotTo(HaveOccurred())
			Expect(base.Workspaces).NotTo(HaveKey("worktree"))
		})
	})
})
