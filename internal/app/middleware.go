package app

import (
	"net/http"

	"github.com/Rowlyge/kuflow/internal/config"
	"github.com/Rowlyge/kuflow/internal/connectionlimit"
	"github.com/Rowlyge/kuflow/internal/middleware"
)

// Middlewares объединяет middleware приложения.
type Middlewares struct {
	Logger          middleware.Middleware
	RequestID       middleware.Middleware
	Telemetry       middleware.Middleware
	Auth            middleware.Middleware
	RateLimit       middleware.Middleware
	ConnectionLimit middleware.Middleware
}

// NewMiddlewares создаёт middleware приложения.
func NewMiddlewares(
	services *Services,
	cfg *config.Config,
) (*Middlewares, error) {
	return &Middlewares{
		Logger: middleware.NewLogger(),

		RequestID: middleware.NewRequestID(),

		Telemetry: middleware.NewTelemetry(
			services.Telemetry,
		),

		Auth: middleware.NewAuth(
			services.Auth,
		),

		RateLimit: middleware.NewRateLimit(
			services.RateLimiter,
			services.Telemetry,
		),

		ConnectionLimit: func(next http.Handler) http.Handler {
			return connectionlimit.NewMiddleware(
				services.ConnectionLimiter,
				cfg.Auth,
			).Wrap(next)
		},
	}, nil
}
