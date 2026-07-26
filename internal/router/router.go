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

	// Все запросы /proxy/... пересылаются
	// целевому серверу.
	mux.Handle(
		"/proxy/",
		middleware.Default(
			app.Services.Proxy,

			app.Middlewares.Logger,
			app.Middlewares.RequestID,
			app.Middlewares.Telemetry,
		),
	)

	return mux
}
