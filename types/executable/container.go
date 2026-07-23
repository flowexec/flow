package executable

import (
	"fmt"
	"path"
	"strings"
)

const (
	// DefaultContainerMount is the container-side path where the workspace root is
	// mounted when mountWorkspace is not specified.
	DefaultContainerMount = "/workspace"
	// DefaultContainerShell is the entrypoint used when entrypoint is not specified.
	DefaultContainerShell = "sh"
	// DefaultContainerRuntime is the runtime preference used when runtime is unset.
	DefaultContainerRuntime = ExecContainerRuntimeAuto
)

// SetDefaults fills in unset container fields. It is called from
// Executable.SetDefaults so that defaults are visible to display commands
// (browse/get) as well as at run time.
//
// Note: fields that are meaningful when explicitly empty (entrypoint, user,
// inheritEnv) are pointers and are intentionally left nil here so the runner can
// distinguish "unset" from "set to empty".
func (c *ExecContainer) SetDefaults() {
	if c == nil {
		return
	}
	if c.Runtime == "" {
		c.Runtime = DefaultContainerRuntime
	}
	if c.MountWorkspace == "" {
		c.MountWorkspace = DefaultContainerMount
	}
}

// EnvInherited reports whether the flow-resolved environment should be passed
// into the container. It defaults to true when inheritEnv is unset.
func (c *ExecContainer) EnvInherited() bool {
	if c == nil || c.InheritEnv == nil {
		return true
	}
	return *c.InheritEnv
}

// ResolveEntrypoint returns the configured entrypoint and whether flow should
// override the image's ENTRYPOINT. When unset, flow defaults to
// DefaultContainerShell and overrides. When explicitly set to "", flow uses the
// image's own ENTRYPOINT.
func (c *ExecContainer) ResolveEntrypoint() (entrypoint string, override bool) {
	if c == nil || c.Entrypoint == nil {
		return DefaultContainerShell, true
	}
	if *c.Entrypoint == "" {
		return "", false
	}
	return *c.Entrypoint, true
}

// Validate performs semantic validation that the JSON schema cannot express.
// It is only invoked when a container block is present.
func (c *ExecContainer) Validate() error {
	if c == nil {
		return nil
	}
	if strings.TrimSpace(c.Image) == "" {
		return fmt.Errorf("container image cannot be empty")
	}
	switch c.Runtime {
	case "", ExecContainerRuntimeAuto, ExecContainerRuntimeDocker, ExecContainerRuntimePodman:
	default:
		return fmt.Errorf("invalid container runtime %q (must be auto, docker, or podman)", c.Runtime)
	}
	if c.Workdir != "" && !path.IsAbs(c.Workdir) {
		return fmt.Errorf("container workdir must be an absolute path, got %q", c.Workdir)
	}
	if c.MountWorkspace != "" && !path.IsAbs(c.MountWorkspace) {
		return fmt.Errorf("container mountWorkspace must be an absolute path, got %q", c.MountWorkspace)
	}
	for _, v := range c.Volumes {
		if _, _, _, err := v.Parts(); err != nil {
			return err
		}
	}
	return nil
}

// Parts splits a volume string into its host path, container path, and optional
// mount options. It validates structure but does not expand the host path. A
// leading Windows drive letter (e.g. `C:\data`) is kept with the host segment so
// its colon is not mistaken for the host:container separator.
func (v ExecContainerVolume) Parts() (host, container, options string, err error) {
	s := string(v)
	var drive string
	if len(s) >= 2 && s[1] == ':' && isDriveLetter(s[0]) {
		drive, s = s[:2], s[2:]
	}
	segments := strings.SplitN(s, ":", 3)
	if len(segments) < 2 {
		return "", "", "", fmt.Errorf("invalid volume %q (expected host:container[:options])", string(v))
	}
	host = drive + strings.TrimSpace(segments[0])
	container = strings.TrimSpace(segments[1])
	if len(segments) == 3 {
		options = strings.TrimSpace(segments[2])
	}
	if strings.TrimSpace(host) == "" || container == "" {
		return "", "", "", fmt.Errorf("invalid volume %q (host and container paths are required)", string(v))
	}
	if !path.IsAbs(container) {
		return "", "", "", fmt.Errorf("invalid volume %q (container path must be absolute)", string(v))
	}
	return host, container, options, nil
}

func isDriveLetter(b byte) bool {
	return (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z')
}

// ContainerConfigMarkdown renders the container configuration as a markdown
// block (a bold label followed by a bullet list). It returns "" when no
// container is configured, so callers can conditionally append it. Effective
// values are shown, so defaults applied by SetDefaults (runtime, workspace
// mount) are visible.
func ContainerConfigMarkdown(c *ExecContainer) string {
	if c == nil {
		return ""
	}
	var b strings.Builder
	b.WriteString("**Container**\n")
	fmt.Fprintf(&b, "- Image: `%s`\n", c.Image)
	if c.Runtime != "" {
		fmt.Fprintf(&b, "- Runtime: %s\n", c.Runtime)
	}
	if c.Workdir != "" {
		fmt.Fprintf(&b, "- Workdir: `%s`\n", c.Workdir)
	}
	if c.MountWorkspace != "" {
		fmt.Fprintf(&b, "- Workspace mount: `%s`\n", c.MountWorkspace)
	}
	for _, v := range c.Volumes {
		fmt.Fprintf(&b, "- Volume: `%s`\n", string(v))
	}
	if c.User != nil {
		fmt.Fprintf(&b, "- User: `%s`\n", *c.User)
	}
	if c.Network != "" {
		fmt.Fprintf(&b, "- Network: `%s`\n", c.Network)
	}
	if c.Entrypoint != nil {
		if *c.Entrypoint == "" {
			b.WriteString("- Entrypoint: image default\n")
		} else {
			fmt.Fprintf(&b, "- Entrypoint: `%s`\n", *c.Entrypoint)
		}
	}
	if !c.EnvInherited() {
		b.WriteString("- Inherit env: false\n")
	}
	return b.String()
}
