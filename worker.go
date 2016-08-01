package main

//import "time"
import (
	"os"
)

type Worker struct {
	jobsChannel         chan *TestJob
	reportsChannel      chan *TestJob
	workerIdlingChannel chan bool
	logger              Logger
	client              *APIClient
}

// NewWorker should be used to create a Worker instances. It ensures the correct
// initialization of all fields.
func NewWorker(jobsChannel chan *TestJob, reportsChannel chan *TestJob, workerIdlingChannel chan bool) *Worker {
	logger := Logger{"Worker", os.Stdout}
	return &Worker{
		jobsChannel:         jobsChannel,
		reportsChannel:      reportsChannel,
		workerIdlingChannel: workerIdlingChannel,
		logger:              logger,
		client:              NewClient(logger),
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
	// TODO: Prepare the repo (fetch and stuff) when the test run is changed
	nextJob.Run(w.logger)

	// Inform manager that we are done in order to set
	// workerCurrentJobCostPredictionSeconds back to zero
	w.workerIdlingChannel <- true

	go func() {
		w.reportsChannel <- nextJob
	}()
}
