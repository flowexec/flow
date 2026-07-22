package exec

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"github.com/pkg/errors"

	"github.com/flowexec/flow/v2/internal/services/run"
	"github.com/flowexec/flow/v2/pkg/logger"
	"github.com/flowexec/flow/v2/types/executable"
)

// containerScriptMount is the container path where a script that lives outside
// any workspace/volume mount is bind-mounted.
const containerScriptMount = "/flow/script"

// containerFallbackWorkdir is used when the executable's target directory cannot
// be located inside the workspace mount (e.g. f:tmp or an absolute path
// elsewhere); the directory is mounted here directly.
const containerFallbackWorkdir = "/flow/workdir"

// buildContainerSpec translates a host-side exec into a container.RunContainer
// spec: resolving the runtime, building mounts, translating the working
// directory and environment, and choosing between cmd and file execution.
func buildContainerSpec(
	e *executable.Executable,
	targetDir string,
	envMap map[string]string,
) (run.ContainerSpec, error) {
	c := e.Exec.Container

	rt, err := resolveRuntimeFn(string(c.Runtime))
	if err != nil {
		return run.ContainerSpec{}, err
	}

	if e.Exec.File != "" {
		switch strings.ToLower(filepath.Ext(e.Exec.File)) {
		case ".bat", ".cmd", ".ps1":
			return run.ContainerSpec{}, errors.Errorf(
				"container execution does not support %s files", filepath.Ext(e.Exec.File))
		}
	}

	wsRoot := e.WorkspacePath()
	mountPoint := c.MountWorkspace
	if mountPoint == "" {
		mountPoint = executable.DefaultContainerMount
	}

	mounts := []run.Mount{{HostPath: wsRoot, ContainerPath: mountPoint}}

	// Resolve the container-side working directory.
	workdir := c.Workdir
	if workdir == "" {
		if cp, ok := containerPathUnder(wsRoot, mountPoint, targetDir); ok {
			workdir = cp
		} else {
			// targetDir is outside the workspace mount (f:tmp, absolute path, ...).
			// Mount it directly so the command still sees its files.
			mounts = append(mounts, run.Mount{HostPath: targetDir, ContainerPath: containerFallbackWorkdir})
			workdir = containerFallbackWorkdir
			logger.Log().Debugf("mounting out-of-workspace dir %s at %s", targetDir, containerFallbackWorkdir)
		}
	}

	// User-supplied volumes.
	for _, v := range c.Volumes {
		m, err := parseVolume(string(v), wsRoot)
		if err != nil {
			return run.ContainerSpec{}, err
		}
		mounts = append(mounts, m)
	}

	spec := run.ContainerSpec{
		Runtime: rt,
		Image:   c.Image,
		Name:    containerName(e),
		Labels: map[string]string{
			"flow.executable": e.Ref().String(),
			"flow.pid":        strconv.Itoa(os.Getpid()),
		},
		Workdir: workdir,
		Mounts:  mounts,
		Network: c.Network,
	}
	spec.Entrypoint, spec.OverrideEntry = c.ResolveEntrypoint()
	spec.User = resolveUser(c)

	if c.EnvInherited() {
		spec.Env = translateEnv(envMap, mounts, wsRoot, mountPoint)
	}

	switch {
	case e.Exec.Cmd != "":
		spec.Cmd = e.Exec.Cmd
	case e.Exec.File != "":
		scriptHost := filepath.Join(targetDir, e.Exec.File)
		if cp, ok := containerPathFor(scriptHost, mounts); ok {
			spec.Script = cp
		} else {
			// Script lives outside every mount; bind it in read-only.
			containerScript := path.Join(containerScriptMount, filepath.Base(e.Exec.File))
			spec.Mounts = append(spec.Mounts, run.Mount{
				HostPath:      scriptHost,
				ContainerPath: containerScript,
				ReadOnly:      true,
			})
			spec.Script = containerScript
		}
	}

	return spec, nil
}

// containerPathUnder returns the container-side path for hostPath if it resolves
// inside root (mounted at mountPoint), accounting for symlinked temp dirs.
func containerPathUnder(root, mountPoint, hostPath string) (string, bool) {
	rootReal := evalSymlinks(root)
	targetReal := evalSymlinks(hostPath)
	rel, err := filepath.Rel(rootReal, targetReal)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", false
	}
	if rel == "." {
		return mountPoint, true
	}
	return path.Join(mountPoint, filepath.ToSlash(rel)), true
}

// containerPathFor maps a host path to its container path using the given mounts,
// returning false if it is not reachable through any mount.
func containerPathFor(hostPath string, mounts []run.Mount) (string, bool) {
	targetReal := evalSymlinks(hostPath)
	for _, m := range mounts {
		rel, err := filepath.Rel(evalSymlinks(m.HostPath), targetReal)
		if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
			continue
		}
		if rel == "." {
			return m.ContainerPath, true
		}
		return path.Join(m.ContainerPath, filepath.ToSlash(rel)), true
	}
	return "", false
}

