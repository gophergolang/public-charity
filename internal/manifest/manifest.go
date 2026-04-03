package manifest

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/gophergolang/public-charity/internal/grid"
	"github.com/gophergolang/public-charity/internal/storage"
)

type Offer struct {
	Category    string `json:"category"`
	Description string `json:"description"`
	Available   bool   `json:"available"`
}

type Surplus struct {
	Category    string `json:"category"`
	Description string `json:"description"`
	Expires     string `json:"expires,omitempty"`
}

type TimelineEntry struct {
	Day         string `json:"day"`
	Time        string `json:"time"`
	Description string `json:"description"`
	Category    string `json:"category"`
}

type User struct {
	Version      int             `json:"version"`
	Username     string          `json:"username"`
	CreatedAt    time.Time       `json:"created_at"`
	UpdatedAt    time.Time       `json:"updated_at"`
	Email        string          `json:"email"`
	Bio          string          `json:"bio"`
	Lat          float64         `json:"lat"`
	Lng          float64         `json:"lng"`
	CellID       string          `json:"cell_id"`
	NeedScores   NeedScores      `json:"need_scores"`
	Interests    []string        `json:"interests"`
	Offers       []Offer         `json:"offers,omitempty"`
	Surplus      []Surplus       `json:"surplus,omitempty"`
	Timeline     []TimelineEntry `json:"timeline,omitempty"`
	ContactPrefs []string        `json:"contact_prefs"`
	AccountType  string          `json:"account_type"`
}

type BusinessRule struct {
	ID                string  `json:"id"`
	Category          string  `json:"category"`
	MinScore          float64 `json:"min_score"`
	RadiusCells       int     `json:"radius_cells"`
	MatchType         string  `json:"match_type"`
	InterestOverlapMin int    `json:"interest_overlap_min,omitempty"`
	OfferTemplate     string  `json:"offer_template"`
	MaxSendsPerDay    int     `json:"max_sends_per_day"`
	Active            bool    `json:"active"`
}

type Business struct {
	BizID       string         `json:"biz_id"`
	Name        string         `json:"name"`
	AccountType string         `json:"account_type"`
	FreeTier    bool           `json:"free_tier"`
	CellID      string         `json:"cell_id"`
	Lat         float64        `json:"lat"`
	Lng         float64        `json:"lng"`
	Rules       []BusinessRule `json:"rules"`
}

func userDir(username string) string {
	return filepath.Join("manifests", username)
}

func userCurrent(username string) string {
	return filepath.Join(userDir(username), "current.json")
}

func Create(u *User) error {
	if storage.Exists(userDir(u.Username)) {
		return fmt.Errorf("user %s already exists", u.Username)
	}
	now := time.Now().UTC()
	u.Version = 1
	u.CreatedAt = now
	u.UpdatedAt = now
	u.CellID = grid.CellID(u.Lat, u.Lng)
	if u.AccountType == "" {
		u.AccountType = "personal"
	}
	return storage.WriteJSON(userCurrent(u.Username), u)
}

func Get(username string) (*User, error) {
	var u User
	if err := storage.ReadJSON(userCurrent(username), &u); err != nil {
		return nil, err
	}
	return &u, nil
}

func Update(u *User) error {
	old, err := Get(u.Username)
	if err != nil {
		return fmt.Errorf("read current: %w", err)
	}
	historyFile := filepath.Join(userDir(u.Username), "history",
		fmt.Sprintf("v%d_%s.json", old.Version, old.UpdatedAt.Format("2006-01-02")))
	if err := storage.WriteJSON(historyFile, old); err != nil {
		return fmt.Errorf("write history: %w", err)
	}
	u.Version = old.Version + 1
	u.CreatedAt = old.CreatedAt
	u.UpdatedAt = time.Now().UTC()
	u.CellID = grid.CellID(u.Lat, u.Lng)
	return storage.WriteJSON(userCurrent(u.Username), u)
}

func Delete(username string) error {
	return storage.Delete(userDir(username))
}

func ListUsers() ([]string, error) {
	return storage.ListDir("manifests")
}

func ListUsersInCell(cellID string) ([]*User, error) {
	usernames, err := ListUsers()
	if err != nil {
		return nil, err
	}
	var users []*User
	for _, name := range usernames {
		u, err := Get(name)
		if err != nil {
			continue
		}
		if u.CellID == cellID {
			users = append(users, u)
		}
	}
	return users, nil
}

func bizDir(bizID string) string {
	return filepath.Join("business", bizID)
}

func SaveBusiness(b *Business) error {
	b.CellID = grid.CellID(b.Lat, b.Lng)
	return storage.WriteJSON(filepath.Join(bizDir(b.BizID), "manifest.json"), b)
}

func GetBusiness(bizID string) (*Business, error) {
	var b Business
	if err := storage.ReadJSON(filepath.Join(bizDir(bizID), "manifest.json"), &b); err != nil {
		return nil, err
	}
	return &b, nil
}

func ListBusinesses() ([]string, error) {
	return storage.ListDir("business")
}
