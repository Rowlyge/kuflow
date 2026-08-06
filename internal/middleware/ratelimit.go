package middleware

import (
	"net/http"
	"strconv"

	"github.com/Rowlyge/kuflow/internal/clientip"
	"github.com/Rowlyge/kuflow/internal/ratelimit"
	"github.com/Rowlyge/kuflow/internal/service"
)

// RateLimit ограничивает количество запросов
// для одного клиента.
type RateLimit struct {
	limiter   *ratelimit.Limiter
	telemetry *service.TelemetryService
}

// NewRateLimit создаёт middleware Rate Limiter.
func NewRateLimit(
	limiter *ratelimit.Limiter,
	telemetry *service.TelemetryService,
) Middleware {
	return func(next http.Handler) http.Handler {
		rl := &RateLimit{
			limiter:   limiter,
			telemetry: telemetry,
		}
		return rl.Wrap(next)
	}
}

// Wrap реализует Middleware.
func (r *RateLimit) Wrap(
	next http.Handler,
) http.Handler {
	return http.HandlerFunc(func(
		w http.ResponseWriter,
		req *http.Request,
	) {
		apiKey := req.Header.Get("X-API-Key")
		ip := clientip.Get(req)
		key := apiKey + ":" + ip

		decision := r.limiter.Allow(key)

		w.Header().Set(
			"X-RateLimit-Limit",
			strconv.Itoa(decision.Limit),
		)
		w.Header().Set(
			"X-RateLimit-Remaining",
			strconv.Itoa(decision.Remaining),
		)
		w.Header().Set(
			"X-RateLimit-Retry-After",
			strconv.Itoa(int(decision.RetryAfter.Seconds())),
		)

		if !decision.Allowed {
			if r.telemetry != nil {
				r.telemetry.RecordRateLimit()
			}

			w.Header().Set(
				"Retry-After",
				strconv.Itoa(int(decision.RetryAfter.Seconds())),
			)

			http.Error(
				w,
				"Too Many Requests",
				http.StatusTooManyRequests,
			)
			return
		}

		next.ServeHTTP(w, req)
	})
}
