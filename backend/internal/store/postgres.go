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
	if err := ensureTables(ctx, pool); err != nil {
		pool.Close()
		return nil, err
	}
	return &PostgresStore{pool: pool}, nil
}

func ensureTables(ctx context.Context, pool *pgxpool.Pool) error {
	sql := []string{
		`CREATE TABLE IF NOT EXISTS osm_extracts (
			bbox_key TEXT PRIMARY KEY,
			buildings BYTEA NOT NULL,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS horizon_profiles (
			cache_key TEXT PRIMARY KEY,
			profile BYTEA NOT NULL,
			lat DOUBLE PRECISION NOT NULL DEFAULT 0,
			lng DOUBLE PRECISION NOT NULL DEFAULT 0,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS users (
			id BIGSERIAL PRIMARY KEY,
			email TEXT NOT NULL UNIQUE,
			name TEXT NOT NULL DEFAULT '',
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)`,
		`CREATE INDEX IF NOT EXISTS idx_users_email ON users (email)`,
		`CREATE TABLE IF NOT EXISTS auth_sessions (
			id BIGSERIAL PRIMARY KEY,
			user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			token_hash TEXT NOT NULL UNIQUE,
			expires_at TIMESTAMPTZ NOT NULL,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)`,
		`CREATE INDEX IF NOT EXISTS idx_auth_sessions_token_hash ON auth_sessions (token_hash)`,
		`CREATE TABLE IF NOT EXISTS magic_links (
			id BIGSERIAL PRIMARY KEY,
			email TEXT NOT NULL,
			code TEXT NOT NULL UNIQUE,
			used BOOLEAN NOT NULL DEFAULT FALSE,
			expires_at TIMESTAMPTZ NOT NULL,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)`,
		`CREATE INDEX IF NOT EXISTS idx_magic_links_code ON magic_links (code)`,
		`CREATE TABLE IF NOT EXISTS saved_projects (
			id BIGSERIAL PRIMARY KEY,
			user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			name TEXT NOT NULL,
			lat DOUBLE PRECISION NOT NULL,
			lng DOUBLE PRECISION NOT NULL,
			height DOUBLE PRECISION NOT NULL DEFAULT 1.5,
			use_dsm BOOLEAN NOT NULL DEFAULT FALSE,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)`,
		`CREATE INDEX IF NOT EXISTS idx_saved_projects_user_id ON saved_projects (user_id)`,
	}
	for _, q := range sql {
		if _, err := pool.Exec(ctx, q); err != nil {
			return err
		}
	}
	return nil
}

