package filesystem_test

import (
	"os"
	"path/filepath"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"

	"github.com/flowexec/flow/pkg/filesystem"
	"github.com/flowexec/flow/types/executable"
	"github.com/flowexec/flow/types/workspace"
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
