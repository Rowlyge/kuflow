CREATE TABLE IF NOT EXISTS requests
(
    id BIGSERIAL PRIMARY KEY,

    method VARCHAR(16) NOT NULL,
    path TEXT NOT NULL,

    status_code INTEGER NOT NULL,

    duration_ms BIGINT NOT NULL,

    response_size BIGINT NOT NULL,

    upstream VARCHAR(128) NOT NULL,

    client_ip VARCHAR(64) NOT NULL,

    user_agent TEXT NOT NULL,

    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_requests_created_at
ON requests(created_at);

CREATE INDEX IF NOT EXISTS idx_requests_status_code
ON requests(status_code);

CREATE INDEX IF NOT EXISTS idx_requests_path
ON requests(path);

CREATE INDEX IF NOT EXISTS idx_requests_upstream
ON requests(upstream);