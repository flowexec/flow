//nolint:nilerr
package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func addWorktreeTools(srv *server.MCPServer, executor CommandExecutor) {
	listWorktrees := mcp.NewTool("list_worktrees",
		mcp.WithDescription("List all git worktrees for the current repository with status, size, "+
			"and merge info. Use branch_filter to narrow results by substring match on branch name."),
		mcp.WithString("branch_filter", mcp.Description(
			"Optional substring to filter worktrees by branch name (e.g. 'feature' or 'worktree-')")),
		mcp.WithString("cursor", mcp.Description("Next-page cursor")),
	)
	listWorktrees.Annotations = mcp.ToolAnnotation{
		Title:           "List worktrees",
		DestructiveHint: boolPtr(false), ReadOnlyHint: boolPtr(true),
		IdempotentHint: boolPtr(true), OpenWorldHint: boolPtr(true),
	}
	srv.AddTool(listWorktrees, listWorktreesHandler(executor))

	addWorktree := mcp.NewTool("add_worktree",
		mcp.WithDescription("Create a new git worktree and branch. Returns the absolute path of the "+
			"new worktree, which can be passed as 'dir' to execute or run_command."),
		mcp.WithString("branch", mcp.Required(), mcp.Description(
			"Branch name for the worktree. Created automatically if it does not exist.")),
		mcp.WithString("base", mcp.Description(
			"Base branch or commit to create the new branch from (defaults to HEAD)")),
		mcp.WithString("dir", mcp.Description(
			"Parent directory for the worktree (defaults to the workspace's worktreeDir or .worktrees/)")),
	)
	addWorktree.Annotations = mcp.ToolAnnotation{
		Title:           "Create a worktree",
		DestructiveHint: boolPtr(true), ReadOnlyHint: boolPtr(false),
		IdempotentHint: boolPtr(false), OpenWorldHint: boolPtr(false),
	}
	srv.AddTool(addWorktree, addWorktreeHandler(executor))

	removeWorktree := mcp.NewTool("remove_worktree",
		mcp.WithDescription("Remove a git worktree by branch name or path. The main worktree cannot be removed."),
		mcp.WithString("branch_or_path", mcp.Required(), mcp.Description(
			"Branch name or filesystem path of the worktree to remove")),
		mcp.WithBoolean("force", mcp.Description(
			"Remove even if the worktree has uncommitted changes")),
	)
	removeWorktree.Annotations = mcp.ToolAnnotation{
		Title:           "Remove a worktree",
		DestructiveHint: boolPtr(true), ReadOnlyHint: boolPtr(false),
		IdempotentHint: boolPtr(false), OpenWorldHint: boolPtr(false),
	}
	srv.AddTool(removeWorktree, removeWorktreeHandler(executor))

	pruneWorktrees := mcp.NewTool("prune_worktrees",
		mcp.WithDescription("Clean up stale worktree entries (directories that no longer exist on disk). "+
			"With merged=true, also remove worktrees whose branches are fully merged into the default branch. "+
			"Set dry_run to preview what would be removed without acting."),
		mcp.WithBoolean("merged", mcp.Description(
			"Also remove worktrees whose branches are fully merged into the default branch")),
		mcp.WithBoolean("dry_run", mcp.Description(
			"Preview what would be pruned without removing anything (default false)")),
	)
	pruneWorktrees.Annotations = mcp.ToolAnnotation{
		Title:           "Prune stale and merged worktrees",
		DestructiveHint: boolPtr(true), ReadOnlyHint: boolPtr(false),
		IdempotentHint: boolPtr(false), OpenWorldHint: boolPtr(false),
	}
	srv.AddTool(pruneWorktrees, pruneWorktreesHandler(executor))
}

type worktreeEntry struct {
	Branch    string `json:"branch"`
	Path      string `json:"path"`
	HEAD      string `json:"head"`
	IsMain    bool   `json:"isMain,omitempty"`
	IsCurrent bool   `json:"isCurrent,omitempty"`
	Detached  bool   `json:"detached,omitempty"`
	Dirty     bool   `json:"dirty"`
	Changed   int    `json:"filesChanged,omitempty"`
	Ahead     int    `json:"ahead"`
	Behind    int    `json:"behind"`
	Merged    bool   `json:"merged,omitempty"`
	SizeBytes int64  `json:"sizeBytes"`
	Size      string `json:"size"`
}

type worktreeListOutput struct {
	Worktrees  []worktreeEntry `json:"worktrees"`
	NextCursor string          `json:"nextCursor,omitempty"`
	TotalCount int             `json:"totalCount"`
}

