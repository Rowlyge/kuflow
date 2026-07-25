package proxy

import (
	"net/http"
	"net/url"
)

// newDirector подготавливает запрос
// перед отправкой на upstream.
func newDirector(
	target *url.URL,
) func(*http.Request) {

	return func(r *http.Request) {

		// Сохраняем Host, к которому обратился клиент.
		originalHost := r.Host

		// Настраиваем URL целевого сервера.
		r.URL.Scheme = target.Scheme
		r.URL.Host = target.Host

		// Изменяем путь.
		rewritePath(r)

		// Удаляем запрещённые заголовки.
		removeHopHeaders(r.Header)

		// Восстанавливаем Host для формирования
		// X-Forwarded-* заголовков.
		r.Host = originalHost

		addForwardHeaders(r)

		// После этого Host меняется на upstream.
		r.Host = target.Host
	}
}
