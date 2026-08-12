package connectionlimit

import (
	"net/http"

	"github.com/Rowlyge/kuflow/internal/config"
)

// Middleware ограничивает количество одновременно
// выполняющихся запросов для каждого API key.
type Middleware struct {
	limiter *Limiter
	header  string
}

// NewMiddleware создаёт Connection Limit Middleware.
func NewMiddleware(
	limiter *Limiter,
	cfg config.AuthConfig,
) *Middleware {
	return &Middleware{
		limiter: limiter,
		header:  cfg.APIKeyHeader,
	}
}

// Wrap оборачивает HTTP handler ограничением
// количества одновременно активных запросов.
func (m *Middleware) Wrap(
	next http.Handler,
) http.Handler {
	return http.HandlerFunc(func(
		w http.ResponseWriter,
		r *http.Request,
	) {
		apiKey := r.Header.Get(m.header)

		if !m.limiter.Acquire(apiKey) {
			http.Error(
				w,
				"Too Many Connections",
				http.StatusTooManyRequests,
			)

			return
		}

		defer m.limiter.Release(apiKey)

		next.ServeHTTP(w, r)
	})
}
