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

const requestInterval = 100 * time.Millisecond

func main() {
	// Use the new rand.New approach instead of deprecated rand.Seed
	source := rand.NewSource(time.Now().UnixNano())
	rng := rand.New(source)

	config, err := LoadConfig("config.json")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Convert duration from seconds to Duration
	config.Duration = time.Duration(config.Duration) * time.Second

	stats := &RequestStats{
		Results: make([]RequestResult, 0),
	}

	fmt.Printf("Starting test at: %s\n", time.Now().UTC().Format(time.RFC3339))
	fmt.Printf("Configuration:\n")
	fmt.Printf("Duration: %v\n", config.Duration)
	fmt.Printf("POST ratio: %.2f\n", config.PostRatio)
	fmt.Printf("Base URL: %s\n", config.BaseURL)
	fmt.Printf("Request interval: %v\n", requestInterval)

	transport := &http.Transport{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 100,
		IdleConnTimeout:     90 * time.Second,
		DisableKeepAlives:   false,
		DialContext: (&net.Dialer{
			Timeout:   10 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		TLSHandshakeTimeout:   10 * time.Second,
		ResponseHeaderTimeout: 10 * time.Second,
	}

	client := &http.Client{
		Timeout:   10 * time.Second,
		Transport: transport,
	}

	// Start the request sender goroutine
	var wg sync.WaitGroup
	wg.Add(1)
	go requestSender(config, stats, client, rng, &wg)

	// Wait for the goroutine to finish
	wg.Wait()

	fmt.Printf("\nTest completed after %v\n", config.Duration)
	fmt.Printf("Total requests sent: %d\n", len(stats.Results))

	printResults(stats)
	verifyResults(config.BaseURL, stats, client)
	saveResults(stats)
}

// requestSender sends requests every 100ms for the specified duration
func requestSender(config *Config, stats *RequestStats, client *http.Client, rng *rand.Rand, wg *sync.WaitGroup) {
	defer wg.Done()

	ticker := time.NewTicker(requestInterval)
	defer ticker.Stop()

	timeout := time.After(config.Duration)
	baseURL := config.BaseURL
	requestCount := 0

	var reqWG sync.WaitGroup

loop:
	for {
		select {
		case <-timeout:
			fmt.Printf("\nBenchmark duration completed. Stopping...\n")
			break loop
		case <-ticker.C:
			var requestType string
			if rng.Float64() < config.PostRatio {
				requestType = "POST"
			} else {
				requestType = "GET"
			}

			reqWG.Add(1)
			go func(reqType string, reqCount int, currentBaseURL string) {
				defer reqWG.Done()
				var err error
				if reqType == "POST" {
					_, err = makePostRequest(currentBaseURL, stats, client)
				} else {
					_, err = makeGetRequest(currentBaseURL, stats, client)
				}
				if err != nil {
					log.Printf("Error making %s request #%d: %v", reqType, reqCount, err)
				}
			}(requestType, requestCount, baseURL)

			requestCount++
		}
	}

	reqWG.Wait()
}

func cleanupClient(client *http.Client) {
	if client != nil && client.Transport != nil {
		if transport, ok := client.Transport.(*http.Transport); ok {
			transport.CloseIdleConnections()
		}
	}
}
