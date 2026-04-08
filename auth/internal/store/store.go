// Package store holds pending magic link tokens in a SQLite database.
// Tokens are single-use: Consume returns the token and removes it.
package store

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/alexbreadman/public-charity/auth/internal/token"
)

type Store struct {
	db *sql.DB
}

func New(db *sql.DB) *Store {
	return &Store{db: db}
}

func (s *Store) Put(t *token.Token) error {
	_, err := s.db.Exec(
		"INSERT INTO magic_tokens (value, email, created_at, expires_at) VALUES (?, ?, ?, ?)",
		t.Value, t.Email, t.CreatedAt.Unix(), t.ExpiresAt.Unix(),
	)
	return err
}

// Consume returns and removes the token if present and not expired.
// The expiry check is in the WHERE clause so expired tokens are never
// deleted — they stay in the DB for Sweep to clean up.
func (s *Store) Consume(value string) (*token.Token, error) {
	now := time.Now().UTC().Unix()
	row := s.db.QueryRow(
		"DELETE FROM magic_tokens WHERE value = ? AND expires_at > ? RETURNING email, created_at, expires_at",
		value, now,
	)
	var email string
	var createdUnix, expiresUnix int64
	if err := row.Scan(&email, &createdUnix, &expiresUnix); err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("token not found or expired")
		}
		return nil, err
	}
	return &token.Token{
		Value:     value,
		Email:     email,
		CreatedAt: time.Unix(createdUnix, 0).UTC(),
		ExpiresAt: time.Unix(expiresUnix, 0).UTC(),
	}, nil
}

// Sweep removes expired tokens. Safe to call periodically.
func (s *Store) Sweep() (int, error) {
	res, err := s.db.Exec("DELETE FROM magic_tokens WHERE expires_at < ?", time.Now().UTC().Unix())
	if err != nil {
		return 0, err
	}
	n, _ := res.RowsAffected()
	return int(n), nil
}
