package logger_test

import (
	"os"
	"runtime"
	"testing"

	"github.com/flowexec/tuikit/io"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/flowexec/flow/v2/pkg/logger"
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

// The test-mode fallback logger once wrapped descriptor 0 with os.NewFile, which adopts
// standard input rather than opening the null device, and attaches a finalizer that closes
// whatever descriptor it holds. Every Log() call left another finalizer waiting to close
// fd 0, so the runtime would close standard input part-way through a test run and the next
// file the process opened inherited the freed number — under -coverprofile, that was the
// coverage profile, and the write failed with EBADF while every test passed.
//
// The descriptor the sink actually holds is asserted in discard_internal_test.go, which
// can reach it; this covers the consequence, which is what anyone hitting the bug sees.
var _ = Describe("Test-mode fallback logger", func() {

	// A fresh file per call is what accumulates finalizers. Reusing one keeps the count at
	// one for the life of the process, however many times Log() is called.
	It("survives repeated use with the garbage collector running", func() {
		for range 200 {
			logger.Log().Debugf("keeping the fallback busy")
		}
		runtime.GC()
		runtime.GC()

		// Standard input must still be open. If a finalizer had closed it, this fails —
		// and so would anything the process opened afterwards.
		_, err := os.Stdin.Stat()
		Expect(err).ToNot(HaveOccurred(), "standard input was closed by a finalizer")

		// A file opened now must be usable, which is precisely what the coverage writer
		// needs at process exit.
		f, err := os.CreateTemp(GinkgoT().TempDir(), "after-gc-*")
		Expect(err).ToNot(HaveOccurred())
		defer f.Close()
		_, err = f.WriteString("still writable")
		Expect(err).ToNot(HaveOccurred(), "a file opened after GC was closed underneath us")
	})
})
