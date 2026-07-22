package imports_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/flowexec/tuikit/io/mocks"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/mock/gomock"

	"github.com/flowexec/flow/v2/pkg/imports"
	"github.com/flowexec/flow/v2/pkg/logger"
	"github.com/flowexec/flow/v2/types/executable"
)

func TestImports(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Imports Suite")
}

var _ = Describe("ExecutablesFromImports", func() {
	var (
		ctrl       *gomock.Controller
		mockLogger *mocks.MockLogger
		dir        string
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		mockLogger = mocks.NewMockLogger(ctrl)
		logger.Init(logger.InitOptions{Logger: mockLogger, TestingTB: GinkgoTB()})
		mockLogger.EXPECT().Debugf(gomock.Any(), gomock.Any()).AnyTimes()

		dir = GinkgoT().TempDir()
		Expect(os.WriteFile(filepath.Join(dir, "Makefile"), []byte("build:\n\techo build\n"), 0o600)).To(Succeed())
	})

	It("generates executables from an in-memory flow file's imports", func() {
		ff := &executable.FlowFile{Imports: executable.Imports{"Makefile"}}
		ff.SetContext("ws", dir, filepath.Join(dir, "virtual"+executable.FlowFileExt))

		result, err := imports.ExecutablesFromImports("ws", ff)
		Expect(err).ToNot(HaveOccurred())
		Expect(len(result)).To(BeNumerically(">", 0))
		for _, e := range result {
			Expect(e.Exec).ToNot(BeNil())
		}
	})
})
