package run

import (
	stdctx "context"
	"fmt"
	"os"
	osexec "os/exec"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/flowexec/tuikit/io"

	flowerrors "github.com/flowexec/flow/v2/pkg/errors"
)

// Mount is a single bind mount passed to the container runtime.
type Mount struct {
	HostPath      string
	ContainerPath string
	ReadOnly      bool
	Options       string // extra options carried through from a user-supplied volume string
}

// ContainerSpec fully describes a containerized execution. All host->container
// translation is expected to be done by the caller; this package only builds and
// runs the runtime invocation.
type ContainerSpec struct {
	Runtime       string // "docker" or "podman"; already resolved (never "auto")
	Image         string
	Name          string            // container name for cleanup/identification
	Labels        map[string]string // emitted in sorted key order
	Workdir       string            // container-side working directory
	Mounts        []Mount           // emitted in slice order
	Env           map[string]string // container env; empty => no --env-file
	User          string
	Network       string
	Entrypoint    string
	OverrideEntry bool
	Cmd           string // XOR Script
	Script        string // container-side path to a script; XOR Cmd
}

// runtime detection cache for the "auto" preference. Only a successful detection
// is memoized: a negative result must not be cached, otherwise a long-lived
// process (e.g. `flow mcp`) that starts before the runtime is available would
// keep failing even after the runtime comes up.
var (
	lookPath      = osexec.LookPath // test seam
	autoRuntimeMu sync.Mutex
	autoRuntime   string
)

// ResolveRuntime resolves a runtime preference ("auto", "docker", "podman", or
// "") to a concrete runtime binary available on the PATH.
func ResolveRuntime(pref string) (string, error) {
	switch pref {
	case "docker", "podman":
		if _, err := lookPath(pref); err != nil {
			return "", flowerrors.NewContainerRuntimeError(pref,
				fmt.Errorf("%s is not installed or not in PATH", pref))
		}
		return pref, nil
	case "", "auto":
		autoRuntimeMu.Lock()
		defer autoRuntimeMu.Unlock()
		if autoRuntime != "" {
			return autoRuntime, nil
		}
		for _, rt := range []string{"docker", "podman"} {
			if _, err := lookPath(rt); err == nil {
				autoRuntime = rt
				return rt, nil
			}
		}
		return "", flowerrors.NewContainerRuntimeError("",
			fmt.Errorf("no container runtime found: install docker or podman, "+
				"or set exec.container.runtime"))
	default:
		return "", flowerrors.NewContainerRuntimeError(pref,
			fmt.Errorf("unknown container runtime %q", pref))
	}
}

// RunContainer runs the command or script described by spec inside a container.
func RunContainer(
	ctx stdctx.Context,
	spec ContainerSpec,
	logMode io.LogMode,
	logger io.Logger,
	stdIn *os.File,
	logFields map[string]interface{},
	task *io.TaskContext,
) error {
	var envFile string
	if len(spec.Env) > 0 {
		f, err := writeEnvFile(spec.Env)
		if err != nil {
			return err
		}
		envFile = f
		defer func() { _ = os.Remove(envFile) }()
	}

	args := buildRunArgs(spec, envFile)
	logger.Debugf("running container (%s):\n%s %s", spec.Image, spec.Runtime, strings.Join(args, " "))

	flattenedFields := make([]interface{}, 0, len(logFields)*2)
	for k, v := range logFields {
		flattenedFields = append(flattenedFields, k, v)
	}

	//nolint:gosec // spec.Runtime is a validated runtime name (docker/podman)
	cmd := osexec.CommandContext(ctx, spec.Runtime, args...)
	// Deliberately leave cmd.Env unset (nil) so the runtime client inherits the
	// host process environment (DOCKER_HOST, HOME, PATH, ...) while the container
	// receives its environment exclusively via --env-file.
	cmd.Stdin = stdIn
	cmd.Stdout = stdOutWriter(logMode, logger, task, flattenedFields...)
	cmd.Stderr = stdErrWriter(logMode, logger, task, flattenedFields...)
	cmd.WaitDelay = 5 * time.Second

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("container execution failed - %w", err)
	}
	return nil
}

// ForceRemoveContainer removes a container by name, treating an already-removed
// container as success. Used as a cleanup callback to guard against orphaned
// containers. Because runs use --rm, the container is usually already gone by the
// time this runs, so a "no such container" result is the normal path - and the
// exact wording differs between docker ("No such container") and podman
// ("no such container"), so the match must be case-insensitive.
func ForceRemoveContainer(runtime, name string) error {
	if runtime == "" || name == "" {
		return nil
	}
	out, err := osexec.Command(runtime, "rm", "-f", name).CombinedOutput()
	if err != nil {
		if containerAlreadyGone(string(out)) {
			return nil
		}
		return fmt.Errorf("failed to remove container %s - %w", name, err)
	}
	return nil
}

