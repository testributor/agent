package main

//import "time"
import (
	"os"
)

type Worker struct {
	jobsChannel         chan Job
	reportsChannel      chan Job
	workerIdlingChannel chan bool
	logger              Logger
	client              *APIClient
	lastTestRunId       int
	project             *Project
}

// NewWorker should be used to create a Worker instances. It ensures the correct
// initialization of all fields.
func NewWorker(jobsChannel chan Job, reportsChannel chan Job, workerIdlingChannel chan bool, project *Project) *Worker {
	logger := Logger{"Worker", os.Stdout}
	return &Worker{
		jobsChannel:         jobsChannel,
		reportsChannel:      reportsChannel,
		workerIdlingChannel: workerIdlingChannel,
		logger:              logger,
		client:              NewClient(logger),
		project:             project,
	}
}

func (w *Worker) Start() {
	w.logger.Log("Entering loop")
	for {
		w.RunJob()
	}
}

// RunJobs reads a job from the jobsChannel and runs it.
func (w *Worker) RunJob() {
	nextJob := <-w.jobsChannel

	if w.lastTestRunId != nextJob.GetTestRunId() {
		w.project.SetupTestEnvironment(nextJob.GetCommitSha(), w.logger)
	}

	nextJob.Run(w.logger)

	w.lastTestRunId = nextJob.GetTestRunId()

	// Inform manager that we are done in order to set
	// workerCurrentJobCostPredictionSeconds back to zero
	w.workerIdlingChannel <- true

	go func() {
		w.reportsChannel <- nextJob
	}()
}
