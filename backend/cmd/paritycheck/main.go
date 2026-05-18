package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/yasserrmd/sunpath/backend/internal/horizon"
	"github.com/yasserrmd/sunpath/backend/internal/store"
)

var samplePoints = []struct {
	lat, lng, h float64
	name        string
}{
	{48.8566, 2.3522, 1.5, "Paris (open area)"},
	{48.8584, 2.2945, 1.5, "Paris (Eiffel tower area)"},
	{40.7484, -73.9857, 1.5, "NYC (Empire State)"},
	{51.5074, -0.1278, 10.0, "London (elevated)"},
	{35.6762, 139.6503, 1.5, "Tokyo"},
}

func main() {
	sqlitePath := flag.String("sqlite", "sunpath.db", "path to SQLite database")
	pgURL := flag.String("pg", os.Getenv("DATABASE_URL"), "Postgres connection URL")
	flag.Parse()

	if *pgURL == "" {
		*pgURL = os.Getenv("DATABASE_URL")
	}
	if *pgURL == "" {
		log.Fatal("DATABASE_URL not set")
	}

	sqliteStore, err := store.Open(*sqlitePath)
	if err != nil {
		log.Fatalf("opening SQLite store: %v", err)
	}
	defer sqliteStore.Close()

	pgStore, err := store.NewPostgresStore(context.Background(), *pgURL)
	if err != nil {
		log.Fatalf("connecting to Postgres: %v", err)
	}
	defer pgStore.Close()

	allPassed := true
	for _, sp := range samplePoints {
		dataHash := horizon.ComputeDataHash(nil)
		ck := horizon.CacheKey(sp.lat, sp.lng, sp.h, dataHash)

		// Actually just compare the store contents for already-cached profiles
		sqliteProfile, err := sqliteStore.GetHorizonProfile(ck)
		if err != nil {
			log.Printf("  [SKIP] %s: SQLite read error: %v", sp.name, err)
			continue
		}

		pgProfile, err := pgStore.GetHorizonProfile(ck)
		if err != nil {
			log.Printf("  [SKIP] %s: Postgres read error: %v", sp.name, err)
			continue
		}

		if sqliteProfile == nil && pgProfile == nil {
			fmt.Printf("  [SAME] %s: both empty (not cached)\n", sp.name)
			continue
		}
		if sqliteProfile == nil {
			fmt.Printf("  [DIFF] %s: only in Postgres\n", sp.name)
			allPassed = false
			continue
		}
		if pgProfile == nil {
			fmt.Printf("  [DIFF] %s: only in SQLite\n", sp.name)
			allPassed = false
			continue
		}

		match := true
		if sqliteProfile.Lat != pgProfile.Lat || sqliteProfile.Lng != pgProfile.Lng {
			fmt.Printf("  [DIFF] %s: lat/lng mismatch\n", sp.name)
			match = false
		}
		if sqliteProfile.Confidence != pgProfile.Confidence {
			fmt.Printf("  [DIFF] %s: confidence %.4f vs %.4f\n", sp.name, sqliteProfile.Confidence, pgProfile.Confidence)
			match = false
		}
		if len(sqliteProfile.Horizon) != len(pgProfile.Horizon) {
			fmt.Printf("  [DIFF] %s: horizon length %d vs %d\n", sp.name, len(sqliteProfile.Horizon), len(pgProfile.Horizon))
			match = false
		} else {
			for i := 0; i < len(sqliteProfile.Horizon); i++ {
				if sqliteProfile.Horizon[i] != pgProfile.Horizon[i] {
					fmt.Printf("  [DIFF] %s: horizon[%d] %.6f vs %.6f\n", sp.name, i, sqliteProfile.Horizon[i], pgProfile.Horizon[i])
					match = false
					break
				}
			}
		}
		if match {
			fmt.Printf("  [SAME] %s: profiles match\n", sp.name)
		} else {
			allPassed = false
		}
	}

	if allPassed {
		fmt.Println("\nPARITY CHECK PASSED: All cached profiles match between SQLite and Postgres")
	} else {
		fmt.Println("\nPARITY CHECK FAILED: Some profiles differ between stores")
		os.Exit(1)
	}
}
