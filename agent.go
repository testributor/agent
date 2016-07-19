package main

import "os"

func main() {
	logger := Logger{"Main", os.Stdout}

	printLogo(logger)

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

	project, err := NewProject(logger)
	if err != nil {
		logger.Log(err.Error())
		os.Exit(1)
	}

	err = project.Init(logger)
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

// Because we can
func printLogo(logger Logger) {
	logger.Log(`
 _______________________  _______  __  ____________  ___
/_  __/ __/ __/_  __/ _ \/  _/ _ )/ / / /_  __/ __ \/ _ \
 / / / _/_\ \  / / / , _// // _  / /_/ / / / / /_/ / , _/
/_/ /___/___/ /_/ /_/|_/___/____/\____/ /_/  \____/_/|_|

                              https://www.testributor.com
`)
}
