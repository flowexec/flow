package internal

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	errhandler "github.com/flowexec/flow/v2/cmd/internal/errors"
	"github.com/flowexec/flow/v2/cmd/internal/flags"
	"github.com/flowexec/flow/v2/cmd/internal/response"
	worktreeIO "github.com/flowexec/flow/v2/internal/io/worktree"
	"github.com/flowexec/flow/v2/internal/services/git"
	"github.com/flowexec/flow/v2/pkg/context"
	"github.com/flowexec/flow/v2/pkg/logger"
)

func RegisterWorktreeCmd(ctx *context.Context, rootCmd *cobra.Command) {
	wtCmd := &cobra.Command{
		Use:     "worktree",
		Aliases: []string{"wt"},
		Short:   "Manage git worktrees for the current repository.",
		Long: `Manage git worktrees for the current repository. Worktrees are lightweight
working copies that share the same repository but can have different branches
checked out simultaneously. Each worktree is auto-discovered as a flow workspace.

Use 'worktree add' to create a new branch and worktree, 'worktree list' to see
all worktrees with their status, and 'worktree prune' to clean up stale ones.`,
	}
	registerWorktreeAddCmd(ctx, wtCmd)
	registerWorktreeListCmd(ctx, wtCmd)
	registerWorktreeRemoveCmd(ctx, wtCmd)
	registerWorktreePruneCmd(ctx, wtCmd)
	rootCmd.AddCommand(wtCmd)
}

var WorktreeDirFlag = &flags.Metadata{
	Name:    "dir",
	Usage:   "Parent directory for the new worktree. Defaults to the workspace's worktreeDir or .worktrees/.",
	Default: "",
}

var WorktreeBaseFlag = &flags.Metadata{
	Name:      "base",
	Shorthand: "b",
	Usage:     "Base branch or commit to create the new branch from. Defaults to HEAD.",
	Default:   "",
}

var NoCreateBranchFlag = &flags.Metadata{
	Name:    "no-create-branch",
	Usage:   "Use an existing branch instead of creating a new one.",
	Default: false,
}

var MergedFlag = &flags.Metadata{
	Name:    "merged",
	Usage:   "Also remove worktrees whose branches have been merged into the default branch.",
	Default: false,
}

func registerWorktreeAddCmd(ctx *context.Context, wtCmd *cobra.Command) {
	addCmd := &cobra.Command{
		Use:   "add BRANCH [PATH]",
		Short: "Create a new worktree and branch.",
		Long: `Create a new git worktree for the given branch. If the branch does not exist,
it is created from HEAD (or the ref given by --base). If PATH is omitted, the
worktree is placed under the workspace's worktreeDir (default .worktrees/).

Examples:
  flow worktree add feature/auth
  flow worktree add feature/auth --base develop
  flow worktree add bugfix/login ./tmp/login-fix
  flow worktree add release/v2 --no-create-branch`,
		Args: cobra.RangeArgs(1, 2),
		Run:  func(cmd *cobra.Command, args []string) { worktreeAddFunc(ctx, cmd, args) },
	}
	RegisterFlag(ctx, addCmd, *WorktreeDirFlag)
	RegisterFlag(ctx, addCmd, *WorktreeBaseFlag)
	RegisterFlag(ctx, addCmd, *NoCreateBranchFlag)
	RegisterFlag(ctx, addCmd, *flags.OutputFormatFlag)
	wtCmd.AddCommand(addCmd)
}

func worktreeAddFunc(ctx *context.Context, cmd *cobra.Command, args []string) {
	branch := args[0]

	cwd, err := os.Getwd()
	if err != nil {
		errhandler.HandleFatal(ctx, cmd, fmt.Errorf("unable to determine working directory: %w", err))
		return
	}

	repoRoot, err := git.RepoCommonDir(cwd)
	if err != nil {
		errhandler.HandleFatal(ctx, cmd, err)
		return
	}

	worktreePath := resolveAddPath(ctx, cmd, cwd, repoRoot, branch, args)

	base := flags.ValueFor[string](cmd, *WorktreeBaseFlag, false)
	noCreate := flags.ValueFor[bool](cmd, *NoCreateBranchFlag, false)

	createBranch := !noCreate && !git.BranchExists(repoRoot, branch)

	if err := git.AddWorktree(repoRoot, worktreePath, branch, createBranch, base); err != nil {
		errhandler.HandleFatal(ctx, cmd, fmt.Errorf("creating worktree: %w", err))
		return
	}

	ensureGitignore(repoRoot, worktreePath)

	msg := fmt.Sprintf("Worktree created at %s (branch %s)", worktreePath, branch)
	response.HandleSuccess(ctx, cmd, msg, map[string]any{
		"branch": branch,
		"path":   worktreePath,
	})
}

