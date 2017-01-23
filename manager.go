package main

import (
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	MIN_WORKLOAD_SECONDS                    = 10
	NO_JOBS_ON_TESTRIBUTOR_TIMEOUT_SECONDS  = 5
	REMAINING_WORKLOAD_CHECK_TIMOUT_SECONDS = 5
)

type Manager struct {
	jobsChannel                           chan Job
	newJobsChannel                        chan []Job
	cancelledTestRunIdsChan               chan []int
	workerIdlingChannel                   chan bool
	jobs                                  []Job
	workerCurrentJobCostPredictionSeconds float64
	workerCurrentJobStartedAt             time.Time
	logger                                Logger
	client                                *APIClient
}

// NewManager should be used to create a Manager instances. It ensures the correct
// initialization of all fields.
func NewManager(jobsChannel chan Job, cancelledTestRunIdsChan chan []int) *Manager {
	logger := Logger{"Manager", os.Stdout}
	return &Manager{
		jobsChannel:             jobsChannel,
		cancelledTestRunIdsChan: cancelledTestRunIdsChan,
		newJobsChannel:          make(chan []Job),
		workerIdlingChannel:     make(chan bool),
		logger:                  logger,
		client:                  NewClient(logger),
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
	var jobs = make([]Job, 0, 10)
	switch v := result.(type) {
	case []interface{}:
		for _, job := range v {
			newJob := NewTestJob(job.(map[string]interface{}))
			newJob.SetQueuedAtSecondsSinceEpoch(time.Now().Unix())
			jobs = append(jobs, newJob)
		}
	case map[string]interface{}:
		jobs = append(jobs, NewSetupJob(v))
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
		totalWorkload += job.GetCostPredictionSeconds()
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
		m.workerCurrentJobCostPredictionSeconds = jobToBeAssigned.GetCostPredictionSeconds()
		m.workerCurrentJobStartedAt = time.Now()

		newJobsList := make([]Job, len(m.jobs)-1)
		copy(newJobsList, m.jobs[1:])
		m.jobs = newJobsList

		return true
	} else {
		return false
	}
}

func (m *Manager) CancelTestRuns(ids []int) {
	if len(ids) == 0 {
		return
	}

	newJobsList := []Job{}
	// Uniq set implemented as a map (http://stackoverflow.com/a/9251352)
	cancelledIdsSet := make(map[string]struct{})
	for _, job := range m.jobs {
		for _, id := range ids {
			if job.GetTestRunId() == id {
				cancelledIdsSet[strconv.Itoa(id)] = struct{}{}
			} else {
				newJobsList = append(newJobsList, job)
			}
		}
	}

	if len(cancelledIdsSet) > 0 {
		// Build a slice out of the "set" of cancelled ids
		cancelledIds := make([]string, 0, len(cancelledIdsSet))
		for id := range cancelledIdsSet {
			cancelledIds = append(cancelledIds, id)
		}
		m.logger.Log("Cancelling builds: " + strings.Join(cancelledIds, ", "))
		m.jobs = newJobsList
	}
}

func (m *Manager) ParseChannels() {
	var newJobs []Job

	// TODO: These two selects need DRYing
	if len(m.jobs) > 0 {
		// TODO: we also need to monitor for cancelled jobs
		select {
		case newJobs = <-m.newJobsChannel:
			m.jobs = append(m.jobs, newJobs...)
		case <-m.workerIdlingChannel:
			m.workerCurrentJobCostPredictionSeconds = 0
		case cancelledIds := <-m.cancelledTestRunIdsChan:
			m.CancelTestRuns(cancelledIds)
		case m.jobsChannel <- m.jobs[0]:
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
		case <-m.cancelledTestRunIdsChan:
			// Do nothing, we just read this to let the Reporter continue.
			// The reporter doesn't know if Manager has jobs in queue or not.
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
