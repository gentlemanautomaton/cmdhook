package main

import (
	"bytes"
	"os"
	"os/exec"
	"os/signal"
	"syscall"

	"github.com/gentlemanautomaton/cmdline"
)

func main() {
	preStart := os.Getenv("PRESTART")
	postStart := os.Getenv("POSTSTART")
	sigterm := os.Getenv("SIGTERM")
	postStop := os.Getenv("POSTSTOP")

	var (
		name = os.Args[1]
		args = os.Args[2:]
	)

	// PreStart
	if _, err := executeHook(preStart); err != nil {
		if status, ok := exitStatus(err); ok {
			os.Exit(status)
		}
		os.Exit(PrestartFailure)
	}

	// Start
	process, result, err := executeProgram(name, args)
	if err != nil {
		if status, ok := exitStatus(err); ok {
			os.Exit(status)
		}
		os.Exit(StartFailure)
	}

	// PostStart
	executeHook(postStart)

	// Signal processing
	stopped := make(chan struct{})
	spdone := processSignals(process, stopped, sigterm)

	// Stop
	err = <-result // Wait for program execution to finish
	close(stopped) // Let the signal processor know that the program exited
	<-spdone       // Wait for the signal processor to finish

	// PostStop
	executeHook(postStop)

	// Return the exit status of the program we ran
	if err != nil {
		if status, ok := exitStatus(err); ok {
			os.Exit(status)
		}
		os.Exit(ExecFailure)
	}

	os.Exit(0)
}

func processSignals(process *os.Process, stopped chan struct{}, sigterm string) <-chan struct{} {
	spdone := make(chan struct{})

	sigChan := make(chan os.Signal, 64)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		defer close(spdone)

		for {
			select {
			case <-stopped:
				return
			case sig := <-sigChan:
				switch sig {
				case syscall.SIGINT, syscall.SIGTERM:
					if handled, err := executeHook(sigterm); handled && err == nil {
						break
					}
					fallthrough
				default:
					process.Signal(sig)
				}
			}
		}
	}()

	return spdone
}

func executeProgram(name string, args []string) (process *os.Process, result chan error, err error) {
	result = make(chan error, 1)

	cmd := exec.Command(name, args...)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout

	err = cmd.Start()
	if err != nil {
		close(result)
		return nil, nil, err
	}

	go func() {
		defer close(result)
		result <- cmd.Wait()
	}()

	return cmd.Process, result, err
}

func executeHook(cl string) (executed bool, err error) {
	if cl == "" {
		return false, nil
	}

	var (
		outbuf, errbuf bytes.Buffer
		name, args     = cmdline.SplitCommand(cl)
	)

	cmd := exec.Command(name, args...)
	cmd.Stdout = &outbuf
	cmd.Stderr = &errbuf

	err = cmd.Run()

	outbuf.WriteTo(os.Stdout)
	errbuf.WriteTo(os.Stderr)

	return true, err
}
