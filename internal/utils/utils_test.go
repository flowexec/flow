package utils_test

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/flowexec/tuikit/io/mocks"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/mock/gomock"

	"github.com/flowexec/flow/internal/utils"
	"github.com/flowexec/flow/pkg/logger"
)

func TestUtils(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Utils Suite")
}

var _ = Describe("Utils", func() {
	var (
		testHomeDir = ap("/Users/testuser")
		wsDir       = ap("/workspace")
		execDir     = ap("/execPath")

		mockLogger        *mocks.MockLogger
		testWorkingDir, _ = os.UserConfigDir()
		execPath          = filepath.Join(execDir, "exec.flow")
	)

	type testObj struct{}

	BeforeEach(func() {
		mockLogger = mocks.NewMockLogger(gomock.NewController(GinkgoT()))
		logger.Init(logger.InitOptions{Logger: mockLogger, TestingTB: GinkgoTB()})
		Expect(os.Chdir(testWorkingDir)).To(Succeed())
		Expect(os.Setenv("HOME", testHomeDir)).To(Succeed())
		Expect(os.Setenv("USERPROFILE", testHomeDir)).To(Succeed())
	})

	Describe("ExpandDirectory", func() {
		DescribeTable("with different inputs",
			func(dir string, expected string) {
				returnedDir := utils.ExpandDirectory(dir, wsDir, execPath, nil)
				Expect(returnedDir).To(Equal(expected))
			},
			Entry("empty dir", "", execDir),
			Entry("dir starts with //", "//dir", filepath.Join(wsDir, "dir")),
			Entry("dir is .", ".", testWorkingDir),
			Entry("dir starts with ./", "./dir", filepath.Join(testWorkingDir, "dir")),
			Entry("dir starts with ~/", "~/dir", filepath.Join(testHomeDir, "dir")),
			Entry("dir starts with /", ap("/dir"), ap("/dir")),
			Entry("default case", "dir", filepath.Join(execDir, "dir")),
			Entry("hidden dir with extension-like name", ap("/path/.config"), ap("/path/.config")),
			Entry("file with extension returns parent dir", ap("/path/file.txt"), ap("/path")),
		)

		When("env vars are in the dir", func() {
			It("expands the env vars", func() {
				envMap := map[string]string{"VAR1": "one", "VAR2": "two"}
				Expect(utils.ExpandDirectory(ap("/${VAR1}/${VAR2}"), wsDir, execPath, envMap)).
					To(Equal(ap("/one/two")))
			})
			It("expands the env vars with a ws prefix", func() {
				envMap := map[string]string{"VAR1": "one"}
				Expect(utils.ExpandDirectory("//dir/${VAR1}", wsDir, execPath, envMap)).
					To(Equal(filepath.Join(wsDir, "dir", "one")))
			})
			It("logs a warning if the env var is not found", func() {
				envMap := map[string]string{"VAR1": "one"}
				mockLogger.EXPECT().Warn("unable to find env key in path expansion", "key", "VAR2")
				Expect(utils.ExpandDirectory(ap("/${VAR1}/${VAR2}"), wsDir, execPath, envMap)).
					To(Equal(ap("/one")))
			})
		})
	})

	Describe("PathFromWd", func() {
		When("path is a subdirectory", func() {
			It("returns the relative path", func() {
				result, err := utils.PathFromWd(filepath.Join(testWorkingDir, "subdir"))
				Expect(result).To(Equal("subdir"))
				Expect(err).ToNot(HaveOccurred())
			})
		})
		When("path is a parent directory", func() {
			It("returns the relative path", func() {
				result, err := utils.PathFromWd(testWorkingDir)
				Expect(result).To(Equal("."))
				Expect(err).ToNot(HaveOccurred())
			})
		})
		When("path is a sibling directory", func() {
			It("returns the relative path", func() {
				result, err := utils.PathFromWd(filepath.Join(filepath.Dir(testWorkingDir), "sibling"))
				Expect(result).To(Equal(filepath.Join("..", "sibling")))
				Expect(err).ToNot(HaveOccurred())
			})
		})
	})

	Describe("ValidateOneOf", func() {
		DescribeTable("with different inputs",
			func(fieldName string, vals []interface{}, expectedErr string) {
				err := utils.ValidateOneOf(fieldName, vals...)
				if expectedErr != "" {
					Expect(err.Error()).To(ContainSubstring(expectedErr))
				} else {
					Expect(err).ToNot(HaveOccurred())
				}
			},
			Entry(
				"no values",
				"fieldName",
				[]interface{}{},
				"must define at least one fieldName",
			),
			Entry(
				"one value",
				"fieldName",
				[]interface{}{"value"},
				nil,
			),
			Entry(
				"one value with nils",
				"fieldName",
				[]interface{}{nil, "value", nil},
				nil,
			),
			Entry(
				"pointer value",
				"fieldName",
				[]interface{}{&testObj{}},
				nil,
			),
			Entry(
				"more than one value",
				"fieldName",
				[]interface{}{"value1", "value2"},
				"must define only one fieldName",
			),
		)
	})
})

// vol is the volume name of the working directory on Windows (e.g. "C:").
// Empty on all other platforms.
var vol = func() string {
	if runtime.GOOS == "windows" {
		wd, _ := os.Getwd()
		return filepath.VolumeName(wd)
	}
	return ""
}()

// ap (absolute path) converts a POSIX-style absolute path such as "/foo/bar"
// into a native absolute path. On Windows it prepends the current volume so
// that filepath.IsAbs returns true; on other platforms it is a no-op.
func ap(p string) string {
	return filepath.FromSlash(vol + p)
}
