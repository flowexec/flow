package run

import (
	stdctx "context"
	"fmt"
	"os"
	osexec "os/exec"
	"slices"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/flowexec/tuikit/io"

	flowerrors "github.com/flowexec/flow/v2/pkg/errors"
	"github.com/flowexec/flow/v2/pkg/store"
	"github.com/flowexec/flow/v2/types/config"
	"github.com/flowexec/flow/v2/types/executable"
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
	Runtime       Runtime // already resolved (never "auto")
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

// Runtime is a resolved container runtime: the family a flow file names, and the command and
// global flags this machine runs for it.
type Runtime struct {
	Family  string // "docker" or "podman"
	Command string
	Args    []string
}

// runtime detection cache for the "auto" preference, keyed by the machine config it was
// resolved with. Only a successful detection is memoized: a negative result must not be
// cached, otherwise a long-lived process (e.g. `flow mcp`) that starts before the runtime is
// available would keep failing even after the runtime comes up.
var (
	lookPath      = osexec.LookPath // test seam
	engineRunning = probeEngine     // test seam
	autoRuntimeMu sync.Mutex
	autoRuntime   = map[string]Runtime{}
)

// engineProbeTimeout bounds the `info` probe. A runtime whose engine is a VM (Docker Desktop,
// Colima, a Podman machine) can hang while it starts.
const engineProbeTimeout = 5 * time.Second

// probeEngine reports whether the runtime's CLI reaches an engine: `info` fails when Docker
// Desktop or Colima is stopped, no Podman machine is started, or DOCKER_HOST doesn't answer.
func probeEngine(command string, args []string) bool {
	ctx, cancel := stdctx.WithTimeout(stdctx.Background(), engineProbeTimeout)
	defer cancel()
	return osexec.CommandContext(ctx, command, append(slices.Clone(args), "info")...).Run() == nil
}

// ResolveRuntime resolves a flow file's runtime preference ("auto", "docker", "podman", or "")
// to a runtime installed on this machine, using cfg's container settings: the runtime `auto`
// means here, and the command and global flags for each family.
//
// When `auto` must detect, docker is preferred, then podman. If both are installed, the first
// whose engine answers is chosen, so a stopped Docker Desktop doesn't hide a running Podman;
// with neither answering, the first installed is used and its own error explains why.
func ResolveRuntime(pref string, cfg *config.Config) (Runtime, error) {
	if pref == "" || pref == string(executable.ExecContainerRuntimeAuto) {
		pref = string(cfg.ContainerRuntimePreference())
	}
	switch pref {
	case "docker", "podman":
		rt := runtimeFor(pref, cfg)
		if _, err := lookPath(rt.Command); err != nil {
			return Runtime{}, flowerrors.NewContainerRuntimeError(pref,
				fmt.Errorf("%s is not installed or not in PATH", rt.Command))
		}
		return rt, nil
	case "auto":
		return detectRuntime(cfg)
	default:
		return Runtime{}, flowerrors.NewContainerRuntimeError(pref,
			fmt.Errorf("unknown container runtime %q", pref))
	}
}

func runtimeFor(family string, cfg *config.Config) Runtime {
	command, args := cfg.ContainerCommand(config.ContainerRuntime(family))
	return Runtime{Family: family, Command: command, Args: args}
}

func detectRuntime(cfg *config.Config) (Runtime, error) {
	installed := make([]Runtime, 0, 2)
	for _, family := range []string{"docker", "podman"} {
		rt := runtimeFor(family, cfg)
		if _, err := lookPath(rt.Command); err == nil {
			installed = append(installed, rt)
		}
	}
	if len(installed) == 0 {
		return Runtime{}, flowerrors.NewContainerRuntimeError("",
			fmt.Errorf("no container runtime found: install docker or podman, "+
				"or set exec.container.runtime"))
	}

	key := fmt.Sprint(installed)
	autoRuntimeMu.Lock()
	defer autoRuntimeMu.Unlock()
	if rt, ok := autoRuntime[key]; ok {
		return rt, nil
	}
	chosen := installed[0]
	// Probing costs an `info` round trip, so it's only worth it when there is a choice.
	if len(installed) > 1 {
		for _, rt := range installed {
			if engineRunning(rt.Command, rt.Args) {
				chosen = rt
				break
			}
		}
	}
	autoRuntime[key] = chosen
	return chosen, nil
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
	args = append(slices.Clone(spec.Runtime.Args), args...)
	logger.Debugf("running container (%s):\n%s %s", spec.Image, spec.Runtime.Command, strings.Join(args, " "))

	flattenedFields := make([]interface{}, 0, len(logFields)*2)
	for k, v := range logFields {
		flattenedFields = append(flattenedFields, k, v)
	}

	//nolint:gosec // the runtime command comes from the user's own config or a runtime family name
	cmd := osexec.CommandContext(ctx, spec.Runtime.Command, args...)
	// Deliberately leave cmd.Env unset (nil) so the runtime client inherits the
	// host process environment (DOCKER_HOST, HOME, PATH, ...) while the container
	// receives its environment exclusively via --env-file.
	cmd.Stdin = stdIn
	cmd.Stdout = stdOutWriter(logMode, logger, task, flattenedFields...)
	cmd.Stderr = stdErrWriter(logMode, logger, task, flattenedFields...)
	cmd.WaitDelay = 5 * time.Second

	store.MarkRunStarted()
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
func ForceRemoveContainer(runtime Runtime, name string) error {
	if runtime.Command == "" || name == "" {
		return nil
	}
	//nolint:gosec // the runtime command comes from the user's own config or a runtime family name
	out, err := osexec.Command(runtime.Command, append(slices.Clone(runtime.Args), "rm", "-f", name)...).CombinedOutput()
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
