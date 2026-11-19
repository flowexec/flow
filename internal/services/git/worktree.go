package git

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// WorktreeInfo holds the data parsed from one entry of `git worktree list --porcelain`.
type WorktreeInfo struct {
	Path       string
	HEAD       string
	Branch     string
	Detached   bool
	Bare       bool
	Prunable   bool
	Locked     bool
	LockReason string
}

// WorktreeStatusInfo holds the working-tree state of a single worktree.
type WorktreeStatusInfo struct {
	Dirty        bool
	FilesChanged int
	Ahead        int
	Behind       int
}

// ListWorktrees returns every worktree for the repository that contains dir.
func ListWorktrees(dir string) ([]WorktreeInfo, error) {
	root, err := RepoCommonDir(dir)
	if err != nil {
		return nil, err
	}

	output, err := gitOutput(root, "worktree", "list", "--porcelain")
	if err != nil {
		return nil, fmt.Errorf("listing worktrees: %w", err)
	}
	return parseWorktreeList(output), nil
}

func parseWorktreeList(output string) []WorktreeInfo {
	var worktrees []WorktreeInfo
	var current *WorktreeInfo

	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimRight(line, "\r")
		switch {
		case strings.HasPrefix(line, "worktree "):
			if current != nil {
				worktrees = append(worktrees, *current)
			}
			current = &WorktreeInfo{Path: strings.TrimPrefix(line, "worktree ")}
		case current == nil:
			continue
		case strings.HasPrefix(line, "HEAD "):
			current.HEAD = strings.TrimPrefix(line, "HEAD ")
		case strings.HasPrefix(line, "branch "):
			current.Branch = strings.TrimPrefix(line, "branch refs/heads/")
		case line == "bare":
			current.Bare = true
		case line == "detached":
			current.Detached = true
		case line == "prunable":
			current.Prunable = true
		case line == "locked":
			current.Locked = true
		case strings.HasPrefix(line, "locked "):
			current.Locked = true
			current.LockReason = strings.TrimPrefix(line, "locked ")
		}
	}
	if current != nil {
		worktrees = append(worktrees, *current)
	}
	return worktrees
}

// AddWorktree creates a new worktree. When createBranch is true, the branch is created with -b.
// base is the starting point for the new branch (empty means HEAD).
func AddWorktree(repoDir, worktreePath, branch string, createBranch bool, base string) error {
	root, err := RepoCommonDir(repoDir)
	if err != nil {
		return err
	}

	args := []string{"worktree", "add"}
	if createBranch {
		args = append(args, "-b", branch, worktreePath)
		if base != "" {
			args = append(args, base)
		}
	} else {
		args = append(args, worktreePath, branch)
	}
	return runGit(root, args...)
}

// RemoveWorktree removes a worktree by its path.
func RemoveWorktree(repoDir, worktreePath string, force bool) error {
	root, err := RepoCommonDir(repoDir)
	if err != nil {
		return err
	}

	args := []string{"worktree", "remove"}
	if force {
		args = append(args, "--force")
	}
	args = append(args, worktreePath)
	return runGit(root, args...)
}

// PruneWorktrees runs git worktree prune. When dryRun is true it returns the paths that would be
// pruned without acting.
func PruneWorktrees(repoDir string, dryRun bool) ([]string, error) {
	root, err := RepoCommonDir(repoDir)
	if err != nil {
		return nil, err
	}

	if dryRun {
		output, err := gitOutput(root, "worktree", "prune", "--dry-run", "-v")
		if err != nil {
			return nil, fmt.Errorf("pruning worktrees (dry run): %w", err)
		}
		return parsePruneOutput(output), nil
	}

	if err := runGit(root, "worktree", "prune"); err != nil {
		return nil, fmt.Errorf("pruning worktrees: %w", err)
	}
	return nil, nil
}

func parsePruneOutput(output string) []string {
	var paths []string
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		// git worktree prune --dry-run -v outputs lines like:
		// Removing worktrees/<name>: gitdir file points to non-existing location
		if idx := strings.Index(line, "Removing worktrees/"); idx >= 0 {
			rest := strings.TrimPrefix(line[idx:], "Removing worktrees/")
			if colon := strings.Index(rest, ":"); colon > 0 {
				paths = append(paths, rest[:colon])
			}
		}
	}
	return paths
}

