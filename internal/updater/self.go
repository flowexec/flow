package updater

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

const goosWindows = "windows"

const (
	// maxBinarySize caps the decompressed binary at 100 MB to guard against
	// decompression bombs.
	maxBinarySize = 100 << 20 // 100 MB
)

// These vars are package-level so tests can override them without hitting
// real GitHub URLs or the live executable path.
var (
	githubDownloadBaseURL = "https://github.com/" + githubRepo + "/releases/download"
	getExecutablePath     = os.Executable
)

// Upgrade downloads and installs the release described by info.
// When flow was installed via Homebrew it delegates to `brew upgrade flow`;
// otherwise it downloads the matching GitHub release asset and replaces the
// running binary in-place.
func Upgrade(info *ReleaseInfo) error {
	if isHomebrew() {
		return upgradeViaHomebrew()
	}
	return upgradeViaBinary(info.TagName)
}

func isHomebrew() bool {
	exe, err := getExecutablePath()
	if err != nil {
		return false
	}
	exe, err = filepath.EvalSymlinks(exe)
	if err != nil {
		return false
	}
	if prefix := os.Getenv("HOMEBREW_PREFIX"); prefix != "" && strings.HasPrefix(exe, prefix) {
		return true
	}
	for _, p := range []string{
		"/opt/homebrew/",
		"/usr/local/Cellar/",
		"/usr/local/opt/",
		"/home/linuxbrew/",
	} {
		if strings.HasPrefix(exe, p) {
			return true
		}
	}
	return false
}

func upgradeViaHomebrew() error {
	cmd := exec.Command("brew", "upgrade", "flow")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func upgradeViaBinary(tagName string) error {
	exe, err := getExecutablePath()
	if err != nil {
		return fmt.Errorf("cannot determine current executable path: %w", err)
	}
	exe, err = filepath.EvalSymlinks(exe)
	if err != nil {
		return fmt.Errorf("cannot resolve executable symlinks: %w", err)
	}

	asset := assetFileName(tagName)
	url := fmt.Sprintf("%s/%s/%s", githubDownloadBaseURL, tagName, asset)

	tmpArchive, err := os.CreateTemp("", "flow-update-*.archive")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	tmpArchive.Close()
	defer os.Remove(tmpArchive.Name())

	if err := downloadFile(url, tmpArchive.Name()); err != nil {
		return fmt.Errorf("download failed: %w", err)
	}

	binaryName := "flow"
	if runtime.GOOS == goosWindows {
		binaryName = "flow.exe"
	}
	extracted, err := extractBinary(tmpArchive.Name(), binaryName)
	if err != nil {
		return fmt.Errorf("extraction failed: %w", err)
	}
	defer os.Remove(extracted)

	if err := os.Chmod(extracted, 0755); err != nil { //nolint:mnd
		return fmt.Errorf("chmod failed: %w", err)
	}

	if err := os.Rename(extracted, exe); err != nil {
		// Cross-device rename fails on some filesystems; fall back to copy.
		return copyReplace(extracted, exe)
	}
	return nil
}

// assetFileName returns the goreleaser archive name for the current platform.
// Format: flow_<tag>_<goos>_<goarch>[.tar.gz|.zip]
func assetFileName(tagName string) string {
	ext := ".tar.gz"
	if runtime.GOOS == goosWindows {
		ext = ".zip"
	}
	return fmt.Sprintf("flow_%s_%s_%s%s", tagName, runtime.GOOS, runtime.GOARCH, ext)
}

func downloadFile(url, dest string) error {
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected HTTP %d downloading %s", resp.StatusCode, url)
	}
	f, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = io.Copy(f, io.LimitReader(resp.Body, maxBinarySize))
	return err
}

// extractBinary locates a file named binaryName inside a .tar.gz or .zip
// archive and writes it to a temporary file, returning its path.
func extractBinary(archivePath, binaryName string) (string, error) {
	if strings.HasSuffix(archivePath, ".zip") {
		return extractFromZip(archivePath, binaryName)
	}
	return extractFromTarGz(archivePath, binaryName)
}

func extractFromTarGz(archivePath, binaryName string) (string, error) {
	f, err := os.Open(archivePath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	gz, err := gzip.NewReader(f)
	if err != nil {
		return "", fmt.Errorf("invalid gzip archive: %w", err)
	}
	defer gz.Close()

	tr := tar.NewReader(gz)
	for {
		hdr, err := tr.Next()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return "", fmt.Errorf("error reading tar: %w", err)
		}
		if filepath.Base(hdr.Name) != binaryName {
			continue
		}
		tmp, err := os.CreateTemp("", "flow-new-*")
		if err != nil {
			return "", err
		}
		if _, err := io.Copy(tmp, io.LimitReader(tr, maxBinarySize)); err != nil {
			tmp.Close()
			os.Remove(tmp.Name())
			return "", err
		}
		tmp.Close()
		return tmp.Name(), nil
	}
	return "", fmt.Errorf("binary %q not found in archive", binaryName)
}

func extractFromZip(archivePath, binaryName string) (string, error) {
	zr, err := zip.OpenReader(archivePath)
	if err != nil {
		return "", err
	}
	defer zr.Close()

	for _, zf := range zr.File {
		if filepath.Base(zf.Name) != binaryName {
			continue
		}
		rc, err := zf.Open()
		if err != nil {
			return "", err
		}
		defer rc.Close()

		tmp, err := os.CreateTemp("", "flow-new-*")
		if err != nil {
			return "", err
		}
		if _, err := io.Copy(tmp, io.LimitReader(rc, maxBinarySize)); err != nil {
			tmp.Close()
			os.Remove(tmp.Name())
			return "", err
		}
		tmp.Close()
		return tmp.Name(), nil
	}
	return "", fmt.Errorf("binary %q not found in zip archive", binaryName)
}

// copyReplace copies src to a sibling temp file in dst's directory, then
// renames it to dst. Used as a fallback when os.Rename crosses filesystems.
func copyReplace(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	dir := filepath.Dir(dst)
	tmp, err := os.CreateTemp(dir, ".flow-update-*")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName)

	if _, err := io.Copy(tmp, in); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	if err := os.Chmod(tmpName, 0755); err != nil { //nolint:mnd
		return err
	}
	return os.Rename(tmpName, dst)
}
