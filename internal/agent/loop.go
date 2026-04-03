package agent

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gophergolang/public-charity/internal/bloom"
	"github.com/gophergolang/public-charity/internal/grid"
	"github.com/gophergolang/public-charity/internal/manifest"
	"github.com/gophergolang/public-charity/internal/messages"
)

type Agent struct {
	bloomGrid  *bloom.Grid
	gatewayURL string

	sendCountMu sync.Mutex
	sendCounts  map[string]int // key: "{biz_id}/{rule_id}/{date}" -> count
}

func New(bloomGrid *bloom.Grid, gatewayURL string) *Agent {
	return &Agent{
		bloomGrid:  bloomGrid,
		gatewayURL: gatewayURL,
		sendCounts: make(map[string]int),
	}
}

func (a *Agent) Run(interval time.Duration) {
	log.Printf("agent service started (interval: %s)", interval)
	for {
		a.tick()
		time.Sleep(interval)
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
	cells := grid.CellsInRadius(biz.CellID, rule.RadiusCells)

	// Bloom filter triage: which cells have relevant users?
	var candidateCells []string
	for _, cellID := range cells {
		if a.bloomGrid.HasUsers(cellID, rule.Category) {
			candidateCells = append(candidateCells, cellID)
		}
	}

	if len(candidateCells) == 0 {
		return
	}

	// Read manifests only from candidate cells
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

		for _, u := range match.Users {
			msg := messages.Message{
				From:     biz.BizID,
				Category: rule.Category,
				Subject:  fmt.Sprintf("Offer from %s", biz.Name),
				Body:     offerText,
				RuleID:   rule.ID,
			}
			a.sendMessage(u.Username, &msg)
		}
		a.incrementSendCount(biz.BizID, rule)
	}
}

func (a *Agent) sendMessage(username string, msg *messages.Message) {
	body, _ := json.Marshal(msg)
	url := fmt.Sprintf("%s/api/messages/%s", a.gatewayURL, username)
	resp, err := http.Post(url, "application/json", bytes.NewReader(body))
	if err != nil {
		log.Printf("send message to %s: %v", username, err)
		return
	}
	resp.Body.Close()
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
