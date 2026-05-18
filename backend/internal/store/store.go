package store

import "time"

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
	UseDSM     bool      `json:"use_dsm"`
	CreatedAt  time.Time `json:"created_at"`
}

type UserRecord struct {
	ID        int64     `json:"id"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

type SessionRecord struct {
	ID        int64     `json:"id"`
	UserID    int64     `json:"user_id"`
	TokenHash string    `json:"token_hash"`
	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
}

type ProjectRecord struct {
	ID        int64     `json:"id"`
	UserID    int64     `json:"user_id"`
	Name      string    `json:"name"`
	Lat       float64   `json:"lat"`
	Lng       float64   `json:"lng"`
	Height    float64   `json:"height"`
	UseDSM    bool      `json:"use_dsm"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type CacheStats struct {
	OSMExtracts      int   `json:"osm_extracts"`
	HorizonProfiles  int   `json:"horizon_profiles"`
	Hits             int64 `json:"hits"`
	Misses           int64 `json:"misses"`
	OldestExtractAge int64 `json:"oldest_extract_age_seconds"`
	OldestProfileAge int64 `json:"oldest_profile_age_seconds"`
}
