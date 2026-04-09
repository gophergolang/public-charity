// Package llm generates warm introduction messages via Gemini Flash Lite.
package llm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	apiclient "github.com/alexbreadman/public-charity/agent/internal/client"
	"github.com/alexbreadman/public-charity/agent/internal/matcher"
)

var httpClient = &http.Client{Timeout: 30 * time.Second}

type geminiRequest struct {
	Contents []geminiContent `json:"contents"`
}
type geminiContent struct {
	Parts []geminiPart `json:"parts"`
}
type geminiPart struct {
	Text string `json:"text"`
}
type geminiResponse struct {
	Candidates []struct {
		Content struct {
			Parts []geminiPart `json:"parts"`
		} `json:"content"`
	} `json:"candidates"`
}

// GenerateIntro creates a warm message introducing User B to User A.
// If apiKey is empty or Gemini fails, falls back to a template.
func GenerateIntro(apiKey string, m matcher.Match, forA bool) string {
	recipient, other := m.UserA, m.UserB
	if !forA {
		recipient, other = m.UserB, m.UserA
	}

	recipientName := name(recipient)
	otherName := name(other)

	// Collect what's relevant for this recipient.
	var needDescriptions []string
	for _, p := range m.NeedOfferPairs {
		if p.Needer == recipientName {
			needDescriptions = append(needDescriptions, fmt.Sprintf("%s (%s)", p.Category, p.Offer))
		}
	}
	var offerDescriptions []string
	for _, p := range m.NeedOfferPairs {
		if p.Needer == name(other) {
			offerDescriptions = append(offerDescriptions, fmt.Sprintf("%s (%s)", p.Category, p.Offer))
		}
	}

	if apiKey == "" {
		return fallback(recipientName, otherName, m, needDescriptions, offerDescriptions)
	}

	prompt := fmt.Sprintf(`You are a friendly neighbourhood connector for public.charity.
Write a warm, brief message (3-4 sentences) to %s introducing %s.

%s could help %s with: %s
%s could help %s with: %s
They're both free: %s
Shared interests: %s

Be warm, specific, and low-pressure. Make it easy to say no.
Don't use exclamation marks excessively. Sound human, not corporate.
Return ONLY the message text, nothing else.`,
		recipientName, otherName,
		otherName, recipientName, joinOrNone(needDescriptions),
		recipientName, otherName, joinOrNone(offerDescriptions),
		joinOrNone(m.SharedSlots),
		joinOrNone(m.SharedInterests),
	)

	body, _ := json.Marshal(geminiRequest{
		Contents: []geminiContent{{Parts: []geminiPart{{Text: prompt}}}},
	})
	apiURL := fmt.Sprintf(
		"https://generativelanguage.googleapis.com/v1beta/models/gemini-2.0-flash-lite:generateContent?key=%s",
		apiKey,
	)
	resp, err := httpClient.Post(apiURL, "application/json", bytes.NewReader(body))
	if err != nil {
		return fallback(recipientName, otherName, m, needDescriptions, offerDescriptions)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return fallback(recipientName, otherName, m, needDescriptions, offerDescriptions)
	}

	var result geminiResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil ||
		len(result.Candidates) == 0 ||
		len(result.Candidates[0].Content.Parts) == 0 {
		return fallback(recipientName, otherName, m, needDescriptions, offerDescriptions)
	}
	text := result.Candidates[0].Content.Parts[0].Text
	if strings.TrimSpace(text) == "" {
		return fallback(recipientName, otherName, m, needDescriptions, offerDescriptions)
	}
	// Limit response size.
	if len(text) > 1000 {
		text = text[:1000]
	}
	return text
}

func fallback(recipientName, otherName string, m matcher.Match, needs, offers []string) string {
	var parts []string
	parts = append(parts, fmt.Sprintf("Hey %s — we think you and %s might be able to help each other out.", recipientName, otherName))
	if len(needs) > 0 {
		parts = append(parts, fmt.Sprintf("%s can help you with %s.", otherName, strings.Join(needs, ", ")))
	}
	if len(offers) > 0 {
		parts = append(parts, fmt.Sprintf("And you could help them with %s.", strings.Join(offers, ", ")))
	}
	if len(m.SharedSlots) > 0 {
		parts = append(parts, fmt.Sprintf("You're both free %s.", strings.Join(m.SharedSlots, ", ")))
	}
	if len(m.SharedInterests) > 0 {
		parts = append(parts, fmt.Sprintf("You also both enjoy %s.", strings.Join(m.SharedInterests, ", ")))
	}
	parts = append(parts, "No pressure at all — just reply if you'd like an introduction.")
	return strings.Join(parts, " ")
}

func name(u *apiclient.User) string {
	if u.DisplayName != "" {
		return u.DisplayName
	}
	return u.Email
}

func joinOrNone(s []string) string {
	if len(s) == 0 {
		return "none"
	}
	return strings.Join(s, ", ")
}

// ReadBody is a utility for reading response bodies with a size limit.
func ReadBody(r io.Reader, limit int64) ([]byte, error) {
	return io.ReadAll(io.LimitReader(r, limit))
}
