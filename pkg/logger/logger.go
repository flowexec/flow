package logger

import (
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/flowexec/tuikit/io"
	"github.com/flowexec/tuikit/themes"
)

var (
	globalLogger     io.Logger
	testLoggers      sync.Map
	once             sync.Once
	loggerMutex      sync.RWMutex
	testLoggerEnvKey = "FLOW_TEST_LOGGER"

	discardOnce sync.Once
	discardFile *os.File
)

// discardOutput returns a writer that throws log output away, for tests that have not
// registered a logger of their own.
//
// Falls back to stdout if the device cannot be opened: a noisy test logger is a smaller
// problem than one writing to an arbitrary descriptor.
func discardOutput() *os.File {
	discardOnce.Do(func() {
		f, err := os.OpenFile(os.DevNull, os.O_WRONLY, 0)
		if err != nil {
			discardFile = os.Stdout
			return
		}
		discardFile = f
	})
	return discardFile
}

type InitOptions struct {
	StdOut           *os.File
	ArchiveDirectory string
	ArchiveID        string // Unique identifier embedded in the archive filename for reliable linking.
	LogMode          io.LogMode
	Theme            themes.Theme

	// TestingTB is used to set a testing.T instance for the logger.
	// This is useful for capturing log output in tests.
	TestingTB testing.TB
	Logger    io.Logger // Optional logger to use instead of creating a new one
}

// Init initializes the global logger with the provided options.
// This function is safe to call multiple times - only the first call will initialize the logger.
// When TestingTB is provided, it initializes a test logger specific to that test.
func Init(opts InitOptions) {
	if opts.TestingTB != nil {
		loggerMutex.Lock()
		defer loggerMutex.Unlock()
		initializeTestLogger(opts.TestingTB, opts.StdOut, opts.Logger)
		return
	}

	once.Do(func() {
		if opts.StdOut == nil {
			panic("logger output file is unset")
		}

		loggerOpts := []io.LoggerOptions{
			io.WithOutput(opts.StdOut),
			io.WithMode(opts.LogMode),
		}

		if opts.Theme != nil {
			loggerOpts = append(loggerOpts, io.WithTheme(opts.Theme))
		}
		if opts.ArchiveDirectory != "" {
			loggerOpts = append(loggerOpts, io.WithArchiveDirectory(opts.ArchiveDirectory))
			if opts.ArchiveID != "" {
				loggerOpts = append(loggerOpts, io.WithArchiveID(opts.ArchiveID))
			}
		}

		globalLogger = io.NewLogger(loggerOpts...)
	})
}

// Log returns the global logger instance.
// The logger must be initialized with Init() before calling this function.
func Log() io.Logger {
	if testing.Testing() {
		loggerMutex.RLock()
		defer loggerMutex.RUnlock()

		if name := os.Getenv(testLoggerEnvKey); name != "" {
			if logger, ok := testLoggers.Load(name); ok {
				return logger.(io.Logger) //nolint:errcheck
			}
		}
		return io.NewLogger(io.WithOutput(discardOutput()))
	}

	if globalLogger == nil {
		panic("global logger not initialized - call logger.Init() first")
	}
	return globalLogger
}

func initializeTestLogger(tb testing.TB, stdOut *os.File, logger io.Logger) {
	key := fmt.Sprintf("%s_%d", tb.Name(), time.Now().UnixNano())
	if _, ok := testLoggers.Load(key); ok {
		return
	}

	tb.Setenv(testLoggerEnvKey, key)

	if logger == nil {
		if stdOut == nil {
			tb.Log("no stdOut provided for test logger, using os.Stdout")
			stdOut = os.Stdout
		}
		logger = io.NewLogger(io.WithOutput(stdOut), io.WithMode(io.Text))
	}

	testLoggers.Store(key, logger)

	if t, ok := tb.(*testing.T); ok {
		t.Cleanup(func() {
			loggerMutex.Lock()
			defer loggerMutex.Unlock()
			testLoggers.Delete(key)
		})
	}
}

// Reset resets the global logger to its initial state.
func Reset() {
	loggerMutex.Lock()
	defer loggerMutex.Unlock()

	testLoggers.Range(func(key, value interface{}) bool {
		testLoggers.Delete(key)
		return true
	})

	globalLogger = nil
	once = sync.Once{}
	if testing.Testing() {
		os.Unsetenv(testLoggerEnvKey)
	}
}
