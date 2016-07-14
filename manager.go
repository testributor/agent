package main

import (
	"os"
	"strconv"
	"time"
)

const (
	MIN_WORKLOAD_SECONDS                   = 10
	NO_JOBS_ON_TESTRIBUTOR_TIMEOUT_SECONDS = 5
)

type Manager struct {
	jobsChannel                           chan *TestJob
	newJobsChannel                        chan []TestJob // TODO: Make this a pointer to slice?
	jobs                                  []TestJob
	workerCurrentJobCostPredictionSeconds float64
	workerCurrentJobStartedAt             time.Time
	logger                                Logger
	client                                *APIClient
}

// NewManager should be used to create a Manager instances. It ensures the correct
// initialization of all fields.
func NewManager(jobsChannel chan *TestJob) *Manager {
	logger := Logger{"Manager", os.Stdout}
	return &Manager{
		jobsChannel:    jobsChannel,
		newJobsChannel: make(chan []TestJob),
		logger:         logger,
		client:         NewClient(logger),
	}
}

// FetchJobs makes a call to Testributor api and fetches the next batch of jobs,
// only if the jobs list is running low on workload.
// When finished, it writes the jobs to the newJobsChannel.
func (m *Manager) FetchJobs() {
	result, err := m.client.FetchJobs()
	if err != nil {
		panic("Tried to fetch some jobs but there was an error: " + err.Error())
	}
	var jobs = make([]TestJob, 0, 10)
	for _, job := range result.([]interface{}) {
		jobs = append(jobs, TestJobNew(job.(map[string]interface{})))
	}

	if len(jobs) > 0 {
		m.logger.Log("Fetched " + strconv.Itoa(len(jobs)) + " jobs")
		m.newJobsChannel <- jobs
	} else {
		// TODO: Use exponential backoff here (or something like that)
		time.Sleep(NO_JOBS_ON_TESTRIBUTOR_TIMEOUT_SECONDS * time.Second)
		m.FetchJobs()
	}
}

func (m *Manager) workloadOnWorkerSeconds() float64 {
	secondsLeft :=
		m.workerCurrentJobCostPredictionSeconds - float64(time.Since(m.workerCurrentJobStartedAt))

	if secondsLeft < 0 {
		return 0
	} else {
		return secondsLeft
	}
}

func (m *Manager) TotalWorkloadInQueueSeconds() float64 {
	totalWorkload := float64(0)
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

// AssignJobToWorker removes the first job from the queue and writes the
// workerCurrentJobCostPredictionSeconds and workerCurrentJobStartedAt
// attributes.
func (m *Manager) AssignJobToWorker() bool {
	if length := len(m.jobs); length > 0 {
		jobToBeAssigned := m.jobs[0]
		m.workerCurrentJobCostPredictionSeconds = jobToBeAssigned.costPredictionSeconds
		m.workerCurrentJobStartedAt = time.Now()

		// NOTE: copy to a new slice to avoid growing the underlying array indefinitely
		// This will create garbage to be collected (the old m.jobs slice)
		// so we might want to do it less frequently (for example every 100
		// assignments or something)
		newJobsList := make([]TestJob, len(m.jobs)-1)
		copy(newJobsList, m.jobs[1:])
		m.jobs = newJobsList

		return true
	} else {
		return false
	}
}

// Blocks until either a new job batch has arrived or a job can be sent to the
// worker.
func (m *Manager) ParseChannels() {
	var newJobs []TestJob

	if len(m.jobs) > 0 {
		// TODO: we also need to monitor for cancelled jobs
		select {
		case newJobs = <-m.newJobsChannel:
			// Write the new jobs in the jobs list
			m.jobs = append(m.jobs, newJobs...)
		case m.jobsChannel <- &m.jobs[0]:
			m.AssignJobToWorker()

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
	}
}

// Begins the Manager's main loop which keeps the job list populated and
// feeds the worker with jobs.
func (m *Manager) Start() {
	go m.FetchJobs()

	m.logger.Log("Entering loop")
	for {
		m.ParseChannels()
	}
}
