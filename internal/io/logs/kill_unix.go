//go:build !windows

package logs

import (
	"os"
	"syscall"
)

func killProcess(proc *os.Process) error {
	return proc.Signal(syscall.SIGTERM)
}
