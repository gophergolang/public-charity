// Package email sends magic link emails via Resend's HTTP API.
package email

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const resendEndpoint = "https://api.resend.com/emails"

type Sender struct {
	apiKey string
	from   string
	client *http.Client
}

func NewResend(apiKey, from string) *Sender {
	return &Sender{
		apiKey: apiKey,
		from:   from,
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

type sendRequest struct {
	From    string   `json:"from"`
	To      []string `json:"to"`
	Subject string   `json:"subject"`
	HTML    string   `json:"html"`
	Text    string   `json:"text"`
}

func (s *Sender) SendMagicLink(to, link string) error {
	subject := "Your sign-in link"
	text := fmt.Sprintf("Click to sign in:\n\n%s\n\nThis link expires in 15 minutes.", link)
	html := fmt.Sprintf(
		`<p>Click to sign in:</p><p><a href="%s">Sign in</a></p><p>Or paste this URL: <br><code>%s</code></p><p>This link expires in 15 minutes.</p>`,
		link, link,
	)

	body, err := json.Marshal(sendRequest{
		From: s.from, To: []string{to}, Subject: subject, HTML: html, Text: text,
	})
	if err != nil {
		return err
	}
	req, err := http.NewRequest(http.MethodPost, resendEndpoint, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+s.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("resend request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("resend status %d: %s", resp.StatusCode, string(b))
	}
	return nil
}
