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

func (builder *TestJobBuilder) command() string {
	return (*builder)["command"].(string)
}

func (builder *TestJobBuilder) sentAtSecondsSinceEpoch() int64 {
	return int64((*builder)["sent_at_seconds_since_epoch"].(float64))
}

// NewTestJob is used to create a TestJob from the API response
func NewTestJob(jobData map[string]interface{}) TestJob {
	builder := TestJobBuilder(jobData)

	testJob := TestJob{
		id: builder.id(),
		costPredictionSeconds:   builder.costPredictionSeconds(),
		sentAtSecondsSinceEpoch: builder.sentAtSecondsSinceEpoch(),
		createdAt:               builder.createdAt(),
		command:                 builder.command(),
	}

	return testJob
}

func (testJob *TestJob) Run(logger Logger) {
	testJob.startedAtSecondsSinceEpoch = time.Now().Unix()

	logger.Log("Running " + testJob.command)

	res, err := system_command.Run(testJob.command, logger)
	if err != nil {
		testJob.result = err.Error()
		testJob.resultType = system_command.RESULT_TYPES["error"]
	} else {
		testJob.result = res.CombinedOutput
		testJob.resultType = res.ResultType
	}
}
