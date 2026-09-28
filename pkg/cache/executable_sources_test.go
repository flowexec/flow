package cache_test

import (
	"errors"
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

var _ = Describe("Executable sources", func() {
	const (
		wsName     = "test"
		sourceName = "tasks"
	)
	var (
		ds         store.DataStore
		wsCache    *cacheMocks.MockWorkspaceCache
		wsConfig   *workspace.Workspace
		wsPath     string
		sourceFile string
		calls      int
		sourceErr  error
	)

	newCache := func() cache.ExecutableCache { return cache.NewExecutableCache(wsCache, ds) }

	source := func(ws *workspace.Workspace) (executable.ExecutableList, error) {
		calls++
		Expect(ws.AssignedName()).To(Equal(wsName))
		if sourceErr != nil {
			return nil, sourceErr
		}
		execs := executable.ExecutableList{
			{Verb: executable.VerbUpdate, Name: "app", Exec: &executable.ExecExecutableType{Cmd: "task update"}},
			{Verb: executable.VerbRun, Name: "shadowed", Exec: &executable.ExecExecutableType{Cmd: "task run"}},
		}
		for _, e := range execs {
			e.SetContext("", "", "discovered", "")
			e.SetSourceFilePath(sourceFile)
		}
		return execs, nil
	}

	BeforeEach(func() {
		mockLogger := mocks.NewMockLogger(gomock.NewController(GinkgoT()))
		logger.Init(logger.InitOptions{Logger: mockLogger, TestingTB: GinkgoTB()})
		mockLogger.EXPECT().Debug(gomock.Any(), gomock.Any()).AnyTimes()
		mockLogger.EXPECT().Debugf(gomock.Any(), gomock.Any()).AnyTimes()
		mockLogger.EXPECT().Warn(gomock.Any(), gomock.Any()).AnyTimes()
		mockLogger.EXPECT().Error(gomock.Any(), gomock.Any()).AnyTimes()

		cacheDir := GinkgoT().TempDir()
		GinkgoT().Setenv(filesystem.FlowCacheDirEnvVar, cacheDir)
		GinkgoT().Setenv(filesystem.FlowConfigDirEnvVar, GinkgoT().TempDir())

		var err error
		ds, err = store.NewDataStore(filepath.Join(cacheDir, "test.db"))
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(func() { _ = ds.Close() })

		wsPath = filepath.Join(cacheDir, "workspace")
		Expect(filesystem.InitWorkspaceConfig(wsName, wsPath)).To(Succeed())
		wsConfig, err = filesystem.LoadWorkspaceConfig(wsName, wsPath)
		Expect(err).NotTo(HaveOccurred())
		sourceFile = filepath.Join(wsPath, "Taskfile.yml")
		Expect(os.WriteFile(sourceFile, []byte("version: 3\n"), 0o600)).To(Succeed())

		v := executable.FlowFileVisibility(common.VisibilityPrivate)
		flowFile := &executable.FlowFile{
			Namespace:  "discovered",
			Visibility: &v,
			Executables: executable.ExecutableList{
				{Verb: executable.VerbRun, Name: "shadowed", Exec: &executable.ExecExecutableType{Cmd: "echo flow"}},
			},
		}
		flowFile.SetContext(wsName, wsPath, filepath.Join(wsPath, "test"+executable.FlowFileExt))
		Expect(filesystem.WriteFlowFile(flowFile.ConfigPath(), flowFile)).To(Succeed())

		wsCache = cacheMocks.NewMockWorkspaceCache(gomock.NewController(GinkgoT()))
		wsCache.EXPECT().GetLatestData().Return(&cache.WorkspaceCacheData{
			Workspaces:         map[string]*workspace.Workspace{wsName: wsConfig},
			WorkspaceLocations: map[string]string{wsName: wsPath},
		}, nil).AnyTimes()

		calls, sourceErr = 0, nil
		cache.RegisterExecutableSource(sourceName, source)
		DeferCleanup(func() { cache.UnregisterExecutableSource(sourceName) })
	})

	updateRef := executable.Ref("update test/discovered:app")

	It("lists and resolves sourced executables without calling the source on read", func() {
		Expect(newCache().Update()).To(Succeed())
		Expect(calls).To(Equal(1))

		c := newCache()
		list, err := c.GetExecutableList()
		Expect(err).NotTo(HaveOccurred())
		Expect(list).To(HaveLen(2))

		exec, err := c.GetExecutableByRef(updateRef)
		Expect(err).NotTo(HaveOccurred())
		Expect(exec.Exec.Cmd).To(Equal("task update"))
		Expect(exec.Workspace()).To(Equal(wsName))
		Expect(exec.WorkspacePath()).To(Equal(wsPath))
		Expect(exec.Namespace()).To(Equal("discovered"))
		Expect(exec.SourceFilePath()).To(Equal(sourceFile))
		Expect(exec.FlowFilePath()).To(Equal(sourceFile))
		Expect(calls).To(Equal(1))
	})

	It("resolves sourced executables through related verbs and reports their aliases", func() {
		c := newCache()
		Expect(c.Update()).To(Succeed())

		exec, err := c.GetExecutableByRef("upgrade test/discovered:app")
		Expect(err).NotTo(HaveOccurred())
		Expect(exec.Ref()).To(Equal(updateRef))

		aliases, err := c.AliasRefs()
		Expect(err).NotTo(HaveOccurred())
		Expect(aliases).To(HaveKeyWithValue(executable.Ref("upgrade test/discovered:app"), updateRef))
	})

	It("prefers a flow file's executable over a sourced one with the same ref", func() {
		c := newCache()
		Expect(c.Update()).To(Succeed())

		exec, err := c.GetExecutableByRef("run test/discovered:shadowed")
		Expect(err).NotTo(HaveOccurred())
		Expect(exec.Exec.Cmd).To(Equal("echo flow"))
	})

	It("keeps updating when a source fails", func() {
		sourceErr = errors.New("unreadable")
		c := newCache()
		Expect(c.Update()).To(Succeed())

		list, err := c.GetExecutableList()
		Expect(err).NotTo(HaveOccurred())
		Expect(list).To(HaveLen(1))
	})

	It("hides but keeps the executables of a source that isn't registered", func() {
		Expect(newCache().Update()).To(Succeed())

		cache.UnregisterExecutableSource(sourceName)
		c := newCache()
		Expect(c.Update()).To(Succeed())
		_, err := c.GetExecutableByRef(updateRef)
		Expect(err).To(HaveOccurred())
		list, err := c.GetExecutableList()
		Expect(err).NotTo(HaveOccurred())
		Expect(list).To(HaveLen(1))
		aliases, err := c.AliasRefs()
		Expect(err).NotTo(HaveOccurred())
		Expect(aliases).NotTo(HaveKey(executable.Ref("upgrade test/discovered:app")))

		cache.RegisterExecutableSource(sourceName, source)
		exec, err := newCache().GetExecutableByRef(updateRef)
		Expect(err).NotTo(HaveOccurred())
		Expect(exec.Exec.Cmd).To(Equal("task update"))
	})
})
