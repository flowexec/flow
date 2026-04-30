package filesystem_test

import (
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/flowexec/flow/pkg/filesystem"
)

var _ = Describe("Logs", func() {
	var tmpDir string

	BeforeEach(func() {
		var err error
		tmpDir, err = os.MkdirTemp("", "flow-state-test")
		Expect(err).NotTo(HaveOccurred())
		Expect(os.Setenv(filesystem.FlowStateDirEnvVar, tmpDir)).To(Succeed())
	})

	AfterEach(func() {
		Expect(os.RemoveAll(tmpDir)).To(Succeed())
		Expect(os.Unsetenv(filesystem.FlowStateDirEnvVar)).To(Succeed())
	})

	Describe("StateDirPath", func() {
		It("returns the env override when set", func() {
			Expect(filesystem.StateDirPath()).To(Equal(tmpDir))
		})
	})

	Describe("LogsDir", func() {
		It("returns the correct logs directory path", func() {
			logsDir := filesystem.LogsDir()
			Expect(logsDir).To(Equal(filepath.Join(tmpDir, "logs")))
		})
	})

	Describe("EnsureLogsDir", func() {
		It("creates the logs directory if it does not exist", func() {
			Expect(filesystem.EnsureLogsDir()).To(Succeed())
			_, err := os.Stat(filesystem.LogsDir())
			Expect(err).NotTo(HaveOccurred())
		})
	})
})
