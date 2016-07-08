package main

//import "time"
import (
	"os"
	"strconv"
)

type Worker struct {
	jobsChannel    chan *TestJob
	reportsChannel chan *TestJob
	logger         Logger
	client         *APIClient
}

// NewWorker should be used to create a Worker instances. It ensures the correct
// initialization of all fields.
func NewWorker(jobsChannel chan *TestJob, reportsChannel chan *TestJob) *Worker {
	logger := Logger{"Worker", os.Stdout}
	return &Worker{
		jobsChannel:    jobsChannel,
		reportsChannel: reportsChannel,
		logger:         logger,
		client:         NewClient(logger),
	}
}

func (w *Worker) Start() {
	var nextJob *TestJob
	for {
		nextJob = <-w.jobsChannel
		// Run the job
		w.logger.Log("Cost prediction: " + strconv.Itoa(nextJob.costPredictionSeconds))
	}
}
