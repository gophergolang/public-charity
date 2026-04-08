// Package server wires HTTP handlers for the auth service.
//
// Flow:
//  1. POST /auth/request {email}      -> generate magic token, email link to user
//  2. user clicks link in email, hits DASHBOARD_URL/auth/callback?token=...
//  3. dashboard server-side POSTs /auth/verify {token} -> returns {jwt, email}
//  4. dashboard sets its own HttpOnly cookie on its domain, redirects to /
package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/mail"
	"strings"
	"time"

	"github.com/alexbreadman/public-charity/auth/internal/config"
	"github.com/alexbreadman/public-charity/auth/internal/email"
	"github.com/alexbreadman/public-charity/auth/internal/jwt"
	"github.com/alexbreadman/public-charity/auth/internal/store"
	"github.com/alexbreadman/public-charity/auth/internal/token"
)

const maxBodySize = 1 << 14 // 16KB

type Server struct {
	cfg    *config.Config
	store  *store.Store
	issuer *jwt.Issuer
	mailer *email.Sender
}

func New(cfg *config.Config, s *store.Store, issuer *jwt.Issuer, mailer *email.Sender) *Server {
	return &Server{cfg: cfg, store: s, issuer: issuer, mailer: mailer}
}

func (s *Server) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", s.handleHealth)
	mux.HandleFunc("POST /auth/request", s.handleRequest)
	mux.HandleFunc("POST /auth/verify", s.handleVerify)
	mux.HandleFunc("POST /auth/validate", s.handleValidate)
	return withCORS(s.cfg.DashboardURL, mux)
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
}

func (s *Server) handleRequest(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, maxBodySize)
	var req struct {
		Email string `json:"email"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid json")
		return
	}
	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	addr, err := mail.ParseAddress(req.Email)
	if err != nil {
		writeErr(w, http.StatusBadRequest, "invalid email")
		return
	}
	req.Email = addr.Address

	t, err := token.New(req.Email, time.Duration(s.cfg.TokenTTLMin)*time.Minute)
	if err != nil {
		log.Printf("token generate: %v", err)
		writeErr(w, http.StatusInternalServerError, "could not create token")
		return
	}
	if err := s.store.Put(t); err != nil {
		log.Printf("token store: %v", err)
		writeErr(w, http.StatusInternalServerError, "could not persist token")
		return
	}

	link := fmt.Sprintf("%s/auth/callback?token=%s", strings.TrimRight(s.cfg.DashboardURL, "/"), t.Value)
	if err := s.mailer.SendMagicLink(req.Email, link); err != nil {
		log.Printf("send email: %v", err)
		writeErr(w, http.StatusBadGateway, "could not send email")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "sent"})
}

func (s *Server) handleVerify(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, maxBodySize)
	var req struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid json")
		return
	}
	if req.Token == "" {
		writeErr(w, http.StatusBadRequest, "token required")
		return
	}

	t, err := s.store.Consume(req.Token)
	if err != nil {
		writeErr(w, http.StatusUnauthorized, err.Error())
		return
	}
	jwtStr, err := s.issuer.Issue(t.Email)
	if err != nil {
		log.Printf("issue jwt: %v", err)
		writeErr(w, http.StatusInternalServerError, "could not issue jwt")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{
		"jwt":   jwtStr,
		"email": t.Email,
	})
}

func (s *Server) handleValidate(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, maxBodySize)
	var req struct {
		JWT string `json:"jwt"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid json")
		return
	}
	claims, err := s.issuer.Validate(req.JWT)
	if err != nil {
		writeErr(w, http.StatusUnauthorized, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, claims)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeErr(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

// withCORS allows the dashboard origin to hit /auth/* from the browser.
// The dashboard calls /auth/verify server-side so strictly only /auth/request
// needs CORS, but it's simpler to apply it uniformly.
func withCORS(origin string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Vary", "Origin")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}
