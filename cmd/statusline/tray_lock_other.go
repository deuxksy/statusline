//go:build !windows

package main

func acquireTrayLock() (release func(), acquired bool) {
	return func() {}, true
}
