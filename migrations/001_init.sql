-- =====================================================
-- KuFlow Database Schema v1
-- Initial migration
-- =====================================================

CREATE TABLE requests
(
    id BIGSERIAL PRIMARY KEY,

    method VARCHAR(16) NOT NULL,

    url TEXT NOT NULL,

    client_ip VARCHAR(64) NOT NULL,

    user_agent TEXT NOT NULL,

    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_requests_created_at
ON requests(created_at);

CREATE INDEX idx_requests_method
ON requests(method);