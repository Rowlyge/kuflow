package router

import (
	"net/http"

	"github.com/Rowlyge/kuflow/internal/app"
	"github.com/Rowlyge/kuflow/internal/middleware"
)

// New создаёт маршруты приложения.
func New(app *app.App) *http.ServeMux {

	mux := http.NewServeMux()

	// Проверка состояния сервиса.
	mux.Handle(
		"/health",
		middleware.Default(
			http.HandlerFunc(
				app.Handlers.Health.GetStatus,
			),

			app.Middlewares.Logger,
			app.Middlewares.RequestID,
			app.Middlewares.Telemetry,
		),
	)

	// Reverse Proxy.
	mux.Handle(
		"/proxy/",
		middleware.Default(
			app.Services.Proxy,

			app.Middlewares.Logger,
			app.Middlewares.RequestID,
			app.Middlewares.Telemetry,
		),
	)

	// Runtime-метрики KuFlow (JSON).
	mux.Handle(
		"/metrics",
		app.Handlers.Metrics,
	)

	// Runtime-метрики KuFlow (Prometheus).
	mux.Handle(
		"/metrics/prometheus",
		app.Handlers.Prometheus,
	)

	return mux
}
