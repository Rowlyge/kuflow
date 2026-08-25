package router

import (
	"net/http"
	"strings"

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

			app.Middlewares.ConnectionLimit,

			app.Middlewares.RateLimit,

			app.Middlewares.Telemetry,
		),
	)

	// Admin API Keys
	mux.Handle(
		"/admin/api-keys",
		middleware.Default(
			http.HandlerFunc(func(
				w http.ResponseWriter,
				r *http.Request,
			) {

				switch r.Method {

				case http.MethodGet:
					app.Handlers.APIKey.List(
						w,
						r,
					)

				case http.MethodPost:
					app.Handlers.APIKey.Create(
						w,
						r,
					)

				default:

					http.Error(
						w,
						"method not allowed",
						http.StatusMethodNotAllowed,
					)
				}
			}),

			app.Middlewares.RequestID,
			app.Middlewares.Logger,
		),
	)

	mux.Handle(
		"/admin/api-keys/",
		middleware.Default(
			http.HandlerFunc(func(
				w http.ResponseWriter,
				r *http.Request,
			) {

				if r.Method == http.MethodPatch &&
					strings.HasSuffix(
						r.URL.Path,
						"/disable",
					) {

					app.Handlers.APIKey.Disable(
						w,
						r,
					)

					return
				}

				if r.Method == http.MethodDelete {

					app.Handlers.APIKey.Delete(
						w,
						r,
					)

					return
				}

				http.NotFound(
					w,
					r,
				)
			}),

			app.Middlewares.RequestID,
			app.Middlewares.Logger,
		),
	)

	return mux
}
