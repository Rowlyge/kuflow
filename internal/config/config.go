package config

import (
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Proxy    ProxyConfig
}

type ServerConfig struct {
	Port string
}

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
	SSLMode  string
}

type ProxyConfig struct {
	// Список upstream-серверов.
	Upstreams []string
}

func Load() *Config {

	if err := godotenv.Load(); err != nil {
		log.Println("Config: .env file not found, using system environment")
	}

	return &Config{
		Server: ServerConfig{
			Port: getEnv("PORT", "8080"),
		},

		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     getEnv("DB_USER", "proxy"),
			Password: getEnv("DB_PASSWORD", "proxy"),
			Name:     getEnv("DB_NAME", "proxydb"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
		},

		Proxy: ProxyConfig{
			Upstreams: getUpstreams(),
		},
	}
}

// getUpstreams читает список upstream-серверов
// из переменной окружения PROXY_UPSTREAMS.
func getUpstreams() []string {

	value := getEnv(
		"PROXY_UPSTREAMS",
		"http://localhost:8081",
	)

	parts := strings.Split(value, ",")

	upstreams := make([]string, 0, len(parts))

	for _, part := range parts {

		part = strings.TrimSpace(part)

		if part != "" {
			upstreams = append(upstreams, part)
		}
	}

	return upstreams
}

func getEnv(key, defaultValue string) string {

	if value, exists := os.LookupEnv(key); exists {
		return value
	}

	return defaultValue
}
