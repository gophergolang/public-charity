package auth

import (
	"crypto/rand"
	"fmt"
	"sync"
	"time"
)

type pendingCode struct {
	Code      string
	Email     string
	Username  string
	ExpiresAt time.Time
}

var (
	pendingMu    sync.Mutex
	pendingCodes = make(map[string]*pendingCode) // keyed by email
)

func GenerateCode(email, username string) (string, error) {
	code := make([]byte, 3)
	if _, err := rand.Read(code); err != nil {
		return "", err
	}
	codeStr := fmt.Sprintf("%06d", int(code[0])*10000+int(code[1])*100+int(code[2])%100)
	if len(codeStr) > 6 {
		codeStr = codeStr[:6]
	}

	pendingMu.Lock()
	pendingCodes[email] = &pendingCode{
		Code:      codeStr,
		Email:     email,
		Username:  username,
		ExpiresAt: time.Now().Add(10 * time.Minute),
	}
	pendingMu.Unlock()

	return codeStr, nil
}

func VerifyCode(email, code string) (string, error) {
	pendingMu.Lock()
	defer pendingMu.Unlock()

	p, ok := pendingCodes[email]
	if !ok {
		return "", fmt.Errorf("no pending code for this email")
	}

	if time.Now().After(p.ExpiresAt) {
		delete(pendingCodes, email)
		return "", fmt.Errorf("code expired")
	}

	if p.Code != code {
		return "", fmt.Errorf("invalid code")
	}

	username := p.Username
	delete(pendingCodes, email)

	token, err := IssueToken(username)
	if err != nil {
		return "", err
	}
	return token, nil
}
