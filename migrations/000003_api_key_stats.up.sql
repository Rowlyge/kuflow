CREATE TABLE api_key_stats (
    api_key_id BIGINT PRIMARY KEY
        REFERENCES api_keys(id)
        ON DELETE CASCADE,

    requests_total BIGINT NOT NULL DEFAULT 0,

    last_used_at TIMESTAMP,

    last_seen_ip TEXT
);