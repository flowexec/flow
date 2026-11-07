package config

import (
	"encoding/json"
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"time"

	tuikitIO "github.com/flowexec/tuikit/io"
	"gopkg.in/yaml.v3"

	"github.com/flowexec/flow/v2/types/executable"
)

//go:generate go run github.com/atombender/go-jsonschema@v0.16.0 -et --only-models -p config -o config.gen.go schema.yaml

func (c *Config) Validate() error {
	if c.CurrentWorkspace != "" {
		if _, wsFound := c.Workspaces[c.CurrentWorkspace]; !wsFound {
			return fmt.Errorf("current workspace %s does not exist", c.CurrentWorkspace)
		}
	}
	if c.WorkspaceMode != "" &&
		c.WorkspaceMode != ConfigWorkspaceModeFixed &&
		c.WorkspaceMode != ConfigWorkspaceModeDynamic {
		return fmt.Errorf("invalid workspace mode %s", c.WorkspaceMode)
	}
	if err := c.DefaultLogMode.Validate(); err != nil {
		return err
	}
	if err := c.LogRetention.validate(); err != nil {
		return err
	}
	if err := c.Container.validate(); err != nil {
		return err
	}
	for i, v := range c.CustomVerbs {
		if err := executable.ValidateCustomVerbName(v); err != nil {
			return fmt.Errorf("invalid custom verb: %w", err)
		}
		if slices.Contains(c.CustomVerbs[:i], v) {
			return fmt.Errorf("duplicate custom verb %s", v)
		}
	}

	return nil
}

func (c *Config) SetDefaults() {
	if c.Workspaces == nil {
		c.Workspaces = make(map[string]string)
	}
	if c.CurrentWorkspace == "" && len(c.Workspaces) > 0 {
		// Sorted, not arbitrary map order: an unset current workspace would otherwise land on a
		// different workspace each run.
		c.CurrentWorkspace = slices.Sorted(maps.Keys(c.Workspaces))[0]
	}
	if c.WorkspaceMode == "" {
		c.WorkspaceMode = ConfigWorkspaceModeDynamic
	}
	if c.DefaultLogMode == "" {
		c.DefaultLogMode = tuikitIO.Logfmt
	}
}

func (c *Config) ShowTUI() bool {
	return c.Interactive != nil && c.Interactive.Enabled
}

func (c *Config) CurrentVaultName() string {
	if c.CurrentVault == nil {
		return ""
	}
	return *c.CurrentVault
}

// normalizePath cleans a path and, on macOS, strips the "/private" prefix. Paths under /tmp and
// friends are symlinks into /private there and the OS hands back either form, so both sides of
// a comparison have to be normalized or the same directory compares unequal to itself.
//
// pkg/filesystem.NormalizePath is the same function; it cannot be reused here because
// pkg/filesystem imports this package.
func normalizePath(p string) string {
	if p == "" {
		return ""
	}
	p = filepath.Clean(p)
	if runtime.GOOS == "darwin" {
		if p == "/private" {
			return "/"
		}
		p = strings.TrimPrefix(p, "/private/")
		if !strings.HasPrefix(p, "/") {
			p = "/" + p
		}
	}
	return p
}

// NameForWorkspacePath returns the registered workspace whose path is exactly path.
func (c *Config) NameForWorkspacePath(path string) (string, bool) {
	target := normalizePath(path)
	if target == "" {
		return "", false
	}
	for _, name := range slices.Sorted(maps.Keys(c.Workspaces)) {
		if normalizePath(c.Workspaces[name]) == target {
			return name, true
		}
	}
	return "", false
}

// WorkspaceForPath returns the registered workspace containing dir, preferring the longest
// matching path so a workspace nested inside another wins. Iteration is over sorted names
// because Go map order would otherwise make ties nondeterministic between runs.
func (c *Config) WorkspaceForPath(dir string) (string, bool) {
	target := normalizePath(dir)
	if target == "" {
		return "", false
	}

	var bestName, bestPath string
	for _, name := range slices.Sorted(maps.Keys(c.Workspaces)) {
		path := normalizePath(c.Workspaces[name])
		if path == "" {
			continue
		}
		rel, err := filepath.Rel(path, target)
		if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
			continue
		}
		if bestName == "" || len(path) > len(bestPath) {
			bestName, bestPath = name, path
		}
	}
	return bestName, bestName != ""
}

