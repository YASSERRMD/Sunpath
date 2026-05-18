-- +goose Up
CREATE TABLE IF NOT EXISTS users (
    id          BIGSERIAL PRIMARY KEY,
    email       TEXT NOT NULL UNIQUE,
    name        TEXT NOT NULL DEFAULT '',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_users_email ON users (email);

CREATE TABLE IF NOT EXISTS auth_sessions (
    id          BIGSERIAL PRIMARY KEY,
    user_id     BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash  TEXT NOT NULL UNIQUE,
    expires_at  TIMESTAMPTZ NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_auth_sessions_token_hash ON auth_sessions (token_hash);
CREATE INDEX idx_auth_sessions_user_id ON auth_sessions (user_id);

CREATE TABLE IF NOT EXISTS magic_links (
    id          BIGSERIAL PRIMARY KEY,
    email       TEXT NOT NULL,
    code        TEXT NOT NULL UNIQUE,
    used        BOOLEAN NOT NULL DEFAULT FALSE,
    expires_at  TIMESTAMPTZ NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_magic_links_code ON magic_links (code);
CREATE INDEX idx_magic_links_email ON magic_links (email);

-- +goose Down
DROP TABLE IF EXISTS magic_links;
DROP TABLE IF EXISTS auth_sessions;
DROP TABLE IF EXISTS users;
