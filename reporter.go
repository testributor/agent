package main

type Reporter struct {
	reportsChannel chan *TestJob
}