func registerWorktreeListCmd(ctx *context.Context, wtCmd *cobra.Command) {
	listCmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List all worktrees for the current repository.",
		Args:    cobra.NoArgs,
		PreRun:  func(cmd *cobra.Command, args []string) { StartTUI(ctx, cmd) },
		PostRun: func(cmd *cobra.Command, args []string) { WaitForTUI(ctx, cmd) },
		Run:     func(cmd *cobra.Command, args []string) { worktreeListFunc(ctx, cmd) },
	}
	RegisterFlag(ctx, listCmd, *flags.OutputFormatFlag)
	wtCmd.AddCommand(listCmd)
}

func worktreeListFunc(ctx *context.Context, cmd *cobra.Command) {
	cwd, err := os.Getwd()
	if err != nil {
		errhandler.HandleFatal(ctx, cmd, fmt.Errorf("unable to determine working directory: %w", err))
		return
	}

	repoRoot, err := git.RepoCommonDir(cwd)
	if err != nil {
		errhandler.HandleFatal(ctx, cmd, err)
		return
	}

	worktrees, err := git.ListWorktrees(repoRoot)
	if err != nil {
		errhandler.HandleFatal(ctx, cmd, err)
		return
	}

	if len(worktrees) == 0 {
		errhandler.HandleFatal(ctx, cmd, fmt.Errorf("no worktrees found"))
		return
	}

	cwdRoot, _ := git.WorktreeTopLevel(cwd)
	displayList := make([]worktreeIO.DisplayInfo, 0, len(worktrees))
	for i, wt := range worktrees {
		display := worktreeIO.DisplayInfo{
			Branch:   wt.Branch,
			Path:     wt.Path,
			HEAD:     wt.HEAD,
			IsMain:   i == 0,
			Detached: wt.Detached,
			Locked:   wt.Locked,
			Prunable: wt.Prunable,
		}
		if cwdRoot != "" && pathsEqual(cwdRoot, wt.Path) {
			display.IsCurrent = true
		}
		enrichWorktreeDisplay(&display, repoRoot)
		displayList = append(displayList, display)
	}

	// Sort: current first, then by size descending.
	sort.SliceStable(displayList, func(i, j int) bool {
		if displayList[i].IsCurrent != displayList[j].IsCurrent {
			return displayList[i].IsCurrent
		}
		if displayList[i].IsMain != displayList[j].IsMain {
			return displayList[i].IsMain
		}
		return displayList[i].SizeBytes > displayList[j].SizeBytes
	})

	if TUIEnabled(ctx, cmd) {
		view := worktreeIO.NewWorktreeListView(ctx, displayList)
		SetView(ctx, cmd, view)
	} else {
		worktreeIO.PrintWorktreeList(
			flags.ValueFor[string](cmd, *flags.OutputFormatFlag, false),
			displayList,
		)
	}
}

func enrichWorktreeDisplay(d *worktreeIO.DisplayInfo, repoRoot string) {
	if !d.Prunable {
		if status, err := git.GetWorktreeStatus(d.Path); err == nil {
			d.Dirty = status.Dirty
			d.Changed = status.FilesChanged
			d.Ahead = status.Ahead
			d.Behind = status.Behind
		}
		if bytes, err := git.DiskUsage(d.Path); err == nil {
			d.SizeBytes = bytes
			d.Size = git.FormatSize(bytes)
		}
	}
	if d.Branch != "" && !d.IsMain {
		if merged, err := git.BranchMerged(repoRoot, d.Branch); err == nil {
			d.Merged = merged
		}
	}
}

func registerWorktreeRemoveCmd(ctx *context.Context, wtCmd *cobra.Command) {
	removeCmd := &cobra.Command{
		Use:     "remove BRANCH_OR_PATH",
		Aliases: []string{"rm"},
		Short:   "Remove a worktree.",
		Long: `Remove a git worktree by branch name or path. The main worktree cannot be removed.
Use --force to remove a worktree with uncommitted changes.

Examples:
  flow worktree remove feature/auth
  flow worktree remove .worktrees/feature--auth
  flow worktree remove feature/auth --force`,
		Args: cobra.ExactArgs(1),
		Run:  func(cmd *cobra.Command, args []string) { worktreeRemoveFunc(ctx, cmd, args) },
	}
	RegisterFlag(ctx, removeCmd, *flags.ForceFlag)
	RegisterFlag(ctx, removeCmd, *flags.YesFlag)
	RegisterFlag(ctx, removeCmd, *flags.OutputFormatFlag)
	wtCmd.AddCommand(removeCmd)
}

