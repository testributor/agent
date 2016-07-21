package main

import (
	"github.com/ispyropoulos/agent/system_command"
	"strconv"
	"time"
)

type TestJob struct {
	id                         int
	costPredictionSeconds      float64
	sentAtSecondsSinceEpoch    int64
	startedAtSecondsSinceEpoch int64
	createdAt                  time.Time
	command                    string
	result                     string
	resultType                 int
}

// This is a custom type based on the type return my APIClient's FetchJobs
// function. We add methods on this type to parse the various fields and return
// them in a format suitable for TestJob fields.
type TestJobBuilder map[string]interface{}

func (builder *TestJobBuilder) id() int {
	return int((*builder)["id"].(float64))
}

func (builder *TestJobBuilder) costPredictionSeconds() float64 {
	switch (*builder)["cost_prediction"].(type) {
	case string:
		costPredictionSeconds, err :=
			strconv.ParseFloat((*builder)["cost_prediction"].(string), 64)
		if err != nil {
			panic("Invalid format for cost prediction: " + err.Error())
		}

		return costPredictionSeconds
	default:
		return 0
	}
}

func (builder *TestJobBuilder) createdAt() time.Time {
	createdAt, err := time.Parse(time.RFC3339, (*builder)["created_at"].(string))
	if err != nil {
		createdAt = *new(time.Time)
	}

	return createdAt
}

func (builder *TestJobBuilder) sentAtSecondsSinceEpoch() float64 {
	return (*builder)["sent_at_seconds_since_epoch"].(float64)
}

// TestJobNew is used to create a TestJob from the API response
func TestJobNew(jobData map[string]interface{}) TestJob {
	builder := TestJobBuilder(jobData)

	testJob := TestJob{
		id: builder.id(),
		costPredictionSeconds:   builder.costPredictionSeconds(),
		sentAtSecondsSinceEpoch: builder.sentAtSecondsSinceEpoch(),
		createdAt:               builder.createdAt(),
	}

	return testJob
}
