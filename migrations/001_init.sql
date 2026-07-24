-- ==========================================
-- Migration: 001_init.sql
-- Description: Create requests table
-- ==========================================

CREATE TABLE requests (
    id BIGSERIAL PRIMARY KEY,

    method VARCHAR(10) NOT NULL,

    host VARCHAR(255) NOT NULL,

    path TEXT NOT NULL,

    status_code SMALLINT NOT NULL,

    request_size INTEGER NOT NULL,

    response_size INTEGER NOT NULL,

    latency_ms INTEGER NOT NULL,

    client_ip INET,

    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_requests_host
ON requests(host);

CREATE INDEX idx_requests_created_at
ON requests(created_at);

CREATE INDEX idx_requests_status
ON requests(status_code);