func (c *Config) CurrentWorkspaceName() (string, error) {
	var ws string
	if c.WorkspaceMode == ConfigWorkspaceModeDynamic {
		wd, err := os.Getwd()
		if err != nil {
			return "", err
		}
		ws, _ = c.WorkspaceForPath(wd)
	}
	if ws == "" {
		ws = c.CurrentWorkspace
	}
	if ws == "" {
		return "", fmt.Errorf("current workspace not found")
	}

	return ws, nil
}

func (c *Config) SendTextNotification() bool {
	return c.Interactive != nil && c.Interactive.Enabled &&
		c.Interactive.NotifyOnCompletion != nil && *c.Interactive.NotifyOnCompletion
}

func (c *Config) SendSoundNotification() bool {
	return c.Interactive != nil && c.Interactive.Enabled &&
		c.Interactive.SoundOnCompletion != nil && *c.Interactive.SoundOnCompletion
}

// MouseEnabled reports whether the interactive UI should respond to the mouse. It is on unless
// explicitly disabled.
func (c *Config) MouseEnabled() bool {
	return c.Interactive == nil || c.Interactive.Mouse == nil || *c.Interactive.Mouse
}

func (c *Config) YAML() (string, error) {
	yamlBytes, err := yaml.Marshal(c)
	if err != nil {
		return "", fmt.Errorf("failed to marshal user config - %w", err)
	}
	return string(yamlBytes), nil
}

func (c *Config) JSON() (string, error) {
	jsonBytes, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal user config - %w", err)
	}
	return string(jsonBytes), nil
}

func (c *Config) Markdown() string {
	var sections []string

	// General settings
	general := "## General\n"
	general += fmt.Sprintf("**Workspace:** `%s`\n\n", c.CurrentWorkspace)
	if c.CurrentNamespace != "" {
		general += fmt.Sprintf("**Namespace:** `%s`\n\n", c.CurrentNamespace)
	}

	mode := string(c.WorkspaceMode)
	if mode == "" {
		mode = "dynamic"
	}
	general += fmt.Sprintf("**Workspace Mode:** %s\n\n", mode)

	if c.Theme != "" {
		general += fmt.Sprintf("**Theme:** %s\n\n", c.Theme)
	}
	if c.DefaultTimeout != 0 {
		general += fmt.Sprintf("**Default Timeout:** %s\n\n", c.DefaultTimeout)
	}
	if c.DefaultLogMode != "" {
		general += fmt.Sprintf("**Log Mode:** %s\n\n", c.DefaultLogMode)
	}
	general += fmt.Sprintf("**Log Retention:** %s\n\n", c.LogRetentionSummary())
	if c.Container != nil {
		general += fmt.Sprintf("**Containers:** %s\n\n", c.ContainerSummary())
	}
	if len(c.CustomVerbs) > 0 {
		general += fmt.Sprintf("**Custom Verbs:** %s\n\n", strings.Join(c.CustomVerbs, ", "))
	}
	sections = append(sections, general)

	if c.Interactive != nil {
		sections = append(sections, c.interactiveMarkdown())
	}

	// Workspaces
	if len(c.Workspaces) > 0 {
		ws := fmt.Sprintf("## Workspaces (%d)\n", len(c.Workspaces))
		allWs := make([]string, 0, len(c.Workspaces))
		for name := range c.Workspaces {
			allWs = append(allWs, name)
		}
		slices.Sort(allWs)
		for _, name := range allWs {
			ws += fmt.Sprintf("- **%s** — %s\n", name, c.Workspaces[name])
		}
		sections = append(sections, ws)
	}

	// Templates
	if len(c.Templates) > 0 {
		tmpl := fmt.Sprintf("## Templates (%d)\n", len(c.Templates))
		allTmpl := make([]string, 0, len(c.Templates))
		for name := range c.Templates {
			allTmpl = append(allTmpl, name)
		}
		slices.Sort(allTmpl)
		for _, name := range allTmpl {
			tmpl += fmt.Sprintf("- **%s** — %s\n", name, c.Templates[name])
		}
		sections = append(sections, tmpl)
	}

	return strings.Join(sections, "\n")
}

