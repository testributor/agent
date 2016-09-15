package main

type Job interface {
	Run(Logger)
	GetCostPredictionSeconds() float64
	GetTestRunId() int
	GetId() string
	GetCommitSha() string
}
