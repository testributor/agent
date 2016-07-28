package main

import (
	"testing"
)

func TestParseChannelsWhenThereIsANewReport(t *testing.T) {
	reportsChan := make(chan *TestJob)
	r := NewReporter(reportsChan)

	go func() {
		reportsChan <- &TestJob{Id: 123}
	}()

	r.ParseChannels()
	if len(r.reports) < 1 || r.reports[0].Id != 123 {
		t.Error("It should put the new TestJob in the reports list")
	}
}

func TestParseChannelsWhenActiveServerIsDone(t *testing.T) {
	reportsChan := make(chan *TestJob)
	r := NewReporter(reportsChan)
	r.activeSenders = 2

	go func() {
		r.activeSenderDone <- true
	}()

	r.ParseChannels()
	if r.activeSenders != 1 {
		t.Error("It should decrement activeSenders")
	}
}
