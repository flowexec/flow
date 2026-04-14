package render_test

import (
	stdCtx "context"
	"os"
	"path/filepath"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/mock/gomock"

	flowIO "github.com/flowexec/flow/internal/io"
	"github.com/flowexec/flow/internal/runner"
	"github.com/flowexec/flow/internal/runner/engine/mocks"
	"github.com/flowexec/flow/internal/runner/render"
	testUtils "github.com/flowexec/flow/tests/utils"
	"github.com/flowexec/flow/types/executable"
)

func TestRender(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Render Suite")
}

var _ = Describe("Render Runner", func() {
	var (
		renderRnr  runner.Runner
		ctx        *testUtils.ContextWithMocks
		mockEngine *mocks.MockEngine
	)

	BeforeEach(func() {
		// Ensure DISABLE_FLOW_INTERACTIVE is not leaked in from the harness.
		GinkgoTB().Setenv(flowIO.DisableInteractiveEnvKey, "")
		ctx = testUtils.NewContextWithMocks(stdCtx.Background(), GinkgoTB())
		renderRnr = render.NewRunner()
		mockEngine = mocks.NewMockEngine(gomock.NewController(GinkgoT()))
	})

	Context("Name", func() {
		It("returns 'render'", func() {
			Expect(renderRnr.Name()).To(Equal("render"))
		})
	})

	Context("IsCompatible", func() {
		It("is false when executable is nil", func() {
			Expect(renderRnr.IsCompatible(nil)).To(BeFalse())
		})
		It("is false when Render spec is nil", func() {
			Expect(renderRnr.IsCompatible(&executable.Executable{})).To(BeFalse())
		})
		It("is true when Render spec is set", func() {
			e := &executable.Executable{Render: &executable.RenderExecutableType{}}
			Expect(renderRnr.IsCompatible(e)).To(BeTrue())
		})
	})

	Describe("Exec (plain mode)", func() {
		var wsDir string

		BeforeEach(func() {
			wsDir = ctx.Ctx.CurrentWorkspace.Location()
		})

		writeFile := func(name, content string) string {
			path := filepath.Join(wsDir, name)
			Expect(os.WriteFile(path, []byte(content), 0o600)).To(Succeed())
			return path
		}

		newExec := func(spec *executable.RenderExecutableType) *executable.Executable {
			e := &executable.Executable{Render: spec}
			e.SetContext(ctx.Ctx.CurrentWorkspace.AssignedName(), wsDir, "", filepath.Join(wsDir, "examples.flow"))
			return e
		}

		It("emits rendered content bracketed by parseable markers", func() {
			writeFile("tmpl.md", "# Hello\n\nworld\n")
			e := newExec(&executable.RenderExecutableType{TemplateFile: "tmpl.md"})

			// Template-parse info log.
			ctx.Logger.EXPECT().Infof(gomock.Any(), gomock.Any()).Times(1)
			// Begin marker + rendered content + end marker, in order.
			gomock.InOrder(
				ctx.Logger.EXPECT().Print(gomock.Regex("^"+regexEscape(render.PlainBeginMarker)+" file=tmpl.md$")),
				ctx.Logger.EXPECT().Print(gomock.Eq("# Hello\n\nworld\n")),
				ctx.Logger.EXPECT().Print(gomock.Eq(render.PlainEndMarker)),
			)

			Expect(renderRnr.Exec(ctx.Ctx, e, mockEngine, map[string]string{}, nil)).To(Succeed())
		})

		It("substitutes data from a JSON template data file", func() {
			writeFile("tmpl.md", `Name: {{ data["name"] }}`)
			writeFile("data.json", `{"name":"flow"}`)
			e := newExec(&executable.RenderExecutableType{
				TemplateFile:     "tmpl.md",
				TemplateDataFile: "data.json",
			})

			ctx.Logger.EXPECT().Infof(gomock.Any(), gomock.Any()).Times(1)
			gomock.InOrder(
				ctx.Logger.EXPECT().Print(gomock.Regex(regexEscape(render.PlainBeginMarker))),
				ctx.Logger.EXPECT().Print(gomock.Eq("Name: flow")),
				ctx.Logger.EXPECT().Print(gomock.Eq(render.PlainEndMarker)),
			)

			Expect(renderRnr.Exec(ctx.Ctx, e, mockEngine, map[string]string{}, nil)).To(Succeed())
		})

		It("substitutes data from a YAML template data file", func() {
			writeFile("tmpl.md", `Env: {{ data["env"] }}`)
			writeFile("data.yaml", "env: prod\n")
			e := newExec(&executable.RenderExecutableType{
				TemplateFile:     "tmpl.md",
				TemplateDataFile: "data.yaml",
			})

			ctx.Logger.EXPECT().Infof(gomock.Any(), gomock.Any()).Times(1)
			gomock.InOrder(
				ctx.Logger.EXPECT().Print(gomock.Regex(regexEscape(render.PlainBeginMarker))),
				ctx.Logger.EXPECT().Print(gomock.Eq("Env: prod")),
				ctx.Logger.EXPECT().Print(gomock.Eq(render.PlainEndMarker)),
			)

			Expect(renderRnr.Exec(ctx.Ctx, e, mockEngine, map[string]string{}, nil)).To(Succeed())
		})

		It("returns an error when the template data file is missing", func() {
			writeFile("tmpl.md", "x")
			e := newExec(&executable.RenderExecutableType{
				TemplateFile:     "tmpl.md",
				TemplateDataFile: "does-not-exist.json",
			})

			err := renderRnr.Exec(ctx.Ctx, e, mockEngine, map[string]string{}, nil)
			Expect(err).To(MatchError(ContainSubstring("does not exist")))
		})

		It("returns an error when the template file is missing", func() {
			e := newExec(&executable.RenderExecutableType{TemplateFile: "no-such.md"})

			ctx.Logger.EXPECT().Infof(gomock.Any(), gomock.Any()).AnyTimes()
			err := renderRnr.Exec(ctx.Ctx, e, mockEngine, map[string]string{}, nil)
			Expect(err).To(HaveOccurred())
		})

		It("uses plain mode even with TUI enabled when DISABLE_FLOW_INTERACTIVE is set", func() {
			GinkgoTB().Setenv(flowIO.DisableInteractiveEnvKey, "true")
			// Force config TUI on so the only reason plain mode triggers is the env var.
			ctx.Ctx.Config.Interactive.Enabled = true

			writeFile("tmpl.md", "hello")
			e := newExec(&executable.RenderExecutableType{TemplateFile: "tmpl.md"})

			ctx.Logger.EXPECT().Infof(gomock.Any(), gomock.Any()).Times(1)
			gomock.InOrder(
				ctx.Logger.EXPECT().Print(gomock.Regex(regexEscape(render.PlainBeginMarker))),
				ctx.Logger.EXPECT().Print(gomock.Eq("hello")),
				ctx.Logger.EXPECT().Print(gomock.Eq(render.PlainEndMarker)),
			)

			Expect(renderRnr.Exec(ctx.Ctx, e, mockEngine, map[string]string{}, nil)).To(Succeed())
		})
	})
})

// regexEscape escapes regex metacharacters in the marker constants for use
// inside gomock.Regex matchers. Accepts a string to keep callsites readable
// even though today it's only ever called with the begin marker.
func regexEscape(s string) string { //nolint:unparam
	var out []byte
	for _, r := range []byte(s) {
		switch r {
		case '\\', '.', '+', '*', '?', '(', ')', '|', '[', ']', '{', '}', '^', '$':
			out = append(out, '\\', r)
		default:
			out = append(out, r)
		}
	}
	return string(out)
}
