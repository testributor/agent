package main

import "os"

func main() {
	// Check if env vars are set, use defaults if not (or exit if needed)
	// and initialize oauth token.
	SetupClientData()
	newJobsChannel := make(chan []TestJob)
	jobsChannel := make(chan *TestJob)
	reportsChannel := make(chan *TestJob)

	manager := Manager{
		newJobsChannel: newJobsChannel,
		jobsChannel:    jobsChannel,
		logger:         Logger{"Manager", os.Stdout},
	}
	worker := Worker{
		jobsChannel: jobsChannel,
		logger:      Logger{"Worker", os.Stdout},
	}
	reporter := Reporter{reportsChannel}

	go worker.Start()
	go reporter.Start()
	manager.Start()
}
