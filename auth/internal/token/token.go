// Package token generates and validates single-use magic link tokens.
package token

import (
	"crypto/rand"
	"encoding/base64"
	"time"
)

type Token struct {
	Value     string    `json:"value"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
	ExpiresAt time.Time `json:"expires_at"`
}

func New(email string, ttl time.Duration) (*Token, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	return &Token{
		Value:     base64.RawURLEncoding.EncodeToString(b),
		Email:     email,
		CreatedAt: now,
		ExpiresAt: now.Add(ttl),
	}, nil
}

func (t *Token) Expired() bool {
	return time.Now().UTC().After(t.ExpiresAt)
}
