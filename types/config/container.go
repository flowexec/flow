package config

import (
	"fmt"
	"slices"
	"strings"
)

// ContainerRuntimePreference is the runtime an executable set to `auto` uses on this machine:
// "auto" when the config does not choose one.
func (c *Config) ContainerRuntimePreference() ContainerRuntime {
	if c == nil || c.Container == nil || c.Container.Runtime == "" {
		return ContainerRuntimeAuto
	}
	return c.Container.Runtime
}

// ContainerCommand returns the command and global flags this machine runs for a runtime family:
// the configured replacement, or the family name with no flags.
func (c *Config) ContainerCommand(family ContainerRuntime) (string, []string) {
	cmd := c.containerRuntimeCommand(family)
	if cmd == nil || cmd.Command == "" {
		var args []string
		if cmd != nil {
			args = cmd.Args
		}
		return string(family), args
	}
	return cmd.Command, cmd.Args
}

// SetContainerCommand sets the command and global flags for a runtime family. Setting the family's
// own name with no flags removes the entry, since that is what runs without one.
func (c *Config) SetContainerCommand(family ContainerRuntime, command string, args []string) error {
	if err := validateFamily(family); err != nil {
		return err
	}
	command = strings.TrimSpace(command)
	if command == "" {
		return fmt.Errorf("container command for %s cannot be empty", family)
	}
	if c.Container == nil {
		c.Container = &Container{}
	}
	if c.Container.Runtimes == nil {
		c.Container.Runtimes = &ContainerRuntimes{}
	}
	var entry *ContainerRuntimeCommand
	if command != string(family) || len(args) > 0 {
		entry = &ContainerRuntimeCommand{Command: command, Args: args}
		if command == string(family) {
			entry.Command = ""
		}
	}
	switch family {
	case ContainerRuntimeDocker:
		c.Container.Runtimes.Docker = entry
	case ContainerRuntimePodman:
		c.Container.Runtimes.Podman = entry
	}
	if c.Container.Runtimes.Docker == nil && c.Container.Runtimes.Podman == nil {
		c.Container.Runtimes = nil
	}
	return nil
}

func (c *Config) containerRuntimeCommand(family ContainerRuntime) *ContainerRuntimeCommand {
	if c == nil || c.Container == nil || c.Container.Runtimes == nil {
		return nil
	}
	switch family {
	case ContainerRuntimeDocker:
		return c.Container.Runtimes.Docker
	case ContainerRuntimePodman:
		return c.Container.Runtimes.Podman
	default:
		return nil
	}
}

func validateFamily(family ContainerRuntime) error {
	switch family {
	case ContainerRuntimeDocker, ContainerRuntimePodman:
		return nil
	default:
		return fmt.Errorf("invalid container runtime %q (must be docker or podman)", family)
	}
}

// Clone returns a deep copy, or nil for nil.
func (c *Container) Clone() *Container {
	if c == nil {
		return nil
	}
	out := &Container{Runtime: c.Runtime}
	if c.Runtimes != nil {
		out.Runtimes = &ContainerRuntimes{Docker: c.Runtimes.Docker.clone(), Podman: c.Runtimes.Podman.clone()}
	}
	return out
}

func (c *ContainerRuntimeCommand) clone() *ContainerRuntimeCommand {
	if c == nil {
		return nil
	}
	return &ContainerRuntimeCommand{Command: c.Command, Args: slices.Clone(c.Args)}
}

func (c *Container) validate() error {
	if c == nil {
		return nil
	}
	switch c.Runtime {
	case "", ContainerRuntimeAuto, ContainerRuntimeDocker, ContainerRuntimePodman:
	default:
		return fmt.Errorf("invalid container runtime %q (must be auto, docker, or podman)", c.Runtime)
	}
	if c.Runtimes == nil {
		return nil
	}
	for family, cmd := range map[ContainerRuntime]*ContainerRuntimeCommand{
		ContainerRuntimeDocker: c.Runtimes.Docker,
		ContainerRuntimePodman: c.Runtimes.Podman,
	} {
		if cmd == nil {
			continue
		}
		for _, arg := range cmd.Args {
			if strings.TrimSpace(arg) == "" {
				return fmt.Errorf("container args for %s cannot contain an empty flag", family)
			}
		}
	}
	return nil
}

// ContainerSummary describes how container executables run on this machine.
func (c *Config) ContainerSummary() string {
	parts := []string{"runtime " + string(c.ContainerRuntimePreference())}
	for _, family := range []ContainerRuntime{ContainerRuntimeDocker, ContainerRuntimePodman} {
		if c.containerRuntimeCommand(family) == nil {
			continue
		}
		cmd, args := c.ContainerCommand(family)
		parts = append(parts, fmt.Sprintf("%s runs `%s`", family, strings.TrimSpace(cmd+" "+strings.Join(args, " "))))
	}
	return strings.Join(parts, ", ")
}
