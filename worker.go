package main

type Worker struct {
	jobsChannel chan *TestJob
}
