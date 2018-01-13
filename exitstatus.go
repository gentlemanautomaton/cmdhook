package main

const (
	// PrestartFailure is returned when a prestart hook fails to execute and its
	// exit status cannot be determined.
	PrestartFailure = 417000 + iota // arbitrary

	// StartFailure is returned when a the command fails to start and its
	// exit status cannot be determined.
	StartFailure

	// ExecFailure is returned when a the command encounters an error and its
	// exit status cannot be determined.
	ExecFailure
)
