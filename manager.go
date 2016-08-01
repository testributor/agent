package main

import (
	"os"
	"strconv"
	"time"
)

const (
	MIN_WORKLOAD_SECONDS                    = 10
	NO_JOBS_ON_TESTRIBUTOR_TIMEOUT_SECONDS  = 5
	REMAINING_WORKLOAD_CHECK_TIMOUT_SECONDS = 5
)

type Manager struct {
	jobsChannel                           chan *TestJob
	newJobsChannel                        chan []TestJob // TODO: Make this a pointer to slice?
	workerIdlingChannel                   chan bool
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
		jobsChannel:         jobsChannel,
		newJobsChannel:      make(chan []TestJob),
		workerIdlingChannel: make(chan bool),
		logger:              logger,
		client:              NewClient(logger),
	}
}

// FetchJobs makes a call to Testributor api and fetches the next batch of jobs.
// When finished, it writes the jobs to the newJobsChannel.
// If no pending jobs are found on server, it schedules an other call to FetchJobs.
// If there are jobs, it schedules a call to checkWorkload and exits.
// checkWorkload will call FetchJobs again when needed.
func (m *Manager) FetchJobs() {
	result, err := m.client.FetchJobs()
	if err != nil {
		panic("Tried to fetch some jobs but there was an error: " + err.Error())
	}
	var jobs = make([]TestJob, 0, 10)
	for _, job := range result.([]interface{}) {
		testJob := NewTestJob(job.(map[string]interface{}))
		testJob.QueuedAtSecondsSinceEpoch = time.Now().Unix()
		jobs = append(jobs, testJob)
	}

	if len(jobs) > 0 {
		m.logger.Log("Fetched " + strconv.Itoa(len(jobs)) + " jobs")
		m.newJobsChannel <- jobs
		// Schedule next check of remaining workload
		go func() {
			<-time.After(REMAINING_WORKLOAD_CHECK_TIMOUT_SECONDS * time.Second)
			m.checkWorkload()
		}()
	} else {
		// Schedule FetchJobs again and again until we get some jobs.
		// TODO: Use exponential backoff here (or something like that)
		go func() {
			<-time.After(NO_JOBS_ON_TESTRIBUTOR_TIMEOUT_SECONDS * time.Second)
			m.FetchJobs()
		}()
	}
}

// checkWorkload checks the remaining workload. If it isn't "low" it schedules
// another call to checkWorkload. If it is low it runs FetchJobs and exits.
// FetchJobs will schedule checkWorkload again when it successfully fetches jobs.
func (m *Manager) checkWorkload() {
	if m.LowWorkload() {
		go m.FetchJobs()
	} else {
		go func() {
			<-time.After(REMAINING_WORKLOAD_CHECK_TIMOUT_SECONDS * time.Second)
			m.checkWorkload()
		}()
	}
}

// workloadOnWorkerSeconds returns the remaining seconds of workload on the
// worker (minimum 0)
func (m *Manager) workloadOnWorkerSeconds() float64 {
	secondsLeft :=
		m.workerCurrentJobCostPredictionSeconds - float64(time.Since(m.workerCurrentJobStartedAt))

	if secondsLeft < 0 {
		return 0
	} else {
		return secondsLeft
	}
}

// TotalWorkloadInQueueSeconds return the sum of CostPredictionSeconds of
// all jobs in queue plus the workloadOnWorkerSeconds.
func (m *Manager) TotalWorkloadInQueueSeconds() float64 {
	totalWorkload := float64(0)

	for _, job := range m.jobs {
		totalWorkload += job.CostPredictionSeconds
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
		m.workerCurrentJobCostPredictionSeconds = jobToBeAssigned.CostPredictionSeconds
		m.workerCurrentJobStartedAt = time.Now()

		newJobsList := make([]TestJob, len(m.jobs)-1)
		copy(newJobsList, m.jobs[1:])
		m.jobs = newJobsList

		return true
	} else {
		return false
	}
}

func (m *Manager) ParseChannels() {
	var newJobs []TestJob

	if len(m.jobs) > 0 {
		// TODO: we also need to monitor for cancelled jobs
		select {
		case newJobs = <-m.newJobsChannel:
			m.jobs = append(m.jobs, newJobs...)
		case <-m.workerIdlingChannel:
			m.workerCurrentJobCostPredictionSeconds = 0
		case m.jobsChannel <- &m.jobs[0]:
			m.AssignJobToWorker()
		}
	} else {
		// If there are no jobs left in the list, we don't want to try to push
		// a job to the worker
		select {
		case newJobs = <-m.newJobsChannel:
			m.jobs = append(m.jobs, newJobs...)
		case <-m.workerIdlingChannel:
			m.workerCurrentJobCostPredictionSeconds = 0
		}
	}
}

// Begins the Manager's main loop which keeps the job list populated and
// feeds the worker with jobs.
func (m *Manager) Start() {
	go m.FetchJobs() // Starts the FetchJobs-checkWorkload "loop"

	m.logger.Log("Entering loop")
	for {
		m.ParseChannels()
	}
}
