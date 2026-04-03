package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/gophergolang/public-charity/internal/auth"
)

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("/auth/signup", handleSignup)
	mux.HandleFunc("/auth/verify", handleVerify)
	mux.HandleFunc("/auth/validate", handleValidate)
	mux.HandleFunc("/auth/oauth/authorize", handleOAuthAuthorize)
	mux.HandleFunc("/auth/oauth/token", handleOAuthToken)
	mux.HandleFunc("/auth/oauth/userinfo", handleUserInfo)

	log.Println("auth service listening on :8081")
	log.Fatal(http.ListenAndServe(":8081", mux))
}

func handleSignup(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST only", http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		Email    string `json:"email"`
		Username string `json:"username"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad json", http.StatusBadRequest)
		return
	}

	code, err := auth.GenerateCode(req.Email, req.Username)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// TODO: send email with code. For now, return it directly (dev mode).
	log.Printf("magic link code for %s: %s", req.Email, code)
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "code_sent",
		"dev_code": code,
	})
}

func handleVerify(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST only", http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		Email string `json:"email"`
		Code  string `json:"code"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad json", http.StatusBadRequest)
		return
	}

	token, err := auth.VerifyCode(req.Email, req.Code)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"token": token})
}

func handleValidate(w http.ResponseWriter, r *http.Request) {
	token := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
	claims, err := auth.ValidateToken(token)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	json.NewEncoder(w).Encode(claims)
}

func handleOAuthAuthorize(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST only", http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		ClientID    string   `json:"client_id"`
		RedirectURI string   `json:"redirect_uri"`
		Username    string   `json:"username"`
		Scopes      []string `json:"scopes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad json", http.StatusBadRequest)
		return
	}

	code, err := auth.CreateAuthCode(req.ClientID, req.Username, req.RedirectURI, req.Scopes)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{
		"code":         code,
		"redirect_uri": req.RedirectURI,
	})
}

func handleOAuthToken(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST only", http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		Code         string `json:"code"`
		ClientID     string `json:"client_id"`
		ClientSecret string `json:"client_secret"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad json", http.StatusBadRequest)
		return
	}

	token, err := auth.ExchangeCode(req.Code, req.ClientID, req.ClientSecret)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{
		"access_token": token,
		"token_type":   "Bearer",
	})
}

func handleUserInfo(w http.ResponseWriter, r *http.Request) {
	token := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
	claims, err := auth.ValidateToken(token)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{
		"sub":      claims.Username,
		"username": claims.Username,
	})
}
