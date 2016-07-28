package main

import (
	"fmt"
	"io"
	"time"
)

type Logger struct {
	prefix string // The name of the "thread"
	writer io.Writer
}

// Write is implemented as part of the io.Writer interface
func (l Logger) Write(p []byte) (n int, err error) {
	now := time.Now().UTC()
	short_uuid := WorkerUUIDShort

	// TODO: ljust
	prefix := now.Format(
		"[15:04:05 Mon 02 Jan UTC]") +
		"[" + short_uuid + "]" +
		"[" + l.prefix + "]"

	prefix = fmt.Sprintf("%-40s ", prefix)

	//TODO We used STDOUT.flush in ruby, is it needed here too?
	return l.writer.Write([]byte(prefix + string(p) + "\n"))
}

// This should be used to write strings instead for byte arrays (which is what
// Write methods expects)
func (l Logger) Log(message string) {
	l.Write([]byte(message))
}
