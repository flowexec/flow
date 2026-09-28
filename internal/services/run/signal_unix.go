//go:build !windows

package run

import (
	"os"
	"time"
)

// killTimeout is how long a cancelled child gets to exit after being interrupted before it is killed.
const killTimeout = 2 * time.Second

func interruptProcess(proc *os.Process) error {
	return proc.Signal(os.Interrupt)
}
