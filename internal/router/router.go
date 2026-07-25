package router

import (
	"net/http"

	"github.com/Rowlyge/kuflow/internal/app"
	"github.com/Rowlyge/kuflow/internal/middleware"
)

// New создаёт маршруты приложения.
func New(
	handlers *app.Handlers,
	services *app.Services,
) *http.ServeMux {

	mux := http.NewServeMux()

	// Проверка состояния сервиса.
	mux.Handle(
		"/health",
		middleware.Default(
			http.HandlerFunc(
				handlers.Health.GetStatus,
			),
			services.Telemetry,
		),
	)

	// Все запросы /proxy/... пересылаются целевому серверу.
	mux.Handle(
		"/proxy/",
		middleware.Default(
			services.Proxy,
			services.Telemetry,
		),
	)

	return mux
}
