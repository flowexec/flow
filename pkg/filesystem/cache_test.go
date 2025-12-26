package filesystem_test

import (
	"os"
	"path/filepath"

	"github.com/flowexec/flow/pkg/filesystem"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Cache", func() {
	var (
		tmpDir string
	)

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

	Describe("LatestCachedDataDir", func() {
		It("returns the correct path", func() {
			Expect(filesystem.LatestCachedDataDir()).To(Equal(filepath.Join(tmpDir, "latestcache")))
		})
	})

	Describe("EnsureCachedDataDir", func() {
		It("creates the directory if it does not exist", func() {
			Expect(filesystem.EnsureCachedDataDir()).To(Succeed())
			_, err := os.Stat(filesystem.LatestCachedDataDir())
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Describe("WriteLatestCachedData and LoadLatestCachedData", func() {
		It("writes and reads data correctly", func() {
			cacheKey := "test"
			data := []byte("test data")

			Expect(filesystem.WriteLatestCachedData(cacheKey, data)).To(Succeed())

			readData, err := filesystem.LoadLatestCachedData(cacheKey)
			Expect(err).NotTo(HaveOccurred())
			Expect(string(readData)).To(Equal(string(data)))
		})
	})
})
