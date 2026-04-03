package main

import (
	"log"
	"os"
	"time"

	"github.com/gophergolang/public-charity/internal/agent"
	"github.com/gophergolang/public-charity/internal/bloom"
)

func main() {
	gatewayURL := os.Getenv("GATEWAY_URL")
	if gatewayURL == "" {
		gatewayURL = "http://localhost:8080"
	}

	interval := 5 * time.Minute
	if v := os.Getenv("AGENT_INTERVAL"); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			interval = d
		}
	}

	bloomGrid := bloom.NewGrid()
	if err := bloomGrid.Load(); err != nil {
		log.Printf("bloom load: %v (starting fresh)", err)
	}

	a := agent.New(bloomGrid, gatewayURL)
	a.Run(interval)
}
