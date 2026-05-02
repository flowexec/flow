package updater_test

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/flowexec/flow/v2/internal/updater"
)

func makeTarGz(binaryName string, content []byte) string {
	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gw)

	hdr := &tar.Header{
		Name: "flow_v1.0.0_linux_amd64/" + binaryName,
		Mode: 0755,
		Size: int64(len(content)),
	}
	Expect(tw.WriteHeader(hdr)).To(Succeed())
	_, err := tw.Write(content)
	Expect(err).NotTo(HaveOccurred())
	Expect(tw.Close()).To(Succeed())
	Expect(gw.Close()).To(Succeed())

	tmp, err := os.CreateTemp("", "test-*.tar.gz")
	Expect(err).NotTo(HaveOccurred())
	_, err = tmp.Write(buf.Bytes())
	Expect(err).NotTo(HaveOccurred())
	Expect(tmp.Close()).To(Succeed())
	return tmp.Name()
}

func makeZip(binaryName string, content []byte) string {
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)

	f, err := zw.Create("flow_v1.0.0_windows_amd64/" + binaryName)
	Expect(err).NotTo(HaveOccurred())
	_, err = f.Write(content)
	Expect(err).NotTo(HaveOccurred())
	Expect(zw.Close()).To(Succeed())

	tmp, err := os.CreateTemp("", "test-*.zip")
	Expect(err).NotTo(HaveOccurred())
	_, err = tmp.Write(buf.Bytes())
	Expect(err).NotTo(HaveOccurred())
	Expect(tmp.Close()).To(Succeed())
	return tmp.Name()
}

var _ = Describe("assetFileName", func() {
	It("matches the goreleaser name template for the current platform", func() {
		name := updater.AssetFileName("v2.0.0")
		Expect(name).To(HavePrefix("flow_v2.0.0_" + runtime.GOOS + "_" + runtime.GOARCH))
		if runtime.GOOS == "windows" {
			Expect(name).To(HaveSuffix(".zip"))
		} else {
			Expect(name).To(HaveSuffix(".tar.gz"))
		}
	})
})

var _ = Describe("isHomebrew", func() {
	var origPrefix string

	BeforeEach(func() { origPrefix = os.Getenv("HOMEBREW_PREFIX") })
	AfterEach(func() {
		if origPrefix == "" {
			_ = os.Unsetenv("HOMEBREW_PREFIX")
		} else {
			_ = os.Setenv("HOMEBREW_PREFIX", origPrefix)
		}
	})

	It("returns false when binary is not under a Homebrew path", func() {
		_ = os.Unsetenv("HOMEBREW_PREFIX")
		Expect(updater.IsHomebrew()).To(BeFalse())
	})

	It("returns true when binary is under HOMEBREW_PREFIX", func() {
		fakePrefix, err := os.MkdirTemp("", "fake-homebrew-*")
		Expect(err).NotTo(HaveOccurred())
		defer os.RemoveAll(fakePrefix)

		// Resolve symlinks so the comparison works on macOS (/tmp → /private/tmp).
		fakePrefix, err = filepath.EvalSymlinks(fakePrefix)
		Expect(err).NotTo(HaveOccurred())

		fakeBin := filepath.Join(fakePrefix, "bin", "flow")
		Expect(os.MkdirAll(filepath.Dir(fakeBin), 0755)).To(Succeed())
		Expect(os.WriteFile(fakeBin, []byte("#!/bin/sh"), 0755)).To(Succeed())

		_ = os.Setenv("HOMEBREW_PREFIX", fakePrefix)
		origFn := *updater.GetExecutablePath
		*updater.GetExecutablePath = func() (string, error) { return fakeBin, nil } //nolint:unparam
		defer func() { *updater.GetExecutablePath = origFn }()

		Expect(updater.IsHomebrew()).To(BeTrue())
	})
})

