package breaker

import "time"

// Config описывает параметры Circuit Breaker.
type Config struct {

	// Сколько подряд ошибок нужно,
	// чтобы открыть Circuit.
	FailureThreshold int

	// Сколько успешных запросов подряд
	// требуется для закрытия Circuit
	// после Half-Open.
	SuccessThreshold int

	// Через сколько времени
	// после открытия разрешить
	// тестовый запрос.
	OpenTimeout time.Duration
}

// DefaultConfig возвращает
// рекомендуемую конфигурацию.
func DefaultConfig() Config {

	return Config{

		FailureThreshold: 5,

		SuccessThreshold: 2,

		OpenTimeout: 30 * time.Second,
	}
}
