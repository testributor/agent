package main

import (
	"testing"
	"time"
)

func TestAssignJobToWorkerWhenNonEmpty(t *testing.T) {
	manager := Manager{}
	expectedJob := TestJob{1, 4, 100, time.Now()}

	manager.jobs = []TestJob{
		expectedJob,
		TestJob{2, 1, 100, time.Now()},
		TestJob{3, 2, 100, time.Now()},
		TestJob{4, 3, 100, time.Now()},
	}

	oldJobs := len(manager.jobs)
	assigned := manager.AssignJobToWorker()
	if !assigned {
		t.Error("It should return true")
	}

	if len(manager.jobs) != oldJobs-1 {
		t.Error("Should pop a job from the queue")
	}

	if manager.workerCurrentJobCostPredictionSeconds !=
		expectedJob.costPredictionSeconds {
		t.Error("Should pop the first job in queue")
	}
}

func TestAssignJobToWorkerWhenEmpty(t *testing.T) {
	manager := Manager{}
	manager.jobs = []TestJob{}

	oldJobs := len(manager.jobs)
	assigned := manager.AssignJobToWorker()
	if assigned {
		t.Error("Should return false")
	}

	if oldJobs != 0 || len(manager.jobs) != 0 {
		t.Error("jobs queue should be empty")
	}
}

func TestTotalWorkloadInQueueSeconds(t *testing.T) {
	manager := Manager{
		workerCurrentJobCostPredictionSeconds: 1,
		jobs: []TestJob{
			TestJob{2, 2, 100, time.Now()},
			TestJob{3, 10, 100, time.Now()},
			TestJob{4, 100, 100, time.Now()},
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
			TestJob{2, 1, 2, time.Now()},
			TestJob{3, 2, 10, time.Now()},
			TestJob{4, 3, 100, time.Now()},
		},
	}

	if !manager.LowWorkload() {
		t.Error("LowWorkload should return true when workload is 6")
	}
}

func TestParseChannelsWhenNoJobExists(t *testing.T) {
	newJobsChannel := make(chan []TestJob)
	manager := Manager{jobs: []TestJob{}, newJobsChannel: newJobsChannel}
	go func() {
		manager.newJobsChannel <- []TestJob{
			TestJob{},
		}
	}()

	if len(manager.jobs) != 0 {
		t.Error("There should be no jobs in queue")
	}
	manager.ParseChannels()
	if len(manager.jobs) != 1 {
		t.Error("Expected to find 1 job but found: ", manager.jobs)
	}
}

func TestParseChannelsWhenJobsExists(t *testing.T) {
	jobsChannel := make(chan *TestJob)
	newJobsChannel := make(chan []TestJob)
	var newJob *TestJob

	manager := Manager{
		jobs: []TestJob{
			// Keep the cost predicton big enough to avoid a call to FetchJobs
			TestJob{2, 100, 2, time.Now()},
			TestJob{3, 200, 10, time.Now()},
		},
		jobsChannel:    jobsChannel,
		newJobsChannel: newJobsChannel,
	}

	go func() {
		newJob = <-manager.jobsChannel
	}()

	if newJob != nil {
		t.Error("No job should be read yet")
	}
	oldJobs := len(manager.jobs)
	manager.ParseChannels()
	if newJob == nil {
		t.Error("A Job should be read from jobsChannel")
	}
	if len(manager.jobs) != oldJobs-1 {
		t.Error("A Job should be removed from jobs list")
	}

	go func() {
		manager.newJobsChannel <- []TestJob{
			TestJob{},
		}
	}()
	oldJobs = len(manager.jobs)
	manager.ParseChannels()
	if len(manager.jobs) != oldJobs+1 {
		t.Error("Expected jobs to be ", oldJobs+1, "but found: ", len(manager.jobs))
	}
}