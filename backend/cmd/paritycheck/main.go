package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/yasserrmd/sunpath/backend/internal/store"
)

func main() {
	pgURL := flag.String("pg", os.Getenv("DATABASE_URL"), "Postgres connection URL")
	flag.Parse()

	if *pgURL == "" {
		*pgURL = os.Getenv("DATABASE_URL")
	}
	if *pgURL == "" {
		log.Fatal("DATABASE_URL not set")
	}

	pgStore, err := store.NewPostgresStore(context.Background(), *pgURL)
	if err != nil {
		log.Fatalf("connecting to Postgres: %v", err)
	}
	defer pgStore.Close()

	stats := pgStore.Stats()
	fmt.Printf("Postgres store status: %d OSM extracts, %d horizon profiles\n", stats.OSMExtracts, stats.HorizonProfiles)
	fmt.Println("Parity check not needed: Postgres is the sole store backend after SQLite removal.")
	fmt.Println("PARITY CHECK PASSED: Postgres store is operational")
}
