package config

import (
	"log"
	"os"
	"strconv"
	"time"
)

// getEnv возвращает значение переменной окружения
// либо значение по умолчанию.
func getEnv(
	key string,
	fallback string,
) string {

	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	return value
}

// mustBool читает bool из переменной окружения.
// При ошибке завершает приложение.
func mustBool(
	key string,
	fallback bool,
) bool {

	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	result, err := strconv.ParseBool(value)
	if err != nil {
		log.Fatalf(
			"invalid value for %s: %v",
			key,
			err,
		)
	}

	return result
}

// mustDuration читает time.Duration
// из переменной окружения.
//
// Примеры:
//
//	5s
//	500ms
//	1m
func mustDuration(
	key string,
	fallback time.Duration,
) time.Duration {

	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	result, err := time.ParseDuration(value)
	if err != nil {
		log.Fatalf(
			"invalid duration for %s: %v",
			key,
			err,
		)
	}

	return result
}
