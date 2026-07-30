package app

import (
	"github.com/Rowlyge/kuflow/internal/balancer"
	"github.com/Rowlyge/kuflow/internal/config"
	"github.com/Rowlyge/kuflow/internal/service"
)

// Services объединяет бизнес-логику приложения.
type Services struct {
	Health    *service.HealthService
	Proxy     *service.ProxyService
	Telemetry *service.TelemetryService
}

// NewServices создаёт сервисы приложения.
func NewServices(
	cfg *config.Config,
	repositories *Repositories,
	infrastructure *Infrastructure,
) (*Services, error) {

	// Создаём балансировщик.
	proxyBalancer := balancer.NewRoundRobin(
		infrastructure.Upstreams,
	)

	// Создаём сервис Reverse Proxy.
	proxyService, err := service.NewProxyService(
		proxyBalancer,
	)
	if err != nil {
		return nil, err
	}

	return &Services{

		Health: service.NewHealthService(),

		Proxy: proxyService,

		Telemetry: service.NewTelemetryService(
			repositories.Request,
		),
	}, nil
}
