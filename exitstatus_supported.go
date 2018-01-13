// +build linux windows darwin

package main

import (
	"os/exec"
	"syscall"
)

func exitStatus(err error) (status int, received bool) {
	if exiterr, ok := err.(*exec.ExitError); ok {
		if ws, ok := exiterr.Sys().(syscall.WaitStatus); ok {
			return ws.ExitStatus(), true
		}
	}
	return 0, false
}
