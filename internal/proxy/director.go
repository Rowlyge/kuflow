package proxy

import (
	"net/http/httputil"

	"github.com/Rowlyge/kuflow/internal/balancer"
)

// newDirector создаёт Rewrite для Reverse Proxy.
//
// Rewrite вызывается перед каждым запросом.
// Он выбирает upstream через балансировщик,
// проверяет Circuit Breaker и изменяет исходящий запрос
// так, чтобы он был отправлен выбранному серверу.
//
// Если доступного upstream нет или его Circuit Breaker
// запрещает запрос, запрос помечается как недоступный.
// ReverseProxy передаст ошибку в ErrorHandler,
// который вернёт клиенту HTTP 503.
func newDirector(
	b balancer.Balancer,
) func(*httputil.ProxyRequest) {

	return func(pr *httputil.ProxyRequest) {

		req := pr.In

		// Получаем следующий upstream через балансировщик.
		upstream, err := b.Next()

		if err != nil || upstream == nil {
			ctx := MarkUpstreamUnavailable(
				req.Context(),
			)

			// Важно: помечаем именно исходящий запрос.
			//
			// ReverseProxy использует pr.Out для вызова
			// Transport и передаёт его context в ErrorHandler.
			pr.Out = pr.Out.WithContext(ctx)

			return
		}

		// Проверяем Circuit Breaker выбранного upstream.
		if upstream.Breaker == nil || !upstream.Breaker.Allow() {
			ctx := MarkUpstreamUnavailable(
				req.Context(),
			)

			pr.Out = pr.Out.WithContext(ctx)

			return
		}

		// Сохраняем выбранный upstream в Context.
		//
		// Context нужен:
		// - ModifyResponse для OnSuccess/OnFailure;
		// - ErrorHandler для transport failure;
		// - telemetry;
		// - Circuit Breaker.
		ctx := IntoContext(
			req.Context(),
			upstream,
		)

		pr.Out = pr.Out.WithContext(ctx)

		originalHost := req.Host

		// Перенаправляем исходящий запрос на выбранный upstream.
		pr.SetURL(upstream.URL)

		// Host исходящего запроса.
		pr.Out.Host = upstream.URL.Host

		// Передаём исходный Host клиента.
		pr.Out.Header.Set(
			"X-Forwarded-Host",
			originalHost,
		)

		// Передаём схему upstream.
		pr.Out.Header.Set(
			"X-Forwarded-Proto",
			upstream.URL.Scheme,
		)

		// Передаём адрес клиента.
		pr.Out.Header.Set(
			"X-Forwarded-For",
			req.RemoteAddr,
		)
	}
}
