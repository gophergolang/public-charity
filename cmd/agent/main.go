package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gophergolang/public-charity/internal/agent"
	"github.com/gophergolang/public-charity/internal/bloom"
	"github.com/gophergolang/public-charity/internal/messages"
)

func main() {
	interval := 5 * time.Minute
	if v := os.Getenv("AGENT_INTERVAL"); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			interval = d
		}
	}

	flushInterval := 333 * time.Millisecond // tsdb's ingestion interval
	if v := os.Getenv("FLUSH_INTERVAL"); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			flushInterval = d
		}
	}

	bloomGrid := bloom.NewGrid()
	if err := bloomGrid.Load(); err != nil {
		log.Printf("bloom load: %v (starting fresh)", err)
	}

	// Channel-based message bucket with background flush (tsdb batchLogs pattern)
	bucket := messages.NewBucket(flushInterval)

	a := agent.New(bloomGrid, bucket)

	// Graceful shutdown — flush remaining messages
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sig
		log.Println("shutting down, flushing messages...")
		bucket.Stop()
		os.Exit(0)
	}()

	a.Run(interval)
}
