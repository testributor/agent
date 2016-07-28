package main

import (
	"github.com/ispyropoulos/agent/system_command"
	"strconv"
	"time"
)

type TestJob struct {
	Id                         int       `json:"id"`
	CostPredictionSeconds      float64   `json:"cost_prediction_seconds"`
	SentAtSecondsSinceEpoch    int64     `json:"sent_at_seconds_since_epoch"`
	StartedAtSecondsSinceEpoch int64     `json:"started_at_seconds_since_epoch"`
	CreatedAt                  time.Time `json:"created_at"`
	Command                    string    `json:"command"`
	Result                     string    `json:"result"`
	ResultType                 int       `json:"status"`
	TestRunId                  int       `json:"test_run_id"`
	WorkerInQueueSeconds       int64     `json:"worker_in_queue_seconds"`
	WorkerCommandRunSeconds    int64     `json:"worker_command_run_seconds"`
	QueuedAtSecondsSinceEpoch  int64
}

// This is a custom type based on the type return my APIClient's FetchJobs
// function. We add methods on this type to parse the various fields and return
// them in a format suitable for TestJob fields.
type TestJobBuilder map[string]interface{}

func (builder *TestJobBuilder) id() int {
	return int((*builder)["id"].(float64))
}

func (builder *TestJobBuilder) testRunId() int {
	testRun := (*builder)["test_run"].(map[string]interface{})

	return int(testRun["id"].(float64))
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
		Id:                      builder.id(),
		TestRunId:               builder.testRunId(),
		CostPredictionSeconds:   builder.costPredictionSeconds(),
		SentAtSecondsSinceEpoch: builder.sentAtSecondsSinceEpoch(),
		CreatedAt:               builder.createdAt(),
		Command:                 builder.command(),
	}

	return testJob
}

func (testJob *TestJob) Run(logger Logger) {
	testJob.StartedAtSecondsSinceEpoch = time.Now().Unix()

	logger.Log("Running " + testJob.Command)

	res, err := system_command.Run(testJob.Command, logger)

	if err != nil {
		testJob.Result = err.Error()
		testJob.ResultType = system_command.RESULT_TYPES["error"]
	} else {
		testJob.Result = res.CombinedOutput
		testJob.ResultType = res.ResultType
	}

	testJob.WorkerInQueueSeconds =
		testJob.StartedAtSecondsSinceEpoch - testJob.QueuedAtSecondsSinceEpoch
	testJob.WorkerCommandRunSeconds = int64(res.DurationSeconds)
}
