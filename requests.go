package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

func makePostRequest(baseURL string, stats *RequestStats, client *http.Client) (string, error) {
	msg := Message{
		ID:        fmt.Sprintf("msg-%d", time.Now().UnixNano()),
		Content:   fmt.Sprintf("Content generated at %v", time.Now().UTC()),
		Method:    "POST",
		Timestamp: time.Now().UTC(),
	}

	contentOnly := struct {
		Content string `json:"content"`
	}{
		Content: msg.Content,
	}

	jsonData, err := json.Marshal(contentOnly)
	if err != nil {
		recordResult(stats, RequestResult{
			Message: msg,
			Success: false,
		})
		return baseURL, nil
	}

	start := time.Now()
	resp, err := client.Post("http://"+baseURL+"/api/v1/messages", "application/json", bytes.NewBuffer(jsonData))
	duration := time.Since(start).Milliseconds()

	result := RequestResult{
		Message:      msg,
		ResponseTime: float64(duration),
		Success:      false,
	}

	if err != nil {
		recordResult(stats, result)
		return baseURL, nil // Continue with existing baseURL on error
	}

	defer func() {
		if resp != nil && resp.Body != nil {
			resp.Body.Close()
		}
	}()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		result.Success = true
	}

	var hostIP string
	if result.Success {
		body, err := io.ReadAll(resp.Body)
		if err == nil {
			var responseData MessageResponse
			if err := json.Unmarshal(body, &responseData); err == nil {
				hostIP = responseData.HostIP
			}
		}
	}

	recordResult(stats, result)
	if hostIP != "" {
		return hostIP, nil
	}
	return baseURL, nil
}

func makeGetRequest(baseURL string, stats *RequestStats, client *http.Client) (string, error) {
	msg := Message{
		ID:        fmt.Sprintf("msg-%d", time.Now().UnixNano()),
		Method:    "GET",
		Timestamp: time.Now().UTC(),
	}

	start := time.Now()
	resp, err := client.Get("http://" + baseURL + "/api/v1/messages")
	duration := time.Since(start).Milliseconds()

	result := RequestResult{
		Message:      msg,
		ResponseTime: float64(duration),
		Success:      false,
	}

	if err != nil {
		recordResult(stats, result)
		return baseURL, nil // Continue with existing baseURL on error
	}

	// Ensure response body is closed
	defer func() {
		if resp != nil && resp.Body != nil {
			resp.Body.Close()
		}
	}()

	// Check status code
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		result.Success = true
	}

	// Try to decode response, but don't fail if we can't
	var hostIP string
	if result.Success {
		body, err := io.ReadAll(resp.Body)
		if err == nil {
			var responseData []MessageResponse
			if err := json.Unmarshal(body, &responseData); err == nil && len(responseData) > 0 {
				hostIP = responseData[0].HostIP
			}
		}
	}

	recordResult(stats, result)
	if hostIP != "" {
		return hostIP, nil
	}
	return baseURL, nil // Keep existing baseURL if no valid hostIP received
}

func recordResult(stats *RequestStats, result RequestResult) {
	stats.mu.Lock()
	defer stats.mu.Unlock()
	stats.Results = append(stats.Results, result)
	stats.TotalRequests++
	stats.TotalResponseTime += result.ResponseTime
	switch result.Message.Method {
	case "POST":
		if result.Success {
			stats.SuccessfulPosts++
		} else {
			stats.FailedPosts++
		}
	case "GET":
		if result.Success {
			stats.SuccessfulGets++
		} else {
			stats.FailedGets++
		}
	default:
		// For unexpected message types, count as failed
		if result.Message.Method != "" {
			stats.FailedGets++ // Default to counting as failed GET
		}
	}
}
