//go:build windows

package main

import (
	"syscall"
	"unsafe"
)

const (
	trayMutexName      = `Local\statusline-tray`
	errAlreadyExists   = syscall.Errno(183)
	kernel32CreateMute = "CreateMutexW"
)

var (
	kernel32DLL      = syscall.NewLazyDLL("kernel32.dll")
	procCreateMutexW = kernel32DLL.NewProc(kernel32CreateMute)
)

// acquireTrayLock — Windows Named Mutex로 tray 중복 실행을 방지한다.
// 이미 실행 중이면 acquired=false를 반환하며, 프로세스 종료 시 OS가 핸들을 자동 해제한다.
func acquireTrayLock() (release func(), acquired bool) {
	return acquireNamedTrayLock(trayMutexName)
}

func acquireNamedTrayLock(name string) (release func(), acquired bool) {
	ptr, err := syscall.UTF16PtrFromString(name)
	if err != nil {
		return func() {}, true // fail-soft
	}
	handle, _, callErr := procCreateMutexW.Call(0, 0, uintptr(unsafe.Pointer(ptr)))
	if handle == 0 {
		return func() {}, true // fail-soft
	}
	if callErr == errAlreadyExists {
		_ = syscall.CloseHandle(syscall.Handle(handle))
		return nil, false
	}
	return func() {
		_ = syscall.CloseHandle(syscall.Handle(handle))
	}, true
}
