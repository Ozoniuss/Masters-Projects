package main

import (
	"fmt"
	"sync"
	"time"
)

// Module 1.1, page 13
/*
Module:
	Name: JobHandler, instance jh.
Events:
	Request: <jh, Submit | job> : Requests a job to be processed.
	Indication: <jh, Confirm | job? : Confirms that the given job has been (or will be) processed.
Properties:
	JH1: Guaranteed response: Every submitted job is eventually confirmed.
*/

// Algorithm 1.1, page 14
/*
Implements:
	JobHandler, instance jh.
upon event <jh, Init> do
	buffer := ∅;
upon event <jh, Submit | job> do
	buffer := buffer ∪ {job};
	trigger <jh, Confirm | job>;

(internal event)
upon buffer <> ∅ do
	job := selectjob(buffer);
	process(job);
	buffer := buffer \ {job};
*/

type Job struct {
	id       int
	duration int
}

// Only 10 jobs in this example, but theoretically this processor could run
// infinitely.
const NO_JOBS = 10

// handleJobsSynchronously handles jobs synchronously, as described by algorithm
// 1.1, page 14.
//
// Implements JobHandler module, jh instance.
func handleJobsAsynchronously(submissions chan Job, confirmations [NO_JOBS]chan Job) {

	fmt.Println("start asynchronous handler")

	// upon event <jh, Init> do
	processingJobs := make([]Job, 0, 10)

	// The property JH1 is satisfied due to having an infinite for loop, and
	// storing all jobs in a buffer. See remarks about infinite loop from
	// synchronous algorithm, and storing all jobs in a (theoretically infinite)
	// buffer guarantees that all of them execute.
	for {
		select {
		// upon event  jh, Submit | job> do
		case job := <-submissions:
			fmt.Printf("received submission %d\n", job.id)
			processingJobs = append(processingJobs, job)
			confirmations[job.id] <- job
		// upon buffer <> ∅ do
		default:
			if len(processingJobs) != 0 {
				j := processingJobs[0] // fifo selection
				processingJobs = processingJobs[1:]
				processJob(j)
				fmt.Printf("job %d done\n", j.id)
			}
		}
	}
}

// processJob simulates a job being processed.
func processJob(payload Job) {
	t := time.Duration(payload.duration) * time.Second
	time.Sleep(t)
}

// See README for running instructions and analysis.
func main() {
	fmt.Println("Main process started.")

	t := time.Now()

	// Mimics unlimited buffer, there can be theoretically infinite job
	// submissions.
	submissions := make(chan Job, NO_JOBS)

	// Each job will be individually sent a confirmation when it completes, to
	// simulate a distributed system.
	var confirmations [NO_JOBS]chan Job

	// Init jobs and confirmation channels.

	var jobs [NO_JOBS]Job
	for i := 0; i < NO_JOBS; i++ {
		jobs[i] = Job{id: i, duration: (i % 3) + 1}
		confirmations[i] = make(chan Job)
	}

	// Start synchronous job handler.
	go handleJobsAsynchronously(submissions, confirmations)

	// This is only used to not exit the main goroutine only after the child
	// goroutines complete.
	w := &sync.WaitGroup{}
	w.Add(NO_JOBS)

	// Submit 10 jobs.
	for i := 0; i < NO_JOBS; i++ {
		go func(i int) {
			// Send job
			submissions <- jobs[i]
			//Listen for confirmation
			c := <-confirmations[i]
			fmt.Printf("confirmed job %d\n", c.id)
			w.Done()
		}(i)
	}

	// Exit program only after all jobs finished.
	w.Wait()

	fmt.Printf("time elapsed: %s\n", time.Since(t))

	// Wait for jobs to complete now...
	time.Sleep(30 * time.Second)
}
