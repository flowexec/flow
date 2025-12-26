package filesystem_test

import (
	"os"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/flowexec/flow/pkg/filesystem"
)

var _ = Describe("Logs", func() {
	Describe("LogsDir", func() {
		It("returns the correct logs directory path", func() {
			logsDir := filesystem.LogsDir()
			Expect(logsDir).To(HaveSuffix("/logs"))
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
