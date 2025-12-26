package filesystem_test

import (
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/flowexec/flow/pkg/filesystem"
	"github.com/flowexec/flow/types/config"
)

var _ = Describe("FlowFile", func() {
	var (
		tmpDir string
	)

	BeforeEach(func() {
		var err error
		tmpDir, err = os.MkdirTemp("", "flow-config-test")
		Expect(err).NotTo(HaveOccurred())
		Expect(os.Setenv(filesystem.FlowConfigDirEnvVar, tmpDir)).To(Succeed())
	})

	AfterEach(func() {
		Expect(os.RemoveAll(tmpDir)).To(Succeed())
		Expect(os.Unsetenv(filesystem.FlowConfigDirEnvVar)).To(Succeed())
	})

	Describe("ConfigDirPath", func() {
		It("returns the correct path", func() {
			Expect(filesystem.ConfigDirPath()).To(Equal(tmpDir))
		})
	})

	Describe("UserConfigFilePath", func() {
		It("returns the correct path", func() {
			Expect(filesystem.UserConfigFilePath()).To(Equal(filepath.Join(tmpDir, "config.yaml")))
		})
	})

	Describe("EnsureConfigDir", func() {
		It("creates the directory if it does not exist", func() {
			Expect(filesystem.EnsureConfigDir()).To(Succeed())
			_, err := os.Stat(filesystem.ConfigDirPath())
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Describe("WriteConfig and LoadConfig", func() {
		It("writes and reads config correctly", func() {
			userConfig := &config.Config{
				Workspaces:       map[string]string{"default": tmpDir},
				CurrentWorkspace: "default",
				WorkspaceMode:    config.ConfigWorkspaceModeDynamic,
				Interactive: &config.Interactive{
					Enabled: true,
				},
				DefaultLogMode: "logfmt",
			}

			Expect(filesystem.WriteConfig(userConfig)).To(Succeed())

			readConfig, err := filesystem.LoadConfig()
			Expect(err).NotTo(HaveOccurred())
			Expect(readConfig).To(Equal(userConfig))
		})
	})
})
