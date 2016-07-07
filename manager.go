package main

import (
	"errors"
	"strconv"
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
	logger                                Logger
}

// FetchJobs makes a call to Testributor api and fetches the next batch of jobs,
// only if the jobs list is running low on workload.
// When finished, it writes the jobs to the newJobsChannel.
func (m *Manager) FetchJobs() {
	jobs := []TestJob{
		TestJob{10},
		TestJob{2},
	}

	m.logger.Log("Fetched " + strconv.Itoa(len(jobs)) + " jobs")
	m.newJobsChannel <- jobs
}

func (m *Manager) workloadOnWorkerSeconds() int {
	secondsLeft :=
		m.workerCurrentJobCostPredictionSeconds - int(time.Since(m.workerCurrentJobStartedAt))

	if secondsLeft < 0 {
		return 0
	} else {
		return secondsLeft
	}
}

func (m *Manager) TotalWorkloadInQueueSeconds() int {
	totalWorkload := 0
	for _, job := range m.jobs {
		totalWorkload += job.costPredictionSeconds
	}

	totalWorkload += m.workloadOnWorkerSeconds()

	return totalWorkload
}

// LowWorkload returns true when the total workload (the one the list + the one
// already on the worker) is lower than MIN_WORKLOAD_SECONDS.
func (m *Manager) LowWorkload() bool {
	return m.TotalWorkloadInQueueSeconds() <= MIN_WORKLOAD_SECONDS
}

func (m *Manager) PopJob() (TestJob, error) {
	if length := len(m.jobs); length > 0 {
		nextJob := m.jobs[0]
		// TODO: Investigate if we have to use "copy" to make the underlying array smalller
		m.jobs = m.jobs[1:]

		return nextJob, nil
	} else {
		return TestJob{}, errors.New("No jobs left")
	}
}

// Blocks until either a new job batch has arrived or a job can be sent to the
// worker.
func (m *Manager) ParseChannels() {
	var newJobs []TestJob
	nextJob, err := m.PopJob()

	if err == nil {
		// TODO: we also need to monitor for cancelled jobs
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
		m.jobs = append(m.jobs, newJobs...)
		// Write the new jobs in the jobs list
	}
}

// Begins the Manager's main loop which keeps the job list populated and
// feeds the worker with jobs.
func (m *Manager) Start() {
	go m.FetchJobs()

	for {
		m.ParseChannels()
	}
}
