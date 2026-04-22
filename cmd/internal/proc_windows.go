//go:build windows

package internal

import (
	"os/exec"
	"syscall"
)

const createNewProcessGroup = 0x00000200

func setSysProcAttr(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{CreationFlags: createNewProcessGroup}
}
