package router

import (
	"net/http"

	"github.com/Rowlyge/kuflow/internal/app"
	"github.com/Rowlyge/kuflow/internal/middleware"
)

// New создаёт маршруты приложения.
func New(app *app.App) *http.ServeMux {

	mux := http.NewServeMux()

	// Health
	mux.Handle(
		"/health",
		middleware.Default(
			http.HandlerFunc(
				app.Handlers.Health.GetStatus,
			),

			app.Middlewares.RequestID,
			app.Middlewares.Logger,
			app.Middlewares.Telemetry,
		),
	)

	// Runtime JSON
	mux.Handle(
		"/metrics",
		app.Handlers.Metrics,
	)

	// Prometheus
	mux.Handle(
		"/metrics/prometheus",
		app.Handlers.Prometheus,
	)

	// Любой другой запрос считается Proxy-запросом.
	mux.Handle(
		"/",
		middleware.Default(
			app.Services.Proxy,

			app.Middlewares.RequestID,

			app.Middlewares.Logger,

			app.Middlewares.Auth,

			app.Middlewares.RateLimit,

			app.Middlewares.Telemetry,
		),
	)

	return mux
}