func (s *PostgresStore) Pool() *pgxpool.Pool {
	return s.pool
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

func (s *PostgresStore) CreateUser(ctx context.Context, email, name string) (*UserRecord, error) {
	var u UserRecord
	err := s.pool.QueryRow(ctx,
		`INSERT INTO users (email, name) VALUES ($1, $2)
		 ON CONFLICT (email) DO UPDATE SET name = $2
		 RETURNING id, email, name, created_at`,
		email, name).Scan(&u.ID, &u.Email, &u.Name, &u.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (s *PostgresStore) GetUserByEmail(ctx context.Context, email string) (*UserRecord, error) {
	var u UserRecord
	err := s.pool.QueryRow(ctx,
		"SELECT id, email, name, created_at FROM users WHERE email = $1", email).
		Scan(&u.ID, &u.Email, &u.Name, &u.CreatedAt)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (s *PostgresStore) GetUserByID(ctx context.Context, id int64) (*UserRecord, error) {
	var u UserRecord
	err := s.pool.QueryRow(ctx,
		"SELECT id, email, name, created_at FROM users WHERE id = $1", id).
		Scan(&u.ID, &u.Email, &u.Name, &u.CreatedAt)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (s *PostgresStore) CreateSession(ctx context.Context, userID int64, tokenHash string, expiresAt time.Time) (*SessionRecord, error) {
	var sess SessionRecord
	err := s.pool.QueryRow(ctx,
		`INSERT INTO auth_sessions (user_id, token_hash, expires_at)
		 VALUES ($1, $2, $3)
		 RETURNING id, user_id, token_hash, expires_at, created_at`,
		userID, tokenHash, expiresAt).
		Scan(&sess.ID, &sess.UserID, &sess.TokenHash, &sess.ExpiresAt, &sess.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &sess, nil
}

func (s *PostgresStore) GetSessionByTokenHash(ctx context.Context, tokenHash string) (*SessionRecord, error) {
	var sess SessionRecord
	err := s.pool.QueryRow(ctx,
		"SELECT id, user_id, token_hash, expires_at, created_at FROM auth_sessions WHERE token_hash = $1", tokenHash).
		Scan(&sess.ID, &sess.UserID, &sess.TokenHash, &sess.ExpiresAt, &sess.CreatedAt)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &sess, nil
}

func (s *PostgresStore) DeleteExpiredSessions(ctx context.Context) error {
	_, err := s.pool.Exec(ctx, "DELETE FROM auth_sessions WHERE expires_at < NOW()")
	return err
}

func (s *PostgresStore) CreateMagicLink(ctx context.Context, email, code string, expiresAt time.Time) error {
	_, err := s.pool.Exec(ctx,
		`INSERT INTO magic_links (email, code, expires_at) VALUES ($1, $2, $3)
		 ON CONFLICT (code) DO NOTHING`,
		email, code, expiresAt)
	return err
}

func (s *PostgresStore) ConsumeMagicLink(ctx context.Context, code string) (*string, error) {
	var email string
	err := s.pool.QueryRow(ctx,
		`UPDATE magic_links SET used = TRUE
		 WHERE code = $1 AND used = FALSE AND expires_at > NOW()
		 RETURNING email`,
		code).Scan(&email)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &email, nil
}

func (s *PostgresStore) ListProjects(ctx context.Context, userID int64) ([]ProjectRecord, error) {
	rows, err := s.pool.Query(ctx,
		`SELECT id, user_id, name, lat, lng, height, use_dsm, created_at, updated_at
		 FROM saved_projects WHERE user_id = $1 ORDER BY updated_at DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var projects []ProjectRecord
	for rows.Next() {
		var p ProjectRecord
		if err := rows.Scan(&p.ID, &p.UserID, &p.Name, &p.Lat, &p.Lng, &p.Height, &p.UseDSM, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, err
		}
		projects = append(projects, p)
	}
	return projects, nil
}

func (s *PostgresStore) GetProject(ctx context.Context, projectID, userID int64) (*ProjectRecord, error) {
	var p ProjectRecord
	err := s.pool.QueryRow(ctx,
		`SELECT id, user_id, name, lat, lng, height, use_dsm, created_at, updated_at
		 FROM saved_projects WHERE id = $1 AND user_id = $2`,
		projectID, userID).Scan(&p.ID, &p.UserID, &p.Name, &p.Lat, &p.Lng, &p.Height, &p.UseDSM, &p.CreatedAt, &p.UpdatedAt)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (s *PostgresStore) CreateProject(ctx context.Context, userID int64, name string, lat, lng, height float64, useDSM bool) (*ProjectRecord, error) {
	var p ProjectRecord
	err := s.pool.QueryRow(ctx,
		`INSERT INTO saved_projects (user_id, name, lat, lng, height, use_dsm)
		 VALUES ($1, $2, $3, $4, $5, $6)
		 RETURNING id, user_id, name, lat, lng, height, use_dsm, created_at, updated_at`,
		userID, name, lat, lng, height, useDSM).
		Scan(&p.ID, &p.UserID, &p.Name, &p.Lat, &p.Lng, &p.Height, &p.UseDSM, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (s *PostgresStore) UpdateProject(ctx context.Context, projectID, userID int64, name string, lat, lng, height float64, useDSM bool) (*ProjectRecord, error) {
	var p ProjectRecord
	err := s.pool.QueryRow(ctx,
		`UPDATE saved_projects
		 SET name = $1, lat = $2, lng = $3, height = $4, use_dsm = $5, updated_at = NOW()
		 WHERE id = $6 AND user_id = $7
		 RETURNING id, user_id, name, lat, lng, height, use_dsm, created_at, updated_at`,
		name, lat, lng, height, useDSM, projectID, userID).
		Scan(&p.ID, &p.UserID, &p.Name, &p.Lat, &p.Lng, &p.Height, &p.UseDSM, &p.CreatedAt, &p.UpdatedAt)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (s *PostgresStore) DeleteProject(ctx context.Context, projectID, userID int64) error {
	tag, err := s.pool.Exec(ctx,
		"DELETE FROM saved_projects WHERE id = $1 AND user_id = $2", projectID, userID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
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
