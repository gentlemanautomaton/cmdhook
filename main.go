package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"

	"github.com/gentlemanautomaton/cmdline"
)

func main() {
	preStart := os.Getenv("PRESTART")
	postStart := os.Getenv("POSTSTART")
	sigterm := os.Getenv("SIGTERM")
	postStop := os.Getenv("POSTSTOP")

	var (
		args    = os.Args[1:]
		verbose = false
	)

	if len(args) > 1 && args[0] == "-v" {
		verbose = true
		args = args[1:]
	}

	name := args[0]
	args = args[1:]

	// PreStart
	if _, err := executeHook("PRESTART", preStart, verbose); err != nil {
		if status, ok := exitStatus(err); ok {
			os.Exit(status)
		}
		os.Exit(PrestartFailure)
	}

	// Start
	process, result, err := executeProgram(name, args, verbose)
	if err != nil {
		if status, ok := exitStatus(err); ok {
			os.Exit(status)
		}
		os.Exit(StartFailure)
	}

	// PostStart
	executeHook("POSTSTART", postStart, verbose)

	// Signal processing
	stopped := make(chan struct{})
	spdone := processSignals(process, stopped, sigterm, verbose)

	// Stop
	err = <-result // Wait for program execution to finish
	close(stopped) // Let the signal processor know that the program exited
	<-spdone       // Wait for the signal processor to finish

	// PostStop
	executeHook("POSTSTOP", postStop, verbose)

	// Return the exit status of the program we ran
	if err != nil {
		if status, ok := exitStatus(err); ok {
			os.Exit(status)
		}
		os.Exit(ExecFailure)
	}

	os.Exit(0)
}

func processSignals(process *os.Process, stopped chan struct{}, sigterm string, verbose bool) <-chan struct{} {
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
					if handled, err := executeHook("SIGTERM", sigterm, verbose); handled && err == nil {
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

func executeProgram(name string, args []string, verbose bool) (process *os.Process, result chan error, err error) {
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

func executeHook(hook, cl string, verbose bool) (executed bool, err error) {
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

	if verbose {
		description := fmtHook(name, args)

		if err == nil {
			fmt.Fprintf(os.Stdout, "%s SUCCESS: %s\n", hook, description)
		} else {
			fmt.Fprintf(os.Stderr, "%s FAILURE: %s\n  %v\n", hook, description, err)
		}
	}

	return true, err
}

func fmtHook(name string, args []string) string {
	all := make([]string, len(args)+1)
	all[0] = name
	copy(all[1:], args)
	output := fmt.Sprintf("%#v", all)
	output = strings.TrimPrefix(output, "[]string{")
	output = strings.TrimSuffix(output, "}")
	return "[" + output + "]"
}
