package run_test

import (
	"errors"
	"os"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	run "github.com/flowexec/flow/v2/internal/services/run"
	pkgerrors "github.com/flowexec/flow/v2/pkg/errors"
)

var _ = Describe("Container backend", func() {
	Describe("buildRunArgs", func() {
		It("builds a minimal cmd invocation with entrypoint override", func() {
			spec := run.ContainerSpec{
				Runtime:       "docker",
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
				"alpine:3",
				"-c", "echo hi",
			}))
		})

		It("omits --env-file when no env file is provided", func() {
			spec := run.ContainerSpec{Runtime: "docker", Image: "alpine:3", Cmd: "true", OverrideEntry: true, Entrypoint: "sh"}
			Expect(run.BuildRunArgsForTest(spec, "")).NotTo(ContainElement("--env-file"))
		})

		It("includes --env-file when provided", func() {
			spec := run.ContainerSpec{Runtime: "docker", Image: "alpine:3", Cmd: "true", OverrideEntry: true, Entrypoint: "sh"}
			args := run.BuildRunArgsForTest(spec, "/tmp/flow-env")
			Expect(args).To(ContainElements("--env-file", "/tmp/flow-env"))
		})

		It("emits labels sorted by key", func() {
			spec := run.ContainerSpec{
				Runtime:       "docker",
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
				Runtime: "docker",
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
			spec := run.ContainerSpec{Runtime: "docker", Image: "alpine:3", Cmd: "true", OverrideEntry: true, Entrypoint: "sh"}
			args := run.BuildRunArgsForTest(spec, "")
			Expect(args).NotTo(ContainElement("--user"))
			Expect(args).NotTo(ContainElement("--network"))
		})

		It("runs a script as the final argument with no -c", func() {
			spec := run.ContainerSpec{
				Runtime:       "docker",
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
			spec := run.ContainerSpec{Runtime: "docker", Image: "node:18", Cmd: "npm test", OverrideEntry: false}
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
			Expect(info.Mode().Perm()).To(Equal(os.FileMode(0600)))

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
			rt, err := run.ResolveRuntime("auto")
			Expect(err).NotTo(HaveOccurred())
			Expect(rt).To(Equal("docker"))
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
			rt, err := run.ResolveRuntime("auto")
			Expect(err).NotTo(HaveOccurred())
			Expect(rt).To(Equal("podman"))
		})

		It("returns an EXECUTION_FAILED error when neither is present", func() {
			run.ResetRuntimeCacheForTest()
			restore := run.SetLookPathForTest(func(string) (string, error) {
				return "", errors.New("not found")
			})
			defer restore()
			_, err := run.ResolveRuntime("auto")
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
			_, err := run.ResolveRuntime("podman")
			Expect(err).To(HaveOccurred())
		})
	})
})