func worktreeRemoveFunc(ctx *context.Context, cmd *cobra.Command, args []string) {
	target := args[0]

	cwd, err := os.Getwd()
	if err != nil {
		errhandler.HandleFatal(ctx, cmd, fmt.Errorf("unable to determine working directory: %w", err))
		return
	}

	repoRoot, err := git.RepoCommonDir(cwd)
	if err != nil {
		errhandler.HandleFatal(ctx, cmd, err)
		return
	}

	worktreePath, branch, err := resolveWorktreeTarget(repoRoot, target)
	if err != nil {
		errhandler.HandleFatal(ctx, cmd, err)
		return
	}

	force := flags.ValueFor[bool](cmd, *flags.ForceFlag, false)

	label := branch
	if label == "" {
		label = worktreePath
	}
	if !confirmed(ctx, cmd, fmt.Sprintf("Remove worktree '%s'?", label)) {
		return
	}

	if err := git.RemoveWorktree(repoRoot, worktreePath, force); err != nil {
		if !force {
			logger.Log().Warnf("Hint: use --force to remove a worktree with uncommitted changes")
		}
		errhandler.HandleFatal(ctx, cmd, fmt.Errorf("removing worktree: %w", err))
		return
	}

	response.HandleSuccess(ctx, cmd, fmt.Sprintf("Worktree '%s' removed", label), map[string]any{
		"branch": branch,
		"path":   worktreePath,
	})
}

func registerWorktreePruneCmd(ctx *context.Context, wtCmd *cobra.Command) {
	pruneCmd := &cobra.Command{
		Use:   "prune",
		Short: "Remove stale worktree entries and optionally clean up merged branches.",
		Long: `Clean up stale worktree entries whose directories no longer exist. With --merged,
also remove worktrees whose branches have been merged into the default branch.

Examples:
  flow worktree prune --dry-run
  flow worktree prune --merged
  flow worktree prune --merged --yes`,
		Args: cobra.NoArgs,
		Run:  func(cmd *cobra.Command, args []string) { worktreePruneFunc(ctx, cmd, args) },
	}
	RegisterFlag(ctx, pruneCmd, *flags.DryRunFlag)
	RegisterFlag(ctx, pruneCmd, *MergedFlag)
	RegisterFlag(ctx, pruneCmd, *flags.YesFlag)
	RegisterFlag(ctx, pruneCmd, *flags.OutputFormatFlag)
	wtCmd.AddCommand(pruneCmd)
}

func worktreePruneFunc(ctx *context.Context, cmd *cobra.Command, _ []string) {
	cwd, err := os.Getwd()
	if err != nil {
		errhandler.HandleFatal(ctx, cmd, fmt.Errorf("unable to determine working directory: %w", err))
		return
	}

	repoRoot, err := git.RepoCommonDir(cwd)
	if err != nil {
		errhandler.HandleFatal(ctx, cmd, err)
		return
	}

	dryRun := flags.ValueFor[bool](cmd, *flags.DryRunFlag, false)
	merged := flags.ValueFor[bool](cmd, *MergedFlag, false)

	// Discover stale entries (directories that no longer exist).
	stale, err := git.PruneWorktrees(repoRoot, true)
	if err != nil {
		errhandler.HandleFatal(ctx, cmd, err)
		return
	}

	// Discover merged worktrees.
	var mergedWorktrees []string
	if merged {
		worktrees, err := git.ListWorktrees(repoRoot)
		if err != nil {
			errhandler.HandleFatal(ctx, cmd, err)
			return
		}
		for i, wt := range worktrees {
			if i == 0 || wt.Branch == "" {
				continue // skip main worktree and detached heads
			}
			if isMerged, err := git.BranchMerged(repoRoot, wt.Branch); err == nil && isMerged {
				mergedWorktrees = append(mergedWorktrees, wt.Path)
			}
		}
	}

	if len(stale) == 0 && len(mergedWorktrees) == 0 {
		response.HandleSuccess(ctx, cmd, "Nothing to prune", nil)
		return
	}

	// Report what would be or will be removed.
	for _, s := range stale {
		logger.Log().Infof("Stale: %s", s)
	}
	for _, m := range mergedWorktrees {
		logger.Log().Infof("Merged: %s", m)
	}

	if dryRun {
		response.HandleSuccess(ctx, cmd,
			fmt.Sprintf("Would prune %d stale + %d merged worktrees", len(stale), len(mergedWorktrees)),
			map[string]any{"stale": stale, "merged": mergedWorktrees})
		return
	}

	if !confirmed(ctx, cmd, fmt.Sprintf("Prune %d stale + %d merged worktrees?", len(stale), len(mergedWorktrees))) {
		return
	}

	// Prune stale entries via git.
	if len(stale) > 0 {
		if _, err := git.PruneWorktrees(repoRoot, false); err != nil {
			errhandler.HandleFatal(ctx, cmd, fmt.Errorf("pruning stale worktrees: %w", err))
			return
		}
	}

	// Remove merged worktrees.
	var removed int
	for _, path := range mergedWorktrees {
		if err := git.RemoveWorktree(repoRoot, path, false); err != nil {
			logger.Log().Warnf("Failed to remove %s: %v", path, err)
			continue
		}
		removed++
	}

	response.HandleSuccess(ctx, cmd,
		fmt.Sprintf("Pruned %d stale + %d merged worktrees", len(stale), removed),
		map[string]any{"stale": len(stale), "merged": removed})
}