const (
	DefaultLogRetentionMaxAge    = 30 * 24 * time.Hour
	DefaultLogRetentionMaxSizeMB = 500
	DefaultLogRetentionMaxCount  = 1000
)

// LogRetentionPolicy returns the archive limits to apply to run logs, filling unset fields with
// their defaults.
func (c *Config) LogRetentionPolicy() tuikitIO.RetentionPolicy {
	policy := tuikitIO.RetentionPolicy{
		MaxFiles:     DefaultLogRetentionMaxCount,
		MaxAge:       DefaultLogRetentionMaxAge,
		MaxTotalSize: DefaultLogRetentionMaxSizeMB * bytesPerMB,
	}
	if c == nil || c.LogRetention == nil {
		return policy
	}
	if r := c.LogRetention; r.MaxCount != nil {
		policy.MaxFiles = *r.MaxCount
	}
	if r := c.LogRetention; r.MaxAge != nil {
		policy.MaxAge = *r.MaxAge
	}
	if r := c.LogRetention; r.MaxSizeMB != nil {
		policy.MaxTotalSize = int64(*r.MaxSizeMB) * bytesPerMB
	}
	return policy
}

const bytesPerMB = 1024 * 1024

func (r *LogRetention) validate() error {
	if r == nil {
		return nil
	}
	if r.MaxAge != nil && *r.MaxAge < 0 {
		return fmt.Errorf("log retention maxAge must not be negative")
	}
	if r.MaxSizeMB != nil && *r.MaxSizeMB < 0 {
		return fmt.Errorf("log retention maxSizeMB must not be negative")
	}
	if r.MaxCount != nil && *r.MaxCount < 0 {
		return fmt.Errorf("log retention maxCount must not be negative")
	}
	return nil
}

// LogRetentionSummary describes the effective log retention limits.
func (c *Config) LogRetentionSummary() string {
	p := c.LogRetentionPolicy()
	limit := func(v int64, text string) string {
		if v <= 0 {
			return "unlimited"
		}
		return text
	}
	return fmt.Sprintf("max age %s, max size %s, max count %s",
		limit(int64(p.MaxAge), p.MaxAge.String()),
		limit(p.MaxTotalSize, fmt.Sprintf("%d MB", p.MaxTotalSize/bytesPerMB)),
		limit(int64(p.MaxFiles), fmt.Sprintf("%d", p.MaxFiles)))
}

func (ct ConfigTheme) String() string {
	return string(ct)
}

func (c *Config) interactiveMarkdown() string {
	interactive := "## Interactive\n"
	if !c.Interactive.Enabled {
		return interactive + "**Enabled:** no\n\n"
	}
	interactive += "**Enabled:** yes\n\n"
	if c.Interactive.NotifyOnCompletion != nil && *c.Interactive.NotifyOnCompletion {
		interactive += "**Notify on Completion:** yes\n\n"
	}
	if c.Interactive.SoundOnCompletion != nil && *c.Interactive.SoundOnCompletion {
		interactive += "**Sound on Completion:** yes\n\n"
	}
	if !c.MouseEnabled() {
		interactive += "**Mouse:** no\n\n"
	}
	return interactive
}

// MarshalJSON writes durations the way the config file and its schema spell them ("5m0s"),
// not as the nanosecond integers encoding/json gives a time.Duration.
func (c Config) MarshalJSON() ([]byte, error) {
	type plain Config
	var timeout string
	if c.DefaultTimeout != 0 {
		timeout = c.DefaultTimeout.String()
	}
	return json.Marshal(struct {
		plain
		DefaultTimeout string `json:"defaultTimeout,omitempty"`
	}{plain(c), timeout})
}

// MarshalJSON writes maxAge as a duration string; zero, meaning no limit, stays "0s".
func (r LogRetention) MarshalJSON() ([]byte, error) {
	type plain LogRetention
	var maxAge *string
	if r.MaxAge != nil {
		s := r.MaxAge.String()
		maxAge = &s
	}
	return json.Marshal(struct {
		plain
		MaxAge *string `json:"maxAge,omitempty"`
	}{plain(r), maxAge})
}
