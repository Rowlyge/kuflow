package app

import (
	"time"

	authcache "github.com/Rowlyge/kuflow/internal/auth/cache"
	"github.com/Rowlyge/kuflow/internal/balancer"
	"github.com/Rowlyge/kuflow/internal/config"
	"github.com/Rowlyge/kuflow/internal/service"
	authservice "github.com/Rowlyge/kuflow/internal/service/auth"
)

type Services struct {
	Health    *service.HealthService
	Proxy     *service.ProxyService
	Telemetry *service.TelemetryService

	Auth *authservice.Service

	AuthCache     *authcache.Cache
	AuthLoader    *authcache.Loader
	AuthRefresher *authcache.Refresher
}

func NewServices(
	cfg *config.Config,
	repositories *Repositories,
	infrastructure *Infrastructure,
) (*Services, error) {

	proxyBalancer, err := balancer.New(
		cfg.Proxy.Balancer,
		infrastructure.Upstreams,
	)
	if err != nil {
		return nil, err
	}

	proxyService, err := service.NewProxyService(
		proxyBalancer,
		cfg.Proxy,
	)
	if err != nil {
		return nil, err
	}

	// -------------------------
	// Runtime API Key Cache
	// -------------------------

	cache := authcache.New()

	loader := authcache.NewLoader(
		repositories.APIKey,
		cache,
	)

	refresher := authcache.NewRefresher(
		loader,
		10*time.Second,
	)

	auth := authservice.New(
		authservice.NewValidator(cache),
	)

	return &Services{

		Health: service.NewHealthService(),

		Proxy: proxyService,

		Telemetry: service.NewTelemetryService(
			repositories.Request,
			infrastructure.Collector,
		),

		Auth: auth,

		AuthCache: cache,

		AuthLoader: loader,

		AuthRefresher: refresher,
	}, nil
}
