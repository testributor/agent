package main

import (
	"io/ioutil"
	"testing"
)

func TestSystemCommandWhenCommandReturnsAnError(t *testing.T) {
	result, err := SystemCommand([]string{"some non existing command"}, ioutil.Discard)

	if err == nil {
		t.Error("SystemCommand should return and error")
	}

	if result != (CommandResult{}) {
		t.Error("SystemCommand should return an empty CommandResult but got: ", result)
	}
}

func TestSystemCommandWhenCommandExists(t *testing.T) {
	// NOTE: This makes and external call to GitHub. If we can find an other
	// command that will be available on any system and has predictable output
	// we should better change this.
	result, err := SystemCommand([]string{"git", "ls-remote"}, ioutil.Discard)

	if err != nil {
		t.Error("SystemCommand should not return and error but got: ", err)
	}

	if result.output == "" {
		t.Error("SystemCommand should set the output but got empty string")
	}

	if result.errors == "" {
		t.Error("SystemCommand should set the errors but got empty string")
	}

	if result.combinedOutput == "" {
		t.Error("SystemCommand should set the combinedOutput but got empty string")
	}

	if result.durationSeconds == 0 {
		t.Error("SystemCommand should set the durationSeconds but got 0")
	}

	if !result.success {
		t.Error("SystemCommand should set success to true")
	}

	if result.resultType != RESULT_TYPES["passed"] {
		t.Error("SystemCommand should set result type 'passed' but got: ", result.resultType)
	}
}
