package main

import (
	"io/ioutil"
	"testing"
	"time"
)

func TestRunJobSendingToWorkerIdlingChannel(t *testing.T) {
	jobsChannel := make(chan *TestJob)
	reportsChannel := make(chan *TestJob)
	workerIdlingChannel := make(chan bool)
	worker := NewWorker(jobsChannel, reportsChannel, workerIdlingChannel, &Project{})
	worker.logger = Logger{"", ioutil.Discard}
	workerIdling := false

	go func() {
		jobsChannel <- &TestJob{Command: "ls"}
	}()

	go worker.RunJob()

	// Read workerIdling or timeout after 1 second
	timer := time.NewTimer(time.Second * 1).C
	select {
	case <-timer:
	case workerIdling = <-workerIdlingChannel:
	}

	if !workerIdling {
		t.Error("It should send true to worker idling channel")
	}
}
