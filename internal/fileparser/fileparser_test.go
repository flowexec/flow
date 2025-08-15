package fileparser_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/flowexec/tuikit/io/mocks"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/mock/gomock"

	"github.com/flowexec/flow/internal/fileparser"
	"github.com/flowexec/flow/internal/logger"
	"github.com/flowexec/flow/types/executable"
)

func TestFileParser(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "FileParser Suite")
}

var _ = Describe("ExecutablesFromImports", func() {
	var (
		ctrl       *gomock.Controller
		mockLogger *mocks.MockLogger
		flowFile   *executable.FlowFile
	)
	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		mockLogger = mocks.NewMockLogger(ctrl)
		logger.Init(logger.InitOptions{Logger: mockLogger, TestingTB: GinkgoTB()})
		mockLogger.EXPECT().Debugf(gomock.Any(), gomock.Any()).AnyTimes()

		wd, err := os.Getwd()
		Expect(err).ToNot(HaveOccurred())

		ff := filepath.Join(wd, "testdata", "test"+executable.FlowFileExt)
		flowFile = &executable.FlowFile{Imports: make(executable.FromFile, 0)}
		flowFile.SetContext("ws", filepath.Join(wd, "testdata"), ff)
	})

	It("should return executables from imports", func() {
		flowFile.Imports = append(
			flowFile.Imports,
			"Makefile",
			"package.json",
			"docker-compose.yml",
			"complex.sh",
		)

		result, err := fileparser.ExecutablesFromImports("ws", flowFile)
		Expect(err).NotTo(HaveOccurred())
		Expect(len(result)).To(BeNumerically(">", 10))

		for _, e := range result {
			Expect(e.Exec).ToNot(BeNil())
			Expect(e.Exec.Dir).To(Equal(executable.Directory("//")))
		}
	})

	It("should log a warning for invalid file type", func() {
		mockLogger.EXPECT().Warnx(gomock.Any(), "file", "invalidfile").AnyTimes()
		flowFile.Imports = append(flowFile.Imports, "invalidfile")
		result, err := fileparser.ExecutablesFromImports("ws", flowFile)
		Expect(err).NotTo(HaveOccurred())
		Expect(result).To(BeEmpty())
	})

	It("should log an error for dir instead of file", func() {
		mockLogger.EXPECT().Errorx(gomock.Any(), "err", "invaliddir is not a file").AnyTimes()
		flowFile.Imports = append(flowFile.Imports, "invaliddir")
		result, err := fileparser.ExecutablesFromImports("ws", flowFile)
		Expect(err).NotTo(HaveOccurred())
		Expect(result).To(BeEmpty())
	})

	It("should log an error for non-existent file", func() {
		mockLogger.EXPECT().Error(gomock.Any(), gomock.Any()).AnyTimes()
		flowFile.Imports = append(flowFile.Imports, "nonexistent.sh")
		result, err := fileparser.ExecutablesFromImports("ws", flowFile)
		Expect(err).NotTo(HaveOccurred())
		Expect(result).To(BeEmpty())
	})

	It("should log an error when configuration key is not recognized", func() {
		mockLogger.EXPECT().Error(gomock.Any(), gomock.Any()).AnyTimes()
		flowFile.Imports = append(flowFile.Imports, "unknownkey.sh")
		result, err := fileparser.ExecutablesFromImports("ws", flowFile)
		Expect(err).ToNot(HaveOccurred())
		Expect(result).ToNot(BeNil())
	})
})