var _ = Describe("extractFromTarGz", func() {
	It("extracts the named binary", func() {
		archive := makeTarGz("flow", []byte("binary-content"))
		defer os.Remove(archive)

		out, err := updater.ExtractFromTarGz(archive, "flow")
		Expect(err).NotTo(HaveOccurred())
		defer os.Remove(out)

		got, err := os.ReadFile(out)
		Expect(err).NotTo(HaveOccurred())
		Expect(string(got)).To(Equal("binary-content"))
	})

	It("returns an error when the binary is absent", func() {
		archive := makeTarGz("other", []byte("content"))
		defer os.Remove(archive)

		_, err := updater.ExtractFromTarGz(archive, "flow")
		Expect(err).To(HaveOccurred())
	})
})

var _ = Describe("extractFromZip", func() {
	It("extracts the named binary", func() {
		archive := makeZip("flow.exe", []byte("zip-content"))
		defer os.Remove(archive)

		out, err := updater.ExtractFromZip(archive, "flow.exe")
		Expect(err).NotTo(HaveOccurred())
		defer os.Remove(out)

		got, err := os.ReadFile(out)
		Expect(err).NotTo(HaveOccurred())
		Expect(string(got)).To(Equal("zip-content"))
	})

	It("returns an error when the binary is absent", func() {
		archive := makeZip("other.exe", []byte("content"))
		defer os.Remove(archive)

		_, err := updater.ExtractFromZip(archive, "flow.exe")
		Expect(err).To(HaveOccurred())
	})
})

var _ = Describe("copyReplace", func() {
	It("copies source content to destination", func() {
		src, err := os.CreateTemp("", "src-*")
		Expect(err).NotTo(HaveOccurred())
		_, _ = src.WriteString("hello world")
		src.Close()
		defer os.Remove(src.Name())

		dst, err := os.CreateTemp("", "dst-*")
		Expect(err).NotTo(HaveOccurred())
		dst.Close()
		defer os.Remove(dst.Name())

		Expect(updater.CopyReplace(src.Name(), dst.Name())).To(Succeed())

		got, err := os.ReadFile(dst.Name())
		Expect(err).NotTo(HaveOccurred())
		Expect(string(got)).To(Equal("hello world"))
	})
})

var _ = Describe("upgradeViaBinary", func() {
	var (
		origDownloadBase string
		origExePath      func() (string, error)
		destBin          string
	)

	BeforeEach(func() {
		origDownloadBase = *updater.GithubDownloadBaseURL
		origExePath = *updater.GetExecutablePath

		tmp, err := os.CreateTemp("", "flow-dest-*")
		Expect(err).NotTo(HaveOccurred())
		_, _ = tmp.WriteString("old binary")
		tmp.Close()
		destBin = tmp.Name()

		*updater.GetExecutablePath = func() (string, error) { return destBin, nil } //nolint:unparam
	})

	AfterEach(func() {
		*updater.GithubDownloadBaseURL = origDownloadBase
		*updater.GetExecutablePath = origExePath
		_ = os.Remove(destBin)
	})

	It("downloads, extracts, and replaces the target binary", func() {
		binaryName := "flow"
		if runtime.GOOS == "windows" {
			binaryName = "flow.exe"
		}
		archive := tarGzBytes(binaryName, []byte("new binary content"))

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, _ = w.Write(archive)
		}))
		defer server.Close()
		*updater.GithubDownloadBaseURL = server.URL

		Expect(updater.UpgradeViaBinary("v2.0.0")).To(Succeed())

		got, err := os.ReadFile(destBin)
		Expect(err).NotTo(HaveOccurred())
		Expect(string(got)).To(Equal("new binary content"))
	})

	It("returns an error when the download fails", func() {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		}))
		defer server.Close()
		*updater.GithubDownloadBaseURL = server.URL

		Expect(updater.UpgradeViaBinary("v2.0.0")).To(HaveOccurred())
	})
})

func tarGzBytes(binaryName string, content []byte) []byte {
	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gw)
	hdr := &tar.Header{
		Name: "flow_v2.0.0_linux_amd64/" + binaryName,
		Mode: 0755,
		Size: int64(len(content)),
	}
	_ = tw.WriteHeader(hdr)
	_, _ = tw.Write(content)
	_ = tw.Close()
	_ = gw.Close()
	return buf.Bytes()
}
