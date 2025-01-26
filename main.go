package main

import (
	"fmt"
	"log"
	"math/rand"
	"net"
	"net/http"
	"sync"
	"time"
)

var shutdown = make(chan struct{})

type RequestType struct {
	Type     string // "POST" or "GET"
	JobIndex int
}

func main() {
	rand.Seed(time.Now().UnixNano())

	config, err := LoadConfig("config.json")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	stats := &RequestStats{
		Results: make([]RequestResult, 0, config.TotalRequests),
	}

	fmt.Printf("Starting test at: %s\n", time.Now().UTC().Format(time.RFC3339))
	fmt.Printf("Configuration:\n")
	fmt.Printf("Total requests: %d\n", config.TotalRequests)
	fmt.Printf("POST ratio: %.2f\n", config.PostRatio)
	fmt.Printf("Base URL: %s\n", config.BaseURL)
	fmt.Printf("Concurrency: %d\n", config.WorkerNumber)
	fmt.Printf("Timeout: %s\n", config.Timeout)
	fmt.Println("Press Enter to stop the test...")

	transport := &http.Transport{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 100,
		IdleConnTimeout:     90 * time.Second,
		DisableKeepAlives:   false,
		DialContext: (&net.Dialer{
			Timeout:   2 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		TLSHandshakeTimeout:   2 * time.Second,
		ResponseHeaderTimeout: 2 * time.Second,
	}

	client := &http.Client{
		Timeout:   3 * time.Second,
		Transport: transport,
	}

	jobs := make(chan RequestType, config.TotalRequests)
	var activeWorkers sync.WaitGroup

	go func() {
		var input string
		fmt.Scanln(&input)
		close(shutdown)
	}()

	// Start worker pool
	for w := 1; w <= config.WorkerNumber; w++ {
		activeWorkers.Add(1)
		go worker(jobs, stats, config.BaseURL, client, &activeWorkers)
	}

	// Create batches of requests
	jobCount := 0
	batchSize := 10
outer:
	for jobCount < config.TotalRequests {
		// Create a batch of requests
		batch := createBatch(batchSize, config.PostRatio)

		// Send the batch
		for _, requestType := range batch {
			select {
			case <-shutdown:
				break outer
			default:
				stats.wg.Add(1)
				jobs <- RequestType{Type: requestType, JobIndex: jobCount}
				jobCount++

				if jobCount >= config.TotalRequests {
					break
				}
			}
		}
	}
	close(jobs)

	activeWorkers.Wait()
	stats.wg.Wait()

	fmt.Printf("\nTest stopped after %d requests\n", jobCount)

	printResults(stats)
	verifyResults(config.BaseURL, stats, client)
	saveResults(stats)
}

// createBatch creates a slice of request types based on the POST ratio
func createBatch(size int, postRatio float64) []string {
	batch := make([]string, size)
	postsInBatch := int(float64(size) * postRatio)

	// Fill the batch with POST requests first
	for i := 0; i < postsInBatch; i++ {
		batch[i] = "POST"
	}

	// Fill the remaining slots with GET requests
	for i := postsInBatch; i < size; i++ {
		batch[i] = "GET"
	}

	// Shuffle the batch
	rand.Shuffle(len(batch), func(i, j int) {
		batch[i], batch[j] = batch[j], batch[i]
	})

	return batch
}

func worker(jobs <-chan RequestType, stats *RequestStats, baseURL string, client *http.Client, activeWorkers *sync.WaitGroup) {
	defer activeWorkers.Done()

	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for job := range jobs {
		select {
		case <-shutdown:
			stats.wg.Done()
			continue
		default:
			var err error
			if job.Type == "POST" {
				baseURL, err = makePostRequest(baseURL, stats, client)
			} else {
				baseURL, err = makeGetRequest(baseURL, stats, client)
			}
			if err != nil {
				log.Printf("Error making request: %v", err)
			}
			stats.wg.Done()

			// Wait for next tick to maintain 100ms interval
			<-ticker.C
		}
	}
}

func cleanupClient(client *http.Client) {
	if client != nil && client.Transport != nil {
		if transport, ok := client.Transport.(*http.Transport); ok {
			transport.CloseIdleConnections()
		}
	}
}
