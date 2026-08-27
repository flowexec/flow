//go:build unit

package run_test

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/flowexec/flow/v2/internal/services/run"
	flowErrors "github.com/flowexec/flow/v2/pkg/errors"
)

// writeFakePython creates an executable stub at the conventional interpreter path
// inside venvDir and returns that path.
func writeFakePython(venvDir string) string {
	path := run.VenvPythonForTest(venvDir)
	Expect(os.MkdirAll(filepath.Dir(path), 0750)).To(Succeed())
	Expect(os.WriteFile(path, []byte("#!/bin/sh\nexit 0\n"), 0700)).To(Succeed())
	return path
}

var _ = Describe("Python", func() {
	Describe("envValue", func() {
		It("prefers the last matching entry, since later entries win at exec time", func() {
			list := []string{"FOO=first", "BAR=x", "FOO=second"}
			Expect(run.EnvValueForTest(list, "FOO")).To(Equal("second"))
		})

		It("falls back to the process environment", func() {
			GinkgoTB().Setenv("FLOW_PY_TEST_ONLY", "from-process")
			Expect(run.EnvValueForTest(nil, "FLOW_PY_TEST_ONLY")).To(Equal("from-process"))
		})

		It("returns empty for an unset key", func() {
			Expect(run.EnvValueForTest([]string{"A=1"}, "MISSING_ENTIRELY")).To(BeEmpty())
		})
	})

	Describe("pythonEnv", func() {
		It("adds unbuffered and no-bytecode defaults", func() {
			out := strings.Join(run.PythonEnvForTest([]string{"A=1"}), "\n")
			Expect(out).To(ContainSubstring("PYTHONUNBUFFERED=1"))
			Expect(out).To(ContainSubstring("PYTHONDONTWRITEBYTECODE=1"))
		})

		It("does not override an explicit value from the run environment", func() {
			out := run.PythonEnvForTest([]string{"PYTHONUNBUFFERED=0"})
			Expect(out).NotTo(ContainElement("PYTHONUNBUFFERED=1"))
			Expect(out).To(ContainElement("PYTHONUNBUFFERED=0"))
		})

		It("does not mutate the caller's slice", func() {
			in := []string{"A=1"}
			run.PythonEnvForTest(in)
			Expect(in).To(HaveLen(1))
		})
	})

	Describe("pathCandidates", func() {
		It("prefers python over python3 on Windows to dodge the Store alias stub", func() {
			if runtime.GOOS == "windows" {
				Expect(run.PathCandidatesForTest()).To(Equal([]string{"python", "python3"}))
			} else {
				Expect(run.PathCandidatesForTest()).To(Equal([]string{"python3", "python"}))
			}
		})
	})

	Describe("ResolvePython", func() {
		It("honours FLOW_PYTHON_BIN when it points at a real file", func() {
			dir := GinkgoTB().TempDir()
			bin := writeFakePython(filepath.Join(dir, "venv"))

			got, err := run.ResolvePython([]string{run.PythonBinEnv + "=" + bin})
			Expect(err).NotTo(HaveOccurred())
			Expect(got).To(Equal(bin))
		})

		It("fails loudly rather than falling through when FLOW_PYTHON_BIN is wrong", func() {
			// A deliberately-chosen interpreter must never silently resolve to a
			// different one.
			_, err := run.ResolvePython([]string{run.PythonBinEnv + "=/nonexistent/python"})
			Expect(err).To(HaveOccurred())
			var notFound flowErrors.InterpreterNotFoundError
			Expect(err).To(BeAssignableToTypeOf(notFound))
			Expect(err.Error()).To(ContainSubstring("/nonexistent/python"))
		})

		It("uses an active VIRTUAL_ENV", func() {
			venv := filepath.Join(GinkgoTB().TempDir(), "venv")
			bin := writeFakePython(venv)

			got, err := run.ResolvePython([]string{"VIRTUAL_ENV=" + venv})
			Expect(err).NotTo(HaveOccurred())
			Expect(got).To(Equal(bin))
		})

		It("falls back to the workspace .venv when no virtualenv is active", func() {
			ws := GinkgoTB().TempDir()
			bin := writeFakePython(filepath.Join(ws, ".venv"))

			got, err := run.ResolvePython([]string{"FLOW_WORKSPACE_PATH=" + ws})
			Expect(err).NotTo(HaveOccurred())
			Expect(got).To(Equal(bin))
		})

		It("prefers an active VIRTUAL_ENV over the workspace .venv", func() {
			ws := GinkgoTB().TempDir()
			writeFakePython(filepath.Join(ws, ".venv"))
			active := filepath.Join(GinkgoTB().TempDir(), "active")
			activeBin := writeFakePython(active)

			got, err := run.ResolvePython([]string{
				"FLOW_WORKSPACE_PATH=" + ws,
				"VIRTUAL_ENV=" + active,
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(got).To(Equal(activeBin))
		})

		It("skips a virtualenv path that does not exist and keeps searching", func() {
			// A stale VIRTUAL_ENV should not be fatal while a real interpreter exists.
			got, err := run.ResolvePython([]string{"VIRTUAL_ENV=/nonexistent/venv"})
			if err != nil {
				Skip("no python on PATH to fall back to")
			}
			Expect(got).NotTo(ContainSubstring("/nonexistent/venv"))
		})
	})

	Describe("WritePythonScript", func() {
		It("writes the code to a .py file and cleans up", func() {
			path, cleanup, err := run.WritePythonScript("print('hi')\n")
			Expect(err).NotTo(HaveOccurred())
			Expect(filepath.Ext(path)).To(Equal(".py"))

			content, err := os.ReadFile(path) //nolint:gosec
			Expect(err).NotTo(HaveOccurred())
			Expect(string(content)).To(Equal("print('hi')\n"))

			cleanup()
			_, statErr := os.Stat(path)
			Expect(os.IsNotExist(statErr)).To(BeTrue())
		})

		It("restricts permissions so script contents are not world-readable", func() {
			if runtime.GOOS == "windows" {
				Skip("unix file modes are ACL-mapped on Windows")
			}
			path, cleanup, err := run.WritePythonScript("print(1)\n")
			Expect(err).NotTo(HaveOccurred())
			defer cleanup()

			info, err := os.Stat(path)
			Expect(err).NotTo(HaveOccurred())
			Expect(info.Mode().Perm()).To(Equal(os.FileMode(0600)))
		})
	})
})
