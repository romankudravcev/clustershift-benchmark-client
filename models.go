package main

import (
	"sync"
	"time"
)

type Message struct {
	ID        string    `json:"id"`
	Content   string    `json:"content"`
	Method    string    `json:"method"`
	Timestamp time.Time `json:"timestamp"`
}

type MessageResponse struct {
	ID        int64     `json:"id"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
	HostIP    string    `json:"host_ip"`
}

type RequestResult struct {
	Message      Message `json:"message"`
	Success      bool    `json:"success"`
	ResponseTime float64 `json:"response_time_ms"`
}

type RequestStats struct {
	TotalRequests     int             `json:"total_requests"`
	SuccessfulPosts   int             `json:"successful_posts"`
	FailedPosts       int             `json:"failed_posts"`
	SuccessfulGets    int             `json:"successful_gets"`
	FailedGets        int             `json:"failed_gets"`
	TotalResponseTime float64         `json:"total_response_time_ms"`
	Results           []RequestResult `json:"results"`
	mu                sync.Mutex
	wg                sync.WaitGroup
}
