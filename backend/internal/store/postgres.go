package store

import (
	"context"
	"encoding/json"
	"sync/atomic"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresStore struct {
	pool   *pgxpool.Pool
	hits   int64
	misses int64
}

func NewPostgresStore(ctx context.Context, databaseURL string) (*PostgresStore, error) {
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		return nil, err
	}
	return &PostgresStore{pool: pool}, nil
}

func (s *PostgresStore) Close() error {
	s.pool.Close()
	return nil
}

func (s *PostgresStore) GetOSMExtract(bboxKey string) ([]BuildingRecord, error) {
	var data []byte
	err := s.pool.QueryRow(context.Background(),
		"SELECT buildings FROM osm_extracts WHERE bbox_key = $1", bboxKey).Scan(&data)
	if err == pgx.ErrNoRows {
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

func (s *PostgresStore) PutOSMExtract(bboxKey string, buildings []BuildingRecord) error {
	data, err := json.Marshal(buildings)
	if err != nil {
		return err
	}
	_, err = s.pool.Exec(context.Background(),
		`INSERT INTO osm_extracts (bbox_key, buildings, created_at)
		 VALUES ($1, $2, NOW())
		 ON CONFLICT (bbox_key) DO UPDATE SET buildings = $2, created_at = NOW()`, bboxKey, data)
	return err
}

func (s *PostgresStore) GetHorizonProfile(cacheKey string) (*HorizonRecord, error) {
	var data []byte
	err := s.pool.QueryRow(context.Background(),
		"SELECT profile FROM horizon_profiles WHERE cache_key = $1", cacheKey).Scan(&data)
	if err == pgx.ErrNoRows {
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

func (s *PostgresStore) PutHorizonProfile(cacheKey string, profile *HorizonRecord) error {
	data, err := json.Marshal(profile)
	if err != nil {
		return err
	}
	_, err = s.pool.Exec(context.Background(),
		`INSERT INTO horizon_profiles (cache_key, profile, lat, lng, created_at)
		 VALUES ($1, $2, $3, $4, NOW())
		 ON CONFLICT (cache_key) DO UPDATE SET profile = $2, lat = $3, lng = $4, created_at = NOW()`,
		cacheKey, data, profile.Lat, profile.Lng)
	return err
}

func (s *PostgresStore) EvictOlderThan(age time.Duration) (int, error) {
	cutoff := time.Now().Add(-age)
	tag1, err := s.pool.Exec(context.Background(),
		"DELETE FROM osm_extracts WHERE created_at < $1", cutoff)
	if err != nil {
		return 0, err
	}
	n1 := int(tag1.RowsAffected())

	tag2, err := s.pool.Exec(context.Background(),
		"DELETE FROM horizon_profiles WHERE created_at < $1", cutoff)
	if err != nil {
		return 0, err
	}
	n2 := int(tag2.RowsAffected())

	return n1 + n2, nil
}

func (s *PostgresStore) Stats() CacheStats {
	var st CacheStats

	s.pool.QueryRow(context.Background(), "SELECT COUNT(*) FROM osm_extracts").Scan(&st.OSMExtracts)
	s.pool.QueryRow(context.Background(), "SELECT COUNT(*) FROM horizon_profiles").Scan(&st.HorizonProfiles)

	s.pool.QueryRow(context.Background(),
		"SELECT COALESCE(EXTRACT(EPOCH FROM (NOW() - MIN(created_at)))::BIGINT, 0) FROM osm_extracts").Scan(&st.OldestExtractAge)
	s.pool.QueryRow(context.Background(),
		"SELECT COALESCE(EXTRACT(EPOCH FROM (NOW() - MIN(created_at)))::BIGINT, 0) FROM horizon_profiles").Scan(&st.OldestProfileAge)

	st.Hits = atomic.LoadInt64(&s.hits)
	st.Misses = atomic.LoadInt64(&s.misses)
	return st
}
