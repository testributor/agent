package main

import (
	"errors"
	"fmt"
	"time"
)

const (
	MIN_WORKLOAD_SECONDS = 10
)

type Manager struct {
	newJobsChannel                        chan []TestJob
	jobsChannel                           chan *TestJob
	jobs                                  []TestJob
	workerCurrentJobCostPredictionSeconds int
	workerCurrentJobStartedAt             time.Time
}

// FetchJobs makes a call to Testributor api and fetches the next batch of jobs,
// only if the jobs list is running low on workload.
// When finished, it writes the jobs to the newJobsChannel.
func (m *Manager) FetchJobs() {
	jobs := []TestJob{
		TestJob{},
		TestJob{},
	}
	m.newJobsChannel <- jobs
}

func (m *Manager) workloadOnWorkerSeconds() int {
	return m.workerCurrentJobCostPredictionSeconds - int(time.Since(m.workerCurrentJobStartedAt))
}

// LowWorkload returns true when the total workload (the one the list + the one
// already on the worker) is lower than MIN_WORKLOAD_SECONDS.
func (m *Manager) LowWorkload() bool {
	totalWorkload := 0
	for _, job := range m.jobs {
		totalWorkload += job.costPredictionSeconds
	}

	totalWorkload += m.workloadOnWorkerSeconds()

	return totalWorkload <= MIN_WORKLOAD_SECONDS
}

func (m *Manager) PopJob() (TestJob, error) {
	if length := len(m.jobs); length > 0 {
		nextJob := m.jobs[length-1]
		// TODO: Investigate if we have to use "copy" to make the underlying array smalller
		m.jobs = m.jobs[:length-1]

		return nextJob, nil
	} else {
		return TestJob{}, errors.New("No jobs left")
	}
}

func (m *Manager) Start() {
	var (
		newJobs []TestJob
		nextJob *TestJob
	)
	fmt.Println(nextJob)

	go m.FetchJobs()

	for {
		nextJob, err := m.PopJob()

		if err == nil {
			select {
			case newJobs = <-m.newJobsChannel:
				// Write the new jobs in the jobs list
			case m.jobsChannel <- &nextJob:
				// Send a job to the worker and remove it from the list
				if m.LowWorkload() {
					go m.FetchJobs()
				}
			}
		} else {
			// If there are no jobs left in the list, we only need to wait until
			// new jobs are sent to the newJobsChannel.
			newJobs = <-m.newJobsChannel
			fmt.Println(newJobs)
			// Write the new jobs in the jobs list
		}
	}
}
