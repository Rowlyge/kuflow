package config

import (
	"testing"
)

func TestValidateSuccess(t *testing.T) {

	cfg := &Config{
		Server: ServerConfig{
			Port: "8080",
		},

		Database: DatabaseConfig{
			URL: "postgres://proxy:proxy@localhost:5432/proxydb?sslmode=disable",
		},

		Proxy: ProxyConfig{
			Upstreams: []string{
				"http://localhost:8081",
			},
		},

		Auth: AuthConfig{
			APIKeyHeader: "X-API-Key",
		},
	}

	if err := cfg.Validate(); err != nil {
		t.Fatalf(
			"expected valid config, got error: %v",
			err,
		)
	}
}

func TestValidateMissingDatabaseURL(t *testing.T) {

	cfg := &Config{
		Server: ServerConfig{
			Port: "8080",
		},

		Proxy: ProxyConfig{
			Upstreams: []string{
				"http://localhost:8081",
			},
		},

		Auth: AuthConfig{
			APIKeyHeader: "X-API-Key",
		},
	}

	if err := cfg.Validate(); err == nil {
		t.Fatal(
			"expected validation error for DATABASE_URL",
		)
	}
}

func TestValidateMissingUpstreams(t *testing.T) {

	cfg := &Config{
		Server: ServerConfig{
			Port: "8080",
		},

		Database: DatabaseConfig{
			URL: "postgres://proxy:proxy@localhost:5432/proxydb?sslmode=disable",
		},

		Auth: AuthConfig{
			APIKeyHeader: "X-API-Key",
		},
	}

	if err := cfg.Validate(); err == nil {
		t.Fatal(
			"expected validation error for PROXY_UPSTREAMS",
		)
	}
}

func TestValidateMissingAPIKeyHeader(t *testing.T) {

	cfg := &Config{
		Server: ServerConfig{
			Port: "8080",
		},

		Database: DatabaseConfig{
			URL: "postgres://proxy:proxy@localhost:5432/proxydb?sslmode=disable",
		},

		Proxy: ProxyConfig{
			Upstreams: []string{
				"http://localhost:8081",
			},
		},
	}

	if err := cfg.Validate(); err == nil {
		t.Fatal(
			"expected validation error for API key header",
		)
	}
}

func TestValidateMissingServerPort(t *testing.T) {

	cfg := &Config{
		Database: DatabaseConfig{
			URL: "postgres://proxy:proxy@localhost:5432/proxydb?sslmode=disable",
		},

		Proxy: ProxyConfig{
			Upstreams: []string{
				"http://localhost:8081",
			},
		},

		Auth: AuthConfig{
			APIKeyHeader: "X-API-Key",
		},
	}

	if err := cfg.Validate(); err == nil {
		t.Fatal(
			"expected validation error for SERVER_PORT",
		)
	}
}
