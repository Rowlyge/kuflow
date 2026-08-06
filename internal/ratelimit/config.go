package ratelimit

import "time"

// Config содержит настройки Rate Limiter.
type Config struct {

	// Максимальное количество токенов.
	Capacity int

	// Сколько токенов восстанавливать.
	RefillTokens int

	// Интервал восстановления.
	RefillInterval time.Duration
}
