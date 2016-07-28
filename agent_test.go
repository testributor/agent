package main

import (
	"bytes"
	"regexp"
	"testing"
)

func TestSetWorkerUuid(t *testing.T) {
	setWorkerUuid()
	var b bytes.Buffer
	logger := Logger{"test_logger", &b}
	logger.Log("something")
	result := make([]byte, 100)
	b.Read(result)

	regex := regexp.MustCompile(`\[.+\]\[\w{8}\]\[test_logger\].*`)
	if match := regex.Find(result); match == nil {
		t.Error("It should create a UUID but log message didn't show any: ", string(result))
	}
}
