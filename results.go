package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

func printResults(stats *RequestStats) {
	fmt.Printf("\nTest Results:\n")
	fmt.Printf("Total Requests: %d\n", stats.TotalRequests)
	fmt.Printf("Successful POST requests: %d\n", stats.SuccessfulPosts)
	fmt.Printf("Failed POST requests: %d\n", stats.FailedPosts)
	fmt.Printf("Successful GET requests: %d\n", stats.SuccessfulGets)
	fmt.Printf("Failed GET requests: %d\n", stats.FailedGets)

	if stats.TotalRequests > 0 {
		avgResponseTime := stats.TotalResponseTime / float64(stats.TotalRequests)
		fmt.Printf("Average Response Time: %.2f ms\n", avgResponseTime)
	}
}

func verifyResults(baseURL string, stats *RequestStats, client *http.Client) {
	resp, err := client.Get("http://" + baseURL + "/api/v1/messages")
	if err != nil {
		log.Printf("Error getting statistics from server: %v", err)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error reading response body: %v", err)
		return
	}

	var serverStats []MessageResponse

	if err := json.Unmarshal(body, &serverStats); err != nil {
		log.Printf("Error unmarshaling server stats: %v", err)
		return
	}

	fmt.Printf("\nServer Verification:\n")
	fmt.Printf("Server successful POST requests: %d (Client: %d)\n",
		len(serverStats), stats.SuccessfulPosts)
	fmt.Printf("Client successful GET requests: %d \n",
		stats.SuccessfulGets)
}

func saveResults(stats *RequestStats) {
	filename := fmt.Sprintf("http_test_results_%s.json",
		time.Now().UTC().Format("20060102_150405"))

	data, err := json.MarshalIndent(stats, "", "  ")
	if err != nil {
		log.Printf("Error marshaling results: %v", err)
		return
	}

	if err := writeFile(filename, data); err != nil {
		log.Printf("Error saving results: %v", err)
		return
	}

	fmt.Printf("\nResults saved to: %s\n", filename)
}

func writeFile(filename string, data []byte) error {
	return os.WriteFile(filename, data, 0644)
}
