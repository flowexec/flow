package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strings"

	tuikitIO "github.com/flowexec/tuikit/io"
	"golang.org/x/exp/maps"
	"gopkg.in/yaml.v3"
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

	return nil
}

func (c *Config) SetDefaults() {
	if c.Workspaces == nil {
		c.Workspaces = make(map[string]string)
	}
	if c.CurrentWorkspace == "" && len(c.Workspaces) > 0 {
		c.CurrentWorkspace = maps.Keys(c.Workspaces)[0]
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

func (c *Config) CurrentWorkspaceName() (string, error) {
	var ws string
	mode := c.WorkspaceMode

	switch mode {
	case ConfigWorkspaceModeDynamic:
		wd, err := os.Getwd()
		if err != nil {
			return "", err
		}
		if runtime.GOOS == "darwin" {
			// On macOS, paths that start with /tmp (and some other system directories)
			// are actually symbolic links to paths under /private. The OS may return
			// either form of the path - e.g., both "/tmp/file" and "/private/tmp/file"
			// refer to the same location. We strip the "/private" prefix for consistent
			// path comparison, while preserving the original paths for filesystem operations.
			wd = strings.TrimPrefix(wd, "/private")
		}

		for wsName, path := range c.Workspaces {
			rel, err := filepath.Rel(filepath.Clean(path), filepath.Clean(wd))
			if err != nil {
				return "", err
			}
			if !strings.HasPrefix(rel, "..") {
				ws = wsName
				break
			}
		}
		fallthrough
	case ConfigWorkspaceModeFixed:
		if ws != "" {
			break
		}
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
	sections = append(sections, general)

	// Interactive settings
	if c.Interactive != nil { //nolint:nestif
		interactive := "## Interactive\n"
		if c.Interactive.Enabled {
			interactive += "**Enabled:** yes\n\n"
			if c.Interactive.NotifyOnCompletion != nil && *c.Interactive.NotifyOnCompletion {
				interactive += "**Notify on Completion:** yes\n\n"
			}
			if c.Interactive.SoundOnCompletion != nil && *c.Interactive.SoundOnCompletion {
				interactive += "**Sound on Completion:** yes\n\n"
			}
		} else {
			interactive += "**Enabled:** no\n\n"
		}
		sections = append(sections, interactive)
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

func (ct ConfigTheme) String() string {
	return string(ct)
}