func listWorktreesHandler(executor CommandExecutor) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		branchFilter := request.GetString("branch_filter", "")
		cursor := request.GetString("cursor", "")

		output, err := executor.Execute(worktreeCmd, "list", outputFlag, outputJSON)
		if err != nil {
			return toolError(ErrCodeExecutionFailed, fmt.Sprintf("Failed to list worktrees: %s", output)), nil
		}

		var entries []worktreeEntry
		if err := json.Unmarshal([]byte(output), &entries); err != nil {
			return mcp.NewToolResultText(output), nil
		}

		// Apply branch filter.
		if branchFilter != "" {
			filtered := make([]worktreeEntry, 0)
			for _, e := range entries {
				if strings.Contains(e.Branch, branchFilter) || strings.Contains(e.Path, branchFilter) {
					filtered = append(filtered, e)
				}
			}
			entries = filtered
		}

		page, nextCursor, totalCount, err := paginate(entries, cursor, defaultPageSize)
		if err != nil {
			return toolError(ErrCodeInvalidInput, err.Error()), nil
		}

		result := worktreeListOutput{
			Worktrees:  page,
			NextCursor: nextCursor,
			TotalCount: totalCount,
		}
		jsonData, _ := json.Marshal(result)
		return mcp.NewToolResultStructured(result, string(jsonData)), nil
	}
}

type successEnvelope struct {
	Result struct {
		Message string         `json:"message"`
		Data    map[string]any `json:"data"`
	} `json:"result"`
}

type addWorktreeOutput struct {
	Branch string `json:"branch"`
	Path   string `json:"path"`
}

func addWorktreeHandler(executor CommandExecutor) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		branch, err := request.RequireString("branch")
		if err != nil {
			return toolError(ErrCodeInvalidInput, "branch is required"), nil
		}

		args := []string{worktreeCmd, "add", branch, outputFlag, outputJSON}

		if base := request.GetString("base", ""); base != "" {
			args = append(args, "--base", base)
		}
		if dir := request.GetString("dir", ""); dir != "" {
			args = append(args, "--dir", dir)
		}

		output, err := executor.Execute(args...)
		if err != nil {
			return toolError(ErrCodeExecutionFailed, fmt.Sprintf("Failed to create worktree: %s", output)), nil
		}

		result, err := extractEnvelopeData[addWorktreeOutput](output)
		if err != nil {
			return mcp.NewToolResultText(output), nil
		}

		jsonData, _ := json.Marshal(result)
		return mcp.NewToolResultStructured(result, string(jsonData)), nil
	}
}

type removeWorktreeOutput struct {
	Branch string `json:"branch"`
	Path   string `json:"path"`
}

func removeWorktreeHandler(executor CommandExecutor) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		target, err := request.RequireString("branch_or_path")
		if err != nil {
			return toolError(ErrCodeInvalidInput, "branch_or_path is required"), nil
		}

		args := []string{worktreeCmd, "remove", target, "--yes", outputFlag, outputJSON}

		if request.GetBool("force", false) {
			args = append(args, "--force")
		}

		output, err := executor.Execute(args...)
		if err != nil {
			return toolError(ErrCodeExecutionFailed, fmt.Sprintf("Failed to remove worktree: %s", output)), nil
		}

		result, err := extractEnvelopeData[removeWorktreeOutput](output)
		if err != nil {
			return mcp.NewToolResultText(output), nil
		}

		jsonData, _ := json.Marshal(result)
		return mcp.NewToolResultStructured(result, string(jsonData)), nil
	}
}

type pruneWorktreesOutput struct {
	Stale   any    `json:"stale"`
	Merged  any    `json:"merged"`
	Message string `json:"message"`
	DryRun  bool   `json:"dryRun"`
}

func pruneWorktreesHandler(executor CommandExecutor) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		merged := request.GetBool("merged", false)
		dryRun := request.GetBool("dry_run", false)

		args := []string{worktreeCmd, "prune", "--yes", outputFlag, outputJSON}
		if merged {
			args = append(args, "--merged")
		}
		if dryRun {
			args = append(args, "--dry-run")
		}

		output, err := executor.Execute(args...)
		if err != nil {
			return toolError(ErrCodeExecutionFailed, fmt.Sprintf("Failed to prune worktrees: %s", output)), nil
		}

		var env successEnvelope
		if err := json.Unmarshal([]byte(output), &env); err != nil {
			return mcp.NewToolResultText(output), nil
		}

		result := pruneWorktreesOutput{
			Message: env.Result.Message,
			DryRun:  dryRun,
			Stale:   env.Result.Data["stale"],
			Merged:  env.Result.Data["merged"],
		}
		jsonData, _ := json.Marshal(result)
		return mcp.NewToolResultStructured(result, string(jsonData)), nil
	}
}

func extractEnvelopeData[T any](output string) (T, error) {
	var env successEnvelope
	var zero T
	if err := json.Unmarshal([]byte(output), &env); err != nil {
		return zero, err
	}
	dataBytes, err := json.Marshal(env.Result.Data)
	if err != nil {
		return zero, err
	}
	var result T
	if err := json.Unmarshal(dataBytes, &result); err != nil {
		return zero, err
	}
	return result, nil
}
