package proxy

import (
	"net/http"

	"github.com/Rowlyge/kuflow/internal/balancer"
)

// newDirector создаёт Director для Reverse Proxy.
//
// Director вызывается перед каждым запросом.
// Он выбирает upstream через балансировщик
// и изменяет запрос так, чтобы он был отправлен
// выбранному серверу.
func newDirector(
	b balancer.Balancer,
) func(*http.Request) {

	return func(req *http.Request) {

		// Получаем следующий доступный upstream.
		upstream, err := b.Next()

		if err != nil {

			// Не удалось выбрать ни одного
			// доступного upstream.
			//
			// ErrorHandler позже вернёт 503.
			*req = *req.WithContext(
				MarkUpstreamUnavailable(
					req.Context(),
				),
			)

			return
		}

		// Сохраняем выбранный upstream
		// в Context для telemetry и Circuit Breaker.
		*req = *req.WithContext(
			IntoContext(
				req.Context(),
				upstream,
			),
		)

		originalHost := req.Host

		// Меняем адрес назначения.
		req.URL.Scheme = upstream.URL.Scheme
		req.URL.Host = upstream.URL.Host
		req.Host = upstream.URL.Host

		req.Header.Set(
			"X-Forwarded-Host",
			originalHost,
		)

		// Передаём схему запроса.
		req.Header.Set(
			"X-Forwarded-Proto",
			req.URL.Scheme,
		)

		// Передаём адрес клиента.
		req.Header.Set(
			"X-Forwarded-For",
			req.RemoteAddr,
		)
	}
}
