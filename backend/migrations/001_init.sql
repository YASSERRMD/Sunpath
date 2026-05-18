-- +goose Up
CREATE EXTENSION IF NOT EXISTS postgis;

CREATE TABLE IF NOT EXISTS osm_extracts (
    bbox_key TEXT PRIMARY KEY,
    buildings JSONB NOT NULL,
    bbox geometry(POLYGON, 4326),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_osm_extracts_bbox ON osm_extracts USING GIST (bbox);

CREATE TABLE IF NOT EXISTS horizon_profiles (
    cache_key TEXT PRIMARY KEY,
    profile JSONB NOT NULL,
    lat DOUBLE PRECISION,
    lng DOUBLE PRECISION,
    location geometry(POINT, 4326),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_horizon_profiles_location ON horizon_profiles USING GIST (location);

-- +goose Down
DROP TABLE IF EXISTS horizon_profiles;
DROP TABLE IF EXISTS osm_extracts;
DROP EXTENSION IF EXISTS postgis;
