package main

import "os"

type Reporter struct {
	reportsChannel chan *TestJob
	logger         Logger
	client         *APIClient
}

// NewReporter should be used to create a Reporter instances. It ensures the correct
// initialization of all fields.
func NewReporter(reportsChannel chan *TestJob) *Reporter {
	logger := Logger{"Reporter", os.Stdout}
	return &Reporter{
		reportsChannel: reportsChannel,
		logger:         logger,
		client:         NewClient(logger),
	}
}

func (r *Reporter) Start() {
	r.logger.Log("Entering loop")
}
