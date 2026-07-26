package filesystem_test

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/flowexec/tuikit/io/mocks"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"
	"go.uber.org/mock/gomock"
	"gopkg.in/yaml.v3"

	"github.com/flowexec/flow/v2/pkg/filesystem"
	"github.com/flowexec/flow/v2/pkg/logger"
	"github.com/flowexec/flow/v2/types/executable"
	"github.com/flowexec/flow/v2/types/workspace"
)

var _ = Describe("Templates", func() {
	var (
		tmpDir string
	)

	BeforeEach(func() {
		var err error
		tmpDir, err = os.MkdirTemp("", "flow-templates-test")
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		Expect(os.RemoveAll(tmpDir)).To(Succeed())
	})

	Describe("WriteFlowFileTemplate", func() {
		It("writes the flowfile successfully", func() {
			ff := &executable.FlowFile{
				Namespace: "test",
				Executables: executable.ExecutableList{
					{Verb: "run", Name: "test-executable", Description: "{{ .key }}"},
				},
			}
			ffStr, err := ff.YAML()
			Expect(err).NotTo(HaveOccurred())
			template := &executable.Template{
				Template: ffStr,
				Form: executable.FormFields{
					{Key: "key", Prompt: "enter key", Default: "value"},
				},
			}
			templatePath := templateFullPath(tmpDir, "test")
			template.SetContext("test", templatePath)

			workspaceConfig := workspace.DefaultWorkspaceConfig("test")
			workspaceConfig.SetContext("test", tmpDir)

			err = WriteFlowFileTemplate(template.Location(), template)
			Expect(err).ToNot(HaveOccurred())
			_, err = os.Stat(filepath.Join(tmpDir, "test"+executable.FlowFileTemplateExt))
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Describe("LoadFlowFileTemplate", func() {
		It("loads the template correctly", func() {
			ff := &executable.FlowFile{
				Namespace: "test",
				Executables: executable.ExecutableList{
					{Verb: "run", Name: "test-executable", Description: "{{ .key }}"},
				},
			}
			ffStr, err := ff.YAML()
			Expect(err).NotTo(HaveOccurred())
			template := &executable.Template{
				Template: ffStr,
				Form: executable.FormFields{
					{Key: "key", Prompt: "enter key", Default: "value"},
				},
			}
			templatePath := templateFullPath(tmpDir, "test")
			template.SetContext("test", templatePath)
			Expect(WriteFlowFileTemplate(templatePath, template)).To(Succeed())

			readTemplate, err := filesystem.LoadFlowFileTemplate("test", templatePath)
			Expect(err).NotTo(HaveOccurred())
			Expect(readTemplate).To(Equal(template))
			Expect(readTemplate.Location()).To(Equal(templatePath))
			Expect(readTemplate.Name()).To(Equal("test"))
		})
	})

	Describe("LoadWorkspaceFlowFileTemplates", func() {
		BeforeEach(func() {
			mockLogger := mocks.NewMockLogger(gomock.NewController(GinkgoT()))
			logger.Init(logger.InitOptions{Logger: mockLogger, TestingTB: GinkgoTB()})
			mockLogger.EXPECT().Debug(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
			mockLogger.EXPECT().Error(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
		})

		write := func(name, content string) {
			Expect(os.WriteFile(filepath.Join(tmpDir, name), []byte(content), 0600)).To(Succeed())
		}

		It("discovers only *.flow.tmpl files, naming them from their filename", func() {
			write("webapp.flow.tmpl", "template: |\n  namespace: test\n")
			write("api.flow.tmpl.yaml", "template: |\n  namespace: test\n")
			write("regular.flow", "namespace: test\n")          // executable, not a template
			write("deployment.yaml.tmpl", "kind: Deployment\n") // artifact partial, not a template
			write("notes.txt", "hi\n")                          // unrelated

			workspaceCfg := &workspace.Workspace{}
			workspaceCfg.SetContext("test", tmpDir)

			templates, err := filesystem.LoadWorkspaceFlowFileTemplates(workspaceCfg)
			Expect(err).NotTo(HaveOccurred())
			names := make([]string, 0, len(templates))
			for _, t := range templates {
				names = append(names, t.Name())
			}
			Expect(names).To(ConsistOf("webapp", "api"))
		})

		It("respects excluded paths", func() {
			excludedDir, err := os.MkdirTemp(tmpDir, "excluded")
			Expect(err).NotTo(HaveOccurred())
			Expect(os.WriteFile(
				filepath.Join(excludedDir, "hidden.flow.tmpl"), []byte("template: |\n  namespace: test\n"), 0600,
			)).To(Succeed())
			write("visible.flow.tmpl", "template: |\n  namespace: test\n")

			workspaceCfg := &workspace.Workspace{
				Templates: &workspace.ExecutableFilter{
					Included: []string{tmpDir},
					Excluded: []string{excludedDir},
				},
			}
			workspaceCfg.SetContext("test", tmpDir)

			templates, err := filesystem.LoadWorkspaceFlowFileTemplates(workspaceCfg)
			Expect(err).NotTo(HaveOccurred())
			Expect(templates).To(HaveLen(1))
			Expect(templates[0].Name()).To(Equal("visible"))
		})
	})
})

func WriteFlowFileTemplate(templateFilePath string, template *executable.Template) error {
	file, err := os.Create(filepath.Clean(templateFilePath))
	if err != nil {
		return errors.Wrap(err, "unable to create template file")
	}
	defer file.Close()

	if err := yaml.NewEncoder(file).Encode(template); err != nil {
		return errors.Wrap(err, "unable to encode template file")
	}
	return nil
}

func templateFullPath(templateDir, templateName string) string {
	templatePath := filepath.Join(templateDir, templateName)
	if strings.HasSuffix(templateName, executable.FlowFileTemplateExt) {
		return templatePath
	} else if strings.HasSuffix(templatePath, executable.FlowFileExt) {
		return strings.TrimSuffix(templatePath, executable.FlowFileExt) + executable.FlowFileTemplateExt
	}
	return templatePath + executable.FlowFileTemplateExt
}
