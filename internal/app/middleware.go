package app

import "github.com/Rowlyge/kuflow/internal/middleware"

// Middlewares объединяет middleware приложения.
type Middlewares struct {
	Logger    middleware.Middleware
	RequestID middleware.Middleware
	Telemetry middleware.Middleware

	Auth middleware.Middleware

	RateLimit middleware.Middleware
}

// NewMiddlewares создаёт middleware приложения.
func NewMiddlewares(
	services *Services,
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
	}, nil
}
