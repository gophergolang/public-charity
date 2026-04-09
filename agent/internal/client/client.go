// Package client talks to the dashboard's HTTP API.
package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type User struct {
	ID          string             `json:"id"`
	Email       string             `json:"email"`
	DisplayName string             `json:"display_name"`
	CellID      string             `json:"cell_id"`
	Interests   []string           `json:"interests"`
	NeedScores  map[string]float64 `json:"need_scores"`
	Offers      []Offer            `json:"offers"`
	Avail       []string           `json:"availability"` // "tue-morning", etc.
}

type Offer struct {
	Category    string `json:"category"`
	Description string `json:"description"`
}

type SendMessageReq struct {
	RecipientEmail string `json:"recipient_email"`
	SenderType     string `json:"sender_type"`
	Category       string `json:"category,omitempty"`
	Subject        string `json:"subject"`
	Body           string `json:"body"`
	RuleID         string `json:"rule_id,omitempty"`
}

type Client struct {
	baseURL string
	apiKey  string
	http    *http.Client
}

func New(baseURL, apiKey string) *Client {
	return &Client{
		baseURL: baseURL,
		apiKey:  apiKey,
		http:    &http.Client{Timeout: 30 * time.Second},
	}
}

func (c *Client) FetchUsers() ([]User, error) {
	req, err := http.NewRequest("GET", c.baseURL+"/api/users", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch users: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("fetch users: status %d: %s", resp.StatusCode, string(b))
	}

	var users []User
	if err := json.NewDecoder(resp.Body).Decode(&users); err != nil {
		return nil, fmt.Errorf("decode users: %w", err)
	}
	return users, nil
}

func (c *Client) SendMessage(msg SendMessageReq) error {
	body, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	req, err := http.NewRequest("POST", c.baseURL+"/api/messages", bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("send message: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("send message: status %d: %s", resp.StatusCode, string(b))
	}
	return nil
}
