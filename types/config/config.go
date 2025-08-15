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
	if c.CurrentWorkspace == "" {
		if _, found := c.Workspaces["default"]; found {
			c.CurrentWorkspace = "default"
		} else {
			return fmt.Errorf("current workspace is not set")
		}
	}
	if _, wsFound := c.Workspaces[c.CurrentWorkspace]; !wsFound {
		return fmt.Errorf("current workspace %s does not exist", c.CurrentWorkspace)
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
	mkdwn := "# Global Configurations\n"
	mkdwn += fmt.Sprintf("**Current workspace:** `%s`\n", c.CurrentWorkspace)
	switch c.WorkspaceMode {
	case ConfigWorkspaceModeFixed:
		mkdwn += "*Workspace mode is set to fixed. This means that your working directory will have no impact on the " +
			"current workspace.*\n\n"
	case ConfigWorkspaceModeDynamic:
		mkdwn += "*Workspace mode is set to dynamic. This means that your current workspace is also determined by " +
			"your working directory.*\n\n"
	}

	if c.CurrentNamespace != "" {
		mkdwn += fmt.Sprintf("**Current namespace**: %s\n\n", c.CurrentNamespace)
	} else {
		mkdwn += "*No namespace is set*\n\n"
	}
	if c.DefaultTimeout != 0 {
		mkdwn += fmt.Sprintf("**Default timeout**: %s\n", c.DefaultTimeout)
	}
	if c.Theme != "" {
		mkdwn += fmt.Sprintf("**Theme**: %s\n", c.Theme)
	}
	if c.Interactive != nil { //nolint:nestif
		mkdwn += "## Interactivity Settings\n"
		if c.Interactive.Enabled {
			mkdwn += "**Interactive mode is enabled**\n"
			if c.Interactive.NotifyOnCompletion != nil {
				mkdwn += "*Notify on completion is enabled*\n"
			}
			if c.Interactive.SoundOnCompletion != nil {
				mkdwn += "*Sound on completion is enabled*\n"
			}
		} else {
			mkdwn += "**Interactive mode is disabled**\n"
		}
	}
	mkdwn += "## Registered Workspaces\n"
	allWs := make([]string, 0, len(c.Workspaces))
	for name := range c.Workspaces {
		allWs = append(allWs, name)
	}
	slices.Sort(allWs)
	for _, name := range allWs {
		mkdwn += fmt.Sprintf("- %s: %s\n", name, c.Workspaces[name])
	}

	if len(c.Templates) > 0 {
		mkdwn += "## Registered Templates\n"
		allTmpl := make([]string, 0, len(c.Templates))
		for name := range c.Templates {
			allTmpl = append(allTmpl, name)
		}
		slices.Sort(allTmpl)
		for _, name := range allTmpl {
			mkdwn += fmt.Sprintf("- %s: %s\n", name, c.Templates[name])
		}
	}

	return mkdwn
}

func (ct ConfigTheme) String() string {
	return string(ct)
}