// containerAlreadyGone reports whether a `rm -f` failure is just the container
// being absent (the normal case under --rm). Docker says "No such container" and
// podman "no such container", so the match is case-insensitive.
func containerAlreadyGone(output string) bool {
	return strings.Contains(strings.ToLower(output), "no such container")
}

// buildRunArgs constructs the runtime argv (excluding the runtime binary itself).
// It is pure so tests can assert the exact slice.
func buildRunArgs(spec ContainerSpec, envFile string) []string {
	args := []string{"run", "--rm", "-i"}
	if spec.Name != "" {
		args = append(args, "--name", spec.Name)
	}

	labelKeys := make([]string, 0, len(spec.Labels))
	for k := range spec.Labels {
		labelKeys = append(labelKeys, k)
	}
	sort.Strings(labelKeys)
	for _, k := range labelKeys {
		args = append(args, "--label", fmt.Sprintf("%s=%s", k, spec.Labels[k]))
	}

	if envFile != "" {
		args = append(args, "--env-file", envFile)
	}

	for _, m := range spec.Mounts {
		args = append(args, "-v", mountArg(m))
	}

	if spec.Workdir != "" {
		args = append(args, "-w", spec.Workdir)
	}
	if spec.User != "" {
		args = append(args, "--user", spec.User)
	}
	if spec.Network != "" {
		args = append(args, "--network", spec.Network)
	}
	if spec.OverrideEntry {
		args = append(args, "--entrypoint", spec.Entrypoint)
	}

	// End-of-options separator: ensures a user-controlled image (or any value that
	// happens to start with "-") is treated as the image positional and never
	// parsed as a runtime flag.
	args = append(args, "--", spec.Image)

	switch {
	case spec.Script != "":
		if spec.OverrideEntry {
			args = append(args, spec.Script)
		} else {
			args = append(args, DefaultContainerShellArg, spec.Script)
		}
	case spec.Cmd != "":
		if spec.OverrideEntry {
			args = append(args, "-c", spec.Cmd)
		} else {
			args = append(args, DefaultContainerShellArg, "-c", spec.Cmd)
		}
	}
	return args
}

// DefaultContainerShellArg is the shell used to interpret cmd/scripts when flow
// does not override the image entrypoint.
const DefaultContainerShellArg = "sh"

func mountArg(m Mount) string {
	arg := m.HostPath + ":" + m.ContainerPath
	opts := m.Options
	if m.ReadOnly && !hasOption(opts, "ro") {
		if opts == "" {
			opts = "ro"
		} else {
			opts += ",ro"
		}
	}
	if opts != "" {
		arg += ":" + opts
	}
	return arg
}

func hasOption(options, want string) bool {
	for _, o := range strings.Split(options, ",") {
		if strings.TrimSpace(o) == want {
			return true
		}
	}
	return false
}

// writeEnvFile writes env to a 0600 temp file in docker --env-file format. It
// rejects keys/values that the format cannot represent rather than silently
// corrupting them.
func writeEnvFile(env map[string]string) (string, error) {
	keys := make([]string, 0, len(env))
	for k := range env {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var b strings.Builder
	for _, k := range keys {
		if k == "" || strings.HasPrefix(k, "#") {
			// A bare key means "inherit from host" and a #-leading key is parsed as
			// a comment; neither is what the user intends, so skip.
			continue
		}
		v := env[k]
		if strings.ContainsAny(v, "\n\r") {
			return "", flowerrors.NewValidationError(
				fmt.Sprintf("environment variable %q has a multi-line value, which cannot be "+
					"passed to a container via --env-file", k),
				map[string]any{"key": k},
			)
		}
		b.WriteString(k)
		b.WriteByte('=')
		b.WriteString(v)
		b.WriteByte('\n')
	}

	f, err := os.CreateTemp("", "flow-container-env-*")
	if err != nil {
		return "", fmt.Errorf("unable to create env file - %w", err)
	}
	if _, err := f.WriteString(b.String()); err != nil {
		_ = f.Close()
		_ = os.Remove(f.Name())
		return "", fmt.Errorf("unable to write env file - %w", err)
	}
	if err := f.Close(); err != nil {
		_ = os.Remove(f.Name())
		return "", fmt.Errorf("unable to close env file - %w", err)
	}
	return f.Name(), nil
}
