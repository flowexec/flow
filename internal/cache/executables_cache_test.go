package cache_test

import (
	"os"
	"path/filepath"
	"runtime"

	"github.com/flowexec/tuikit/io/mocks"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/mock/gomock"

	"github.com/flowexec/flow/internal/cache"
	cacheMocks "github.com/flowexec/flow/internal/cache/mocks"
	"github.com/flowexec/flow/internal/filesystem"
	"github.com/flowexec/flow/internal/logger"
	"github.com/flowexec/flow/types/common"
	"github.com/flowexec/flow/types/executable"
	"github.com/flowexec/flow/types/workspace"
)

var _ = Describe("ExecutableCacheImpl", func() {
	var (
		mockLogger          *mocks.MockLogger
		execCache           *cache.ExecutableCacheImpl
		wsCache             *cacheMocks.MockWorkspaceCache
		wsName, wsPath      string
		cacheDir, configDir string
	)

	BeforeEach(func() {
		mockLogger = mocks.NewMockLogger(gomock.NewController(GinkgoT()))
		logger.Init(logger.InitOptions{Logger: mockLogger, TestingTB: GinkgoTB()})

		var err error
		cacheDir, err = os.MkdirTemp("", "flow-cache-test")
		Expect(err).NotTo(HaveOccurred())
		Expect(os.Setenv(filesystem.FlowCacheDirEnvVar, cacheDir)).To(Succeed())
		configDir, err = os.MkdirTemp("", "flow-config-test")
		Expect(err).NotTo(HaveOccurred())
		Expect(os.Setenv(filesystem.FlowConfigDirEnvVar, configDir)).To(Succeed())

		wsName = "test"
		wsPath = filepath.Join(cacheDir, "workspace")
		err = filesystem.InitWorkspaceConfig(wsName, wsPath)
		Expect(err).NotTo(HaveOccurred())
		wsConfig, err := filesystem.LoadWorkspaceConfig(wsName, wsPath)
		Expect(err).NotTo(HaveOccurred())

		v := executable.FlowFileVisibility(common.VisibilityPrivate)
		execCfg := &executable.FlowFile{
			Namespace:  "testdata",
			Visibility: &v,
			Executables: executable.ExecutableList{
				{Verb: "run", Name: "exec"},
			},
		}
		execCfg.SetContext(wsName, wsPath, filepath.Join(wsPath, "test"+executable.FlowFileExt))
		err = filesystem.WriteFlowFile(execCfg.ConfigPath(), execCfg)
		Expect(err).NotTo(HaveOccurred())
		execCacheData := &cache.ExecutableCacheData{
			ExecutableMap: make(map[executable.Ref]string),
			AliasMap:      make(map[executable.Ref]executable.Ref),
			ConfigMap:     make(map[string]cache.WorkspaceInfo),
		}
		wsCache = cacheMocks.NewMockWorkspaceCache(gomock.NewController(GinkgoT()))
		wsCache.EXPECT().GetLatestData().Return(&cache.WorkspaceCacheData{
			Workspaces:         map[string]*workspace.Workspace{wsName: wsConfig},
			WorkspaceLocations: map[string]string{wsName: wsPath},
		}, nil).AnyTimes()
		execCache = &cache.ExecutableCacheImpl{
			Data:           execCacheData,
			WorkspaceCache: wsCache,
		}
	})

	AfterEach(func() {
		Expect(os.RemoveAll(cacheDir)).To(Succeed())
		Expect(os.Unsetenv(filesystem.FlowCacheDirEnvVar)).To(Succeed())
	})

	Describe("Update and GetExecutableList", func() {
		It("should update the executable cache from filesystem and retrieve the expected data", func() {
			mockLogger.EXPECT().Debugf(gomock.Any()).Times(1)
			mockLogger.EXPECT().Debugx(gomock.Any(), "workspace", wsName).Times(1)
			mockLogger.EXPECT().Debugx(gomock.Any(), "count", 1).Times(1)
			err := execCache.Update()
			Expect(err).ToNot(HaveOccurred())

			var readData executable.ExecutableList
			readData, err = execCache.GetExecutableList()
			Expect(err).ToNot(HaveOccurred())
			Expect(readData).ToNot(BeNil())
			execs := readData.FilterByWorkspace(wsName)
			Expect(execs).To(HaveLen(1))
		})

		When("generated executables are expected", func() {
			It("should generate the expected executables", func() {
				_, f, _, _ := runtime.Caller(0)
				err := filesystem.CopyFile(
					filepath.Join(filepath.Dir(f), "testdata", "from-file.sh"),
					filepath.Join(wsPath, "from-file.sh"),
				)
				Expect(err).NotTo(HaveOccurred())
				v := executable.FlowFileVisibility(common.VisibilityPrivate)
				execCfg := &executable.FlowFile{
					Namespace:  "testdata",
					Visibility: &v,
					FromFile:   []string{"from-file.sh"},
				}
				execCfg.SetContext(wsName, wsPath, filepath.Join(wsPath, "test"+executable.FlowFileExt))
				err = filesystem.WriteFlowFile(execCfg.ConfigPath(), execCfg)
				Expect(err).NotTo(HaveOccurred())

				mockLogger.EXPECT().Debugf(gomock.Any()).Times(1)
				mockLogger.EXPECT().Debugx(gomock.Any(), "workspace", wsName).Times(1)
				mockLogger.EXPECT().Debugx(gomock.Any(), "count", 1).Times(1)
				err = execCache.Update()
				Expect(err).ToNot(HaveOccurred())

				var readData executable.ExecutableList
				readData, err = execCache.GetExecutableList()
				Expect(err).ToNot(HaveOccurred())
				Expect(readData).ToNot(BeNil())
				execs := readData.FilterByWorkspace(wsName)
				Expect(execs).To(HaveLen(1))
			})
		})
	})

	Describe("Update and GetExecutableList", func() {
		It("should update the executable cache from filesystem and retrieve the expected data", func() {
			mockLogger.EXPECT().Debugf(gomock.Any()).Times(1)
			mockLogger.EXPECT().Debugx(gomock.Any(), "workspace", wsName).Times(1)
			mockLogger.EXPECT().Debugx(gomock.Any(), "count", 1).Times(1)
			err := execCache.Update()
			Expect(err).ToNot(HaveOccurred())

			var readData *executable.Executable
			ref := executable.Ref("run test/testdata:exec")
			readData, err = execCache.GetExecutableByRef(ref)
			Expect(err).ToNot(HaveOccurred())
			Expect(readData).ToNot(BeNil())
			Expect(readData.Ref()).To(Equal(ref))
		})
	})

	Describe("Verb aliases behavior", func() {
		BeforeEach(func() {
			v := executable.FlowFileVisibility(common.VisibilityPrivate)
			execCfg := &executable.FlowFile{
				Namespace:  "testdata",
				Visibility: &v,
				Executables: executable.ExecutableList{
					{Verb: "run", Name: "test-alias", Aliases: []string{"alias1"}},
				},
			}
			execCfg.SetContext(wsName, wsPath, filepath.Join(wsPath, "aliases-test"+executable.FlowFileExt))

			err := filesystem.WriteFlowFile(execCfg.ConfigPath(), execCfg)
			Expect(err).NotTo(HaveOccurred())

			wsCache = cacheMocks.NewMockWorkspaceCache(gomock.NewController(GinkgoT()))
			execCache.WorkspaceCache = wsCache
		})

		Context("when the executable has verbAliases configured", func() {
			It("should be accessible from the alias", func() {
				wsConfig := &workspace.Workspace{VerbAliases: nil}
				wsConfig.SetContext(wsName, wsPath)
				wsCache.EXPECT().GetLatestData().Return(&cache.WorkspaceCacheData{
					Workspaces:         map[string]*workspace.Workspace{wsName: wsConfig},
					WorkspaceLocations: map[string]string{wsName: wsPath},
				}, nil).AnyTimes()

				v := executable.FlowFileVisibility(common.VisibilityPrivate)
				execCfg := &executable.FlowFile{
					Namespace:  "testdata",
					Visibility: &v,
					Executables: executable.ExecutableList{
						{Verb: "run", Name: "run-alias", VerbAliases: []executable.Verb{executable.VerbActivate}},
					},
				}
				execCfg.SetContext(wsName, wsPath, filepath.Join(wsPath, "verbaliases"+executable.FlowFileExt))
				err := filesystem.WriteFlowFile(execCfg.ConfigPath(), execCfg)
				Expect(err).NotTo(HaveOccurred())

				mockLogger.EXPECT().Debugf(gomock.Any()).AnyTimes()
				mockLogger.EXPECT().Debugx(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
				Expect(execCache.Update()).To(Succeed())

				// Should be able to access via verb alias "activate"
				execRef := executable.Ref("activate test/testdata:run-alias")
				exec, err := execCache.GetExecutableByRef(execRef)
				Expect(err).NotTo(HaveOccurred())
				Expect(exec).NotTo(BeNil())
				Expect(exec.Name).To(Equal("run-alias"))

			})
		})

		Context("when workspace has no verbAliases configured (nil)", func() {
			It("should allow access via default verb aliases", func() {
				wsConfig := &workspace.Workspace{VerbAliases: nil}
				wsConfig.SetContext(wsName, wsPath)
				wsCache.EXPECT().GetLatestData().Return(&cache.WorkspaceCacheData{
					Workspaces:         map[string]*workspace.Workspace{wsName: wsConfig},
					WorkspaceLocations: map[string]string{wsName: wsPath},
				}, nil).AnyTimes()

				mockLogger.EXPECT().Debugf(gomock.Any()).AnyTimes()
				mockLogger.EXPECT().Debugx(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
				Expect(execCache.Update()).To(Succeed())

				// Should be able to access via default verb "run"
				runRef := executable.Ref("run test/testdata:test-alias")
				exec, err := execCache.GetExecutableByRef(runRef)
				Expect(err).NotTo(HaveOccurred())
				Expect(exec).NotTo(BeNil())
				Expect(exec.Name).To(Equal("test-alias"))

				// Should also be able to access via default verb alias "exec"
				execRef := executable.Ref("exec test/testdata:test-alias")
				exec, err = execCache.GetExecutableByRef(execRef)
				Expect(err).NotTo(HaveOccurred())
				Expect(exec).NotTo(BeNil())
				Expect(exec.Name).To(Equal("test-alias"))

				// Should also be able to access via alias
				aliasRef := executable.Ref("exec test/testdata:alias1")
				execFromAlias, err := execCache.GetExecutableByRef(aliasRef)
				Expect(err).NotTo(HaveOccurred())
				Expect(execFromAlias).NotTo(BeNil())
				Expect(execFromAlias.Name).To(Equal("test-alias"))
			})
		})

		Context("when workspace has empty verbAliases map", func() {
			It("should disable all verb aliases", func() {
				emptyAliases := &workspace.WorkspaceVerbAliases{}
				wsConfig := &workspace.Workspace{VerbAliases: emptyAliases}
				wsConfig.SetContext(wsName, wsPath)
				wsCache.EXPECT().GetLatestData().Return(&cache.WorkspaceCacheData{
					Workspaces:         map[string]*workspace.Workspace{wsName: wsConfig},
					WorkspaceLocations: map[string]string{wsName: wsPath},
				}, nil).AnyTimes()

				mockLogger.EXPECT().Debugf(gomock.Any()).AnyTimes()
				mockLogger.EXPECT().Debugx(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
				Expect(execCache.Update()).To(Succeed())

				// Should NOT be able to access via default aliases like "exec"
				execRef := executable.Ref("exec test/testdata:test-alias")
				_, err := execCache.GetExecutableByRef(execRef)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("unable to find executable"))

				// Should still be able to access via primary verb "run"
				runRef := executable.Ref("run test/testdata:test-alias")
				exec, err := execCache.GetExecutableByRef(runRef)
				Expect(err).NotTo(HaveOccurred())
				Expect(exec).NotTo(BeNil())
				Expect(exec.Name).To(Equal("test-alias"))
			})
		})

		Context("when workspace has custom verbAliases", func() {
			It("should only allow access via custom aliases", func() {
				customAliases := &workspace.WorkspaceVerbAliases{"run": {"exec", "start"}}
				wsConfig := &workspace.Workspace{VerbAliases: customAliases}
				wsConfig.SetContext(wsName, wsPath)
				wsCache.EXPECT().GetLatestData().Return(&cache.WorkspaceCacheData{
					Workspaces:         map[string]*workspace.Workspace{wsName: wsConfig},
					WorkspaceLocations: map[string]string{wsName: wsPath},
				}, nil).AnyTimes()

				mockLogger.EXPECT().Debugf(gomock.Any()).AnyTimes()
				mockLogger.EXPECT().Debugx(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
				Expect(execCache.Update()).To(Succeed())

				// Should be able to access via custom aliases
				execAliasRef := executable.Ref("exec test/testdata:test-alias")
				exec, err := execCache.GetExecutableByRef(execAliasRef)
				Expect(err).NotTo(HaveOccurred())
				Expect(exec).NotTo(BeNil())
				Expect(exec.Name).To(Equal("test-alias"))

				startRef := executable.Ref("start test/testdata:test-alias")
				exec, err = execCache.GetExecutableByRef(startRef)
				Expect(err).NotTo(HaveOccurred())
				Expect(exec).NotTo(BeNil())
				Expect(exec.Name).To(Equal("test-alias"))

				// Should also work with executable aliases
				aliasExecRef := executable.Ref("exec test/testdata:alias1")
				execFromAlias, err := execCache.GetExecutableByRef(aliasExecRef)
				Expect(err).NotTo(HaveOccurred())
				Expect(execFromAlias).NotTo(BeNil())
				Expect(execFromAlias.Name).To(Equal("test-alias"))

				// Should NOT be able to access via default aliases like "execute"
				executeRef := executable.Ref("execute test/testdata:test-alias")
				_, err = execCache.GetExecutableByRef(executeRef)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("unable to find executable"))
			})
		})

		Context("when workspace has custom verbAliases for different verb", func() {
			It("should not provide any aliases for unmatched verbs", func() {
				customAliases := &workspace.WorkspaceVerbAliases{"view": {"show", "display"}}
				wsConfig := &workspace.Workspace{VerbAliases: customAliases}
				wsConfig.SetContext(wsName, wsPath)
				wsCache.EXPECT().GetLatestData().Return(&cache.WorkspaceCacheData{
					Workspaces:         map[string]*workspace.Workspace{wsName: wsConfig},
					WorkspaceLocations: map[string]string{wsName: wsPath},
				}, nil).AnyTimes()

				mockLogger.EXPECT().Debugf(gomock.Any()).AnyTimes()
				mockLogger.EXPECT().Debugx(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
				Expect(execCache.Update()).To(Succeed())

				// Should NOT be able to access via any aliases since no aliases are configured for "run"
				execRef := executable.Ref("exec test/testdata:test-alias")
				_, err := execCache.GetExecutableByRef(execRef)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("unable to find executable"))

				// Should still be able to access via primary verb "run"
				runRef := executable.Ref("run test/testdata:test-alias")
				exec, err := execCache.GetExecutableByRef(runRef)
				Expect(err).NotTo(HaveOccurred())
				Expect(exec).NotTo(BeNil())
				Expect(exec.Name).To(Equal("test-alias"))
			})
		})
	})

	Describe("Duplicate executable warnings", func() {
		Context("when duplicate executables exist in different flow files", func() {
			It("should log warning for duplicate executable", func() {
				v := executable.FlowFileVisibility(common.VisibilityPrivate)
				execCfg1 := &executable.FlowFile{
					Namespace:  "testdata",
					Visibility: &v,
					Executables: executable.ExecutableList{
						{Verb: "run", Name: "duplicate-exec"},
					},
				}
				execCfg1.SetContext(wsName, wsPath, filepath.Join(wsPath, "first"+executable.FlowFileExt))
				err := filesystem.WriteFlowFile(execCfg1.ConfigPath(), execCfg1)
				Expect(err).NotTo(HaveOccurred())

				execCfg2 := &executable.FlowFile{
					Namespace:  "testdata",
					Visibility: &v,
					Executables: executable.ExecutableList{
						{Verb: "run", Name: "duplicate-exec"},
					},
				}
				execCfg2.SetContext(wsName, wsPath, filepath.Join(wsPath, "second"+executable.FlowFileExt))
				err = filesystem.WriteFlowFile(execCfg2.ConfigPath(), execCfg2)
				Expect(err).NotTo(HaveOccurred())

				mockLogger.EXPECT().Debugf(gomock.Any()).AnyTimes()
				mockLogger.EXPECT().Debugx(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
				mockLogger.EXPECT().Warnx(
					"duplicate executable found during cache update",
					"ref", "run test/testdata:duplicate-exec",
					"conflictPath", execCfg1.ConfigPath(),
					"newPath", execCfg2.ConfigPath(),
					"workspace", wsName,
				).Times(1)

				err = execCache.Update()
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("when duplicate executable aliases exist", func() {
			It("should log warning for duplicate executable alias", func() {
				v := executable.FlowFileVisibility(common.VisibilityPrivate)
				execCfg1 := &executable.FlowFile{
					Namespace:  "testdata",
					Visibility: &v,
					Executables: executable.ExecutableList{
						{Verb: "run", Name: "first-exec"},
					},
				}
				execCfg1.SetContext(wsName, wsPath, filepath.Join(wsPath, "first"+executable.FlowFileExt))
				err := filesystem.WriteFlowFile(execCfg1.ConfigPath(), execCfg1)
				Expect(err).NotTo(HaveOccurred())

				execCfg2 := &executable.FlowFile{
					Namespace:  "testdata",
					Visibility: &v,
					Executables: executable.ExecutableList{
						{Verb: "exec", Name: "first-exec"},
					},
				}
				execCfg2.SetContext(wsName, wsPath, filepath.Join(wsPath, "second"+executable.FlowFileExt))
				err = filesystem.WriteFlowFile(execCfg2.ConfigPath(), execCfg2)
				Expect(err).NotTo(HaveOccurred())

				mockLogger.EXPECT().Debugf(gomock.Any()).AnyTimes()
				mockLogger.EXPECT().Debugx(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
				mockLogger.EXPECT().Warnx(
					"duplicate executable alias found during cache update",
					gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
				).AnyTimes()

				err = execCache.Update()
				Expect(err).NotTo(HaveOccurred())
			})

			It("should log warning for duplicate name aliases", func() {
				v := executable.FlowFileVisibility(common.VisibilityPrivate)
				execCfg1 := &executable.FlowFile{
					Namespace:  "testdata",
					Visibility: &v,
					Executables: executable.ExecutableList{
						{Verb: "run", Name: "first-exec", Aliases: []string{"shared-alias"}},
					},
				}
				execCfg1.SetContext(wsName, wsPath, filepath.Join(wsPath, "first"+executable.FlowFileExt))
				err := filesystem.WriteFlowFile(execCfg1.ConfigPath(), execCfg1)
				Expect(err).NotTo(HaveOccurred())

				execCfg2 := &executable.FlowFile{
					Namespace:  "testdata",
					Visibility: &v,
					Executables: executable.ExecutableList{
						{Verb: "run", Name: "shared-alias"},
					},
				}
				execCfg2.SetContext(wsName, wsPath, filepath.Join(wsPath, "second"+executable.FlowFileExt))
				err = filesystem.WriteFlowFile(execCfg2.ConfigPath(), execCfg2)
				Expect(err).NotTo(HaveOccurred())

				mockLogger.EXPECT().Debugf(gomock.Any()).AnyTimes()
				mockLogger.EXPECT().Debugx(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
				mockLogger.EXPECT().Warnx(
					"duplicate executable alias found during cache update",
					"aliasRef", gomock.Any(),
					"conflictRef", gomock.Any(),
					"primaryRef", gomock.Any(),
					"workspace", wsName,
				).AnyTimes()

				err = execCache.Update()
				Expect(err).NotTo(HaveOccurred())
			})
		})
	})
})
