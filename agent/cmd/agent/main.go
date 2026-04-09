package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/alexbreadman/public-charity/agent/internal/client"
	"github.com/alexbreadman/public-charity/agent/internal/config"
	"github.com/alexbreadman/public-charity/agent/internal/history"
	"github.com/alexbreadman/public-charity/agent/internal/llm"
	"github.com/alexbreadman/public-charity/agent/internal/matcher"
)

func main() {
	cfg, err := config.FromEnv()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	hist, err := history.Open(cfg.DataDir)
	if err != nil {
		log.Fatalf("history: %v", err)
	}
	defer hist.Close()

	cl := client.New(cfg.DashboardURL, cfg.DashboardAPIKey)

	if cfg.GeminiAPIKey == "" {
		log.Println("WARNING: GEMINI_API_KEY not set — using template messages (no AI personalization)")
	}
	log.Printf("agent started (interval=%s, max=%d)", cfg.MatchInterval, cfg.MaxMatches)

	// Run once immediately, then on interval.
	tick(cfg, cl, hist)

	ticker := time.NewTicker(cfg.MatchInterval)
	defer ticker.Stop()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)

	for {
		select {
		case <-ticker.C:
			tick(cfg, cl, hist)
		case <-sig:
			log.Println("shutting down")
			return
		}
	}
}

func tick(cfg *config.Config, cl *client.Client, hist *history.Store) {
	start := time.Now()
	log.Println("matching run started")

	users, err := cl.FetchUsers()
	if err != nil {
		log.Printf("fetch users: %v", err)
		return
	}
	log.Printf("fetched %d users", len(users))

	if len(users) < 2 {
		log.Println("not enough users to match")
		return
	}

	matches := matcher.FindMatches(users, cfg.MaxMatches, hist.IsRecent)
	log.Printf("found %d candidate matches", len(matches))

	sent := 0
	for _, m := range matches {
		// Generate and send message to User A about User B.
		msgA := llm.GenerateIntro(cfg.GeminiAPIKey, m, true)
		errA := cl.SendMessage(client.SendMessageReq{
			RecipientEmail: m.UserA.Email,
			SenderType:     "ai_agent",
			Category:       "matching",
			Subject:        fmt.Sprintf("Meet %s — you could help each other", displayName(m.UserB)),
			Body:           msgA,
			RuleID:         "match_v1",
		})
		if errA != nil {
			log.Printf("send to %s: %v", m.UserA.Email, errA)
			continue
		}

		// Generate and send message to User B about User A.
		msgB := llm.GenerateIntro(cfg.GeminiAPIKey, m, false)
		errB := cl.SendMessage(client.SendMessageReq{
			RecipientEmail: m.UserB.Email,
			SenderType:     "ai_agent",
			Category:       "matching",
			Subject:        fmt.Sprintf("Meet %s — you could help each other", displayName(m.UserA)),
			Body:           msgB,
			RuleID:         "match_v1",
		})
		if errB != nil {
			log.Printf("send to %s: %v", m.UserB.Email, errB)
			continue // don't record — retry next run
		}

		hist.Record(m.UserA.ID, m.UserB.ID, "match_v1")
		sent++
		log.Printf("matched: %s <-> %s (score=%.2f)", m.UserA.Email, m.UserB.Email, m.Score)
	}

	hist.Cleanup()
	log.Printf("matching run done: %d matches sent in %s", sent, time.Since(start).Round(time.Millisecond))
}

func displayName(u *client.User) string {
	if u.DisplayName != "" {
		return u.DisplayName
	}
	return u.Email
}
