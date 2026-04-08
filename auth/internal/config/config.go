package config

import (
	"fmt"
	"os"
)

type Config struct {
	Port         string
	JWTSecret    []byte
	ResendAPIKey string
	EmailFrom    string
	DashboardURL string // where magic links land, e.g. https://dashboard.example.com
	DatabasePath string // SQLite file path, e.g. ./data/auth.db
	TokenTTLMin  int    // magic link lifetime, minutes
	JWTTTLHours  int    // issued JWT lifetime, hours
}

func FromEnv() (*Config, error) {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		return nil, fmt.Errorf("JWT_SECRET is required")
	}
	apiKey := os.Getenv("RESEND_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("RESEND_API_KEY is required")
	}
	from := os.Getenv("EMAIL_FROM")
	if from == "" {
		return nil, fmt.Errorf("EMAIL_FROM is required (e.g. 'Public Charity <auth@public.charity>')")
	}
	dash := os.Getenv("DASHBOARD_URL")
	if dash == "" {
		return nil, fmt.Errorf("DASHBOARD_URL is required (e.g. http://localhost:3000)")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	dbPath := os.Getenv("DATABASE_PATH")
	if dbPath == "" {
		dbPath = "./data/auth.db"
	}

	return &Config{
		Port:         port,
		JWTSecret:    []byte(secret),
		ResendAPIKey: apiKey,
		EmailFrom:    from,
		DashboardURL: dash,
		DatabasePath: dbPath,
		TokenTTLMin:  15,
		JWTTTLHours:  24 * 30,
	}, nil
}
