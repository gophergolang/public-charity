// Package jwt implements minimal HS256 JWT issuance and validation.
package jwt

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

type Claims struct {
	Email    string `json:"email"`
	Subject  string `json:"sub"`
	IssuedAt int64  `json:"iat"`
	Expires  int64  `json:"exp"`
}

type Issuer struct {
	secret []byte
	ttl    time.Duration
}

func NewIssuer(secret []byte, ttl time.Duration) *Issuer {
	return &Issuer{secret: secret, ttl: ttl}
}

func (i *Issuer) Issue(email string) (string, error) {
	now := time.Now().UTC()
	claims := Claims{
		Email:    email,
		Subject:  email,
		IssuedAt: now.Unix(),
		Expires:  now.Add(i.ttl).Unix(),
	}

	header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"HS256","typ":"JWT"}`))
	payload, err := json.Marshal(claims)
	if err != nil {
		return "", err
	}
	payloadEnc := base64.RawURLEncoding.EncodeToString(payload)
	signingInput := header + "." + payloadEnc
	sig := i.sign(signingInput)
	return signingInput + "." + sig, nil
}

func (i *Issuer) Validate(token string) (*Claims, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid token format")
	}

	// Reject any alg other than HS256 to prevent alg:none attacks.
	headerBytes, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return nil, fmt.Errorf("decode header: %w", err)
	}
	var header struct {
		Alg string `json:"alg"`
	}
	if err := json.Unmarshal(headerBytes, &header); err != nil {
		return nil, fmt.Errorf("parse header: %w", err)
	}
	if header.Alg != "HS256" {
		return nil, fmt.Errorf("unsupported alg: %s", header.Alg)
	}

	expected := i.sign(parts[0] + "." + parts[1])
	if !hmac.Equal([]byte(parts[2]), []byte(expected)) {
		return nil, fmt.Errorf("invalid signature")
	}

	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, fmt.Errorf("decode payload: %w", err)
	}
	var claims Claims
	if err := json.Unmarshal(payload, &claims); err != nil {
		return nil, fmt.Errorf("parse payload: %w", err)
	}
	if time.Now().Unix() > claims.Expires {
		return nil, fmt.Errorf("token expired")
	}
	return &claims, nil
}

func (i *Issuer) sign(input string) string {
	mac := hmac.New(sha256.New, i.secret)
	mac.Write([]byte(input))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}
