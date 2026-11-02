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
	rel, err := filepath.Rel(resolvePath(clonesDir()), resolvePath(repoDir))
	return err == nil && rel != "." && rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator))
}

// IsCloneOf reports whether repoDir is the root of a repository whose origin is remote. It
// recognizes a managed clone after the cache directory's path has changed.
func IsCloneOf(repoDir, remote string) bool {
	return IsRepoRoot(repoDir) && originURL(repoDir) == remote
}

// originURL returns origin as configured, not with url.insteadOf applied, which could add
// credentials that shouldn't be recorded.
func originURL(repoDir string) string {
	remote, err := gitOutput(repoDir, "config", "--get", "remote.origin.url")
	if err != nil {
		return ""
	}
	return strings.TrimSpace(remote)
}

func resolvePath(path string) string {
	if resolved, err := filepath.EvalSymlinks(path); err == nil {
		path = resolved
	}
	return filesystem.NormalizePath(path)
}

// DetectSource reads a clone's remote and what it has checked out: its branch, or the tag at a
// detached HEAD. Earlier versions of flow could lose this metadata, leaving the clone as the only
// record of it.
func DetectSource(repoDir string) (remote, ref, refType string, err error) {
	remote = originURL(repoDir)
	if remote == "" {
		return "", "", "", fmt.Errorf("%s has no origin remote", repoDir)
	}
	if branch, err := gitOutput(repoDir, "symbolic-ref", "--short", "-q", "HEAD"); err == nil {
		return remote, strings.TrimSpace(branch), "branch", nil
	}
	// A commit can carry several tags, and picking one would have a force update follow a guess.
	tags, err := gitOutput(repoDir, "tag", "--points-at", "HEAD")
	if err == nil && len(strings.Fields(tags)) == 1 {
		return remote, strings.TrimSpace(tags), refTypeTag, nil
	}
	return remote, "", "", nil
}

// IsRepoRoot reports whether dir is the top level of its own git repository. Git walks up from a
// directory that isn't, so a pull or hard reset there would act on whichever repository encloses it.
func IsRepoRoot(dir string) bool {
	top, err := gitOutput(dir, "rev-parse", "--show-toplevel")
	if err != nil {
		return false
	}
	return resolvePath(dir) == resolvePath(strings.TrimSpace(top))
}

// RepairWorkspaceConfig undoes what flow did to flow.yaml in a managed clone, so that it can't block
// a pull. A tracked file that is exactly flow's old rewrite of HEAD is restored. An untracked file
// that is exactly what flow generated is removed, leaving room for upstream to add its own; flow
// generates it again if upstream hasn't. Any other content is the user's and stays as it is, and
// git refuses a pull that would overwrite it. It reports whether it restored a tracked file.
func RepairWorkspaceConfig(repoDir, workspaceName string) (bool, error) {
	status, err := gitOutput(repoDir, "status", "--porcelain", "--", filesystem.WorkspaceConfigFileName)
	if err != nil {
		return false, err
	}
	untracked := strings.HasPrefix(status, "??")
	if !untracked && !strings.HasPrefix(status, " M") {
		return false, nil
	}

	configFile := filepath.Clean(filepath.Join(repoDir, filesystem.WorkspaceConfigFileName))
	current, err := os.ReadFile(configFile)
	if err != nil {
		return false, err
	}
	if untracked {
		if isGeneratedConfig(current, workspaceName) {
			return false, os.Remove(configFile)
		}
		return false, nil
	}

	committed, err := gitOutput(repoDir, "show", "HEAD:"+filesystem.WorkspaceConfigFileName)
	if err != nil {
		return false, err
	}
	var committedCfg workspace.Workspace
	if err := yaml.Unmarshal([]byte(committed), &committedCfg); err != nil {
		return false, nil //nolint:nilerr // A file flow can't read can't be one flow rewrote.
	}
	if !isEncodingWithGitFields(current, &committedCfg) {
		return false, nil
	}
	if err := runGit(repoDir, "checkout", "HEAD", "--", filesystem.WorkspaceConfigFileName); err != nil {
		return false, err
	}
	return true, nil
}

func isGeneratedConfig(current []byte, workspaceName string) bool {
	return isEncodingWithGitFields(current, workspace.DefaultWorkspaceConfig(workspaceName)) ||
		bytes.Equal(current, encodeWorkspace(workspace.DefaultWorkspaceConfig(workspaceName)))
}

// isEncodingWithGitFields reports whether current is byte for byte what flow writes for base plus
// the git fields current holds. Comparing bytes rather than decoded values means a comment or an
// unknown key the user added counts as an edit.
func isEncodingWithGitFields(current []byte, base *workspace.Workspace) bool {
	var currentCfg workspace.Workspace
	if err := yaml.Unmarshal(current, &currentCfg); err != nil || currentCfg.GitRemote == "" {
		return false
	}
	expected := *base
	expected.GitRemote = currentCfg.GitRemote
	expected.GitRef = currentCfg.GitRef
	expected.GitRefType = currentCfg.GitRefType
	return bytes.Equal(current, encodeWorkspace(&expected))
}

func encodeWorkspace(cfg *workspace.Workspace) []byte {
	encoded, err := yaml.Marshal(cfg)
	if err != nil {
		return nil
	}
	return encoded
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

	if !IsRepoRoot(repoDir) {
		return fmt.Errorf("%s is not the root of a git repository; refusing to reset the repository enclosing it", repoDir)
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
