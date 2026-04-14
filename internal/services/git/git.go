package git

import (
	"bytes"
	"fmt"
	"io"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/pkg/errors"

	"github.com/flowexec/flow/pkg/filesystem"
)

var sshURLPattern = regexp.MustCompile(`^[\w.-]+@[\w.-]+:[\w./-]+$`)

// IsGitURL returns true if the given string looks like a Git remote URL (HTTPS or SSH).
func IsGitURL(s string) bool {
	if sshURLPattern.MatchString(s) {
		return true
	}
	u, err := url.Parse(s)
	if err != nil {
		return false
	}
	return (u.Scheme == "https" || u.Scheme == "http" || u.Scheme == "ssh") &&
		u.Host != "" &&
		strings.HasSuffix(u.Path, ".git")
}

// ClonePath returns the local directory path where a git workspace should be cloned.
// Follows Go module conventions: ~/.cache/flow/git-workspaces/<host>/<path>
func ClonePath(gitURL string) (string, error) {
	host, repoPath, err := parseGitURL(gitURL)
	if err != nil {
		return "", err
	}
	repoPath = strings.TrimSuffix(repoPath, ".git")
	repoPath = strings.TrimPrefix(repoPath, "/")
	return filepath.Join(filesystem.CachedDataDirPath(), "git-workspaces", host, repoPath), nil
}

// Clone clones a git repository to the target directory.
// If branch is non-empty, it checks out that branch.
// If tag is non-empty, it checks out that tag.
// Progress output goes to stderr.
func Clone(gitURL, targetDir, branch, tag string) error {
	if _, err := os.Stat(targetDir); err == nil {
		entries, readErr := os.ReadDir(targetDir)
		if readErr == nil && len(entries) > 0 {
			return fmt.Errorf("target directory %s already exists and is not empty", targetDir)
		}
	}

	args := []string{"clone", "--progress"}
	if branch != "" {
		args = append(args, "--branch", branch)
	} else if tag != "" {
		args = append(args, "--branch", tag)
	}
	args = append(args, gitURL, targetDir)

	if err := runGit("", args...); err != nil {
		return err
	}
	return nil
}

// Pull fetches and pulls the latest changes for a git repository.
// If the workspace was cloned with a specific branch, it pulls that branch.
// If it was cloned with a tag, it fetches tags and checks out the tag.
func Pull(repoDir, ref, refType string) error {
	if info, err := os.Stat(repoDir); err != nil {
		return fmt.Errorf("git repo %s does not exist: %w", repoDir, err)
	} else if !info.IsDir() {
		return fmt.Errorf("git repo %s is not a directory", repoDir)
	}

	if refType == "tag" {
		if err := runGit(repoDir, "fetch", "--tags", "--progress"); err != nil {
			return err
		}
		if ref != "" {
			if err := runGit(repoDir, "checkout", ref); err != nil {
				return err
			}
		}
		return nil
	}

	// For branches (or no ref), do a regular pull
	return runGit(repoDir, "pull", "--progress")
}

// ResetPull performs a force update by resetting the working tree to match the remote.
// For branches, it fetches and does a hard reset to the remote tracking branch.
// For tags, it fetches tags and checks out the specified tag, discarding local changes.
func ResetPull(repoDir, ref, refType string) error {
	if info, err := os.Stat(repoDir); err != nil {
		return fmt.Errorf("git repo %s does not exist: %w", repoDir, err)
	} else if !info.IsDir() {
		return fmt.Errorf("git repo %s is not a directory", repoDir)
	}

	if refType == "tag" {
		if err := runGit(repoDir, "fetch", "--tags", "--force", "--progress"); err != nil {
			return err
		}
		// Discard local changes, then checkout tag
		if err := runGit(repoDir, "checkout", "--force", ref); err != nil {
			return err
		}
		return runGit(repoDir, "clean", "-fd")
	}

	// For branches: fetch, then hard reset to remote
	if err := runGit(repoDir, "fetch", "--progress"); err != nil {
		return err
	}

	// Determine the remote tracking ref
	resetTarget := "FETCH_HEAD"
	if ref != "" {
		resetTarget = "origin/" + ref
	}
	if err := runGit(repoDir, "reset", "--hard", resetTarget); err != nil {
		return err
	}
	return runGit(repoDir, "clean", "-fd")
}

// runGit executes a git command, streams progress to stderr, and captures output for error messages.
func runGit(dir string, args ...string) error {
	cmd := exec.Command("git", args...)
	if dir != "" {
		cmd.Dir = dir
	}

	// Capture stderr for error diagnostics while also streaming it
	var stderrBuf bytes.Buffer
	cmd.Stdout = os.Stderr
	cmd.Stderr = io.MultiWriter(os.Stderr, &stderrBuf)

	if err := cmd.Run(); err != nil {
		stderr := strings.TrimSpace(stderrBuf.String())
		cmdStr := "git " + strings.Join(args, " ")

		if stderr != "" {
			return errors.Wrapf(err, "%s:\n%s", cmdStr, stderr)
		}
		return errors.Wrap(err, cmdStr)
	}
	return nil
}

// parseGitURL extracts the host and path from a git URL (HTTPS or SSH).
func parseGitURL(gitURL string) (host, repoPath string, err error) {
	if sshURLPattern.MatchString(gitURL) {
		// SSH format: git@github.com:org/repo.git
		parts := strings.SplitN(gitURL, ":", 2)
		if len(parts) != 2 {
			return "", "", fmt.Errorf("invalid SSH git URL: %s", gitURL)
		}
		hostPart := parts[0]
		if idx := strings.Index(hostPart, "@"); idx >= 0 {
			hostPart = hostPart[idx+1:]
		}
		return hostPart, parts[1], nil
	}

	u, err := url.Parse(gitURL)
	if err != nil {
		return "", "", fmt.Errorf("invalid git URL: %w", err)
	}
	return u.Host, u.Path, nil
}
