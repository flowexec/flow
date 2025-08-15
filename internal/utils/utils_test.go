package utils_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/flowexec/tuikit/io/mocks"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/mock/gomock"

	"github.com/flowexec/flow/internal/logger"
	"github.com/flowexec/flow/internal/utils"
)

func TestUtils(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Utils Suite")
}

var _ = Describe("Utils", func() {
	const (
		testHomeDir = "/Users/testuser"
		wsDir       = "/workspace"
		execDir     = "/execPath"
	)

	var (
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
			Entry("dir starts with /", "/dir", "/dir"),
			Entry("default case", "dir", filepath.Join(execDir, "dir")),
		)

		When("env vars are in the dir", func() {
			It("expands the env vars", func() {
				envMap := map[string]string{"VAR1": "one", "VAR2": "two"}
				Expect(utils.ExpandDirectory("/${VAR1}/${VAR2}", wsDir, execPath, envMap)).
					To(Equal("/one/two"))
			})
			It("logs a warning if the env var is not found", func() {
				envMap := map[string]string{"VAR1": "one"}
				mockLogger.EXPECT().Warnx("unable to find env key in path expansion", "key", "VAR2")
				Expect(utils.ExpandDirectory("/${VAR1}/${VAR2}", wsDir, execPath, envMap)).
					To(Equal("/one"))
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
				Expect(result).To(Equal("../sibling"))
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
