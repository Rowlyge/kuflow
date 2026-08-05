package app

import (
	"github.com/Rowlyge/kuflow/internal/balancer"
	"github.com/Rowlyge/kuflow/internal/config"
	"github.com/Rowlyge/kuflow/internal/service"
	authservice "github.com/Rowlyge/kuflow/internal/service/auth"
)

// Services объединяет бизнес-логику приложения.
type Services struct {
	Health    *service.HealthService
	Proxy     *service.ProxyService
	Telemetry *service.TelemetryService

	// Авторизация клиентов Proxy.
	Auth *authservice.Service
}

// NewServices создаёт сервисы приложения.
func NewServices(
	cfg *config.Config,
	repositories *Repositories,
	infrastructure *Infrastructure,
) (*Services, error) {

	// Создаём балансировщик.
	proxyBalancer, err := balancer.New(
		cfg.Proxy.Balancer,
		infrastructure.Upstreams,
	)
	if err != nil {
		return nil, err
	}

	// Создаём сервис Reverse Proxy.
	proxyService, err := service.NewProxyService(
		proxyBalancer,
		cfg.Proxy,
	)
	if err != nil {
		return nil, err
	}

	// Создаём сервис авторизации.
	authService := authservice.New(
		repositories.APIKey,
	)

	return &Services{

		Health: service.NewHealthService(),

		Proxy: proxyService,

		Telemetry: service.NewTelemetryService(
			repositories.Request,
			infrastructure.Collector,
		),

		Auth: authService,
	}, nil
}
