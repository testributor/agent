package main

import (
	"os"
	"strconv"
	"time"
)

const (
	REPORTING_FREQUENCY_SECONDS = 5
	ACTIVE_SENDERS_LIMIT        = 3
	BEACON_THRESHOLD_SECONDS    = 12
)

type Reporter struct {
	reportsChannel          chan *TestJob
	logger                  Logger
	client                  *APIClient
	reports                 []TestJob
	lastServerCommunication time.Time
	activeSenders           int // Counts how many go routines are activelly trying to send reports
	tickerChan              <-chan time.Time
	activeSenderDone        chan bool
}

// NewReporter should be used to create a Reporter instances. It ensures the correct
// initialization of all fields.
func NewReporter(reportsChannel chan *TestJob) *Reporter {
	logger := Logger{"Reporter", os.Stdout}
	return &Reporter{
		reportsChannel:   reportsChannel,
		logger:           logger,
		client:           NewClient(logger),
		tickerChan:       time.NewTicker(time.Second * REPORTING_FREQUENCY_SECONDS).C,
		activeSenderDone: make(chan bool),
	}
}

func (r *Reporter) ParseChannels() {
	select {
	case testJob := <-r.reportsChannel:
		r.reports = append(r.reports, *testJob)
	case <-r.activeSenderDone:
		r.activeSenders -= 1
	case <-r.tickerChan:
		if r.activeSenders < ACTIVE_SENDERS_LIMIT && len(r.reports) > 0 {
			go r.SendReports(r.reports)
			r.reports = []TestJob{}
			r.activeSenders += 1
		} else if r.NeedToBeacon() {
			go func() {
				if _, err := r.client.Beacon(); err != nil {
					panic("Tried to beacon but there was an error: " + err.Error())
				}
				r.lastServerCommunication = time.Now()
			}()
		}
	}
}

func (r *Reporter) Start() {
	r.logger.Log("Entering loop")
	for {
		r.ParseChannels()
	}
}

// NeedToBeacon returns true if BEACON_THRESHOLD_SECONDS have passed since the
// last beacon request.
func (r *Reporter) NeedToBeacon() bool {
	return time.Since(r.lastServerCommunication).Seconds() > BEACON_THRESHOLD_SECONDS
}

// SendReports takes a slice of TestJobs and sends it to Testributor. It will
// continue trying until successfully sent. This method should be run as a go
// routine to avoid blocking the worker in case of network issues. This means
// that if manager successfully fetches jobs, but reporter cannot report them
// back (for whatever reason), we will be creating an infinite number of
// background routines trying to send the reports to Testributor. This won't only
// fill the memory at some point, but also take over the network resources
// trying to communicate with Testributor from a large number of different threads.
// To avoid this issue, we keep a track of "active" SendReport routines (using
// a counter which decrements through a channel when routines exit). We apply a
// sane limit to the number of these routines (ACTIVE_SENDERS_LIMIT).
//
// TODO: Handle the cancellation of jobs
func (r *Reporter) SendReports(reports []TestJob) error {
	defer func() { r.activeSenderDone <- true }() // decrement activeSenders

	r.logger.Log("Sending " + strconv.Itoa(len(reports)) + " reports")
	_, err := r.client.UpdateTestJobs(reports)
	if err != nil {
		r.logger.Log(err.Error())
		return err
	}
	r.lastServerCommunication = time.Now()

	return nil
}
