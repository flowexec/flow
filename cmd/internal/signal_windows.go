//go:build windows

package internal

import (
	"os"
	"os/signal"
)

func terminateProcess(proc *os.Process) error {
	return proc.Kill()
}

func notifyTermSignals(ch chan<- os.Signal) {
	signal.Notify(ch, os.Interrupt)
}
