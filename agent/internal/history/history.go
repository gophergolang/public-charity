// Package history tracks sent matches to avoid spamming the same pair.
package history

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

const schema = `
CREATE TABLE IF NOT EXISTS match_history (
	user_a     TEXT NOT NULL,
	user_b     TEXT NOT NULL,
	matched_at TEXT NOT NULL DEFAULT (datetime('now')),
	rule       TEXT
);
CREATE INDEX IF NOT EXISTS idx_match_recent ON match_history(matched_at);
`

type Store struct {
	db *sql.DB
}

func Open(dir string) (*Store, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}
	db, err := sql.Open("sqlite", filepath.Join(dir, "agent.db"))
	if err != nil {
		return nil, err
	}
	if _, err := db.Exec(schema); err != nil {
		return nil, fmt.Errorf("migrate: %w", err)
	}
	return &Store{db: db}, nil
}

func (s *Store) Close() error { return s.db.Close() }

// IsRecent returns true if this pair was matched in the last 7 days.
func (s *Store) IsRecent(a, b string) bool {
	if a > b {
		a, b = b, a
	}
	var count int
	err := s.db.QueryRow(
		`SELECT COUNT(*) FROM match_history
		 WHERE user_a = ? AND user_b = ? AND matched_at > datetime('now', '-7 days')`,
		a, b,
	).Scan(&count)
	return err == nil && count > 0
}

// Record stores a match.
func (s *Store) Record(a, b, rule string) {
	if a > b {
		a, b = b, a
	}
	s.db.Exec(
		`INSERT INTO match_history (user_a, user_b, rule) VALUES (?, ?, ?)`,
		a, b, rule,
	)
}

// Cleanup removes entries older than 30 days.
func (s *Store) Cleanup() {
	s.db.Exec(`DELETE FROM match_history WHERE matched_at < datetime('now', '-30 days')`)
}
