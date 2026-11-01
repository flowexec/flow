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
	"slices"
	"strings"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"

	"github.com/flowexec/flow/v2/pkg/filesystem"
	"github.com/flowexec/flow/v2/types/workspace"
)

const refTypeTag = string(workspace.WorkspaceGitRefTypeTag)

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
	if u.Scheme == "file" && u.Path != "" {
		return true
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
	repoPath = strings.ReplaceAll(repoPath, ":", "_") // sanitize Windows drive-letter colons
	return filepath.Join(clonesDir(), host, repoPath), nil
}

func clonesDir() string {
	return filepath.Join(filesystem.CachedDataDirPath(), "git-workspaces")
}

// IsManagedClone reports whether repoDir is a clone flow made under its cache directory, as opposed
// to a checkout the user owns.
func IsManagedClone(repoDir string) bool {
	rel, err := filepath.Rel(filesystem.NormalizePath(clonesDir()), filesystem.NormalizePath(repoDir))
	return err == nil && rel != "." && rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator))
}

// DetectSource reads a clone's remote and what it has checked out: its branch, or the tag at a
// detached HEAD. Earlier versions of flow could lose this metadata, leaving the clone as the only
// record of it.
func DetectSource(repoDir string) (remote, ref, refType string, err error) {
	remote, err = gitOutput(repoDir, "remote", "get-url", "origin")
	if err != nil {
		return "", "", "", err
	}
	remote = strings.TrimSpace(remote)
	if branch, err := gitOutput(repoDir, "symbolic-ref", "--short", "-q", "HEAD"); err == nil {
		return remote, strings.TrimSpace(branch), "branch", nil
	}
	if tag, err := gitOutput(repoDir, "describe", "--tags", "--exact-match"); err == nil {
		return remote, strings.TrimSpace(tag), refTypeTag, nil
	}
	return remote, "", "", nil
}

// RepairWorkspaceConfig undoes what earlier versions of flow did to flow.yaml in a managed clone, so
// that it can't block a pull. A tracked file that differs from HEAD only by flow's rewrite is
// restored; one with any other change is the user's and stays as it is. It reports whether it
// restored the file.
func RepairWorkspaceConfig(repoDir string) (bool, error) {
	status, err := gitOutput(repoDir, "status", "--porcelain", "--", filesystem.WorkspaceConfigFileName)
	if err != nil {
		return false, err
	}
	switch {
	case strings.HasPrefix(status, "??"):
		return false, ExcludeWorkspaceConfig(repoDir)
	case !strings.HasPrefix(status, " M"):
		return false, nil
	}

	current, err := os.ReadFile(filepath.Clean(filepath.Join(repoDir, filesystem.WorkspaceConfigFileName)))
	if err != nil {
		return false, err
	}
	committed, err := gitOutput(repoDir, "show", "HEAD:"+filesystem.WorkspaceConfigFileName)
	if err != nil {
		return false, err
	}
	if !isGitFieldRewrite(current, []byte(committed)) {
		return false, nil
	}
	if err := runGit(repoDir, "checkout", "HEAD", "--", filesystem.WorkspaceConfigFileName); err != nil {
		return false, err
	}
	return true, nil
}

// isGitFieldRewrite reports whether current is committed as flow used to rewrite it: decoded into
// the workspace type, which loses comments, formatting, and unknown keys, then re-encoded with the
// git fields added.
func isGitFieldRewrite(current, committed []byte) bool {
	var currentCfg, committedCfg workspace.Workspace
	if err := yaml.Unmarshal(current, &currentCfg); err != nil || currentCfg.GitRemote == "" {
		return false
	}
	if err := yaml.Unmarshal(committed, &committedCfg); err != nil {
		return false
	}
	for _, cfg := range []*workspace.Workspace{&currentCfg, &committedCfg} {
		cfg.GitRemote, cfg.GitRef, cfg.GitRefType = "", "", workspace.WorkspaceGitRefTypeBlank
	}
	currentYAML, err := yaml.Marshal(&currentCfg)
	if err != nil {
		return false
	}
	committedYAML, err := yaml.Marshal(&committedCfg)
	if err != nil {
		return false
	}
	return bytes.Equal(currentYAML, committedYAML)
}

// ExcludeWorkspaceConfig has git ignore a flow.yaml that flow generated in a clone of a repository
// without one. The clone stays clean, and git overwrites an ignored file, so upstream can still add
// its own flow.yaml later.
func ExcludeWorkspaceConfig(repoDir string) error {
	excludeFile := filepath.Clean(filepath.Join(repoDir, ".git", "info", "exclude"))
	const pattern = "/" + filesystem.WorkspaceConfigFileName

	existing, err := os.ReadFile(excludeFile)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	if slices.Contains(strings.Split(string(existing), "\n"), pattern) {
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(excludeFile), 0750); err != nil {
		return err
	}
	file, err := os.OpenFile(excludeFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = file.WriteString("\n" + pattern + "\n")
	return err
}

// Clone clones a git repository to the target directory.
// If branch is non-empty, it checks out that branch.
// If tag is non-empty, it checks out that tag.
// If depth > 0, a shallow clone is performed.
// Progress output goes to stderr.
func Clone(gitURL, targetDir, branch, tag string, depth int) error {
	if _, err := os.Stat(targetDir); err == nil {
		entries, readErr := os.ReadDir(targetDir)
		if readErr == nil && len(entries) > 0 {
			return fmt.Errorf("target directory %s already exists and is not empty", targetDir)
		}
	}

	args := []string{"clone", "--progress"}
	if depth > 0 {
		args = append(args, "--depth", fmt.Sprintf("%d", depth))
	}
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

	if refType == refTypeTag {
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

	if refType == refTypeTag {
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

// EnsureInstalled checks that the git binary is available on PATH.
func EnsureInstalled() error {
	if _, err := exec.LookPath("git"); err != nil {
		return fmt.Errorf("git is not installed or not in PATH (required for git workspace features)")
	}
	return nil
}

// runGit executes a git command, streams progress to stderr, and captures output for error messages.
func runGit(dir string, args ...string) error {
	if err := EnsureInstalled(); err != nil {
		return err
	}

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

// gitOutput executes a git command and returns its stdout.
func gitOutput(dir string, args ...string) (string, error) {
	if err := EnsureInstalled(); err != nil {
		return "", err
	}

	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	var stderrBuf bytes.Buffer
	cmd.Stderr = &stderrBuf
	out, err := cmd.Output()
	if err != nil {
		return "", errors.Wrapf(err, "git %s:\n%s", strings.Join(args, " "), strings.TrimSpace(stderrBuf.String()))
	}
	return string(out), nil
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
	host = u.Host
	if host == "" {
		host = "localhost"
	}
	return host, u.Path, nil
}
