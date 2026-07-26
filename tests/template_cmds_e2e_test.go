//go:build e2e

package tests_test

import (
	stdCtx "context"
	"fmt"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/flowexec/flow/v2/tests/utils"
	"github.com/flowexec/flow/v2/types/executable"
)

var _ = Describe("flowfile template commands e2e", Ordered, func() {
	var (
		ctx              *utils.Context
		run              *utils.CommandRunner
		template         *executable.Template
		expectedFlowFile *executable.FlowFile
	)

	BeforeAll(func() {
		ctx = utils.NewContext(stdCtx.Background(), GinkgoTB())
		run = utils.NewE2ECommandRunner()
		workDir, err := os.MkdirTemp("", "flowfile-template-e2e")
		Expect(err).NotTo(HaveOccurred())
		tmpl := executable.FlowFile{
			Namespace:   "test",
			Description: "Template test flowfile",
			Tags:        []string{"test"},
			Executables: []*executable.Executable{
				{
					Verb: "exec",
					Name: "{{ name }}",
					Exec: &executable.ExecExecutableType{Cmd: fmt.Sprintf("echo '%s'", "{{ form['Msg'] }}")}},
			},
		}
		tmplStr, err := tmpl.YAML()
		Expect(err).NotTo(HaveOccurred())
		template = &executable.Template{
			Template: tmplStr,
			Form: executable.FormFields{
				&executable.Field{
					Key:     "Name",
					Prompt:  "Enter a name",
					Default: "test",
				},
				&executable.Field{
					Key:     "Msg",
					Prompt:  "Enter a message",
					Default: "Hello, world!",
				},
			},
			Artifacts: []executable.Artifact{
				{SrcName: "artifact1"},
				{SrcName: "artifact2", DstName: "artifact2-renamed"},
			},
			PreRun: []executable.TemplateRefConfig{
				{Ref: "exec examples:simple-print", Args: []string{"test"}},
			},
			PostRun: []executable.TemplateRefConfig{
				{
					Cmd: "touch {{ name }}",
				},
			},
		}
		template.SetContext("e2e", filepath.Join(workDir, "flowfile.flow.tmpl"))
		data, err := template.YAML()
		Expect(err).NotTo(HaveOccurred())
		Expect(os.WriteFile(filepath.Join(workDir, "flowfile.flow.tmpl"), []byte(data), 0644)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(workDir, "artifact1"), []byte("artifact1"), 0644)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(workDir, "artifact2"), []byte("artifact2"), 0644)).To(Succeed())

		expectedFlowFile = &executable.FlowFile{
			Namespace:   "test",
			Description: "Template test flowfile",
			Tags:        []string{"test"},
			Executables: []*executable.Executable{
				{
					Verb: "exec",
					Name: "test",
					Exec: &executable.ExecExecutableType{Cmd: "echo 'Hello, world!'"}},
			},
		}
		expectedFlowFile.SetContext(
			ctx.CurrentWorkspace.AssignedName(),
			ctx.CurrentWorkspace.Location(),
			filepath.Join(workDir, "flowfile.flow"),
		)
	})

	BeforeEach(func() {
		utils.ResetTestContext(ctx, GinkgoTB())
	})

	AfterEach(func() {
		ctx.Finalize()
	})

	When("registering a new template (flow template add)", func() {
		It("should complete successfully", func() {
			stdOut := ctx.StdOut()
			err := run.Run(ctx.Context, "template", "add", template.Name(), template.Location())
			Expect(err).ToNot(HaveOccurred())
			out, err := readFileContent(stdOut)
			Expect(err).NotTo(HaveOccurred())
			Expect(out).To(ContainSubstring(fmt.Sprintf("Template %s set", template.Name())))
		})
	})

	When("getting a registered template (flow template get)", func() {
		It("should return the template", func() {
			stdOut := ctx.StdOut()
			err := run.Run(ctx.Context, "template", "get", "-t", template.Name(), "-o", "yaml")
			Expect(err).ToNot(HaveOccurred())
			out, err := readFileContent(stdOut)
			Expect(err).NotTo(HaveOccurred())
			str, err := template.YAML()
			Expect(err).NotTo(HaveOccurred())
			Expect(out).To(ContainSubstring(str))
		})
	})

	When("getting a template by path (flow template get)", func() {
		It("should return the template, named after the file", func() {
			stdOut := ctx.StdOut()
			err := run.Run(ctx.Context, "template", "get", "-f", template.Location(), "-o", "yaml")
			Expect(err).ToNot(HaveOccurred())
			out, err := readFileContent(stdOut)
			Expect(err).NotTo(HaveOccurred())

			// A template loaded by path has no registered name, so its name is
			// derived from the filename ("flowfile") rather than the name it was
			// registered under ("e2e"). Compare against a copy carrying that context.
			byPath := *template
			byPath.SetContext("", template.Location())
			str, err := byPath.YAML()
			Expect(err).NotTo(HaveOccurred())
			Expect(out).To(ContainSubstring(str))
			Expect(byPath.Name()).To(Equal("flowfile"))
		})
	})

	When("Listing all registered templates (flow template list)", func() {
		It("should return the list of templates", func() {
			stdOut := ctx.StdOut()
			Expect(run.Run(ctx.Context, "template", "list", "-o", "yaml")).To(Succeed())
			out, err := readFileContent(stdOut)
			Expect(err).NotTo(HaveOccurred())
			// tabs may be present so instead of checking for exact match, we check for length
			str, err := template.YAML()
			Expect(err).NotTo(HaveOccurred())
			Expect(len(out)).To(BeNumerically(">", len(str)))
		})
	})

	When("getting a template that is not registered (flow template get)", func() {
		It("fatals with an unable-to-load message", func() {
			ctx.ExpectFailure()
			stdOut := ctx.StdOut()
			err := run.Run(ctx.Context, "template", "get", "-t", "doesnotexist", "-o", "yaml")
			Expect(err).To(HaveOccurred())
			out, readErr := readFileContent(stdOut)
			Expect(readErr).NotTo(HaveOccurred())
			Expect(out).To(ContainSubstring("unable to load flowfile template"))
		})
	})

	When("Rendering a template (flow template generate)", func() {
		It("should process the template options and render the flowfile", func() {
			name := "test"
			outputDir := filepath.Join(ctx.CurrentWorkspace.Location(), "output")

			Expect(run.Run(
				ctx.Context,
				"template",
				"generate",
				name,
				"-t", template.Name(),
				"-o", outputDir,
				"-s", "name=test",
				"-s", "msg=hello",
			)).To(Succeed())
			out, err := readFileContent(ctx.StdOut())
			Expect(err).NotTo(HaveOccurred())
			Expect(out).To(ContainSubstring(fmt.Sprintf("Template '%s' rendered successfully", name)))
		})
	})

	When("an unregistered template exists in the workspace (auto-discovery)", func() {
		const discoveredTmpl = `form:
  - key: greeting
    prompt: "Greeting?"
    default: "hello"
template: |
  namespace: discovered
  executables:
    - verb: run
      name: "{{ form["greeting"] }}"
      exec:
        cmd: echo hi
`

		BeforeAll(func() {
			// The e2e harness closes the shared IO after each command, so the file is written
			// here (no command) and sync runs as its own spec below. The DataStore persists
			// across specs, so the synced cache is visible to the later list/generate specs.
			tmplPath := filepath.Join(ctx.CurrentWorkspace.Location(), "discovered.flow.tmpl")
			Expect(os.WriteFile(tmplPath, []byte(discoveredTmpl), 0600)).To(Succeed())
		})

		It("discovers the template on sync", func() {
			Expect(run.Run(ctx.Context, "sync")).To(Succeed())
		})

		It("lists the discovered template without registration", func() {
			Expect(run.Run(ctx.Context, "template", "list", "-o", "yaml")).To(Succeed())
			out, err := readFileContent(ctx.StdOut())
			Expect(err).NotTo(HaveOccurred())
			Expect(out).To(ContainSubstring("discovered"))
		})

		It("renders the discovered template by name", func() {
			outputDir := filepath.Join(ctx.CurrentWorkspace.Location(), "discovered-output")
			Expect(run.Run(
				ctx.Context,
				"template",
				"generate",
				"discovered",
				"-t", "discovered",
				"-o", outputDir,
				"-s", "greeting=hi",
			)).To(Succeed())
			out, err := readFileContent(ctx.StdOut())
			Expect(err).NotTo(HaveOccurred())
			Expect(out).To(ContainSubstring("rendered successfully"))
		})
	})
})
