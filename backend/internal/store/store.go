package store

import (
	"database/sql"
	"encoding/json"
	"time"

	_ "modernc.org/sqlite"
)

type BuildingRecord struct {
	OSMID           int64   `json:"osm_id"`
	FootprintJSON   string  `json:"footprint"`
	Height          float64 `json:"height"`
	HeightEstimated bool    `json:"height_estimated"`
}

type HorizonRecord struct {
	Lat        float64   `json:"lat"`
	Lng        float64   `json:"lng"`
	Height     float64   `json:"height"`
	Horizon    []float64 `json:"horizon"`
	Confidence float64   `json:"confidence"`
	BuildCount int       `json:"build_count"`
	EstCount   int       `json:"est_count"`
	DataHash   string    `json:"data_hash"`
	CreatedAt  time.Time `json:"created_at"`
}

type Store struct {
	db *sql.DB
}

func Open(path string) (*Store, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}
	s := &Store{db: db}
	if err := s.migrate(); err != nil {
		return nil, err
	}
	return s, nil
}

func (s *Store) Close() error {
	return s.db.Close()
}

func (s *Store) migrate() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS osm_extracts (
			bbox_key TEXT PRIMARY KEY,
			buildings BLOB NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS horizon_profiles (
			cache_key TEXT PRIMARY KEY,
			profile BLOB NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
	}
	for _, q := range queries {
		if _, err := s.db.Exec(q); err != nil {
			return err
		}
	}
	return nil
}

func (s *Store) GetOSMExtract(bboxKey string) ([]BuildingRecord, error) {
	var data []byte
	err := s.db.QueryRow("SELECT buildings FROM osm_extracts WHERE bbox_key = ?", bboxKey).Scan(&data)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var records []BuildingRecord
	if err := json.Unmarshal(data, &records); err != nil {
		return nil, err
	}
	return records, nil
}

func (s *Store) PutOSMExtract(bboxKey string, buildings []BuildingRecord) error {
	data, err := json.Marshal(buildings)
	if err != nil {
		return err
	}
	_, err = s.db.Exec("INSERT OR REPLACE INTO osm_extracts (bbox_key, buildings, created_at) VALUES (?, ?, CURRENT_TIMESTAMP)", bboxKey, data)
	return err
}

func (s *Store) GetHorizonProfile(cacheKey string) (*HorizonRecord, error) {
	var data []byte
	err := s.db.QueryRow("SELECT profile FROM horizon_profiles WHERE cache_key = ?", cacheKey).Scan(&data)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var rec HorizonRecord
	if err := json.Unmarshal(data, &rec); err != nil {
		return nil, err
	}
	return &rec, nil
}

func (s *Store) PutHorizonProfile(cacheKey string, profile *HorizonRecord) error {
	data, err := json.Marshal(profile)
	if err != nil {
		return err
	}
	_, err = s.db.Exec("INSERT OR REPLACE INTO horizon_profiles (cache_key, profile, created_at) VALUES (?, ?, CURRENT_TIMESTAMP)", cacheKey, data)
	return err
}
