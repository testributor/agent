package main

import (
	"encoding/json"
	"testing"
	"time"
)

var APIresponse []byte = []byte(`
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
}]
`)

var builder TestJobBuilder

func prepareBuilder() {
	var parsedResponse interface{}
	_ = json.Unmarshal(APIresponse, &parsedResponse)

	builder = TestJobBuilder(parsedResponse.([]interface{})[0].(map[string]interface{}))
}

func TestBuilderId(t *testing.T) {
	prepareBuilder()

	if builder.id() != 109136 {
		t.Error("It should return the correct id (109136) but got: ", builder.id())
	}
}

// http://stackoverflow.com/a/522281/974285
// http://stackoverflow.com/a/34422459/974285
func TestBuilderCreatedAt(t *testing.T) {
	prepareBuilder()

	parsedTime, _ := time.Parse(time.RFC3339, "2016-07-09T09:03:05.717Z")
	if builder.createdAt() != parsedTime {
		t.Error("It should parse 2016-07-09T09:03:05.717Z but got: ", builder.createdAt())
	}
}

func TestBuilderSentAtSecondsSinceEpoch(t *testing.T) {
	prepareBuilder()

	if builder.sentAtSecondsSinceEpoch() != 1468054988 {
		t.Error("It should return 1468054988 but got: ", builder.sentAtSecondsSinceEpoch())
	}
}

func TestBuilderCostPredicitonSeconds(t *testing.T) {
	prepareBuilder()

	if builder.costPredictionSeconds() != 1.824951 {
		t.Error("It should return 1.824951 but got: ", builder.costPredictionSeconds())
	}
}
