package mcp

import (
	"context"
	"encoding/json"
	"strconv"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"github.com/flowexec/flow/v2/pkg/filesystem"
)

func addProjectTools(srv *server.MCPServer, executor CommandExecutor) {
	getProjectStatus := mcp.NewTool("get_project_status",
		mcp.WithDescription("Snapshot of the workspace's current state: git branch, dirty files, "+
			"ahead/behind, last commit, and recent run results (pass/fail). One call replaces "+
			"separate git-status + logs queries. Call it at the start of a task to orient, or "+
			"after a sequence of runs to verify everything is green."),
		mcp.WithString("dir", mcp.Description(
			"Directory to check; defaults to the workspace root. Set it to your working directory "+
				"(a worktree or fresh clone) so git state reflects that tree.")),
	)
	getProjectStatus.Annotations = mcp.ToolAnnotation{
		Title:           "Get project status",
		DestructiveHint: boolPtr(false), ReadOnlyHint: boolPtr(true),
		IdempotentHint: boolPtr(true), OpenWorldHint: boolPtr(true),
	}
	srv.AddTool(getProjectStatus, getProjectStatusHandler(executor))
}

// ProjectStatusOutput is the output of the get_project_status tool.
type ProjectStatusOutput struct {
	Git        *GitStatusOutput `json:"git,omitempty"`
	RecentRuns []RunSummary     `json:"recentRuns"`
	Workspace  string           `json:"workspace,omitempty"`
}

// GitStatusOutput summarizes git state for the workspace directory.
type GitStatusOutput struct {
	Branch       string `json:"branch"`
	Dirty        bool   `json:"dirty"`
	FilesChanged int    `json:"filesChanged"`
	Ahead        int    `json:"ahead"`
	Behind       int    `json:"behind"`
	LastCommit   string `json:"lastCommit,omitempty"`
}

// RunSummary is a compact view of one execution log entry.
type RunSummary struct {
	Ref       string `json:"ref"`
	Status    string `json:"status"`
	ExitCode  int    `json:"exitCode"`
	Duration  string `json:"duration,omitempty"`
	StartedAt string `json:"startedAt,omitempty"`
	Error     string `json:"error,omitempty"`
}

func getProjectStatusHandler(executor CommandExecutor) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		dir := request.GetString("dir", "")
		ctx = withRunDir(ctx, dir)

		output := ProjectStatusOutput{
			RecentRuns: []RunSummary{},
		}

		output.Git = collectGitStatus(ctx, executor)
		output.Workspace = collectWorkspaceName()
		output.RecentRuns = collectRecentRuns(executor)

		jsonData, _ := json.Marshal(output)
		return mcp.NewToolResultStructured(output, string(jsonData)), nil
	}
}

func collectWorkspaceName() string {
	cfg, err := filesystem.LoadConfig()
	if err != nil {
		return ""
	}
	res, err := filesystem.ResolveWorkspace(cfg, filesystem.ResolveOptions{})
	if err != nil || res == nil {
		return ""
	}
	return res.Name
}

const gitStatusScript = `branch=$(git rev-parse --abbrev-ref HEAD 2>/dev/null || echo "")
changed=$(git status --porcelain 2>/dev/null | wc -l | tr -d ' ')
ab=$(git rev-list --left-right --count HEAD...@{upstream} 2>/dev/null || echo "0	0")
commit=$(git log -1 --format=%s 2>/dev/null || echo "")
printf '%s\n%s\n%s\n%s' "$branch" "$changed" "$ab" "$commit"`

func collectGitStatus(ctx context.Context, executor CommandExecutor) *GitStatusOutput {
	out, err := executor.ExecuteContext(ctx, execCmd, "--cmd", gitStatusScript,
		"--label", "project-status-git", "--log-mode", "text")
	if err != nil {
		return nil
	}

	lines := strings.Split(strings.TrimSpace(out), "\n")
	if len(lines) < 4 {
		return nil
	}

	status := &GitStatusOutput{Branch: strings.TrimSpace(lines[0])}

	if n, err := strconv.Atoi(strings.TrimSpace(lines[1])); err == nil {
		status.FilesChanged = n
		status.Dirty = n > 0
	}

	abParts := strings.Fields(strings.TrimSpace(lines[2]))
	if len(abParts) == 2 {
		status.Ahead, _ = strconv.Atoi(abParts[0])
		status.Behind, _ = strconv.Atoi(abParts[1])
	}

	status.LastCommit = strings.TrimSpace(lines[3])
	return status
}

func collectRecentRuns(executor CommandExecutor) []RunSummary {
	logsOut, err := executor.Execute("logs", outputFlag, outputJSON)
	if err != nil {
		return nil
	}

	var logsData struct {
		History []LogEntry `json:"history"`
	}
	if json.Unmarshal([]byte(logsOut), &logsData) != nil {
		return nil
	}

	limit := 5
	if len(logsData.History) < limit {
		limit = len(logsData.History)
	}

	runs := make([]RunSummary, 0, limit)
	for _, entry := range logsData.History[:limit] {
		runs = append(runs, RunSummary{
			Ref:       entry.Ref,
			Status:    entry.Status,
			ExitCode:  entry.ExitCode,
			Duration:  entry.Duration,
			StartedAt: entry.StartedAt,
			Error:     entry.Error,
		})
	}
	return runs
}
