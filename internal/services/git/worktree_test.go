package git_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/flowexec/flow/v2/internal/services/git"
)

func initBareRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	gitRun(t, dir, "init", "--initial-branch=main")
	gitRun(t, dir, "config", "user.email", "test@test.com")
	gitRun(t, dir, "config", "user.name", "Test")
	writeFile(t, filepath.Join(dir, "README.md"), "# test\n")
	gitRun(t, dir, "add", ".")
	gitRun(t, dir, "commit", "-m", "initial")
	return dir
}

func gitRun(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v failed: %s\n%s", args, err, out)
	}
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

// parseWorktreeList is not exported, so we test it via ListWorktrees with real repos.

func TestWorktreeDirName(t *testing.T) {
	tests := []struct {
		branch string
		want   string
	}{
		{"feature/x", "feature--x"},
		{"main", "main"},
		{"refs/heads/bug/123", "bug--123"},
		{"deep/nested/branch", "deep--nested--branch"},
	}
	for _, tc := range tests {
		got := git.WorktreeDirName(tc.branch)
		if got != tc.want {
			t.Errorf("WorktreeDirName(%q) = %q, want %q", tc.branch, got, tc.want)
		}
	}
}

func TestFormatSize(t *testing.T) {
	tests := []struct {
		bytes int64
		want  string
	}{
		{500, "500 B"},
		{2048, "2 KB"},
		{1048576, "1 MB"},
		{1073741824, "1.0 GB"},
		{1572864, "2 MB"},
	}
	for _, tc := range tests {
		got := git.FormatSize(tc.bytes)
		if got != tc.want {
			t.Errorf("FormatSize(%d) = %q, want %q", tc.bytes, got, tc.want)
		}
	}
}

func TestListWorktrees(t *testing.T) {
	repo := initBareRepo(t)
	wts, err := git.ListWorktrees(repo)
	if err != nil {
		t.Fatal(err)
	}
	if len(wts) != 1 {
		t.Fatalf("expected 1 worktree, got %d", len(wts))
	}
	if wts[0].Branch != "main" {
		t.Errorf("expected main branch, got %q", wts[0].Branch)
	}
}

func TestAddAndRemoveWorktree(t *testing.T) {
	repo := initBareRepo(t)

	wtPath := filepath.Join(t.TempDir(), "feature--test")
	if err := git.AddWorktree(repo, wtPath, "feature/test", true, ""); err != nil {
		t.Fatalf("AddWorktree: %v", err)
	}

	if _, err := os.Stat(wtPath); os.IsNotExist(err) {
		t.Fatal("worktree directory not created")
	}

	wts, err := git.ListWorktrees(repo)
	if err != nil {
		t.Fatal(err)
	}
	if len(wts) != 2 {
		t.Fatalf("expected 2 worktrees after add, got %d", len(wts))
	}

	found := false
	for _, wt := range wts {
		if wt.Branch == "feature/test" {
			found = true
			break
		}
	}
	if !found {
		t.Error("feature/test worktree not found in list")
	}

	if err := git.RemoveWorktree(repo, wtPath, false); err != nil {
		t.Fatalf("RemoveWorktree: %v", err)
	}

	wts, err = git.ListWorktrees(repo)
	if err != nil {
		t.Fatal(err)
	}
	if len(wts) != 1 {
		t.Fatalf("expected 1 worktree after remove, got %d", len(wts))
	}
}

func TestRepoCommonDir(t *testing.T) {
	repo := initBareRepo(t)

	wtPath := filepath.Join(t.TempDir(), "wt")
	if err := git.AddWorktree(repo, wtPath, "test-common", true, ""); err != nil {
		t.Fatal(err)
	}

	common, err := git.RepoCommonDir(wtPath)
	if err != nil {
		t.Fatal(err)
	}
	repoReal, _ := filepath.EvalSymlinks(repo)
	commonReal, _ := filepath.EvalSymlinks(common)
	if repoReal != commonReal {
		t.Errorf("RepoCommonDir(%q) = %q, want %q", wtPath, commonReal, repoReal)
	}
}

func TestBranchExists(t *testing.T) {
	repo := initBareRepo(t)

	if !git.BranchExists(repo, "main") {
		t.Error("expected main branch to exist")
	}
	if git.BranchExists(repo, "nonexistent") {
		t.Error("expected nonexistent branch to not exist")
	}
}

func TestDiskUsage(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "a.txt"), strings.Repeat("x", 1000))
	writeFile(t, filepath.Join(dir, "b.txt"), strings.Repeat("y", 500))
	if err := os.MkdirAll(filepath.Join(dir, "sub"), 0o755); err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(dir, "sub", "c.txt"), strings.Repeat("z", 200))

	size, err := git.DiskUsage(dir)
	if err != nil {
		t.Fatal(err)
	}
	if size != 1700 {
		t.Errorf("DiskUsage = %d, want 1700", size)
	}
}

func TestDefaultBranch(t *testing.T) {
	repo := initBareRepo(t)
	branch, err := git.DefaultBranch(repo)
	if err != nil {
		t.Fatal(err)
	}
	if branch != "main" {
		t.Errorf("DefaultBranch = %q, want main", branch)
	}
}

func TestBranchMerged(t *testing.T) {
	repo := initBareRepo(t)

	gitRun(t, repo, "checkout", "-b", "merged-branch")
	writeFile(t, filepath.Join(repo, "merged.txt"), "merged\n")
	gitRun(t, repo, "add", ".")
	gitRun(t, repo, "commit", "-m", "on merged branch")
	gitRun(t, repo, "checkout", "main")
	gitRun(t, repo, "merge", "--no-ff", "merged-branch", "-m", "merge")

	merged, err := git.BranchMerged(repo, "merged-branch")
	if err != nil {
		t.Fatal(err)
	}
	if !merged {
		t.Error("expected merged-branch to be merged")
	}

	gitRun(t, repo, "checkout", "-b", "unmerged-branch")
	writeFile(t, filepath.Join(repo, "unmerged.txt"), "unmerged\n")
	gitRun(t, repo, "add", ".")
	gitRun(t, repo, "commit", "-m", "on unmerged branch")
	gitRun(t, repo, "checkout", "main")

	merged, err = git.BranchMerged(repo, "unmerged-branch")
	if err != nil {
		t.Fatal(err)
	}
	if merged {
		t.Error("expected unmerged-branch to not be merged")
	}
}

func TestGetWorktreeStatus(t *testing.T) {
	repo := initBareRepo(t)

	status, err := git.GetWorktreeStatus(repo)
	if err != nil {
		t.Fatal(err)
	}
	if status.Dirty {
		t.Error("expected clean repo")
	}

	writeFile(t, filepath.Join(repo, "new.txt"), "dirty\n")
	status, err = git.GetWorktreeStatus(repo)
	if err != nil {
		t.Fatal(err)
	}
	if !status.Dirty || status.FilesChanged != 1 {
		t.Errorf("expected dirty with 1 changed, got dirty=%v changed=%d", status.Dirty, status.FilesChanged)
	}
}
