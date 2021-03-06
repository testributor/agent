package main

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestParseChannelsWhenThereIsANewReport(t *testing.T) {
	reportsChan := make(chan *TestJob)
	r := NewReporter(reportsChan, make(chan []int))

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
	r := NewReporter(reportsChan, make(chan []int))
	r.activeSenders = 2

	go func() {
		r.activeSenderDone <- true
	}()

	r.ParseChannels()
	if r.activeSenders != 1 {
		t.Error("It should decrement activeSenders")
	}
}

func TestDeleteTestRunIds(t *testing.T) {
	reportsChan := make(chan *TestJob)
	r := NewReporter(reportsChan, make(chan []int))

	responseText := `{"delete_test_runs":[1976]}`
	var result interface{}
	err := json.Unmarshal(([]byte)(responseText), &result)
	if err != nil {
		t.Error(err.Error())
	}

	if !reflect.DeepEqual(r.deleteTestRunIds(result), []int{1976}) {
		t.Error("It should return []int{1976}")
	}
}
