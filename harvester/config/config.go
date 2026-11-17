package config

import (
	"encoding/json"
	"fmt"
	"os"
)

// Config holds the application configuration.
type Config struct {
	AppName           string   `json:"app_name"`
	Api               string   `json:"api"`
	Sources           []string `json:"sources"`
	CaptureIntervalMs int      `json:"capture_interval_milliseconds"`
}

// Load reads a JSON file from the given path and returns a populated Config object.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config JSON: %w", err)
	}

	if len(cfg.Sources) == 0 {
		return nil, fmt.Errorf("No sources specified, service will not be able to run.")
	}

	return &cfg, nil
}
