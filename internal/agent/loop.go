// Package agent implements the continuous rule-processing service.
// Adapted from tsdb patterns:
// - Background goroutine loop (tsdb's Ingest/BackupToDisk intervals)
// - Channel-based message submission via Bucket (tsdb's Client.Publish)
// - Bloom filter triage before manifest reads (tsdb's Period bloom checks)
package agent

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/gophergolang/public-charity/internal/bloom"
	"github.com/gophergolang/public-charity/internal/grid"
	"github.com/gophergolang/public-charity/internal/manifest"
	"github.com/gophergolang/public-charity/internal/messages"
)

type Agent struct {
	bloomGrid *bloom.Grid
	bucket    *messages.Bucket

	sendCountMu sync.Mutex
	sendCounts  map[string]int // key: "{biz_id}/{rule_id}/{date}" -> count
}

// New creates an agent with a channel-based message bucket.
// The bucket uses tsdb's batchLogs pattern: non-blocking Submit,
// background flush goroutine on a timer.
func New(bloomGrid *bloom.Grid, bucket *messages.Bucket) *Agent {
	return &Agent{
		bloomGrid:  bloomGrid,
		bucket:     bucket,
		sendCounts: make(map[string]int),
	}
}

// Run starts the agent loop at the given interval.
// Mirrors tsdb's Ingest goroutine pattern (periodic scan + process).
func (a *Agent) Run(interval time.Duration) {
	log.Printf("agent service started (interval: %s)", interval)
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for range ticker.C {
		a.tick()
	}
}

func (a *Agent) tick() {
	bizIDs, err := manifest.ListBusinesses()
	if err != nil {
		log.Printf("list businesses: %v", err)
		return
	}

	for _, bizID := range bizIDs {
		biz, err := manifest.GetBusiness(bizID)
		if err != nil {
			log.Printf("get business %s: %v", bizID, err)
			continue
		}

		for _, rule := range biz.Rules {
			if !rule.Active {
				continue
			}
			a.processRule(biz, rule)
		}
	}
}

func (a *Agent) processRule(biz *manifest.Business, rule manifest.BusinessRule) {
	// BLOOM TRIAGE: check bloom filters for candidate cells.
	// This mirrors tsdb's approach of checking Period bloom filters before
	// accessing actual data — skip entire cells in nanoseconds.
	cells := grid.CellsInRadius(biz.CellID, rule.RadiusCells)

	var candidateCells []string
	for _, cellID := range cells {
		if a.bloomGrid.HasUsers(cellID, rule.Category) {
			candidateCells = append(candidateCells, cellID)
		}
	}

	if len(candidateCells) == 0 {
		return
	}

	// Only read manifests from cells that passed the bloom filter
	var candidates []*manifest.User
	for _, cellID := range candidateCells {
		users, err := manifest.ListUsersInCell(cellID)
		if err != nil {
			continue
		}
		candidates = append(candidates, users...)
	}

	matches := FindMatches(rule, candidates)

	for _, match := range matches {
		if a.rateLimited(biz.BizID, rule) {
			break
		}

		match.BizName = biz.Name
		offerText, err := GenerateOffer(match, biz.Name)
		if err != nil {
			log.Printf("generate offer: %v", err)
			continue
		}

		// Submit to channel-based bucket (non-blocking, tsdb Client.Publish pattern)
		for _, u := range match.Users {
			msg := &messages.Message{
				From:     biz.BizID,
				Category: rule.Category,
				Subject:  fmt.Sprintf("Offer from %s", biz.Name),
				Body:     offerText,
				RuleID:   rule.ID,
			}
			a.bucket.Submit(u.Username, msg)
		}
		a.incrementSendCount(biz.BizID, rule)
	}
}

func (a *Agent) rateLimited(bizID string, rule manifest.BusinessRule) bool {
	a.sendCountMu.Lock()
	defer a.sendCountMu.Unlock()
	key := fmt.Sprintf("%s/%s/%s", bizID, rule.ID, time.Now().Format("2006-01-02"))
	return a.sendCounts[key] >= rule.MaxSendsPerDay
}

func (a *Agent) incrementSendCount(bizID string, rule manifest.BusinessRule) {
	a.sendCountMu.Lock()
	defer a.sendCountMu.Unlock()
	key := fmt.Sprintf("%s/%s/%s", bizID, rule.ID, time.Now().Format("2006-01-02"))
	a.sendCounts[key]++
}
