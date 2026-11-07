package run_test

import (
	"errors"
	"os"
	"runtime"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	run "github.com/flowexec/flow/v2/internal/services/run"
	pkgerrors "github.com/flowexec/flow/v2/pkg/errors"
	"github.com/flowexec/flow/v2/types/config"
)

var dockerRuntime = run.Runtime{Family: "docker", Command: "docker"}

var _ = Describe("Container backend", func() {
	Describe("buildRunArgs", func() {
		It("builds a minimal cmd invocation with entrypoint override", func() {
			spec := run.ContainerSpec{
				Runtime:       dockerRuntime,
				Image:         "alpine:3",
				Name:          "flow-demo-abcd1234",
				Workdir:       "/workspace",
				Entrypoint:    "sh",
				OverrideEntry: true,
				Cmd:           "echo hi",
			}
			Expect(run.BuildRunArgsForTest(spec, "")).To(Equal([]string{
				"run", "--rm", "-i",
				"--name", "flow-demo-abcd1234",
				"-w", "/workspace",
				"--entrypoint", "sh",
				"--", "alpine:3",
				"-c", "echo hi",
			}))
		})

		It("inserts an end-of-options separator so the image is never parsed as a flag", func() {
			spec := run.ContainerSpec{
				Runtime: dockerRuntime, Image: "-v/etc:/host",
				Cmd: "true", OverrideEntry: true, Entrypoint: "sh",
			}
			args := run.BuildRunArgsForTest(spec, "")
			// "--" must immediately precede the image.
			idx := -1
			for i, a := range args {
				if a == "--" {
					idx = i
					break
				}
			}
			Expect(idx).To(BeNumerically(">=", 0))
			Expect(args[idx+1]).To(Equal("-v/etc:/host"))
		})

		It("omits --env-file when no env file is provided", func() {
			spec := run.ContainerSpec{
				Runtime: dockerRuntime, Image: "alpine:3", Cmd: "true", OverrideEntry: true, Entrypoint: "sh",
			}
			Expect(run.BuildRunArgsForTest(spec, "")).NotTo(ContainElement("--env-file"))
		})

		It("includes --env-file when provided", func() {
			spec := run.ContainerSpec{
				Runtime: dockerRuntime, Image: "alpine:3", Cmd: "true", OverrideEntry: true, Entrypoint: "sh",
			}
			args := run.BuildRunArgsForTest(spec, "/tmp/flow-env")
			Expect(args).To(ContainElements("--env-file", "/tmp/flow-env"))
		})

		It("emits labels sorted by key", func() {
			spec := run.ContainerSpec{
				Runtime:       dockerRuntime,
				Image:         "alpine:3",
				Labels:        map[string]string{"flow.pid": "42", "flow.executable": "run/demo"},
				Cmd:           "true",
				OverrideEntry: true,
				Entrypoint:    "sh",
			}
			args := run.BuildRunArgsForTest(spec, "")
			joined := strings.Join(args, " ")
			Expect(joined).To(ContainSubstring("--label flow.executable=run/demo --label flow.pid=42"))
		})

		It("emits mounts in order with :ro for read-only", func() {
			spec := run.ContainerSpec{
				Runtime: dockerRuntime,
				Image:   "alpine:3",
				Mounts: []run.Mount{
					{HostPath: "/host/ws", ContainerPath: "/workspace"},
					{HostPath: "/host/script.sh", ContainerPath: "/flow/script/script.sh", ReadOnly: true},
				},
				Cmd:           "true",
				OverrideEntry: true,
				Entrypoint:    "sh",
			}
			args := run.BuildRunArgsForTest(spec, "")
			Expect(args).To(ContainElements(
				"-v", "/host/ws:/workspace",
				"-v", "/host/script.sh:/flow/script/script.sh:ro",
			))
		})

		It("omits --user and --network when unset", func() {
			spec := run.ContainerSpec{
				Runtime: dockerRuntime, Image: "alpine:3", Cmd: "true", OverrideEntry: true, Entrypoint: "sh",
			}
			args := run.BuildRunArgsForTest(spec, "")
			Expect(args).NotTo(ContainElement("--user"))
			Expect(args).NotTo(ContainElement("--network"))
		})

		It("runs a script as the final argument with no -c", func() {
			spec := run.ContainerSpec{
				Runtime:       dockerRuntime,
				Image:         "alpine:3",
				Script:        "/workspace/build.sh",
				OverrideEntry: true,
				Entrypoint:    "sh",
			}
			args := run.BuildRunArgsForTest(spec, "")
			Expect(args[len(args)-2:]).To(Equal([]string{"alpine:3", "/workspace/build.sh"}))
			Expect(args).NotTo(ContainElement("-c"))
		})

		It("prefixes cmd with sh -c when the entrypoint is not overridden", func() {
			spec := run.ContainerSpec{Runtime: dockerRuntime, Image: "node:18", Cmd: "npm test", OverrideEntry: false}
			args := run.BuildRunArgsForTest(spec, "")
			Expect(args).NotTo(ContainElement("--entrypoint"))
			Expect(args[len(args)-4:]).To(Equal([]string{"node:18", "sh", "-c", "npm test"}))
		})
	})

	Describe("writeEnvFile", func() {
		It("writes sorted 0600 entries", func() {
			path, err := run.WriteEnvFileForTest(map[string]string{"B": "2", "A": "1"})
			Expect(err).NotTo(HaveOccurred())
			defer os.Remove(path)

			info, err := os.Stat(path)
			Expect(err).NotTo(HaveOccurred())
			// Unix permission bits are not expressible on Windows: Go maps them
			// onto ACLs and the file reads back as 0666. The env file still holds
			// secrets, so the mode is asserted everywhere it means something.
			if runtime.GOOS != "windows" {
				Expect(info.Mode().Perm()).To(Equal(os.FileMode(0600)))
			}

			content, err := os.ReadFile(path)
			Expect(err).NotTo(HaveOccurred())
			Expect(string(content)).To(Equal("A=1\nB=2\n"))
		})

		It("rejects a multi-line value", func() {
			_, err := run.WriteEnvFileForTest(map[string]string{"KEY": "line1\nline2"})
			Expect(err).To(HaveOccurred())
			var verr pkgerrors.ValidationError
			Expect(errors.As(err, &verr)).To(BeTrue())
		})
	})

	Describe("containerAlreadyGone", func() {
		DescribeTable("recognizes an already-removed container across runtimes",
			func(output string, want bool) {
				Expect(run.ContainerAlreadyGoneForTest(output)).To(Equal(want))
			},
			Entry("docker", "Error response from daemon: No such container: flow-x", true),
			Entry("podman", `Error: no container with name or ID "flow-x" found: no such container`, true),
			Entry("unrelated failure", "Error: permission denied while trying to connect", false),
			Entry("empty", "", false),
		)
	})

	Describe("ResolveRuntime", func() {
		AfterEach(func() { run.ResetRuntimeCacheForTest() })

		It("returns docker when docker is present", func() {
			run.ResetRuntimeCacheForTest()
			restore := run.SetLookPathForTest(func(name string) (string, error) {
				if name == "docker" {
					return "/usr/bin/docker", nil
				}
				return "", errors.New("not found")
			})
			defer restore()
			rt, err := run.ResolveRuntime("auto", nil)
			Expect(err).NotTo(HaveOccurred())
			Expect(rt.Family).To(Equal("docker"))
		})

		It("falls back to podman when docker is absent", func() {
			run.ResetRuntimeCacheForTest()
			restore := run.SetLookPathForTest(func(name string) (string, error) {
				if name == "podman" {
					return "/usr/bin/podman", nil
				}
				return "", errors.New("not found")
			})
			defer restore()
			rt, err := run.ResolveRuntime("auto", nil)
			Expect(err).NotTo(HaveOccurred())
			Expect(rt.Family).To(Equal("podman"))
		})

		It("returns an EXECUTION_FAILED error when neither is present", func() {
			run.ResetRuntimeCacheForTest()
			restore := run.SetLookPathForTest(func(string) (string, error) {
				return "", errors.New("not found")
			})
			defer restore()
			_, err := run.ResolveRuntime("auto", nil)
			Expect(err).To(HaveOccurred())
			var rerr pkgerrors.ContainerRuntimeError
			Expect(errors.As(err, &rerr)).To(BeTrue())
			Expect(rerr.Code()).To(Equal("EXECUTION_FAILED"))
		})

		It("errors when an explicitly-requested runtime is missing", func() {
			restore := run.SetLookPathForTest(func(name string) (string, error) {
				if name == "docker" {
					return "/usr/bin/docker", nil
				}
				return "", errors.New("not found")
			})
			defer restore()
			_, err := run.ResolveRuntime("podman", nil)
			Expect(err).To(HaveOccurred())
		})

		It("does not cache a negative result, so detection recovers once a runtime appears", func() {
			run.ResetRuntimeCacheForTest()
			available := false
			restore := run.SetLookPathForTest(func(name string) (string, error) {
				if available && name == "docker" {
					return "/usr/bin/docker", nil
				}
				return "", errors.New("not found")
			})
			defer restore()

			_, err := run.ResolveRuntime("auto", nil)
			Expect(err).To(HaveOccurred()) // runtime not up yet

			available = true // e.g. Docker Desktop finished booting
			rt, err := run.ResolveRuntime("auto", nil)
			Expect(err).NotTo(HaveOccurred())
			Expect(rt.Family).To(Equal("docker"))
		})

		Context("with this machine's container config", func() {
			both := func(name string) (string, error) {
				switch name {
				case "docker", "podman", "finch":
					return "/usr/bin/" + name, nil
				}
				return "", errors.New("not found")
			}

			BeforeEach(func() { run.ResetRuntimeCacheForTest() })

			It("runs a family's configured command and global flags", func() {
				DeferCleanup(run.SetLookPathForTest(both))
				cfg := &config.Config{}
				Expect(cfg.SetContainerCommand(config.ContainerRuntimeDocker, "finch", []string{"--namespace", "dev"})).
					To(Succeed())

				rt, err := run.ResolveRuntime("docker", cfg)
				Expect(err).NotTo(HaveOccurred())
				Expect(rt).To(Equal(run.Runtime{Family: "docker", Command: "finch", Args: []string{"--namespace", "dev"}}))
			})

			It("reports the configured command when it isn't installed", func() {
				DeferCleanup(run.SetLookPathForTest(both))
				cfg := &config.Config{}
				Expect(cfg.SetContainerCommand(config.ContainerRuntimeDocker, "nerdctl", nil)).To(Succeed())

				_, err := run.ResolveRuntime("docker", cfg)
				Expect(err).To(MatchError(ContainSubstring("nerdctl is not installed")))
			})

			It("resolves auto to the machine's chosen runtime without probing", func() {
				DeferCleanup(run.SetLookPathForTest(both))
				DeferCleanup(run.SetEngineRunningForTest(func(string, []string) bool {
					Fail("a chosen runtime needs no probe")
					return false
				}))
				cfg := &config.Config{Container: &config.Container{Runtime: config.ContainerRuntimePodman}}

				rt, err := run.ResolveRuntime("auto", cfg)
				Expect(err).NotTo(HaveOccurred())
				Expect(rt.Family).To(Equal("podman"))
			})

			It("prefers a running engine when both runtimes are installed", func() {
				DeferCleanup(run.SetLookPathForTest(both))
				DeferCleanup(run.SetEngineRunningForTest(func(command string, _ []string) bool {
					return command == "podman" // Docker Desktop is stopped
				}))

				rt, err := run.ResolveRuntime("auto", nil)
				Expect(err).NotTo(HaveOccurred())
				Expect(rt.Family).To(Equal("podman"))
			})

			It("falls back to docker when neither engine answers", func() {
				DeferCleanup(run.SetLookPathForTest(both))
				DeferCleanup(run.SetEngineRunningForTest(func(string, []string) bool { return false }))

				rt, err := run.ResolveRuntime("auto", nil)
				Expect(err).NotTo(HaveOccurred())
				Expect(rt.Family).To(Equal("docker"))
			})

			It("probes with the configured command and flags", func() {
				DeferCleanup(run.SetLookPathForTest(both))
				var probed [][]string
				DeferCleanup(run.SetEngineRunningForTest(func(command string, args []string) bool {
					probed = append(probed, append([]string{command}, args...))
					return true
				}))
				cfg := &config.Config{}
				Expect(cfg.SetContainerCommand(config.ContainerRuntimeDocker, "docker", []string{"--context", "colima"})).
					To(Succeed())

				rt, err := run.ResolveRuntime("auto", cfg)
				Expect(err).NotTo(HaveOccurred())
				Expect(rt.Args).To(Equal([]string{"--context", "colima"}))
				Expect(probed).To(Equal([][]string{{"docker", "--context", "colima"}}))
			})
		})
	})
})
