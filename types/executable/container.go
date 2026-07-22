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
		if _, _, _, err := v.parts(); err != nil {
			return err
		}
	}
	return nil
}

// parts splits a volume string into its host path, container path, and optional
// mount options. It validates structure but does not expand the host path.
func (v ExecContainerVolume) parts() (host, container, options string, err error) {
	segments := strings.SplitN(string(v), ":", 3)
	if len(segments) < 2 {
		return "", "", "", fmt.Errorf("invalid volume %q (expected host:container[:options])", string(v))
	}
	host = strings.TrimSpace(segments[0])
	container = strings.TrimSpace(segments[1])
	if len(segments) == 3 {
		options = strings.TrimSpace(segments[2])
	}
	if host == "" || container == "" {
		return "", "", "", fmt.Errorf("invalid volume %q (host and container paths are required)", string(v))
	}
	if !path.IsAbs(container) {
		return "", "", "", fmt.Errorf("invalid volume %q (container path must be absolute)", string(v))
	}
	return host, container, options, nil
}
