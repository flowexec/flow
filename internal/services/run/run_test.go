package run_test

import (
	"os"
	"path/filepath"
	"testing"

	tuikitIO "github.com/flowexec/tuikit/io"
	"github.com/flowexec/tuikit/io/mocks"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/mock/gomock"

	"github.com/flowexec/flow/internal/services/run"
)

func TestRun(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Run Suite")
}

var _ = Describe("Run", func() {
	var (
		ctrl   *gomock.Controller
		logger *mocks.MockLogger
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		logger = mocks.NewMockLogger(ctrl)
		logger.EXPECT().Debugf(gomock.Any(), gomock.Any()).AnyTimes()
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	Describe("RunCmd", func() {
		When("log mode is hidden", func() {
			It("should not log the command output", func() {
				logger.EXPECT().SetMode(gomock.Any()).AnyTimes()
				logger.EXPECT().LogMode().DoAndReturn(func() tuikitIO.LogMode {
					return tuikitIO.Hidden
				}).AnyTimes()
				err := run.RunCmd("echo \"foo\"", "", nil, tuikitIO.Hidden, logger, os.Stdin, nil, nil)
				Expect(err).NotTo(HaveOccurred())
			})
		})
		When("log mode is text", func() {
			It("should log the command output", func() {
				logger.EXPECT().SetMode(gomock.Any()).AnyTimes()
				logger.EXPECT().LogMode().DoAndReturn(func() tuikitIO.LogMode {
					return tuikitIO.Text
				}).AnyTimes()
				logger.EXPECT().Print("foo").Times(1)
				logger.EXPECT().Print("\n").Times(1)
				err := run.RunCmd("echo \"foo\"", "", nil, tuikitIO.Text, logger, os.Stdin, nil, nil)
				Expect(err).NotTo(HaveOccurred())
			})
		})
		When("log mode is logfmt", func() {
			It("should log the command output", func() {
				logger.EXPECT().SetMode(gomock.Any()).AnyTimes()
				logger.EXPECT().LogMode().DoAndReturn(func() tuikitIO.LogMode {
					return tuikitIO.Logfmt
				}).AnyTimes()
				logger.EXPECT().Info("foo").Times(1)
				err := run.RunCmd("echo \"foo\"", "", nil, tuikitIO.Logfmt, logger, os.Stdin, nil, nil)
				Expect(err).NotTo(HaveOccurred())
			})
		})
		When("log mode is json", func() {
			It("should log the command output", func() {
				logger.EXPECT().SetMode(gomock.Any()).AnyTimes()
				logger.EXPECT().LogMode().DoAndReturn(func() tuikitIO.LogMode {
					return tuikitIO.JSON
				}).AnyTimes()
				logger.EXPECT().Info("foo").Times(1)
				err := run.RunCmd("echo \"foo\"", "", nil, tuikitIO.JSON, logger, os.Stdin, nil, nil)
				Expect(err).NotTo(HaveOccurred())
			})
		})
		When("log fields are provided", func() {
			It("should log the command output with the log fields", func() {
				logger.EXPECT().SetMode(gomock.Any()).AnyTimes()
				logger.EXPECT().LogMode().DoAndReturn(func() tuikitIO.LogMode {
					return tuikitIO.Logfmt
				}).AnyTimes()
				fields := map[string]interface{}{"key": "value"}
				logger.EXPECT().Info("foo", "key", "value").Times(1)
				err := run.RunCmd("echo \"foo\"", "", nil, tuikitIO.JSON, logger, os.Stdin, fields, nil)
				Expect(err).NotTo(HaveOccurred())
			})
		})
		When("env vars are provided", func() {
			It("should log the command output with the env vars", func() {
				logger.EXPECT().SetMode(gomock.Any()).AnyTimes()
				logger.EXPECT().LogMode().DoAndReturn(func() tuikitIO.LogMode {
					return tuikitIO.Logfmt
				}).AnyTimes()
				env := []string{"key=value"}
				logger.EXPECT().Info("value").Times(1)
				err := run.RunCmd("echo \"$key\"", "", env, tuikitIO.JSON, logger, os.Stdin, nil, nil)
				Expect(err).NotTo(HaveOccurred())
			})
		})
	})

	Describe("RunFile", func() {
		var tmpDir string

		BeforeEach(func() {
			var err error
			tmpDir, err = os.MkdirTemp("", "runfile-test")
			Expect(err).NotTo(HaveOccurred())
		})

		AfterEach(func() {
			Expect(os.RemoveAll(tmpDir)).To(Succeed())
		})

		It("should execute a .sh file via the POSIX interpreter", func() {
			err := os.WriteFile(filepath.Join(tmpDir, "test.sh"), []byte("#!/bin/sh\necho foo"), 0644)
			Expect(err).NotTo(HaveOccurred())

			logger.EXPECT().SetMode(gomock.Any()).AnyTimes()
			logger.EXPECT().LogMode().DoAndReturn(func() tuikitIO.LogMode {
				return tuikitIO.Text
			}).AnyTimes()
			logger.EXPECT().Print("foo").Times(1)
			logger.EXPECT().Print("\n").Times(1)
			err = run.RunFile("test.sh", tmpDir, nil, tuikitIO.Logfmt, logger, os.Stdin, nil, nil)
			Expect(err).NotTo(HaveOccurred())
		})

		It("should return an error for a non-existent file", func() {
			logger.EXPECT().SetMode(gomock.Any()).AnyTimes()
			logger.EXPECT().LogMode().AnyTimes()
			err := run.RunFile("missing.sh", tmpDir, nil, tuikitIO.Logfmt, logger, os.Stdin, nil, nil)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("file does not exist"))
		})

		It("should not parse a .bat file as shell syntax", func() {
			// Write batch syntax that would fail the POSIX parser.
			// If routing works correctly, this goes to cmd.exe (not the shell parser).
			err := os.WriteFile(filepath.Join(tmpDir, "test.bat"), []byte("@echo off\r\necho hello"), 0644)
			Expect(err).NotTo(HaveOccurred())

			logger.EXPECT().SetMode(gomock.Any()).AnyTimes()
			logger.EXPECT().LogMode().AnyTimes()
			logger.EXPECT().Print(gomock.Any()).AnyTimes()

			err = run.RunFile("test.bat", tmpDir, nil, tuikitIO.Hidden, logger, os.Stdin, nil, nil)
			// On non-Windows this will fail because cmd.exe doesn't exist, but crucially
			// it should NOT fail with "unable to parse file" (which would mean it hit the shell parser).
			if err != nil {
				Expect(err.Error()).NotTo(ContainSubstring("unable to parse file"))
			}
		})

		It("should not parse a .ps1 file as shell syntax", func() {
			err := os.WriteFile(filepath.Join(tmpDir, "test.ps1"), []byte("Write-Host 'hello'"), 0644)
			Expect(err).NotTo(HaveOccurred())

			logger.EXPECT().SetMode(gomock.Any()).AnyTimes()
			logger.EXPECT().LogMode().AnyTimes()
			logger.EXPECT().Print(gomock.Any()).AnyTimes()

			err = run.RunFile("test.ps1", tmpDir, nil, tuikitIO.Hidden, logger, os.Stdin, nil, nil)
			// Same as above — on non-Windows/non-pwsh systems this may fail,
			// but it must NOT fail with a shell parse error.
			if err != nil {
				Expect(err.Error()).NotTo(ContainSubstring("unable to parse file"))
			}
		})
	})
})
