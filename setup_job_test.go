package main

import (
	"encoding/json"
	"testing"
)

var FetchJobsSetupJobResponse []byte = []byte(`
{ "type":"setup",
	"sent_at_seconds_since_epoch":1485079045,
	"cost_prediction":20,
	"test_run":{
		"id":6,
		"commit_sha":"2912ee5"
	},
	"testributor_yml":"each:\r\n  pattern: 'test\\/.*_test.rb$'\r\n  command: bin/rails runner -e test '$LOAD_PATH.push(\"#{Rails.root}/test\"); require \"%{file}\".gsub(/^test\\//,\"\")'\r\n "
}`)

func prepareSetupJobBuilder() SetupJobBuilder {
	var parsedResponse interface{}
	_ = json.Unmarshal(FetchJobsSetupJobResponse, &parsedResponse)

	return SetupJobBuilder(parsedResponse.(map[string]interface{}))
}

func TestSetupJobBuilderId(t *testing.T) {
	builder := prepareSetupJobBuilder()

	if builder.id() != "setup_job_6" {
		t.Error("builder.id() doesn't match the expected: ", builder.id())
	}
}

func TestSetupJobBuilderCostPredictionSeconds(t *testing.T) {
	builder := prepareSetupJobBuilder()

	if builder.costPredictionSeconds() != 20 {
		t.Error("It should return 20 but got: ", builder.costPredictionSeconds())
	}
}

func TestSetupJobBuilderSentAtSecondsSinceEpoch(t *testing.T) {
	builder := prepareSetupJobBuilder()

	if builder.sentAtSecondsSinceEpoch() != 1485079045 {
		t.Error("It should return 1485079045 but got: ", builder.sentAtSecondsSinceEpoch())
	}
}

func TestSetupJobBuilderTestRunId(t *testing.T) {
	builder := prepareSetupJobBuilder()

	if builder.testRunId() != 6 {
		t.Error("It should return 6 but got: ", builder.testRunId())
	}
}

func TestSetupJobBuilderCommitSha(t *testing.T) {
	builder := prepareSetupJobBuilder()

	if builder.commitSha() != "2912ee5" {
		t.Error("It should return 2912ee5 but got: ", builder.commitSha())
	}
}

func TestSetupJobBuilderTestributorYml(t *testing.T) {
	builder := prepareSetupJobBuilder()
	expectedYml := "each:\r\n  pattern: 'test\\/.*_test.rb$'\r\n  command: bin/rails runner -e test '$LOAD_PATH.push(\"#{Rails.root}/test\"); require \"%{file}\".gsub(/^test\\//,\"\")'\r\n "

	if builder.testributorYml() != expectedYml {
		t.Error("testributorYml doesn't match the expected string")
	}
}
