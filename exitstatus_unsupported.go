// +build !linux,!windows,!darwin

package main

func exitStatus(err error) (status int, received bool) {
	return 0, false
}