// GetWorktreeStatus returns the dirty/ahead/behind state of a worktree.
func GetWorktreeStatus(worktreePath string) (WorktreeStatusInfo, error) {
	var status WorktreeStatusInfo

	output, err := gitOutput(worktreePath, "status", "--porcelain")
	if err != nil {
		return status, fmt.Errorf("checking worktree status: %w", err)
	}
	for _, line := range strings.Split(output, "\n") {
		if strings.TrimSpace(line) != "" {
			status.FilesChanged++
		}
	}
	status.Dirty = status.FilesChanged > 0

	// Ahead/behind: may fail if no upstream is configured.
	abOutput, err := gitOutput(worktreePath, "rev-list", "--left-right", "--count", "HEAD...@{upstream}")
	if err != nil {
		return status, nil //nolint:nilerr // no upstream is not an error
	}
	parts := strings.Fields(strings.TrimSpace(abOutput))
	if len(parts) == 2 {
		status.Ahead, _ = strconv.Atoi(parts[0])
		status.Behind, _ = strconv.Atoi(parts[1])
	}
	return status, nil
}

// BranchMerged reports whether branch has been merged into the repository's default branch.
func BranchMerged(repoDir, branch string) (bool, error) {
	root, err := RepoCommonDir(repoDir)
	if err != nil {
		return false, err
	}
	defaultBr, err := DefaultBranch(root)
	if err != nil {
		return false, err
	}
	output, err := gitOutput(root, "branch", "--merged", defaultBr)
	if err != nil {
		return false, nil //nolint:nilerr
	}
	for _, line := range strings.Split(output, "\n") {
		name := strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(line), "* "))
		if name == branch {
			return true, nil
		}
	}
	return false, nil
}

// DefaultBranch returns the repository's default branch name.
func DefaultBranch(repoDir string) (string, error) {
	// Try the symbolic ref that origin/HEAD points to.
	if ref, err := gitOutput(repoDir, "symbolic-ref", "--short", "refs/remotes/origin/HEAD"); err == nil {
		branch := strings.TrimSpace(ref)
		branch = strings.TrimPrefix(branch, "origin/")
		if branch != "" {
			return branch, nil
		}
	}
	// Fall back to checking if main or master exists.
	for _, candidate := range []string{"main", "master"} {
		if _, err := gitOutput(repoDir, "rev-parse", "--verify", "refs/heads/"+candidate); err == nil {
			return candidate, nil
		}
	}
	return "", fmt.Errorf("unable to determine default branch for %s", repoDir)
}

// RepoCommonDir resolves any directory inside a repository (including inside a worktree) to the
// main worktree's top-level directory.
func RepoCommonDir(dir string) (string, error) {
	commonDir, err := gitOutput(dir, "rev-parse", "--git-common-dir")
	if err != nil {
		return "", fmt.Errorf("not inside a git repository: %w", err)
	}
	commonDir = strings.TrimSpace(commonDir)
	if !filepath.IsAbs(commonDir) {
		commonDir = filepath.Join(dir, commonDir)
	}
	// --git-common-dir returns the .git directory; the repo root is its parent.
	return filepath.Clean(filepath.Dir(commonDir)), nil
}

// WorktreeTopLevel returns the top-level directory of the worktree containing dir.
// Unlike RepoCommonDir (which returns the main repo root), this returns the worktree-specific root.
func WorktreeTopLevel(dir string) (string, error) {
	output, err := gitOutput(dir, "rev-parse", "--show-toplevel")
	if err != nil {
		return "", fmt.Errorf("resolving worktree root: %w", err)
	}
	resolved := strings.TrimSpace(output)
	resolved, err = filepath.EvalSymlinks(resolved)
	if err != nil {
		return resolved, nil //nolint:nilerr
	}
	return filepath.Clean(resolved), nil
}

// WorktreeDirName converts a branch name to a filesystem-safe directory name.
func WorktreeDirName(branch string) string {
	name := strings.TrimPrefix(branch, "refs/heads/")
	name = strings.ReplaceAll(name, "/", "--")
	return name
}

// BranchExists reports whether a local branch with the given name exists.
func BranchExists(repoDir, branch string) bool {
	_, err := gitOutput(repoDir, "rev-parse", "--verify", "refs/heads/"+branch)
	return err == nil
}

// DiskUsage returns the total size in bytes of the directory tree rooted at path.
func DiskUsage(path string) (int64, error) {
	var total int64
	err := filepath.WalkDir(path, func(_ string, d fs.DirEntry, err error) error {
		if err != nil {
			if os.IsNotExist(err) {
				return nil
			}
			return err
		}
		if !d.IsDir() {
			info, err := d.Info()
			if err != nil {
				return nil //nolint:nilerr
			}
			total += info.Size()
		}
		return nil
	})
	return total, err
}

// FormatSize returns a human-readable size string.
func FormatSize(bytes int64) string {
	const (
		kb = 1024
		mb = kb * 1024
		gb = mb * 1024
	)
	switch {
	case bytes >= gb:
		return fmt.Sprintf("%.1f GB", float64(bytes)/float64(gb))
	case bytes >= mb:
		return fmt.Sprintf("%.0f MB", float64(bytes)/float64(mb))
	case bytes >= kb:
		return fmt.Sprintf("%.0f KB", float64(bytes)/float64(kb))
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}
