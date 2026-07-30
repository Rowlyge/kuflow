package middleware

import (
	"log"
	"net/http"
	"time"

	"github.com/Rowlyge/kuflow/internal/clientip"
	"github.com/Rowlyge/kuflow/internal/model"
	"github.com/Rowlyge/kuflow/internal/proxy"
	"github.com/Rowlyge/kuflow/internal/service"
)

// TelemetryMiddleware собирает информацию
// о каждом HTTP-запросе.
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

// Handler собирает телеметрию после обработки запроса,
// формирует полноценную модель и отправляет её в сервис.
func (t *TelemetryMiddleware) Handler(
	next http.Handler,
) http.Handler {

	return http.HandlerFunc(func(
		w http.ResponseWriter,
		r *http.Request,
	) {

		start := time.Now()

		rw := NewResponseWriter(w)

		next.ServeHTTP(rw, r)

		request := &model.Request{
			Method:       r.Method,
			Path:         r.URL.Path,
			StatusCode:   rw.StatusCode(),
			Duration:     time.Since(start),
			ResponseSize: rw.BytesWritten(),
			ClientIP:     clientip.Get(r),
			UserAgent:    r.UserAgent(),

			Upstream: proxy.UpstreamFromContext(
				r.Context(),
			),

			CreatedAt: start,
		}

		// Ошибку логировать будем позже,
		// когда появится общий Logger.
		if err := t.service.Save(
			r.Context(),
			request,
		); err != nil {

			log.Printf(
				"Telemetry: failed to save request: %v",
				err,
			)
		}
	})
}
