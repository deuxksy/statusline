//go:build !windows

package collect

import (
	"os"
	"os/exec"
	"syscall"
)

// SpawnSelfCollect — 자기 자신을 detached로 spawn해 해당 provider를 수집한다 (one-shot).
// 렌더를 차단하지 않는다: Start만 하고 종료를 기다리지 않는다.
func SpawnSelfCollect(provider string) error {
	exe, err := os.Executable()
	if err != nil {
		return err
	}
	cmd := exec.Command(exe, "collect", "--provider="+provider)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
	return cmd.Start()
}
