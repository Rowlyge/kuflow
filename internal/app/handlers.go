package app

import (
	"github.com/Rowlyge/kuflow/internal/handler"
)

// Handlers объединяет HTTP-обработчики.
type Handlers struct {
	Health     *handler.HealthHandler
	Metrics    *handler.MetricsHandler
	Prometheus *handler.PrometheusHandler
}

// NewHandlers создаёт HTTP-обработчики.
func NewHandlers(
	services *Services,
	infrastructure *Infrastructure,
) (*Handlers, error) {

	return &Handlers{

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
