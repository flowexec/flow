//nolint:testpackage
package context

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/flowexec/tuikit/themes"
	"github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/flowexec/flow/types/config"
)

func TestContext(t *testing.T) {
	RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "Context Suite")
}

var _ = ginkgo.Describe("Context", func() {
	ginkgo.Describe("currentWorkspace", func() {
		var (
			cfg    *config.Config
			tmpDir string
		)

		ginkgo.BeforeEach(func() {
			tmpDir = ginkgo.GinkgoT().TempDir()
			cfg = &config.Config{
				Workspaces: map[string]string{
					"ws1": filepath.Clean(filepath.Join(tmpDir, "ws1")),
					"ws2": filepath.Clean(filepath.Join(tmpDir, "ws2")),
				},
				CurrentWorkspace: "ws1",
				WorkspaceMode:    config.ConfigWorkspaceModeFixed,
			}
		})

		ginkgo.AfterEach(func() {
			_ = os.RemoveAll(tmpDir)
		})

		ginkgo.It("should return the current workspace in fixed mode", func() {
			ws, err := currentWorkspace(cfg)
			Expect(err).NotTo(HaveOccurred())
			Expect(ws.AssignedName()).To(Equal("ws1"))
			Expect(ws.Location()).To(Equal(filepath.Join(tmpDir, "ws1")))
		})

		ginkgo.It("should return the current workspace in dynamic mode", func() {
			cfg.WorkspaceMode = config.ConfigWorkspaceModeDynamic
			Expect(os.Mkdir(filepath.Join(tmpDir, "ws2"), 0750)).To(Succeed())
			// os.Setenv("PWD", filepath.Join(tmpDir, "ws2"))
			Expect(os.Chdir(filepath.Join(tmpDir, "ws2"))).To(Succeed())

			ws, err := currentWorkspace(cfg)
			Expect(err).NotTo(HaveOccurred())
			Expect(ws.AssignedName()).To(Equal("ws2"))
			Expect(ws.Location()).To(Equal(filepath.Join(tmpDir, "ws2")))
		})

		ginkgo.It("should return an error if the current workspace is not found", func() {
			cfg.CurrentWorkspace = "ws3"
			_, err := currentWorkspace(cfg)
			Expect(err).To(HaveOccurred())
		})
	})

	ginkgo.Describe("overrideThemeColor", func() {
		var theme themes.Theme
		var palette *config.ColorPalette

		ginkgo.BeforeEach(func() {
			theme = themes.NewTheme("theme", themes.ColorPalette{
				Primary:   "#000000",
				Secondary: "#FFFFFF",
			})
			palette = &config.ColorPalette{
				Primary:   strPtr("#FF0000"),
				Secondary: strPtr("#00FF00"),
			}
		})

		ginkgo.It("should override the theme colors with the palette colors", func() {
			newTheme := overrideThemeColor(theme, palette)
			Expect(newTheme.ColorPalette().PrimaryColor()).To(Equal(lipgloss.Color("#FF0000")))
			Expect(newTheme.ColorPalette().SecondaryColor()).To(Equal(lipgloss.Color("#00FF00")))
		})

		ginkgo.It("should not change the theme if the palette is nil", func() {
			newTheme := overrideThemeColor(theme, nil)
			Expect(newTheme).To(Equal(theme))
		})
	})
})

func strPtr(s string) *string {
	return &s
}
