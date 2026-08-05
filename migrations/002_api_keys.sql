CREATE TABLE IF NOT EXISTS api_keys
(
    id BIGSERIAL PRIMARY KEY,

    api_key TEXT NOT NULL UNIQUE,

    owner VARCHAR(255) NOT NULL,

    enabled BOOLEAN NOT NULL DEFAULT TRUE,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    expires_at TIMESTAMPTZ
);

CREATE INDEX idx_api_keys_enabled
ON api_keys(enabled);

CREATE INDEX idx_api_keys_owner
ON api_keys(owner);