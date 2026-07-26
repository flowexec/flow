package cache_test

import (
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
	"github.com/flowexec/flow/v2/types/workspace"
)

var _ = Describe("TemplateCacheImpl", func() {
	var (
		mockLogger     *mocks.MockLogger
		tmplCache      *cache.TemplateCacheImpl
		wsCache        *cacheMocks.MockWorkspaceCache
		wsName, wsPath string
		cacheDir       string
	)

	// validTemplate is a minimal, form-less template body (valid because an empty form validates).
	const validTemplate = "template: |\n  namespace: test\n  description: from template\n"
	// invalidTemplate has a form field missing both a prompt and a description, which fails validation.
	const invalidTemplate = "form:\n  - key: broken\ntemplate: |\n  namespace: test\n"

	writeTmpl := func(dir, name, content string) string {
		p := filepath.Join(dir, name)
		Expect(os.WriteFile(p, []byte(content), 0600)).To(Succeed())
		return p
	}

	BeforeEach(func() {
		mockLogger = mocks.NewMockLogger(gomock.NewController(GinkgoT()))
		logger.Init(logger.InitOptions{Logger: mockLogger, TestingTB: GinkgoTB()})
		// Discovery/update emits debug lines that aren't the subject of these tests.
		mockLogger.EXPECT().Debugf(gomock.Any()).AnyTimes()
		mockLogger.EXPECT().Debug(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()

		var err error
		cacheDir, err = os.MkdirTemp("", "flow-tmpl-cache-test")
		Expect(err).NotTo(HaveOccurred())
		Expect(os.Setenv(filesystem.FlowCacheDirEnvVar, cacheDir)).To(Succeed())

		ds, err := store.NewDataStore(filepath.Join(cacheDir, "test.db"))
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(func() { _ = ds.Close() })

		wsName = "test"
		wsPath = filepath.Join(cacheDir, "workspace")
		Expect(filesystem.InitWorkspaceConfig(wsName, wsPath)).To(Succeed())
		wsConfig, err := filesystem.LoadWorkspaceConfig(wsName, wsPath)
		Expect(err).NotTo(HaveOccurred())

		wsCache = cacheMocks.NewMockWorkspaceCache(gomock.NewController(GinkgoT()))
		wsCache.EXPECT().GetLatestData().Return(&cache.WorkspaceCacheData{
			Workspaces:         map[string]*workspace.Workspace{wsName: wsConfig},
			WorkspaceLocations: map[string]string{wsName: wsPath},
		}, nil).AnyTimes()

		var ok bool
		tmplCache, ok = cache.NewTemplateCache(wsCache, ds).(*cache.TemplateCacheImpl)
		Expect(ok).To(BeTrue())
	})

	AfterEach(func() {
		Expect(os.RemoveAll(cacheDir)).To(Succeed())
		Expect(os.Unsetenv(filesystem.FlowCacheDirEnvVar)).To(Succeed())
	})

	Describe("Update, GetTemplate, and GetTemplateList", func() {
		It("discovers a workspace template and resolves it by name", func() {
			writeTmpl(wsPath, "webapp.flow.tmpl", validTemplate)

			Expect(tmplCache.Update()).To(Succeed())

			list, err := tmplCache.GetTemplateList()
			Expect(err).NotTo(HaveOccurred())
			Expect(list).To(HaveLen(1))
			Expect(list[0].Name()).To(Equal("webapp"))

			tmpl, err := tmplCache.GetTemplate("webapp")
			Expect(err).NotTo(HaveOccurred())
			Expect(tmpl).NotTo(BeNil())
			Expect(tmpl.Name()).To(Equal("webapp"))
		})

		It("returns an error for an unknown template name", func() {
			Expect(tmplCache.Update()).To(Succeed())
			_, err := tmplCache.GetTemplate("does-not-exist")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("not found"))
		})

		It("does not discover executable or artifact files", func() {
			writeTmpl(wsPath, "webapp.flow.tmpl", validTemplate)
			writeTmpl(wsPath, "regular.flow", "namespace: test\n")
			writeTmpl(wsPath, "deployment.yaml.tmpl", "kind: Deployment\n")

			Expect(tmplCache.Update()).To(Succeed())

			list, err := tmplCache.GetTemplateList()
			Expect(err).NotTo(HaveOccurred())
			Expect(list).To(HaveLen(1))
			Expect(list[0].Name()).To(Equal("webapp"))
		})
	})

	Describe("Invalid templates", func() {
		It("skips and warns on a template that fails validation", func() {
			writeTmpl(wsPath, "broken.flow.tmpl", invalidTemplate)

			mockLogger.EXPECT().Warn(
				"invalid template found during cache update",
				"template", "broken",
				gomock.Any(), gomock.Any(),
				"workspace", wsName,
				"err", gomock.Any(),
			).Times(1)

			Expect(tmplCache.Update()).To(Succeed())

			list, err := tmplCache.GetTemplateList()
			Expect(err).NotTo(HaveOccurred())
			Expect(list).To(BeEmpty())
		})
	})

	Describe("Duplicate template names across workspaces", func() {
		It("warns and keeps a single entry, resolvable via a workspace-qualified reference", func() {
			ws2Name := "other"
			ws2Path := filepath.Join(cacheDir, "workspace2")
			Expect(filesystem.InitWorkspaceConfig(ws2Name, ws2Path)).To(Succeed())
			ws2Config, err := filesystem.LoadWorkspaceConfig(ws2Name, ws2Path)
			Expect(err).NotTo(HaveOccurred())

			wsConfig, err := filesystem.LoadWorkspaceConfig(wsName, wsPath)
			Expect(err).NotTo(HaveOccurred())

			dupWsCache := cacheMocks.NewMockWorkspaceCache(gomock.NewController(GinkgoT()))
			dupWsCache.EXPECT().GetLatestData().Return(&cache.WorkspaceCacheData{
				Workspaces: map[string]*workspace.Workspace{
					wsName:  wsConfig,
					ws2Name: ws2Config,
				},
				WorkspaceLocations: map[string]string{
					wsName:  wsPath,
					ws2Name: ws2Path,
				},
			}, nil).AnyTimes()
			tmplCache.WorkspaceCache = dupWsCache

			writeTmpl(wsPath, "shared.flow.tmpl", validTemplate)
			writeTmpl(ws2Path, "shared.flow.tmpl", validTemplate)

			mockLogger.EXPECT().Warn(
				gomock.Any(),
				"template", "shared",
				"conflictPath", gomock.Any(),
				"newPath", gomock.Any(),
				"workspace", gomock.Any(),
			).Times(1)

			Expect(tmplCache.Update()).To(Succeed())

			// Both workspace-qualified references resolve deterministically.
			t1, err := tmplCache.GetTemplate(wsName + "/shared")
			Expect(err).NotTo(HaveOccurred())
			Expect(t1.Location()).To(Equal(filepath.Join(wsPath, "shared.flow.tmpl")))

			t2, err := tmplCache.GetTemplate(ws2Name + "/shared")
			Expect(err).NotTo(HaveOccurred())
			Expect(t2.Location()).To(Equal(filepath.Join(ws2Path, "shared.flow.tmpl")))
		})
	})
})
