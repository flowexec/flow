//nolint:testpackage
package context

import (
	"os"
	"path/filepath"
	"sync"
	"testing"

	"charm.land/lipgloss/v2"
	"github.com/flowexec/tuikit/themes"
	"github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/flowexec/flow/v2/types/config"
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
			// On Windows the process cannot delete a directory it is cwd'd into,
			// so navigate away first.
			_ = os.Chdir(os.TempDir())
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

func TestCallbacksSurviveShallowCopy(t *testing.T) {
	t.Run("a copy's callbacks reach the original", func(t *testing.T) {
		// A parallel branch runs on a shallow copy. Cleanup it registers — temporary env
		// files, for one — used to land on the copy and never run, because only the root
		// context is ever finalized.
		root := &Context{callbacks: &callbackList{}}
		var ran []string

		root.AddCallback(func(*Context) error { ran = append(ran, "root"); return nil })
		branch := root.ShallowCopy()
		branch.AddCallback(func(*Context) error { ran = append(ran, "branch"); return nil })

		for _, cb := range root.callbacks.drain() {
			_ = cb(root)
		}

		if len(ran) != 2 {
			t.Fatalf("expected both callbacks to run, got %v", ran)
		}
	})

	t.Run("draining twice does not run cleanup twice", func(t *testing.T) {
		root := &Context{callbacks: &callbackList{}}
		count := 0
		root.AddCallback(func(*Context) error { count++; return nil })

		for range 2 {
			for _, cb := range root.callbacks.drain() {
				_ = cb(root)
			}
		}

		if count != 1 {
			t.Errorf("expected the callback to run once, ran %d times", count)
		}
	})

	t.Run("concurrent branches can register at once", func(t *testing.T) {
		// Run with -race: parallel branches register from their own goroutines.
		root := &Context{callbacks: &callbackList{}}
		var wg sync.WaitGroup
		for range 32 {
			wg.Add(1)
			go func() {
				defer wg.Done()
				root.ShallowCopy().AddCallback(func(*Context) error { return nil })
			}()
		}
		wg.Wait()

		if got := len(root.callbacks.drain()); got != 32 {
			t.Errorf("expected 32 callbacks, got %d", got)
		}
	})
}
