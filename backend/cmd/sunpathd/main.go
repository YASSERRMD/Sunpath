package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pressly/goose/v3"
	"github.com/yasserrmd/sunpath/backend/internal/api"
	"github.com/yasserrmd/sunpath/backend/internal/store"
)

func main() {
	listenAddr := getEnv("LISTEN_ADDR", ":8080")
	dbPath := getEnv("DB_PATH", "sunpath.db")
	overpassURL := getEnv("OVERPASS_URL", "https://overpass-api.de/api/interpreter")

	dbStore := getEnv("DB_STORE", "sqlite")
	runMigrations := getEnv("DB_RUN_MIGRATIONS", "false")

	if dbStore == "postgres" || runMigrations == "true" {
		databaseURL := getEnv("DATABASE_URL", "postgres://sunpath:sunpath@localhost:5432/sunpath?sslmode=disable")
		pool, err := pgxpool.New(context.Background(), databaseURL)
		if err != nil {
			log.Fatalf("connecting to postgres: %v", err)
		}
		defer pool.Close()
		if err := goose.RunContext(context.Background(), "up", pool, "migrations"); err != nil {
			log.Fatalf("running migrations: %v", err)
		}
		log.Println("migrations complete")
	}

	st, err := store.Open(dbPath)
	if err != nil {
		log.Fatalf("opening store: %v", err)
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
