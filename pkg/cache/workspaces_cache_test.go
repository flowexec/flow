package cache_test

import (
	"os"
	"path/filepath"

	"github.com/flowexec/tuikit/io/mocks"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/mock/gomock"

	"github.com/flowexec/flow/v2/pkg/cache"
	"github.com/flowexec/flow/v2/pkg/filesystem"
	"github.com/flowexec/flow/v2/pkg/logger"
	"github.com/flowexec/flow/v2/pkg/store"
	"github.com/flowexec/flow/v2/types/workspace"
)

var _ = Describe("WorkspaceCacheImpl", func() {
	var (
		mockLogger *mocks.MockLogger
		wsCache    *cache.WorkspaceCacheImpl
		cacheDir   string
		configDir  string
	)

	BeforeEach(func() {
		ctrl := gomock.NewController(GinkgoT())
		mockLogger = mocks.NewMockLogger(ctrl)
		logger.Init(logger.InitOptions{Logger: mockLogger, TestingTB: GinkgoTB()})

		var err error
		cacheDir, err = os.MkdirTemp("", "flow-cache-test")
		Expect(err).NotTo(HaveOccurred())
		Expect(os.Setenv(filesystem.FlowCacheDirEnvVar, cacheDir)).To(Succeed())
		configDir, err = os.MkdirTemp("", "flow-config-test")
		Expect(err).NotTo(HaveOccurred())
		Expect(os.Setenv(filesystem.FlowConfigDirEnvVar, configDir)).To(Succeed())

		ds, err := store.NewDataStore(filepath.Join(cacheDir, "test.db"))
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(func() { _ = ds.Close() })

		var ok bool
		wsCache, ok = cache.NewWorkspaceCache(ds).(*cache.WorkspaceCacheImpl)
		Expect(ok).To(BeTrue())

		testWs := &workspace.Workspace{}
		testWs.SetContext("test", cacheDir)
		wsCache.Data.Workspaces["test"] = testWs
		wsCache.Data.WorkspaceLocations["test"] = cacheDir
	})

	AfterEach(func() {
		Expect(os.RemoveAll(cacheDir)).To(Succeed())
		Expect(os.Unsetenv(filesystem.FlowCacheDirEnvVar)).To(Succeed())
		Expect(os.RemoveAll(configDir)).To(Succeed())
		Expect(os.Unsetenv(filesystem.FlowConfigDirEnvVar)).To(Succeed())
	})

	Describe("Update and GetLatestData", func() {
		It("should update the workspace cache and retrieve the same data", func() {
			mockLogger.EXPECT().Debugf(gomock.Any()).Times(1)
			mockLogger.EXPECT().Debug(gomock.Any(), "count", 1).Times(1)
			err := wsCache.Update()
			Expect(err).ToNot(HaveOccurred())

			var readData *cache.WorkspaceCacheData
			readData, err = wsCache.GetLatestData()
			Expect(err).ToNot(HaveOccurred())
			Expect(readData).To(Equal(wsCache.Data))
		})
	})

	Describe("GetWorkspaceConfigList", func() {
		It("returns the expected workspace configs", func() {
			list, err := wsCache.GetWorkspaceConfigList()
			Expect(err).ToNot(HaveOccurred())
			Expect(list).To(HaveLen(1))
			Expect(list.FindByName("test")).To(Equal(wsCache.Data.Workspaces["test"]))
		})
	})
})
