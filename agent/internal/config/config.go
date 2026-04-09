package config

import (
	"fmt"
	"os"
	"time"
)

type Config struct {
	DashboardURL    string
	DashboardAPIKey string
	GeminiAPIKey    string
	MatchInterval   time.Duration
	MaxMatches      int
	DataDir         string
}

func FromEnv() (*Config, error) {
	url := os.Getenv("DASHBOARD_URL")
	if url == "" {
		return nil, fmt.Errorf("DASHBOARD_URL is required")
	}
	key := os.Getenv("DASHBOARD_API_KEY")
	if key == "" {
		return nil, fmt.Errorf("DASHBOARD_API_KEY is required")
	}

	interval := 4 * time.Hour
	if v := os.Getenv("MATCH_INTERVAL"); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			interval = d
		}
	}

	maxMatches := 20
	dataDir := os.Getenv("DATA_DIR")
	if dataDir == "" {
		dataDir = "./data"
	}

	return &Config{
		DashboardURL:    url,
		DashboardAPIKey: key,
		GeminiAPIKey:    os.Getenv("GEMINI_API_KEY"),
		MatchInterval:   interval,
		MaxMatches:      maxMatches,
		DataDir:         dataDir,
	}, nil
}
