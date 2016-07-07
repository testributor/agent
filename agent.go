package main

import (
	"fmt"
)

func main() {
	newJobsChannel := make(chan []TestJob)
	jobsChannel := make(chan *TestJob)
	reportsChannel := make(chan *TestJob)

	manager := Manager{newJobsChannel: newJobsChannel, jobsChannel: jobsChannel}
	worker := Worker{jobsChannel}
	reporter := Reporter{reportsChannel}

	fmt.Println(manager, worker, reporter)
}
