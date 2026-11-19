package worktree

import (
	"encoding/json"

	"gopkg.in/yaml.v3"

	"github.com/flowexec/flow/v2/internal/io/common"
	"github.com/flowexec/flow/v2/pkg/logger"
)

// DisplayInfo is the presentation-layer struct for a worktree entry.
type DisplayInfo struct {
	Branch    string `json:"branch"                 yaml:"branch"`
	Path      string `json:"path"                   yaml:"path"`
	HEAD      string `json:"head"                   yaml:"head"`
	IsMain    bool   `json:"isMain,omitempty"       yaml:"isMain,omitempty"`
	IsCurrent bool   `json:"isCurrent,omitempty"    yaml:"isCurrent,omitempty"`
	Detached  bool   `json:"detached,omitempty"     yaml:"detached,omitempty"`
	Dirty     bool   `json:"dirty"                  yaml:"dirty"`
	Changed   int    `json:"filesChanged,omitempty" yaml:"filesChanged,omitempty"`
	Ahead     int    `json:"ahead"                  yaml:"ahead"`
	Behind    int    `json:"behind"                 yaml:"behind"`
	Merged    bool   `json:"merged,omitempty"       yaml:"merged,omitempty"`
	Locked    bool   `json:"locked,omitempty"       yaml:"locked,omitempty"`
	Prunable  bool   `json:"prunable,omitempty"     yaml:"prunable,omitempty"`
	SizeBytes int64  `json:"sizeBytes"              yaml:"sizeBytes"`
	Size      string `json:"size"                   yaml:"size"`
}

// PrintWorktreeList prints the worktree list in the requested format.
func PrintWorktreeList(format string, worktrees []DisplayInfo) {
	switch common.NormalizeFormat(format) {
	case common.YAMLFormat:
		b, err := yaml.Marshal(worktrees)
		if err != nil {
			logger.Log().Fatalf("Failed to marshal worktree list - %v", err)
		}
		logger.Log().Println(string(b))
	case common.JSONFormat:
		b, err := json.Marshal(worktrees)
		if err != nil {
			logger.Log().Fatalf("Failed to marshal worktree list - %v", err)
		}
		logger.Log().Println(string(b))
	}
}
