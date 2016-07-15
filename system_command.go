package main

import (
	"bufio"
	"fmt"
	"io"
	"os/exec"
	"strings"
	"syscall"
	"time"
)

// These should much the codes on the testributor side
var RESULT_TYPES = map[string]int{
	"passed": 3,
	"failed": 4,
	"error":  5,
}

type CommandResult struct {
	output          string
	errors          string
	combinedOutput  string
	resultType      int
	success         bool
	durationSeconds float64
	commandErr      error
	exitCode        int
}

// SystemCommand is used to run system commands. It returns a CommandResult
// struct which holds the stdout, stderr and combined output along with the
// duration, result type (failed, error, success) for testing commands and
// whether the command's exit code was success or not.
// The logger can be any io.Writer but the usual suspects are our Logger
// struct (which formats the output) and ioutil.Discard when we don't want to
// print the output.
func SystemCommand(command []string, logger io.Writer) (CommandResult, error) {
	commandStart := time.Now()
	cmd := exec.Command(command[0], command[1:]...)
	errPipe, err := cmd.StderrPipe()
	if err != nil {
		return CommandResult{}, err
	}
	outPipe, err := cmd.StdoutPipe()
	if err != nil {
		return CommandResult{}, err
	}

	output := ""
	errors := ""
	combined := ""
	errorsDone := make(chan bool)
	outputDone := make(chan bool)
	combinedOutputChannel := make(chan string)

	// Capture the combined output too
	go func(result *string) {
		for {
			newString, more := <-combinedOutputChannel
			*result += newString
			if !more {
				break
			}
		}
	}(&combined)

	go ReadUntilEOF(outPipe, &output, outputDone, combinedOutputChannel, logger)
	go ReadUntilEOF(errPipe, &errors, errorsDone, combinedOutputChannel, logger)

	startErr := cmd.Start()
	if startErr != nil {
		return CommandResult{}, startErr
	}

	// Wait until reading is done before calling Wait()
	// https://golang.org/pkg/os/exec/#Cmd.StdoutPipe
	_ = <-outputDone
	_ = <-errorsDone
	close(combinedOutputChannel) // Nothing more to read. Let the reading go routine exit.

	waitResult := cmd.Wait()

	// http://stackoverflow.com/a/10385867/974285
	var exitCode int
	if exiterr, ok := waitResult.(*exec.ExitError); ok {
		// The program has exited with an exit code != 0

		// This works on both Unix and Windows. Although package
		// syscall is generally platform dependent, WaitStatus is
		// defined for both Unix and Windows and in both cases has
		// an ExitStatus() method with the same signature.
		if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
			exitCode = status.ExitStatus()
		}
	}

	var resultType int
	switch {
	case waitResult == nil:
		resultType = RESULT_TYPES["passed"]
	case strings.TrimSpace(errors) == "":
		resultType = RESULT_TYPES["failed"]
	default:
		resultType = RESULT_TYPES["error"]
	}

	return CommandResult{
		output:          output,
		errors:          errors,
		combinedOutput:  combined,
		resultType:      resultType,
		success:         (waitResult == nil),
		commandErr:      waitResult,
		exitCode:        exitCode,
		durationSeconds: time.Since(commandStart).Seconds(),
	}, nil
}

// Reads from stream until EOF. The result is written on outVar.
// When EOF is reached, true is sent on doneChannel.
// To be used as a go routine.
func ReadUntilEOF(stream io.Reader, outVar *string, doneChannel chan bool, combinedOutputChannel chan string, logger io.Writer) {
	scanner := bufio.NewScanner(stream)
	for scanner.Scan() {
		line := scanner.Text()
		logger.Write(([]byte)(line))
		line = fmt.Sprintf("%v\n", line)
		combinedOutputChannel <- line
		*outVar += line
	}
	if err := scanner.Err(); err != nil {
		logger.Write(([]byte)(err.Error()))
	}

	doneChannel <- true
}
