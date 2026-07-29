package filesystem_test

import (
	"os"
	"path/filepath"

	"github.com/flowexec/tuikit/io/mocks"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/mock/gomock"

	"github.com/flowexec/flow/v2/pkg/filesystem"
	"github.com/flowexec/flow/v2/pkg/logger"
	"github.com/flowexec/flow/v2/types/executable"
	"github.com/flowexec/flow/v2/types/workspace"
)

var _ = Describe("Executables", func() {
	var (
		tmpDir string
	)

	BeforeEach(func() {
		var err error
		tmpDir, err = os.MkdirTemp("", "flow-executables-test")
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		Expect(os.RemoveAll(tmpDir)).To(Succeed())
	})

	Describe("EnsureExecutableDir", func() {
		It("creates the directory if it does not exist", func() {
			Expect(filesystem.EnsureExecutableDir(tmpDir, "subPath")).To(Succeed())
			_, err := os.Stat(filepath.Join(tmpDir, "subPath"))
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Describe("WriteFlowFile and LoadFlowFile", func() {
		It("writes and reads executable definition correctly", func() {
			executableDefinition := &executable.FlowFile{
				Namespace: "test",
				Executables: executable.ExecutableList{
					{
						Verb: "exec",
						Name: "test-executable",
					},
				},
			}

			definitionFile := filepath.Join(tmpDir, "test"+executable.FlowFileExt)
			Expect(filesystem.WriteFlowFile(definitionFile, executableDefinition)).To(Succeed())

			readDefinition, err := filesystem.LoadFlowFile(definitionFile)
			Expect(err).NotTo(HaveOccurred())
			Expect(readDefinition).To(Equal(executableDefinition))
		})
	})

	Describe("LoadWorkspaceFlowFiles", func() {
		It("loads all executable definitions if no paths are set", func() {
			executableDefinition := &executable.FlowFile{
				Namespace: "test",
				Executables: executable.ExecutableList{
					{
						Verb: "exec",
						Name: "test-executable",
					},
				},
			}

			definitionFile := filepath.Join(tmpDir, "test"+executable.FlowFileExt)
			Expect(filesystem.WriteFlowFile(definitionFile, executableDefinition)).To(Succeed())

			workspaceCfg := &workspace.Workspace{}
			workspaceCfg.SetContext("test", tmpDir)

			ctrl := gomock.NewController(GinkgoT())
			logger := mocks.NewMockLogger(ctrl)
			logger.EXPECT().Debug(gomock.Any(), gomock.Any()).AnyTimes()

			definitions, err := filesystem.LoadWorkspaceFlowFiles(workspaceCfg)
			Expect(err).NotTo(HaveOccurred())
			Expect(definitions).To(HaveLen(1))
			Expect(definitions[0].Namespace).To(Equal(executableDefinition.Namespace))
		})
		It("loads executable definitions from the included path", func() {
			executableDefinition := &executable.FlowFile{
				Namespace: "test",
				Executables: executable.ExecutableList{
					{
						Verb: "exec",
						Name: "test-executable",
					},
				},
			}

			definitionFile := filepath.Join(tmpDir, "test"+executable.FlowFileExt)
			Expect(filesystem.WriteFlowFile(definitionFile, executableDefinition)).To(Succeed())

			workspaceCfg := &workspace.Workspace{
				Executables: &workspace.ExecutableFilter{
					Included: []string{tmpDir},
					Excluded: []string{filepath.Join(tmpDir, "excluded")},
				},
			}
			workspaceCfg.SetContext("test", tmpDir)

			ctrl := gomock.NewController(GinkgoT())
			mockLogger := mocks.NewMockLogger(ctrl)
			logger.Init(logger.InitOptions{Logger: mockLogger, TestingTB: GinkgoTB()})
			mockLogger.EXPECT().Debug(gomock.Any(), gomock.Any()).AnyTimes()

			definitions, err := filesystem.LoadWorkspaceFlowFiles(workspaceCfg)
			Expect(err).NotTo(HaveOccurred())
			Expect(definitions).To(HaveLen(1))
			Expect(definitions[0].Namespace).To(Equal(executableDefinition.Namespace))
		})

		It("does not load executable definitions from excluded paths", func() {
			executableDefinition := &executable.FlowFile{
				Namespace: "test",
				Executables: executable.ExecutableList{
					{
						Verb: "exec",
						Name: "test-executable",
					},
				},
			}

			excludedDir, err := os.MkdirTemp(tmpDir, "excluded")
			Expect(err).NotTo(HaveOccurred())

			definitionFile := filepath.Join(excludedDir, "test"+executable.FlowFileExt)
			Expect(filesystem.WriteFlowFile(definitionFile, executableDefinition)).To(Succeed())

			workspaceCfg := &workspace.Workspace{
				Executables: &workspace.ExecutableFilter{
					Included: []string{tmpDir},
					Excluded: []string{excludedDir},
				},
			}
			workspaceCfg.SetContext("test", tmpDir)

			ctrl := gomock.NewController(GinkgoT())
			mockLogger := mocks.NewMockLogger(ctrl)
			logger.Init(logger.InitOptions{Logger: mockLogger, TestingTB: GinkgoTB()})
			mockLogger.EXPECT().Debug(gomock.Any(), gomock.Any()).AnyTimes()

			definitions, err := filesystem.LoadWorkspaceFlowFiles(workspaceCfg)
			Expect(err).NotTo(HaveOccurred())
			Expect(definitions).To(BeEmpty())
		})

		Context("with a nested workspace", func() {
			// nestedSetup writes one flow file at the workspace root and another inside a
			// subdirectory that is itself a workspace root (a git worktree, say).
			nestedSetup := func() string {
				def := &executable.FlowFile{
					Namespace:   "test",
					Executables: executable.ExecutableList{{Verb: "exec", Name: "test-executable"}},
				}
				Expect(filesystem.WriteFlowFile(filepath.Join(tmpDir, "root"+executable.FlowFileExt), def)).To(Succeed())

				nested := filepath.Join(tmpDir, "worktree")
				Expect(os.MkdirAll(nested, 0750)).To(Succeed())
				Expect(os.WriteFile(
					filepath.Join(nested, filesystem.WorkspaceConfigFileName), []byte("{}\n"), 0600,
				)).To(Succeed())
				Expect(filesystem.WriteFlowFile(filepath.Join(nested, "nested"+executable.FlowFileExt), def)).To(Succeed())

				ctrl := gomock.NewController(GinkgoT())
				mockLogger := mocks.NewMockLogger(ctrl)
				logger.Init(logger.InitOptions{Logger: mockLogger, TestingTB: GinkgoTB()})
				mockLogger.EXPECT().Debug(gomock.Any(), gomock.Any()).AnyTimes()
				return nested
			}

			It("does not load executables from a subdirectory that is its own workspace", func() {
				nestedSetup()
				workspaceCfg := &workspace.Workspace{}
				workspaceCfg.SetContext("test", tmpDir)

				definitions, err := filesystem.LoadWorkspaceFlowFiles(workspaceCfg)
				Expect(err).NotTo(HaveOccurred())
				Expect(definitions).To(HaveLen(1))
				Expect(definitions[0].ConfigPath()).To(Equal(filepath.Join(tmpDir, "root"+executable.FlowFileExt)))
			})

			It("still loads them when the nested path is explicitly included", func() {
				nested := nestedSetup()
				workspaceCfg := &workspace.Workspace{
					Executables: &workspace.ExecutableFilter{Included: []string{tmpDir, nested}},
				}
				workspaceCfg.SetContext("test", tmpDir)

				definitions, err := filesystem.LoadWorkspaceFlowFiles(workspaceCfg)
				Expect(err).NotTo(HaveOccurred())
				Expect(definitions).To(HaveLen(2))
			})
		})
	})
})
