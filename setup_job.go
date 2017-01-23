package main

import (
	"strconv"
)

type SetupJob struct {
	Id                        string  `json:"id"`
	CostPredictionSeconds     float64 `json:"cost_prediction_seconds"`
	SentAtSecondsSinceEpoch   int64   `json:"sent_at_seconds_since_epoch"`
	Result                    string  `json:"result"`
	TestRunId                 int     `json:"test_run_id"`
	CommitSha                 string  `json:"commit_sha"`
	TestributorYml            string  `json:"testributor_yml"`
	QueuedAtSecondsSinceEpoch int64
}

// This is a custom type which handles the result of APIClient's FetchJobs
// function. We add methods on this type to parse the various fields and return
// them in a format suitable for SetupJob fields.
type SetupJobBuilder map[string]interface{}

func (builder *SetupJobBuilder) id() string {
	return "setup_job_" + strconv.Itoa(builder.testRunId())
}

// This method is duplicated in test_job.go.
// TODO: DRY
func (builder *SetupJobBuilder) costPredictionSeconds() float64 {
	switch v := (*builder)["cost_prediction"].(type) {
	case float64:
		costPredictionSeconds := v

		// If no prediction is available use the default "huge" value to avoid
		// fetching more jobs.
		if costPredictionSeconds == 0 {
			return NO_PREDICTION_WORKLOAD_SECONDS
		} else {
			return costPredictionSeconds
		}
	default:
		return NO_PREDICTION_WORKLOAD_SECONDS
	}
}

func (builder *SetupJobBuilder) sentAtSecondsSinceEpoch() int64 {
	return int64((*builder)["sent_at_seconds_since_epoch"].(float64))
}

func (builder *SetupJobBuilder) testRunId() int {
	return int((*builder)["test_run"].(map[string]interface{})["id"].(float64))
}

func (builder *SetupJobBuilder) commitSha() string {
	return (*builder)["test_run"].(map[string]interface{})["commit_sha"].(string)
}

func (builder *SetupJobBuilder) testributorYml() string {
	return (*builder)["testributor_yml"].(string)
}

func NewSetupJob(jobData map[string]interface{}) *SetupJob {
	return &SetupJob{}
}

func (setupJob *SetupJob) Run(logger Logger) {
	logger.Log("I'm running a setup Job man!")
}

func (setupJob *SetupJob) GetCostPredictionSeconds() float64 {
	return setupJob.CostPredictionSeconds
}

func (setupJob *SetupJob) GetTestRunId() int {
	return setupJob.TestRunId
}

func (setupJob *SetupJob) GetId() string {
	return setupJob.Id
}

func (setupJob *SetupJob) GetCommitSha() string {
	return setupJob.CommitSha
}

func (setupJob *SetupJob) SetQueuedAtSecondsSinceEpoch(timestamp int64) {
	setupJob.QueuedAtSecondsSinceEpoch = timestamp
}
