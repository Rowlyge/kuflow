package middleware

import (
	"net/http"
	"time"

	"github.com/Rowlyge/kuflow/internal/service"
)

// TelemetryMiddleware собирает информацию о запросах
// и в будущем будет сохранять её через TelemetryService.
type TelemetryMiddleware struct {
	service *service.TelemetryService
}

// NewTelemetry создаёт middleware телеметрии.
func NewTelemetry(
	service *service.TelemetryService,
) Middleware {

	t := &TelemetryMiddleware{
		service: service,
	}

	return t.Handler
}

func (t *TelemetryMiddleware) Handler(next http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		start := time.Now()

		rw := &ResponseWriter{
			ResponseWriter: w,
		}

		next.ServeHTTP(rw, r)

		duration := time.Since(start)

		// На следующем этапе здесь появится сохранение
		// информации о запросе в PostgreSQL.
		_ = duration
		_ = rw
	})
}
