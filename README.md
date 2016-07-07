# Agent

Testributor agent re-written in Go language for maximum portability

## Manager

This is the main thread. It keeps the list of jobs filled and passes jobs to
the worker thread for execution. Its work in pseudocode:

for {
  - Do we need more jobs?
    yes: spawn a go routine to fetch more jobs
    no: do nothing

  select {
    - send jobs to the job channel (read by the worker)
    - read fetched jobs from the jobs channel (written by the go routine spawned previously)
      write the jobs to the job list
  }
}


## Worker

Reads jobs from the jobs channel (written by the Manager thread) and executes them.
When done, the updated job (with results and everything) is written to the reports
channel (blocking until read).


## Reporter

Reads the reports channel and pushes reports in the reports list. When a number
of seconds has passed, the reports are sent back to Katana and they are removed
from the reports list.
