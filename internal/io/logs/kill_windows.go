//go:build windows

package logs

import "os"

func killProcess(proc *os.Process) error {
	return proc.Kill()
}
