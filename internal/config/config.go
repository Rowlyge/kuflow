package config

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// Config объединяет всю конфигурацию приложения.
type Config struct {
	Server    ServerConfig
	Proxy     ProxyConfig
	Health    HealthConfig
	Telemetry TelemetryConfig
	Database  DatabaseConfig
	Auth      AuthConfig
}

// ==========================
// Server
// ==========================

// ServerConfig содержит настройки HTTP-сервера.
type ServerConfig struct {
	Port string

	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
}

// ==========================
// Proxy
// ==========================

// ProxyConfig содержит настройки Reverse Proxy.
type ProxyConfig struct {
	Upstreams []string

	// round_robin, random, least_connections...
	Balancer string

	DialTimeout           time.Duration
	ResponseHeaderTimeout time.Duration
	FlushInterval         time.Duration
}

// ==========================
// Health Checker
// ==========================

// HealthConfig содержит настройки Health Checker.
type HealthConfig struct {
	Enabled bool

	Interval time.Duration
	Timeout  time.Duration

	Path string

	FailureThreshold int
	SuccessThreshold int
}

// ==========================
// Telemetry
// ==========================

// TelemetryConfig содержит настройки телеметрии.
type TelemetryConfig struct {
	Enabled bool

	// Позже позволит писать телеметрию
	// через очередь.
	Async bool
}

// ==========================
// Database
// ==========================

// DatabaseConfig содержит настройки PostgreSQL.
type DatabaseConfig struct {
	URL string
}

// =====================================================
// Load
// =====================================================

// Load загружает всю конфигурацию приложения.
func Load() *Config {

	return &Config{

		Server: loadServerConfig(),

		Proxy: loadProxyConfig(),

		Health: loadHealthConfig(),

		Telemetry: loadTelemetryConfig(),

		Database: loadDatabaseConfig(),

		Auth: loadAuthConfig(),
	}
}

// =====================================================
// Server
// =====================================================

func loadServerConfig() ServerConfig {

	return ServerConfig{

		Port: getEnv(
			"SERVER_PORT",
			"8080",
		),

		ReadTimeout: mustDuration(
			"SERVER_READ_TIMEOUT",
			10*time.Second,
		),

		WriteTimeout: mustDuration(
			"SERVER_WRITE_TIMEOUT",
			30*time.Second,
		),

		IdleTimeout: mustDuration(
			"SERVER_IDLE_TIMEOUT",
			60*time.Second,
		),
	}
}

// =====================================================
// Proxy
// =====================================================

func loadProxyConfig() ProxyConfig {

	return ProxyConfig{

		Upstreams: strings.Split(
			getEnv(
				"PROXY_UPSTREAMS",
				"http://localhost:8081",
			),
			",",
		),

		Balancer: getEnv(
			"PROXY_BALANCER",
			"round_robin",
		),

		DialTimeout: mustDuration(
			"PROXY_DIAL_TIMEOUT",
			5*time.Second,
		),

		ResponseHeaderTimeout: mustDuration(
			"PROXY_RESPONSE_HEADER_TIMEOUT",
			10*time.Second,
		),

		FlushInterval: mustDuration(
			"PROXY_FLUSH_INTERVAL",
			100*time.Millisecond,
		),
	}
}

// =====================================================
// Health Checker
// =====================================================

func loadHealthConfig() HealthConfig {

	return HealthConfig{

		Enabled: mustBool(
			"HEALTH_ENABLED",
			true,
		),

		Interval: mustDuration(
			"HEALTH_INTERVAL",
			5*time.Second,
		),

		Timeout: mustDuration(
			"HEALTH_TIMEOUT",
			2*time.Second,
		),

		Path: getEnv(
			"HEALTH_PATH",
			"/health",
		),

		FailureThreshold: mustInt(
			"HEALTH_FAILURE_THRESHOLD",
			3,
		),

		SuccessThreshold: mustInt(
			"HEALTH_SUCCESS_THRESHOLD",
			2,
		),
	}
}

// =====================================================
// Telemetry
// =====================================================

func loadTelemetryConfig() TelemetryConfig {

	return TelemetryConfig{

		Enabled: mustBool(
			"TELEMETRY_ENABLED",
			true,
		),

		Async: mustBool(
			"TELEMETRY_ASYNC",
			false,
		),
	}
}

func mustInt(
	key string,
	defaultValue int,
) int {
	value := getEnv(
		key,
		strconv.Itoa(defaultValue),
	)

	result, err := strconv.Atoi(value)
	if err != nil {
		panic(fmt.Sprintf(
			"invalid integer value for %s: %q",
			key,
			value,
		))
	}

	return result
}

// =====================================================
// Database
// =====================================================

func loadDatabaseConfig() DatabaseConfig {

	return DatabaseConfig{

		URL: getEnv(
			"DATABASE_URL",
			"",
		),
	}
}
