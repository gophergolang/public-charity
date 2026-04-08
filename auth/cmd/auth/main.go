package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/alexbreadman/public-charity/auth/internal/config"
	"github.com/alexbreadman/public-charity/auth/internal/db"
	"github.com/alexbreadman/public-charity/auth/internal/email"
	"github.com/alexbreadman/public-charity/auth/internal/jwt"
	"github.com/alexbreadman/public-charity/auth/internal/server"
	"github.com/alexbreadman/public-charity/auth/internal/store"
)

func main() {
	cfg, err := config.FromEnv()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	database, err := db.Open(cfg.DatabasePath)
	if err != nil {
		log.Fatalf("open db: %v", err)
	}
	defer database.Close()

	tokStore := store.New(database)
	issuer, err := jwt.NewIssuer(cfg.JWTPrivateKey, time.Duration(cfg.JWTTTLHours)*time.Hour)
	if err != nil {
		log.Fatalf("jwt issuer: %v", err)
	}
	mailer := email.NewResend(cfg.ResendAPIKey, cfg.EmailFrom)
	srv := server.New(cfg, tokStore, issuer, mailer)

	httpSrv := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           srv.Routes(),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	// Periodic expired-token sweep.
	stopSweep := make(chan struct{})
	go func() {
		t := time.NewTicker(5 * time.Minute)
		defer t.Stop()
		for {
			select {
			case <-t.C:
				if n, err := tokStore.Sweep(); err != nil {
					log.Printf("sweep error: %v", err)
				} else if n > 0 {
					log.Printf("swept %d expired tokens", n)
				}
			case <-stopSweep:
				return
			}
		}
	}()

	go func() {
		log.Printf("auth listening on :%s (dashboard=%s, db=%s)", cfg.Port, cfg.DashboardURL, cfg.DatabasePath)
		if err := httpSrv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("listen: %v", err)
		}
	}()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig
	log.Println("shutting down")
	close(stopSweep)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = httpSrv.Shutdown(ctx)
}
