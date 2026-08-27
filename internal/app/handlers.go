package app

import (
	"github.com/Rowlyge/kuflow/internal/handler"
	apikeyhandler "github.com/Rowlyge/kuflow/internal/handler/apikey"
)

// Handlers объединяет HTTP-обработчики.
type Handlers struct {
	Health *handler.HealthHandler

	APIKey      *apikeyhandler.Handler
	APIKeyStats *apikeyhandler.StatsHandler

	Metrics    *handler.MetricsHandler
	Prometheus *handler.PrometheusHandler
}

// NewHandlers создаёт HTTP-обработчики.
func NewHandlers(
	services *Services,
	infrastructure *Infrastructure,
) (*Handlers, error) {

	apiKeyHandler := apikeyhandler.New(
		services.APIKey,
	)

	apiKeyStatsHandler := apikeyhandler.NewStatsHandler(
		services.APIKeyStats,
	)

	return &Handlers{

		APIKey: apiKeyHandler,

		APIKeyStats: apiKeyStatsHandler,

		Health: handler.NewHealthHandler(
			services.Health,
		),

		Metrics: handler.NewMetricsHandler(
			infrastructure.Collector,
		),

		Prometheus: handler.NewPrometheusHandler(
			infrastructure.Collector,
		),
	}, nil
}
