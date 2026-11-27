//go:build windows

package main

import (
	"fmt"
	"os"
	"testing"
)

func TestAcquireNamedTrayLock(t *testing.T) {
	name := fmt.Sprintf(`Local\statusline-tray-test-%d`, os.Getpid())

	release1, ok1 := acquireNamedTrayLock(name)
	if !ok1 {
		t.Fatal("first acquire should succeed")
	}

	release2, ok2 := acquireNamedTrayLock(name)
	if ok2 {
		release2()
		release1()
		t.Fatal("second acquire while held should fail")
	}

	release1()

	release3, ok3 := acquireNamedTrayLock(name)
	if !ok3 {
		t.Fatal("acquire after release should succeed")
	}
	release3()
}
