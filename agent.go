package main

import "os"

func main() {
	logger := Logger{"Main", os.Stdout}
	// Check if env vars are set, use defaults if not (or exit if needed)
	// and initialize oauth token.
	err := SetupClientData()
	if err != nil {
		logger.Log(err.Error())
		os.Exit(1)
	}

	err = EnsureGit(logger)
	if err != nil {
		logger.Log(err.Error())
		os.Exit(1)
	}

	jobsChannel := make(chan *TestJob)
	reportsChannel := make(chan *TestJob)

	manager := NewManager(jobsChannel)
	worker := NewWorker(jobsChannel, reportsChannel)
	reporter := NewReporter(reportsChannel)

	go worker.Start()
	go reporter.Start()
	manager.Start()
}
