package main

import (
	"encoding/json"
	"io/ioutil"
	"strconv"
	"testing"
	"time"
)

var FetchJobsTestJobsResponse []byte = []byte(`
[{
  "command":"bin/rails runner -e test '$LOAD_PATH.push(\"#{Rails.root}/test\"); require \"test/controllers/dashboard_controller_test.rb\".gsub(/^test\\//,\"\")'",
  "created_at":"2016-07-09T09:03:05.717Z",
  "id":109136,
  "cost_prediction":"1.824951",
  "sent_at_seconds_since_epoch":1468054988,
  "test_run":{"commit_sha":"f151713e400ac3d8dc1291fe21a413a6f813072d",
    "id":1915,
    "project":{"repository_ssh_url":"git@github.com:ispyropoulos/katana.git"}
  }
},
{
  "command":"bin/rails runner -e test '$LOAD_PATH.push(\"#{Rails.root}/test\"); require \"test/controllers/dashboard_controller_test.rb\".gsub(/^test\\//,\"\")'",
  "created_at":"2016-07-09T09:03:05.717Z",
  "id":109136,
  "cost_prediction":"0",
  "sent_at_seconds_since_epoch":1468054988,
  "test_run":{"commit_sha":"f151713e400ac3d8dc1291fe21a413a6f813072d",
    "id":1915,
    "project":{"repository_ssh_url":"git@github.com:ispyropoulos/katana.git"}
  }
}]
`)

func prepareTestJobBuilder() TestJobBuilder {
	var parsedResponse interface{}
	_ = json.Unmarshal(FetchJobsTestJobsResponse, &parsedResponse)

	return TestJobBuilder(parsedResponse.([]interface{})[0].(map[string]interface{}))
}

func prepareNoPredictionTestJobBuilder() TestJobBuilder {
	var parsedResponse interface{}
	_ = json.Unmarshal(FetchJobsTestJobsResponse, &parsedResponse)

	return TestJobBuilder(parsedResponse.([]interface{})[1].(map[string]interface{}))
}

func TestBuilderId(t *testing.T) {
	builder := prepareTestJobBuilder()

	if builder.id() != 109136 {
		t.Error("It should return the correct id (109136) but got: ", builder.id())
	}
}

// http://stackoverflow.com/a/522281/974285
// http://stackoverflow.com/a/34422459/974285
func TestBuilderCreatedAt(t *testing.T) {
	builder := prepareTestJobBuilder()

	parsedTime, _ := time.Parse(time.RFC3339, "2016-07-09T09:03:05.717Z")
	if builder.createdAt() != parsedTime {
		t.Error("It should parse 2016-07-09T09:03:05.717Z but got: ", builder.createdAt())
	}
}

func TestBuilderSentAtSecondsSinceEpoch(t *testing.T) {
	builder := prepareTestJobBuilder()

	if builder.sentAtSecondsSinceEpoch() != 1468054988 {
		t.Error("It should return 1468054988 but got: ", builder.sentAtSecondsSinceEpoch())
	}
}

func TestBuilderCostPredictionSeconds(t *testing.T) {
	builder := prepareTestJobBuilder()

	if builder.costPredictionSeconds() != 1.824951 {
		t.Error("It should return 1.824951 but got: ", builder.costPredictionSeconds())
	}
}

func TestBuilderCostPredictionSecondsWhenPredictionIsZero(t *testing.T) {
	builder := prepareNoPredictionTestJobBuilder()

	if builder.costPredictionSeconds() != NO_PREDICTION_WORKLOAD_SECONDS {
		t.Error("It should return "+strconv.Itoa(NO_PREDICTION_WORKLOAD_SECONDS)+" but got: ", builder.costPredictionSeconds())
	}
}

func TestBuilderTestRunId(t *testing.T) {
	builder := prepareTestJobBuilder()

	if builder.testRunId() != 1915 {
		t.Error("It should return 1915 but got: ", builder.testRunId())
	}
}

func TestBuilderCommitSha(t *testing.T) {
	builder := prepareTestJobBuilder()

	if builder.commitSha() != "f151713e400ac3d8dc1291fe21a413a6f813072d" {
		t.Error("It should return f151713e400ac3d8dc1291fe21a413a6f813072d but got: ", builder.commitSha())
	}
}

func TestRunSettingWorkerInQueueSeconds(t *testing.T) {
	testJob := TestJob{
		Id:                        23,
		TestRunId:                 12,
		CostPredictionSeconds:     12,
		SentAtSecondsSinceEpoch:   100,
		Command:                   "ls",
		QueuedAtSecondsSinceEpoch: time.Now().Unix() - 2,
	}
	testJob.Run(Logger{"test", ioutil.Discard})

	// Calling Run should only take some milliseconds so rounded it should be 2 seconds.
	if testJob.WorkerInQueueSeconds != 2 {
		t.Error("It should set WorkerInQueueSeconds to 2 but it is: ", testJob.WorkerInQueueSeconds)
	}
}

func TestRunSettingWorkerCommandRunSeconds(t *testing.T) {
	testJob := TestJob{
		Id:                        23,
		TestRunId:                 12,
		CostPredictionSeconds:     12,
		SentAtSecondsSinceEpoch:   100,
		Command:                   "sleep 1",
		QueuedAtSecondsSinceEpoch: time.Now().Unix() - 2,
	}
	testJob.Run(Logger{"test", ioutil.Discard})

	// Calling Run should only take some milliseconds so rounded it should be 1 seconds.
	if testJob.WorkerCommandRunSeconds != 1 {
		t.Error("It should set WorkerCommandRunSeconds to 1 but it is: ", testJob.WorkerCommandRunSeconds)
	}
}

func TestJsonMarshal(t *testing.T) {
	testJob := TestJob{
		Id:                         23,
		TestRunId:                  12,
		CostPredictionSeconds:      12,
		SentAtSecondsSinceEpoch:    100,
		Command:                    "sleep 1",
		QueuedAtSecondsSinceEpoch:  123,
		WorkerCommandRunSeconds:    100,
		WorkerInQueueSeconds:       20,
		StartedAtSecondsSinceEpoch: 10,
		CommitSha:                  "f151713e400ac3d8dc1291fe21a413a6f813072d",
	}

	jsonData, _ := json.Marshal(testJob)
	expected := `{"id":23,"cost_prediction_seconds":12,"sent_at_seconds_since_epoch":100,"started_at_seconds_since_epoch":10,"created_at":"0001-01-01T00:00:00Z","command":"sleep 1","result":"","status":0,"test_run_id":12,"worker_in_queue_seconds":20,"worker_command_run_seconds":100,"QueuedAtSecondsSinceEpoch":123,"CommitSha":"f151713e400ac3d8dc1291fe21a413a6f813072d"}`

	if string(jsonData) != expected {
		t.Error("Expected: \n", expected, "\nGot: \n", string(jsonData))
	}
}