func resolveAddPath(
	ctx *context.Context, cmd *cobra.Command, cwd, repoRoot, branch string, args []string,
) string {
	if len(args) == 2 {
		p := args[1]
		if !filepath.IsAbs(p) {
			p = filepath.Join(cwd, p)
		}
		return p
	}
	parentDir := resolveWorktreeParent(ctx, cmd, cwd, repoRoot)
	return filepath.Join(parentDir, git.WorktreeDirName(branch))
}

func resolveWorktreeParent(
	ctx *context.Context, cmd *cobra.Command, cwd, repoRoot string,
) string {
	if dirOverride := flags.ValueFor[string](cmd, *WorktreeDirFlag, false); dirOverride != "" {
		if filepath.IsAbs(dirOverride) {
			return dirOverride
		}
		return filepath.Join(cwd, dirOverride)
	}
	return resolveWorktreeDir(ctx, repoRoot)
}

// resolveWorktreeDir returns the parent directory where worktrees should be created.
func resolveWorktreeDir(ctx *context.Context, repoRoot string) string {
	if ctx.CurrentWorkspace != nil && ctx.CurrentWorkspace.WorktreeDir != "" {
		dir := ctx.CurrentWorkspace.WorktreeDir
		if !filepath.IsAbs(dir) {
			dir = filepath.Join(repoRoot, dir)
		}
		return dir
	}
	return filepath.Join(repoRoot, ".worktrees")
}

// resolveWorktreeTarget resolves a branch name or path to a worktree path and branch.
func resolveWorktreeTarget(repoRoot, target string) (path, branch string, err error) {
	worktrees, err := git.ListWorktrees(repoRoot)
	if err != nil {
		return "", "", err
	}

	// Try matching by branch name first.
	for i, wt := range worktrees {
		if wt.Branch == target {
			if i == 0 {
				return "", "", fmt.Errorf("cannot remove the main worktree")
			}
			return wt.Path, wt.Branch, nil
		}
	}

	// Try matching by path (absolute or relative).
	absTarget := target
	if !filepath.IsAbs(absTarget) {
		cwd, _ := os.Getwd()
		absTarget = filepath.Join(cwd, target)
	}
	absTarget = filepath.Clean(absTarget)

	for i, wt := range worktrees {
		if filepath.Clean(wt.Path) == absTarget {
			if i == 0 {
				return "", "", fmt.Errorf("cannot remove the main worktree")
			}
			return wt.Path, wt.Branch, nil
		}
	}

	// Try matching by directory name component.
	for i, wt := range worktrees {
		if filepath.Base(wt.Path) == target || git.WorktreeDirName(wt.Branch) == target {
			if i == 0 {
				return "", "", fmt.Errorf("cannot remove the main worktree")
			}
			return wt.Path, wt.Branch, nil
		}
	}

	return "", "", fmt.Errorf("no worktree found for '%s'", target)
}

// ensureGitignore adds the worktree parent directory to .gitignore if needed.
func ensureGitignore(repoRoot, worktreePath string) {
	// Only manage .gitignore for worktrees inside the repo.
	rel, err := filepath.Rel(repoRoot, worktreePath)
	if err != nil || strings.HasPrefix(rel, "..") {
		return
	}

	// Extract the top-level directory from the relative path.
	parts := strings.SplitN(rel, string(filepath.Separator), 2)
	dirEntry := parts[0] + "/"

	gitignorePath := filepath.Join(repoRoot, ".gitignore")
	if gitignoreContains(gitignorePath, dirEntry) {
		return
	}

	needsNewline := false
	if data, err := os.ReadFile(gitignorePath); err == nil && len(data) > 0 && data[len(data)-1] != '\n' {
		needsNewline = true
	}

	f, err := os.OpenFile(gitignorePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return
	}
	defer f.Close()

	if needsNewline {
		_, _ = f.WriteString("\n")
	}
	_, _ = f.WriteString(dirEntry + "\n")
}

func pathsEqual(a, b string) bool {
	a = filepath.Clean(a)
	b = filepath.Clean(b)
	if a == b {
		return true
	}
	aReal, errA := filepath.EvalSymlinks(a)
	bReal, errB := filepath.EvalSymlinks(b)
	if errA != nil || errB != nil {
		return false
	}
	return aReal == bReal
}

func gitignoreContains(path, entry string) bool {
	f, err := os.Open(path)
	if err != nil {
		return false
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		if strings.TrimSpace(scanner.Text()) == strings.TrimSpace(entry) {
			return true
		}
	}
	return false
}
