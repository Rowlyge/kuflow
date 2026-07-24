package config

import (
	"log"
	"os"

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
	Target string
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
			Target: getEnv("PROXY_TARGET", "https://httpbin.org"),
		},
	}
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}

	return defaultValue
}
