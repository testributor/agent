package main

func main() {
	// Check if env vars are set, use defaults if not (or exit if needed)
	// and initialize oauth token.
	SetupClientData()

	jobsChannel := make(chan *TestJob)
	reportsChannel := make(chan *TestJob)

	manager := NewManager(jobsChannel)
	worker := NewWorker(jobsChannel, reportsChannel)
	reporter := NewReporter(reportsChannel)

	go worker.Start()
	go reporter.Start()
	manager.Start()
}
