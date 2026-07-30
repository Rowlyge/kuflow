package model

import "time"

// Request описывает информацию
// об одном обработанном HTTP-запросе.
type Request struct {

	// Идентификатор записи.
	ID int64

	// HTTP-метод.
	Method string

	// Путь запроса.
	Path string

	// HTTP-статус ответа.
	StatusCode int

	// Время обработки запроса.
	Duration time.Duration

	// Размер ответа в байтах.
	ResponseSize int

	// IP клиента.
	ClientIP string

	// User-Agent клиента.
	UserAgent string

	// Имя upstream-сервера,
	// обработавшего запрос.
	Upstream string

	// Время обработки запроса.
	CreatedAt time.Time
}
