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

		// Получаем следующий upstream.
		upstream, err := b.Next()
		if err != nil {
			// Пока просто ничего не делаем.
			// В будущем здесь появится обработка ситуации,
			// когда нет доступных upstream-серверов.
			return
		}

		// Меняем адрес назначения.
		req.URL.Scheme = upstream.URL.Scheme
		req.URL.Host = upstream.URL.Host

		// Host передаём целевому серверу.
		originalHost := req.Host

		req.URL.Scheme = upstream.URL.Scheme
		req.URL.Host = upstream.URL.Host
		req.Host = upstream.URL.Host

		req.Header.Set(
			"X-Forwarded-Host",
			originalHost,
		)

		// Передаём оригинальную схему.
		req.Header.Set(
			"X-Forwarded-Proto",
			req.URL.Scheme,
		)

		// Если запрос пришёл по HTTP,
		// можно считать RemoteAddr клиентом.
		req.Header.Set(
			"X-Forwarded-For",
			req.RemoteAddr,
		)
	}
}
