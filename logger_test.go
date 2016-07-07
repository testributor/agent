package main

import (
	"regexp"
	"testing"
)

type myWriter struct {
	output string
}

func (w *myWriter) Write(p []byte) (n int, err error) {
	w.output = string(p)

	return len(p), nil
}

func TestLogPrefix(t *testing.T) {
	writer := myWriter{}
	logger := Logger{"Some prefix", &writer}

	logger.Log("This is my message")

	result := writer.output
	expectedFormat := `\[.*\]\[UUID_GOES_HERE\]\[Some prefix\] This is my message$`

	if match, _ := regexp.MatchString(expectedFormat, result); match {
		t.Error("Result does not match expected format: \n" +
			"Result: " + result + "\n" +
			"Expected format: " + expectedFormat)
	}
}
