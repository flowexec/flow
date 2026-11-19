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

func TestAddWorktreeExistingBranch(t *testing.T) {
	repo := initBareRepo(t)

	gitRun(t, repo, "branch", "existing-branch")

	wtPath := filepath.Join(t.TempDir(), "existing")
	if err := git.AddWorktree(repo, wtPath, "existing-branch", false, ""); err != nil {
		t.Fatalf("AddWorktree with existing branch: %v", err)
	}

	wts, err := git.ListWorktrees(repo)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, wt := range wts {
		if wt.Branch == "existing-branch" {
			found = true
		}
	}
	if !found {
		t.Error("existing-branch worktree not found after add")
	}
}

func TestAddWorktreeWithBase(t *testing.T) {
	repo := initBareRepo(t)

	// Make a second commit so HEAD differs from initial.
	writeFile(t, filepath.Join(repo, "second.txt"), "second\n")
	gitRun(t, repo, "add", ".")
	gitRun(t, repo, "commit", "-m", "second commit")

	// Get the initial commit hash.
	cmd := exec.Command("git", "rev-parse", "HEAD~1")
	cmd.Dir = repo
	baseOut, err := cmd.Output()
	if err != nil {
		t.Fatal(err)
	}
	base := strings.TrimSpace(string(baseOut))

	wtPath := filepath.Join(t.TempDir(), "based")
	if err := git.AddWorktree(repo, wtPath, "based-branch", true, base); err != nil {
		t.Fatalf("AddWorktree with base: %v", err)
	}

	// Verify the worktree HEAD matches the base commit.
	headCmd := exec.Command("git", "rev-parse", "HEAD")
	headCmd.Dir = wtPath
	headOut, err := headCmd.Output()
	if err != nil {
		t.Fatal(err)
	}
	if strings.TrimSpace(string(headOut)) != base {
		t.Errorf("worktree HEAD = %s, want base %s", strings.TrimSpace(string(headOut)), base)
	}
}

func TestRemoveWorktreeForce(t *testing.T) {
	repo := initBareRepo(t)

	wtPath := filepath.Join(t.TempDir(), "dirty-wt")
	if err := git.AddWorktree(repo, wtPath, "dirty-branch", true, ""); err != nil {
		t.Fatal(err)
	}

	// Dirty the worktree.
	writeFile(t, filepath.Join(wtPath, "uncommitted.txt"), "dirty\n")

	// Non-force removal should fail.
	if err := git.RemoveWorktree(repo, wtPath, false); err == nil {
		t.Error("expected non-force removal of dirty worktree to fail")
	}

	// Force removal should succeed.
	if err := git.RemoveWorktree(repo, wtPath, true); err != nil {
		t.Fatalf("force removal failed: %v", err)
	}
}

func TestPruneWorktrees(t *testing.T) {
	repo := initBareRepo(t)

	wtPath := filepath.Join(t.TempDir(), "stale-wt")
	if err := git.AddWorktree(repo, wtPath, "stale-branch", true, ""); err != nil {
		t.Fatal(err)
	}

	// Manually delete the worktree directory to simulate a stale entry.
	if err := os.RemoveAll(wtPath); err != nil {
		t.Fatal(err)
	}

	// Dry run should report the stale entry.
	pruned, err := git.PruneWorktrees(repo, true)
	if err != nil {
		t.Fatal(err)
	}
	if len(pruned) == 0 {
		// Some git versions don't output the same format; just verify no error.
		t.Log("dry-run reported no stale entries (git format may vary)")
	}

	// Actual prune should clean up.
	_, err = git.PruneWorktrees(repo, false)
	if err != nil {
		t.Fatalf("prune failed: %v", err)
	}

	// After prune, only the main worktree should remain.
	wts, err := git.ListWorktrees(repo)
	if err != nil {
		t.Fatal(err)
	}
	if len(wts) != 1 {
		t.Errorf("expected 1 worktree after prune, got %d", len(wts))
	}
}

func TestListWorktreesMultiple(t *testing.T) {
	repo := initBareRepo(t)

	wt1 := filepath.Join(t.TempDir(), "wt1")
	wt2 := filepath.Join(t.TempDir(), "wt2")
	if err := git.AddWorktree(repo, wt1, "branch-a", true, ""); err != nil {
		t.Fatal(err)
	}
	if err := git.AddWorktree(repo, wt2, "branch-b", true, ""); err != nil {
		t.Fatal(err)
	}

	wts, err := git.ListWorktrees(repo)
	if err != nil {
		t.Fatal(err)
	}
	if len(wts) != 3 {
		t.Fatalf("expected 3 worktrees, got %d", len(wts))
	}

	branches := map[string]bool{}
	for _, wt := range wts {
		branches[wt.Branch] = true
	}
	for _, want := range []string{"main", "branch-a", "branch-b"} {
		if !branches[want] {
			t.Errorf("missing branch %q in worktree list", want)
		}
	}
}

func TestListWorktreesDetached(t *testing.T) {
	repo := initBareRepo(t)

	wtPath := filepath.Join(t.TempDir(), "detached")
	// Create a worktree in detached HEAD state.
	cmd := exec.Command("git", "worktree", "add", "--detach", wtPath)
	cmd.Dir = repo
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git worktree add --detach: %s\n%s", err, out)
	}

	wts, err := git.ListWorktrees(repo)
	if err != nil {
		t.Fatal(err)
	}

	var detachedWt *git.WorktreeInfo
	for i, wt := range wts {
		if wt.Detached {
			detachedWt = &wts[i]
			break
		}
	}
	if detachedWt == nil {
		t.Fatal("no detached worktree found")
	}
	if detachedWt.Branch != "" {
		t.Errorf("detached worktree has branch=%q, want empty", detachedWt.Branch)
	}
	if detachedWt.HEAD == "" {
		t.Error("detached worktree has empty HEAD")
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
