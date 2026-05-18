package main

import (
	"context"
	"encoding/json"
	"flag"
	"log"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/yasserrmd/sunpath/backend/internal/store"
)

func main() {
	sqlitePath := flag.String("sqlite", "sunpath.db", "path to SQLite database")
	pgURL := flag.String("pg", os.Getenv("DATABASE_URL"), "Postgres connection URL")
	flag.Parse()

	if *pgURL == "" {
		*pgURL = os.Getenv("DATABASE_URL")
	}
	if *pgURL == "" {
		log.Fatal("DATABASE_URL not set; provide --pg flag or set DATABASE_URL env var")
	}

	sqliteStore, err := store.Open(*sqlitePath)
	if err != nil {
		log.Fatalf("opening SQLite store: %v", err)
	}
	defer sqliteStore.Close()

	pgPool, err := pgxpool.New(context.Background(), *pgURL)
	if err != nil {
		log.Fatalf("connecting to Postgres: %v", err)
	}
	defer pgPool.Close()

	stats := sqliteStore.Stats()
	log.Printf("Found %d OSM extracts and %d horizon profiles in SQLite", stats.OSMExtracts, stats.HorizonProfiles)

	// migration for osm_extracts: can't iterate with the interface, so we use a custom approach
	// We'll iterate by reading the raw JSON from the DB
	type osmRow struct {
		key      string
		buildings string
	}
	osmRows, err := sqliteStore.GetAllOSMExtracts()
	if err != nil {
		log.Fatalf("reading OSM extracts: %v", err)
	}

	migrated := 0
	for _, row := range osmRows {
		var buildings []store.BuildingRecord
		if err := json.Unmarshal([]byte(row.BuildingsJSON), &buildings); err != nil {
			log.Printf("skipping OSM extract %s: parse error: %v", row.Key, err)
			continue
		}
		data, _ := json.Marshal(buildings)
		_, err := pgPool.Exec(context.Background(),
			`INSERT INTO osm_extracts (bbox_key, buildings, created_at) VALUES ($1, $2, $3)
			 ON CONFLICT (bbox_key) DO UPDATE SET buildings = $2, created_at = $3`,
			row.Key, data, row.CreatedAt)
		if err != nil {
			log.Printf("error writing OSM extract %s: %v", row.Key, err)
			continue
		}
		migrated++
	}
	log.Printf("Migrated %d/%d OSM extracts to Postgres", migrated, len(osmRows))

	type horizonRow struct {
		key     string
		profile string
	}
	horizonRows, err := sqliteStore.GetAllHorizonProfiles()
	if err != nil {
		log.Fatalf("reading horizon profiles: %v", err)
	}

	migratedH := 0
	for _, row := range horizonRows {
		var profile store.HorizonRecord
		if err := json.Unmarshal([]byte(row.ProfileJSON), &profile); err != nil {
			log.Printf("skipping horizon profile %s: parse error: %v", row.Key, err)
			continue
		}
		data, _ := json.Marshal(profile)
		_, err := pgPool.Exec(context.Background(),
			`INSERT INTO horizon_profiles (cache_key, profile, lat, lng, created_at)
			 VALUES ($1, $2, $3, $4, $5)
			 ON CONFLICT (cache_key) DO UPDATE SET profile = $2, lat = $3, lng = $4, created_at = $5`,
			row.Key, data, profile.Lat, profile.Lng, row.CreatedAt)
		if err != nil {
			log.Printf("error writing horizon profile %s: %v", row.Key, err)
			continue
		}
		migratedH++
	}
	log.Printf("Migrated %d/%d horizon profiles to Postgres", migratedH, len(horizonRows))
	log.Println("Migration complete")
}
