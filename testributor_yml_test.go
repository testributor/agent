package main

import (
	"testing"
)

var testributor_yml_contents = `
worker_init: /bin/bash -c "apt-get install phantomjs"
before: "./scripts/before_build.sh"
each:
  pattern: "test/.*_test.rb$"
  command: 'bin/rake test %{file}'
`

func TestWorkerInit(t *testing.T) {
	testributorYml, err := NewTestributorYml(testributor_yml_contents)
	if err != nil {
		t.Error(err.Error())
	}
	expected := "/bin/bash -c \"apt-get install phantomjs\""
	if init := testributorYml["worker_init"].(string); init != expected {
		t.Error("Expected: \n" + expected + "\nGot: \n" + init)
	}
}

func TestBeforeAll(t *testing.T) {
	testributorYml, err := NewTestributorYml(testributor_yml_contents)
	if err != nil {
		t.Error(err.Error())
	}
	expected := "./scripts/before_build.sh"
	if init := testributorYml["before"].(string); init != expected {
		t.Error("Expected: \n" + expected + "\nGot: \n" + init)
	}
}
