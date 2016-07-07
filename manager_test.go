package main

import (
	"testing"
)

func TestPopJobWhenNonEmpty(t *testing.T) {
	manager := Manager{}
	expectedJob := TestJob{4}
	manager.jobs = []TestJob{
		expectedJob,
		TestJob{1},
		TestJob{2},
		TestJob{3},
	}

	job, err := manager.PopJob()
	if job != expectedJob {
		t.Error("Does not pop the last TestJob")
	}

	if err != nil {
		t.Error("Should not return error")
	}
}

func TestPopJobWhenEmpty(t *testing.T) {
	manager := Manager{}
	manager.jobs = []TestJob{}

	_, err := manager.PopJob()
	if err.Error() != "No jobs left" {
		t.Error("Should return error 'No jobs left', but got: ", err.Error())
	}
}

func TestTotalWorkloadInQueueSeconds(t *testing.T) {
	manager := Manager{
		workerCurrentJobCostPredictionSeconds: 1,
		jobs: []TestJob{
			TestJob{2},
			TestJob{10},
			TestJob{100},
		},
	}

	// seconds left on worker is 0 since the time passed since the default
	// workerCurrentJobStartedAt is a really big number.
	if workload := manager.TotalWorkloadInQueueSeconds(); workload != 112 {
		t.Error("Expected 112, got: ", workload)
	}
}

func TestLowWorkload(t *testing.T) {
	manager := Manager{
		workerCurrentJobCostPredictionSeconds: 1,
		jobs: []TestJob{
			TestJob{1},
			TestJob{2},
			TestJob{3},
		},
	}

	if !manager.LowWorkload() {
		t.Error("LowWorkload should return true when workload is 6")
	}
}

func TestParseChannelsWhenNoJobExists(t *testing.T) {
	manager := Manager{jobs: []TestJob{}}
	go func() {
		manager.newJobsChannel <- []TestJob {
			TestJob{},
		}
	}
	manager.ParseChannels()
}
