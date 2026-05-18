package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/yasserrmd/sunpath/backend/internal/api"
	"github.com/yasserrmd/sunpath/backend/internal/store"
)

func main() {
	listenAddr := getEnv("LISTEN_ADDR", ":8080")
	dbStore := getEnv("DB_STORE", "sqlite")
	dbPath := getEnv("DB_PATH", "sunpath.db")
	databaseURL := getEnv("DATABASE_URL", "postgres://sunpath:sunpath@localhost:5432/sunpath?sslmode=disable")
	overpassURL := getEnv("OVERPASS_URL", "https://overpass-api.de/api/interpreter")

	var st store.Storage
	if dbStore == "postgres" {
		var err error
		st, err = store.NewPostgresStore(context.Background(), databaseURL)
		if err != nil {
			log.Fatalf("opening postgres store: %v", err)
		}
	} else {
		var err error
		st, err = store.Open(dbPath)
		if err != nil {
			log.Fatalf("opening store: %v", err)
		}
	}
	defer st.Close()

	srv := api.NewServer(st, overpassURL)
	httpServer := &http.Server{
		Addr:    listenAddr,
		Handler: srv.Routes(),
	}

	go func() {
		log.Printf("listening on %s", listenAddr)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(ctx); err != nil {
		log.Fatalf("forced shutdown: %v", err)
	}
	log.Println("stopped")
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
