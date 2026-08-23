package config

import (
	"errors"
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

	RateLimit       RateLimitConfig
	ConnectionLimit ConnectionLimitConfig
	AuthCache       AuthCacheConfig
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
		Server:    loadServerConfig(),
		Proxy:     loadProxyConfig(),
		Health:    loadHealthConfig(),
		Telemetry: loadTelemetryConfig(),
		Database:  loadDatabaseConfig(),
		Auth:      loadAuthConfig(),

		RateLimit:       loadRateLimitConfig(),
		ConnectionLimit: loadConnectionLimitConfig(),
		AuthCache:       loadAuthCacheConfig(),
	}
}

// ==========================
// Rate Limiter
// ==========================

type RateLimitConfig struct {
	Capacity       int
	RefillTokens   int
	RefillInterval time.Duration
}

// ==========================
// Connection Limiter
// ==========================

type ConnectionLimitConfig struct {
	MaxConnections int
}

// ==========================
// Auth Cache
// ==========================

type AuthCacheConfig struct {
	RefreshInterval time.Duration
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

func loadRateLimitConfig() RateLimitConfig {
	return RateLimitConfig{
		Capacity: mustInt(
			"RATE_LIMIT_CAPACITY",
			100,
		),

		RefillTokens: mustInt(
			"RATE_LIMIT_REFILL_TOKENS",
			100,
		),

		RefillInterval: mustDuration(
			"RATE_LIMIT_REFILL_INTERVAL",
			time.Minute,
		),
	}
}

func loadConnectionLimitConfig() ConnectionLimitConfig {
	return ConnectionLimitConfig{
		MaxConnections: mustInt(
			"CONNECTION_LIMIT_MAX",
			100,
		),
	}
}

func loadAuthCacheConfig() AuthCacheConfig {
	return AuthCacheConfig{
		RefreshInterval: mustDuration(
			"AUTH_CACHE_REFRESH_INTERVAL",
			10*time.Second,
		),
	}
}

func (c *Config) Validate() error {

	if c == nil {
		return fmt.Errorf(
			"config is nil",
		)
	}

	if c.Database.URL == "" {
		return fmt.Errorf(
			"DATABASE_URL is required",
		)
	}

	if len(c.Proxy.Upstreams) == 0 {
		return fmt.Errorf(
			"PROXY_UPSTREAMS is required",
		)
	}

	for _, upstream := range c.Proxy.Upstreams {

		if strings.TrimSpace(upstream) == "" {
			return fmt.Errorf(
				"PROXY_UPSTREAMS contains empty value",
			)
		}
	}

	if strings.TrimSpace(c.Server.Port) == "" {
		return fmt.Errorf(
			"SERVER_PORT is required",
		)
	}

	if c.Auth.APIKeyHeader == "" {
		return errors.New("AUTH_API_KEY_HEADER is required")
	}

	return nil
}
