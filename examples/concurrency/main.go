package main

import (
	"fmt"
	"runtime"
	"sync"
	"time"
)

// processJob simulates processing a single job by sleeping briefly
// and then returning twice the job ID as the result.
func processJob(jobID int) int {
	time.Sleep(200 * time.Millisecond) // Simulate some work
	return jobID * 2
}

// logResult logs the result of a processed job.
func logResult(workerID, jobID, result int) {
	fmt.Printf("Worker %d processed job %d with result %d\n", workerID, jobID, result)
}

// worker receives jobs from jobsChan, processes them, and sends results to resultsChan.
func worker(id int, jobs <-chan int, results chan<- int) {
	for j := range jobs {
		res := processJob(j)
		logResult(id, j, res)
		results <- res
	}
}

// generateJobs creates a slice of job IDs.
func generateJobs(n int) []int {
	jobs := make([]int, n)
	for i := 0; i < n; i++ {
		jobs[i] = i + 1
	}
	return jobs
}

func main() {
	fmt.Println("Concurrency Example Project (Enhanced)")

	numberOfJobs := 10

	// Create buffered channels for jobs and results.
	jobsChan := make(chan int, numberOfJobs)
	resultsChan := make(chan int, numberOfJobs)

	// Dynamically launch worker goroutines based on the machine's CPU count.
	numWorkers := runtime.NumCPU()
	var wg sync.WaitGroup
	for w := 1; w <= numWorkers; w++ {
		wg.Add(1)
		go func(workerId int) {
			defer wg.Done()
			worker(workerId, jobsChan, resultsChan)
		}(w)
	}

	// Generate and send jobs to the workers.
	jobs := generateJobs(numberOfJobs)
	for _, j := range jobs {
		jobsChan <- j
	}
	close(jobsChan)

	// Wait for all workers to finish processing.
	wg.Wait()
	close(resultsChan)

	// Print the results.
	fmt.Println("Results:")
	for r := range resultsChan {
		fmt.Printf("Result: %d\n", r)
	}
}
