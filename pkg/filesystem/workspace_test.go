package filesystem_test

import (
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/flowexec/flow/pkg/filesystem"
	"github.com/flowexec/flow/types/workspace"
)

var _ = Describe("Workspace", func() {
	var (
		tmpDir string
	)

	BeforeEach(func() {
		var err error
		tmpDir, err = os.MkdirTemp("", "flow-workspace-test")
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		Expect(os.RemoveAll(tmpDir)).To(Succeed())
	})

	Describe("EnsureWorkspaceDir", func() {
		It("creates the directory if it does not exist", func() {
			Expect(filesystem.EnsureWorkspaceDir(tmpDir)).To(Succeed())
			_, err := os.Stat(tmpDir)
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Describe("EnsureWorkspaceConfig", func() {
		It("creates the config file if it does not exist", func() {
			Expect(filesystem.EnsureWorkspaceConfig("test", tmpDir)).To(Succeed())
			_, err := os.Stat(filepath.Join(tmpDir, filesystem.WorkspaceConfigFileName))
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Describe("WriteWorkspaceConfig and LoadWorkspaceConfig", func() {
		It("writes and reads config correctly", func() {
			workspaceConfig := workspace.DefaultWorkspaceConfig("test")
			workspaceConfig.SetContext("test", tmpDir)

			Expect(filesystem.WriteWorkspaceConfig(tmpDir, workspaceConfig)).To(Succeed())

			readConfig, err := filesystem.LoadWorkspaceConfig("test", tmpDir)
			Expect(err).NotTo(HaveOccurred())
			Expect(readConfig).To(Equal(workspaceConfig))
		})
	})
})
