package main

type Reporter struct {
	reportsChannel chan *TestJob
}

func (r *Reporter) Start() {
}
