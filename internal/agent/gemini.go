package agent

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
)

var geminiKey = os.Getenv("GEMINI_API_KEY")

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

func GenerateOffer(match Match, bizName string) (string, error) {
	if geminiKey == "" {
		return fallbackOffer(match, bizName), nil
	}

	var userDescriptions []string
	for _, u := range match.Users {
		userDescriptions = append(userDescriptions, fmt.Sprintf(
			"%s (interests: %s, bio: %s)",
			u.Username, strings.Join(u.Interests, ", "), u.Bio,
		))
	}

	prompt := fmt.Sprintf(`You are a welfare-first neighborhood matching agent for public.charity.
Generate a warm, low-pressure offer message for the following match.

Business: %s
Offer template: %s
Category: %s
Matched users: %s
Shared interests: %s

Write a short, friendly message (3-5 sentences) that:
- Mentions the shared interests naturally
- Includes the business offer
- Makes it easy to say no
- Feels warm, not salesy

Return ONLY the message text, nothing else.`,
		bizName,
		match.Rule.OfferTemplate,
		match.Rule.Category,
		strings.Join(userDescriptions, " and "),
		strings.Join(match.Overlap, ", "),
	)

	body := geminiRequest{
		Contents: []geminiContent{
			{Parts: []geminiPart{{Text: prompt}}},
		},
	}

	jsonBody, _ := json.Marshal(body)
	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/gemini-2.0-flash-lite:generateContent?key=%s", geminiKey)
	resp, err := http.Post(url, "application/json", bytes.NewReader(jsonBody))
	if err != nil {
		return fallbackOffer(match, bizName), nil
	}
	defer resp.Body.Close()

	var result geminiResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil || len(result.Candidates) == 0 {
		return fallbackOffer(match, bizName), nil
	}

	return result.Candidates[0].Content.Parts[0].Text, nil
}

func fallbackOffer(match Match, bizName string) string {
	var names []string
	for _, u := range match.Users {
		names = append(names, u.Username)
	}
	return fmt.Sprintf(
		"Hey %s — %s has an offer for you: %s. Shared interests: %s. Reply YES to accept or ignore to skip.",
		strings.Join(names, " & "),
		bizName,
		match.Rule.OfferTemplate,
		strings.Join(match.Overlap, ", "),
	)
}
