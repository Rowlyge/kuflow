package app

import "github.com/Rowlyge/kuflow/internal/middleware"

// Middlewares объединяет middleware приложения.
type Middlewares struct {
	Logger    middleware.Middleware
	Telemetry middleware.Middleware
}

// NewMiddlewares создаёт middleware приложения.
func NewMiddlewares(
	services *Services,
) (*Middlewares, error) {

	return &Middlewares{

		Logger: middleware.NewLogger(),

		Telemetry: middleware.NewTelemetry(
			services.Telemetry,
		),
	}, nil
}
