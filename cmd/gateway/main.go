package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/gophergolang/public-charity/internal/bloom"
	"github.com/gophergolang/public-charity/internal/manifest"
	"github.com/gophergolang/public-charity/internal/messages"
	"github.com/gophergolang/public-charity/internal/site"
	"github.com/gophergolang/public-charity/internal/storage"

	"github.com/skip2/go-qrcode"
)

var bloomGrid *bloom.Grid

func main() {
	bloomGrid = bloom.NewGrid()
	if err := bloomGrid.Load(); err != nil {
		log.Printf("bloom load: %v (starting fresh)", err)
	}

	mux := http.NewServeMux()

	// API routes
	mux.HandleFunc("/api/signup", handleSignup)
	mux.HandleFunc("/api/users/", handleUser)
	mux.HandleFunc("/api/messages/", handleMessages)
	mux.HandleFunc("/qr/", handleQR)

	// Wrap with host-based routing
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		host := r.Host
		// Strip port if present
		if idx := strings.LastIndex(host, ":"); idx > 0 {
			host = host[:idx]
		}

		// Wildcard subdomain: {username}.public.charity -> serve static site
		if strings.HasSuffix(host, ".public.charity") {
			username := strings.TrimSuffix(host, ".public.charity")
			serveStaticSite(w, r, username)
			return
		}

		// Otherwise serve API
		mux.ServeHTTP(w, r)
	})

	log.Println("gateway listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", handler))
}

func serveStaticSite(w http.ResponseWriter, r *http.Request, username string) {
	path := storage.BasePath + "/sites/" + username + "/index.html"
	http.ServeFile(w, r, path)
}

func handleSignup(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST only", http.StatusMethodNotAllowed)
		return
	}
	var u manifest.User
	if err := json.NewDecoder(r.Body).Decode(&u); err != nil {
		http.Error(w, "bad json: "+err.Error(), http.StatusBadRequest)
		return
	}
	if u.Username == "" || u.Email == "" {
		http.Error(w, "username and email required", http.StatusBadRequest)
		return
	}
	if err := manifest.Create(&u); err != nil {
		http.Error(w, err.Error(), http.StatusConflict)
		return
	}
	bloomGrid.UpdateUser(&u)
	bloomGrid.Persist()
	site.Generate(&u)

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{
		"status": "created",
		"site":   u.Username + ".public.charity",
	})
}

func handleUser(w http.ResponseWriter, r *http.Request) {
	username := strings.TrimPrefix(r.URL.Path, "/api/users/")
	username = strings.TrimSuffix(username, "/")
	if username == "" {
		http.Error(w, "username required", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case http.MethodGet:
		u, err := manifest.Get(username)
		if err != nil {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		json.NewEncoder(w).Encode(u)

	case http.MethodPut:
		var u manifest.User
		if err := json.NewDecoder(r.Body).Decode(&u); err != nil {
			http.Error(w, "bad json", http.StatusBadRequest)
			return
		}
		u.Username = username

		oldUser, err := manifest.Get(username)
		if err != nil {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}

		if err := manifest.Update(&u); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// If cell changed, rebuild old cell bloom filters
		if oldUser.CellID != u.CellID {
			rebuildCellBloom(oldUser.CellID)
		}
		bloomGrid.UpdateUser(&u)
		bloomGrid.Persist()
		site.Generate(&u)

		json.NewEncoder(w).Encode(map[string]string{"status": "updated"})

	case http.MethodDelete:
		u, err := manifest.Get(username)
		if err != nil {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		cellID := u.CellID
		manifest.Delete(username)
		storage.Delete("sites/" + username)
		storage.Delete("messages/" + username)
		rebuildCellBloom(cellID)
		bloomGrid.Persist()

		json.NewEncoder(w).Encode(map[string]string{"status": "deleted"})

	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func handleMessages(w http.ResponseWriter, r *http.Request) {
	username := strings.TrimPrefix(r.URL.Path, "/api/messages/")
	username = strings.TrimSuffix(username, "/")

	switch r.Method {
	case http.MethodGet:
		msgs, err := messages.Pull(username)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if msgs == nil {
			msgs = []messages.Message{}
		}
		json.NewEncoder(w).Encode(msgs)

	case http.MethodPost:
		var msg messages.Message
		if err := json.NewDecoder(r.Body).Decode(&msg); err != nil {
			http.Error(w, "bad json", http.StatusBadRequest)
			return
		}
		if err := messages.Write(username, &msg); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]string{"status": "queued"})

	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func handleQR(w http.ResponseWriter, r *http.Request) {
	username := strings.TrimPrefix(r.URL.Path, "/qr/")
	url := fmt.Sprintf("https://%s.public.charity", username)
	if username == "" || username == "join" {
		url = "https://public.charity"
	}

	png, err := qrcode.Encode(url, qrcode.Medium, 512)
	if err != nil {
		http.Error(w, "qr generation failed", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "image/png")
	w.Write(png)
}

func rebuildCellBloom(cellID string) {
	users, err := manifest.ListUsersInCell(cellID)
	if err != nil {
		return
	}
	bloomGrid.RebuildCell(cellID, users)
}
