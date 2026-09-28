//go:build windows

package collect

import (
	"os"
	"os/exec"
	"syscall"
)

const createNewProcessGroup = 0x00000200

func SpawnSelfCollect(provider string) error {
	exe, err := os.Executable()
	if err != nil {
		return err
	}
	cmd := exec.Command(exe, "collect", "--provider="+provider)
	cmd.SysProcAttr = &syscall.SysProcAttr{CreationFlags: createNewProcessGroup}
	return cmd.Start()
}
