//go:build !windows

package internal

import (
	"os"
	"os/signal"
	"syscall"
)

func terminateProcess(proc *os.Process) error {
	return proc.Signal(syscall.SIGTERM)
}

func notifyTermSignals(ch chan<- os.Signal) {
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
}

func isProcessAlive(pid int) bool {
	proc, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	// On Unix, signal 0 checks for process existence without actually sending a signal.
	return proc.Signal(syscall.Signal(0)) == nil
}
