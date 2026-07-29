package app

import (
	"github.com/Rowlyge/kuflow/internal/balancer"
	"github.com/Rowlyge/kuflow/internal/config"
	"github.com/Rowlyge/kuflow/internal/service"
	"github.com/Rowlyge/kuflow/internal/upstream"
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
) (*Services, error) {

	// Создаём менеджер upstream-серверов.
	upstreamManager, err := upstream.NewManager(
		cfg.Proxy.Upstreams,
	)
	if err != nil {
		return nil, err
	}

	// Создаём балансировщик.
	// Пока используется Round Robin,
	// который работает с одним сервером.
	proxyBalancer := balancer.NewRoundRobin(
		upstreamManager,
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