// evalSymlinks resolves symlinks, falling back to the raw path when the path does
// not yet exist (e.g. a not-yet-created output file destination).
func evalSymlinks(p string) string {
	if resolved, err := filepath.EvalSymlinks(p); err == nil {
		return resolved
	}
	return p
}

// translateEnv rewrites host-path FLOW_* variables to their container-side
// equivalents (or drops them when unreachable) and forwards color settings.
func translateEnv(envMap map[string]string, mounts []run.Mount, wsRoot, mountPoint string) map[string]string {
	out := make(map[string]string, len(envMap)+1)
	for k, v := range envMap {
		switch k {
		case "FLOW_WORKSPACE_PATH":
			if cp, ok := containerPathUnder(wsRoot, mountPoint, v); ok {
				out[k] = cp
			} else {
				out[k] = mountPoint
			}
		case "FLOW_DEFINITION_PATH", "FLOW_DEFINITION_DIR", "FLOW_TMP_DIRECTORY":
			if cp, ok := containerPathFor(v, mounts); ok {
				out[k] = cp
			}
			// else drop: the host path is meaningless inside the container.
		case "FLOW_CONFIG_PATH", "FLOW_CACHE_PATH":
			// Drop: these directories are not mounted.
		default:
			out[k] = v
		}
	}
	out["FLOW_IN_CONTAINER"] = "true"

	// Forward color preferences so containerized tools keep their coloring.
	for _, k := range []string{"TERM", "FORCE_COLOR", "CLICOLOR_FORCE", "NO_COLOR"} {
		if _, set := out[k]; !set {
			if v, ok := os.LookupEnv(k); ok {
				out[k] = v
			}
		}
	}
	return out
}

// parseVolume expands a user-supplied "host:container[:options]" volume string
// into a Mount. Structural parsing (including Windows drive-letter handling) is
// shared with schema validation via ExecContainerVolume.Parts; the host side is
// expanded locally (never via utils.ExpandPath, which fatally exits on failure).
func parseVolume(v, wsRoot string) (run.Mount, error) {
	host, container, options, err := executable.ExecContainerVolume(v).Parts()
	if err != nil {
		return run.Mount{}, err
	}
	hostExpanded, err := expandVolumeHost(host, wsRoot)
	if err != nil {
		return run.Mount{}, err
	}
	return run.Mount{HostPath: hostExpanded, ContainerPath: container, Options: options}, nil
}

func expandVolumeHost(host, wsRoot string) (string, error) {
	switch {
	case strings.HasPrefix(host, "//"):
		return filepath.Join(wsRoot, strings.TrimPrefix(host, "//")), nil
	case strings.HasPrefix(host, "~/"):
		home, err := os.UserHomeDir()
		if err != nil {
			return "", errors.Wrap(err, "unable to resolve home directory for volume path")
		}
		return filepath.Join(home, strings.TrimPrefix(host, "~/")), nil
	case strings.HasPrefix(host, "./"):
		cwd, err := os.Getwd()
		if err != nil {
			return "", errors.Wrap(err, "unable to resolve working directory for volume path")
		}
		return filepath.Join(cwd, strings.TrimPrefix(host, "./")), nil
	case filepath.IsAbs(host):
		return host, nil
	default:
		return "", errors.Errorf("volume host path %q must be absolute, ~/, ./, or //-prefixed", host)
	}
}

// resolveUser returns the --user value. On Linux, when no user is configured, it
// defaults to the current host uid:gid so mounted files are not root-owned.
func resolveUser(c *executable.ExecContainer) string {
	if c.User != nil {
		return *c.User
	}
	if runtime.GOOS == "linux" {
		uid := os.Getuid()
		gid := os.Getgid()
		if uid >= 0 && gid >= 0 {
			return fmt.Sprintf("%d:%d", uid, gid)
		}
	}
	return ""
}

// containerName builds a unique, DNS-safe container name for cleanup.
func containerName(e *executable.Executable) string {
	base := sanitizeName(e.Name)
	if base == "" {
		base = sanitizeName(string(e.Verb))
	}
	if base == "" {
		base = "exec"
	}
	return fmt.Sprintf("flow-%s-%s", base, randomSuffix())
}

func sanitizeName(s string) string {
	var b strings.Builder
	for _, r := range strings.ToLower(s) {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			b.WriteRune(r)
		case r == '-' || r == '_':
			b.WriteRune(r)
		default:
			b.WriteRune('-')
		}
	}
	return strings.Trim(b.String(), "-")
}

func randomSuffix() string {
	buf := make([]byte, 4)
	if _, err := rand.Read(buf); err != nil {
		return "00000000"
	}
	return hex.EncodeToString(buf)
}
