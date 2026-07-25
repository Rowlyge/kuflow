package app

import (
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
) (*Services, error) {

	proxyService, err := service.NewProxyService(
		cfg.Proxy.Target,
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
