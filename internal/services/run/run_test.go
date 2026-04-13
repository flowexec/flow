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
		var testfile *os.File

		BeforeEach(func() {
			var err error
			testfile, err = os.CreateTemp("", "test.sh")
			Expect(err).NotTo(HaveOccurred())
			_, err = testfile.WriteString("#!/bin/sh\necho foo")
			Expect(err).NotTo(HaveOccurred())
		})

		AfterEach(func() {
			Expect(testfile.Close()).To(Succeed())
			Expect(os.Remove(testfile.Name())).To(Succeed())
		})

		It("should log the file execution output", func() {
			logger.EXPECT().SetMode(gomock.Any()).AnyTimes()
			logger.EXPECT().LogMode().DoAndReturn(func() tuikitIO.LogMode {
				return tuikitIO.Text
			}).AnyTimes()
			logger.EXPECT().Print("foo").Times(1)
			logger.EXPECT().Print("\n").Times(1)
			filename := filepath.Base(testfile.Name())
			filedir := filepath.Dir(testfile.Name())
			err := run.RunFile(filename, filedir, nil, tuikitIO.Logfmt, logger, os.Stdin, nil, nil)
			Expect(err).NotTo(HaveOccurred())
		})
	})
})
