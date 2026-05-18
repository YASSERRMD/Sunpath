package store

import (
	"database/sql"
	"encoding/json"
	"sync/atomic"
	"time"

	_ "modernc.org/sqlite"
)

type SQLiteStore struct {
	db     *sql.DB
	hits   int64
	misses int64
}

func Open(path string) (*SQLiteStore, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}
	s := &SQLiteStore{db: db}
	if err := s.migrate(); err != nil {
		return nil, err
	}
	return s, nil
}

func (s *SQLiteStore) Close() error {
	return s.db.Close()
}

func (s *SQLiteStore) migrate() error {
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

func (s *SQLiteStore) GetOSMExtract(bboxKey string) ([]BuildingRecord, error) {
	var data []byte
	err := s.db.QueryRow("SELECT buildings FROM osm_extracts WHERE bbox_key = ?", bboxKey).Scan(&data)
	if err == sql.ErrNoRows {
		atomic.AddInt64(&s.misses, 1)
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	atomic.AddInt64(&s.hits, 1)
	var records []BuildingRecord
	if err := json.Unmarshal(data, &records); err != nil {
		return nil, err
	}
	return records, nil
}

func (s *SQLiteStore) PutOSMExtract(bboxKey string, buildings []BuildingRecord) error {
	data, err := json.Marshal(buildings)
	if err != nil {
		return err
	}
	_, err = s.db.Exec("INSERT OR REPLACE INTO osm_extracts (bbox_key, buildings, created_at) VALUES (?, ?, CURRENT_TIMESTAMP)", bboxKey, data)
	return err
}

func (s *SQLiteStore) GetHorizonProfile(cacheKey string) (*HorizonRecord, error) {
	var data []byte
	err := s.db.QueryRow("SELECT profile FROM horizon_profiles WHERE cache_key = ?", cacheKey).Scan(&data)
	if err == sql.ErrNoRows {
		atomic.AddInt64(&s.misses, 1)
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	atomic.AddInt64(&s.hits, 1)
	var rec HorizonRecord
	if err := json.Unmarshal(data, &rec); err != nil {
		return nil, err
	}
	return &rec, nil
}

func (s *SQLiteStore) PutHorizonProfile(cacheKey string, profile *HorizonRecord) error {
	data, err := json.Marshal(profile)
	if err != nil {
		return err
	}
	_, err = s.db.Exec("INSERT OR REPLACE INTO horizon_profiles (cache_key, profile, created_at) VALUES (?, ?, CURRENT_TIMESTAMP)", cacheKey, data)
	return err
}

func (s *SQLiteStore) EvictOlderThan(age time.Duration) (int, error) {
	cutoff := time.Now().Add(-age).UTC().Format("2006-01-02 15:04:05")

	res1, err := s.db.Exec("DELETE FROM osm_extracts WHERE created_at < ?", cutoff)
	if err != nil {
		return 0, err
	}
	n1, _ := res1.RowsAffected()

	res2, err := s.db.Exec("DELETE FROM horizon_profiles WHERE created_at < ?", cutoff)
	if err != nil {
		return 0, err
	}
	n2, _ := res2.RowsAffected()

	return int(n1 + n2), nil
}

func (s *SQLiteStore) Stats() CacheStats {
	var st CacheStats

	s.db.QueryRow("SELECT COUNT(*) FROM osm_extracts").Scan(&st.OSMExtracts)
	s.db.QueryRow("SELECT COUNT(*) FROM horizon_profiles").Scan(&st.HorizonProfiles)

	s.db.QueryRow("SELECT COALESCE(ROUND((JULIANDAY('now') - JULIANDAY(MIN(created_at))) * 86400), 0) FROM osm_extracts").Scan(&st.OldestExtractAge)
	s.db.QueryRow("SELECT COALESCE(ROUND((JULIANDAY('now') - JULIANDAY(MIN(created_at))) * 86400), 0) FROM horizon_profiles").Scan(&st.OldestProfileAge)

	st.Hits = atomic.LoadInt64(&s.hits)
	st.Misses = atomic.LoadInt64(&s.misses)
	return st
}
