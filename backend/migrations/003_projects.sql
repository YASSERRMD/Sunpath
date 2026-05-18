-- +goose Up
CREATE TABLE IF NOT EXISTS saved_projects (
    id          BIGSERIAL PRIMARY KEY,
    user_id     BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name        TEXT NOT NULL,
    lat         DOUBLE PRECISION NOT NULL,
    lng         DOUBLE PRECISION NOT NULL,
    height      DOUBLE PRECISION NOT NULL DEFAULT 1.5,
    use_dsm     BOOLEAN NOT NULL DEFAULT FALSE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_saved_projects_user_id ON saved_projects (user_id);

-- +goose Down
DROP TABLE IF EXISTS saved_projects;
