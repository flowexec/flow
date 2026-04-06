package filesystem_test

import (
	"os"

	"github.com/flowexec/flow/pkg/filesystem"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Cache", func() {
	var tmpDir string

	BeforeEach(func() {
		var err error
		tmpDir, err = os.MkdirTemp("", "flow-cache-test")
		Expect(err).NotTo(HaveOccurred())
		Expect(os.Setenv(filesystem.FlowCacheDirEnvVar, tmpDir)).To(Succeed())
	})

	AfterEach(func() {
		Expect(os.RemoveAll(tmpDir)).To(Succeed())
		Expect(os.Unsetenv(filesystem.FlowCacheDirEnvVar)).To(Succeed())
	})

	Describe("CachedDataDirPath", func() {
		It("returns the correct path", func() {
			Expect(filesystem.CachedDataDirPath()).To(Equal(tmpDir))
		})
	})
})
