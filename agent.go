package main

import (
	"github.com/tuvistavie/securerandom"
	"os"
	"strings"
)

var WorkerUUID string
var WorkerUUIDShort string

func main() {
	logger := Logger{"Main", os.Stdout}

	printLogo(logger)

	if err := setWorkerUuid(); err != nil {
		logger.Log(err.Error())
		os.Exit(1)
	}

	// Check if env vars are set, use defaults if not (or exit if needed)
	// and initialize oauth token.
	if err := SetupClientData(); err != nil {
		logger.Log(err.Error())
		os.Exit(1)
	}

	if err := EnsureGit(logger); err != nil {
		logger.Log(err.Error())
		os.Exit(1)
	}

	project, err := NewProject(logger)
	if err != nil {
		logger.Log(err.Error())
		os.Exit(1)
	}

	if err := project.Init(logger); err != nil {
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

func setWorkerUuid() error {
	var err error
	WorkerUUID, err = securerandom.Uuid()
	if err != nil {
		return err
	}
	WorkerUUIDShort = strings.Split(WorkerUUID, "-")[0]

	return nil
}
