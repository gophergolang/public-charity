package auth

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"sync"
	"time"
)

type OAuthApp struct {
	ClientID     string   `json:"client_id"`
	ClientSecret string   `json:"client_secret"`
	RedirectURI  string   `json:"redirect_uri"`
	Name         string   `json:"name"`
	Scopes       []string `json:"scopes"`
}

type AuthorizationCode struct {
	Code        string
	ClientID    string
	Username    string
	Scopes      []string
	RedirectURI string
	ExpiresAt   time.Time
}

var (
	appsMu   sync.RWMutex
	apps     = make(map[string]*OAuthApp)     // keyed by client_id
	codesMu  sync.Mutex
	authCodes = make(map[string]*AuthorizationCode)
)

func RegisterApp(app *OAuthApp) error {
	if app.ClientID == "" {
		b := make([]byte, 16)
		rand.Read(b)
		app.ClientID = hex.EncodeToString(b)
	}
	if app.ClientSecret == "" {
		b := make([]byte, 32)
		rand.Read(b)
		app.ClientSecret = hex.EncodeToString(b)
	}
	appsMu.Lock()
	apps[app.ClientID] = app
	appsMu.Unlock()
	return nil
}

func GetApp(clientID string) (*OAuthApp, error) {
	appsMu.RLock()
	defer appsMu.RUnlock()
	app, ok := apps[clientID]
	if !ok {
		return nil, fmt.Errorf("app not found")
	}
	return app, nil
}

func CreateAuthCode(clientID, username, redirectURI string, scopes []string) (string, error) {
	b := make([]byte, 32)
	rand.Read(b)
	code := hex.EncodeToString(b)

	codesMu.Lock()
	authCodes[code] = &AuthorizationCode{
		Code:        code,
		ClientID:    clientID,
		Username:    username,
		Scopes:      scopes,
		RedirectURI: redirectURI,
		ExpiresAt:   time.Now().Add(5 * time.Minute),
	}
	codesMu.Unlock()

	return code, nil
}

func ExchangeCode(code, clientID, clientSecret string) (string, error) {
	codesMu.Lock()
	ac, ok := authCodes[code]
	if ok {
		delete(authCodes, code)
	}
	codesMu.Unlock()

	if !ok {
		return "", fmt.Errorf("invalid code")
	}
	if time.Now().After(ac.ExpiresAt) {
		return "", fmt.Errorf("code expired")
	}
	if ac.ClientID != clientID {
		return "", fmt.Errorf("client mismatch")
	}

	app, err := GetApp(clientID)
	if err != nil {
		return "", err
	}
	if app.ClientSecret != clientSecret {
		return "", fmt.Errorf("invalid client secret")
	}

	return IssueToken(ac.Username)
}
