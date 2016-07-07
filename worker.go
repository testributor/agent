package main

//import "time"
import "strconv"

type Worker struct {
	jobsChannel chan *TestJob
	logger      Logger
}

func (w *Worker) Start() {
	var nextJob *TestJob
	for {
		nextJob = <-w.jobsChannel
		// Run the job
		w.logger.Log("Cost prediction: " + strconv.Itoa(nextJob.costPredictionSeconds))
	}
}
