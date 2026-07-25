//go:build !windows

package process

import (
	"os"
	"syscall"
)

// Alive reports whether a process with the given PID is currently running.
func Alive(pid int) bool {
	if pid <= 0 {
		return false
	}
	proc, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	// On Unix, signal 0 checks for process existence without actually sending a signal.
	return proc.Signal(syscall.Signal(0)) == nil
}
