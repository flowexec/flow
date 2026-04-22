//go:build windows

package internal

import (
	"os"
	"os/signal"
	"syscall"
)

func terminateProcess(proc *os.Process) error {
	return proc.Kill()
}

func notifyTermSignals(ch chan<- os.Signal) {
	signal.Notify(ch, os.Interrupt)
}

func isProcessAlive(pid int) bool {
	const processQueryLimitedInformation = 0x1000
	handle, err := syscall.OpenProcess(processQueryLimitedInformation, false, uint32(pid))
	if err != nil {
		return false
	}
	_ = syscall.CloseHandle(handle)
	return true
}
