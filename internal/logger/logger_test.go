package logger_test

import (
	"os"
	"testing"

	"github.com/flowexec/tuikit/io"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/flowexec/flow/internal/logger"
)

func TestLogger(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Logger Suite")
}

var _ = Describe("Global Logger", func() {
	It("should allow reinitialization after reset", func() {
		opts := logger.InitOptions{
			StdOut:  os.Stdout,
			LogMode: io.Logfmt,
		}

		logger.Init(opts)
		logger1 := logger.Log()

		logger.Reset()

		logger.Init(opts)
		logger2 := logger.Log()

		Expect(logger1).ToNot(Equal(logger2))
	})
})
