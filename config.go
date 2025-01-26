package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

// Config represents the application configuration
type Config struct {
	TotalRequests int           `json:"totalRequests"`
	PostRatio     float64       `json:"postRatio"`
	BaseURL       string        `json:"baseURL"`
	WorkerNumber  int           `json:"workerNumber"`
	Timeout       time.Duration `json:"timeout"`
}

// Validate checks if the configuration values are within acceptable ranges
func (c *Config) Validate() error {
	if c.TotalRequests <= 0 {
		return fmt.Errorf("totalRequests must be greater than 0")
	}
	if c.PostRatio < 0 || c.PostRatio > 1 {
		return fmt.Errorf("postRatio must be between 0 and 1")
	}
	if c.BaseURL == "" {
		return fmt.Errorf("baseURL cannot be empty")
	}
	if c.WorkerNumber <= 0 {
		return fmt.Errorf("workerNumber must be greater than 0")
	}
	if c.Timeout <= 0 {
		return fmt.Errorf("timeout must be greater than 0")
	}
	return nil
}

// LoadConfig loads configuration from a JSON file
func LoadConfig(path string) (*Config, error) {
	file, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("error reading config file: %w", err)
	}

	var config Config
	if err := json.Unmarshal(file, &config); err != nil {
		return nil, fmt.Errorf("error parsing config file: %w", err)
	}

	// Convert timeout from seconds to Duration
	config.Timeout = time.Duration(config.Timeout) * time.Second

	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return &config, nil
}

// NewDefaultConfig creates a new Config with default values
func NewDefaultConfig() *Config {
	return &Config{
		TotalRequests: 100,
		PostRatio:     0.7,
		BaseURL:       "http://localhost:8080",
		WorkerNumber:  10,
		Timeout:       5 * time.Second,
	}
}
