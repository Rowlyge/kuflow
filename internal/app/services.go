package app

import (
	authcache "github.com/Rowlyge/kuflow/internal/auth/cache"
	"github.com/Rowlyge/kuflow/internal/balancer"
	"github.com/Rowlyge/kuflow/internal/config"
	"github.com/Rowlyge/kuflow/internal/connectionlimit"
	"github.com/Rowlyge/kuflow/internal/ratelimit"
	"github.com/Rowlyge/kuflow/internal/service"
	authservice "github.com/Rowlyge/kuflow/internal/service/auth"
)

type Services struct {
	Health    *service.HealthService
	Proxy     *service.ProxyService
	Telemetry *service.TelemetryService

	Auth *authservice.Service

	// =========================
	// Runtime API Key Cache
	// =========================

	AuthCache     *authcache.Cache
	AuthLoader    *authcache.Loader
	AuthRefresher *authcache.Refresher

	// =========================
	// Runtime Rate Limiter
	// =========================

	RateLimiter *ratelimit.Limiter
	RateCleaner *ratelimit.Cleaner

	// =========================
	// Runtime Connection Limiter
	// =========================

	ConnectionLimiter *connectionlimit.Limiter
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

	// =========================
	// Runtime API Key Cache
	// =========================

	cache := authcache.New()

	loader := authcache.NewLoader(
		repositories.APIKey,
		cache,
	)

	refresher := authcache.NewRefresher(
		loader,
		cfg.AuthCache.RefreshInterval,
	)

	auth := authservice.New(
		authservice.NewValidator(cache),
	)

	// =========================
	// Runtime Rate Limiter
	// =========================

	limiter := ratelimit.New(
		ratelimit.Config{
			Capacity: cfg.RateLimit.Capacity,

			RefillTokens: cfg.RateLimit.RefillTokens,

			RefillInterval: cfg.RateLimit.RefillInterval,
		},
	)

	cleaner := ratelimit.NewCleaner(
		limiter.Store(),
		cfg.RateLimit.RefillInterval,
		10*cfg.RateLimit.RefillInterval,
	)

	// =========================
	// Runtime Connection Limiter
	// =========================

	connectionLimiter := connectionlimit.New(
		cfg.ConnectionLimit.MaxConnections,
	)

	return &Services{
		Health: service.NewHealthService(),

		Proxy: proxyService,

		Telemetry: service.NewTelemetryService(
			repositories.Request,
			infrastructure.Collector,
		),

		Auth: auth,

		AuthCache:     cache,
		AuthLoader:    loader,
		AuthRefresher: refresher,

		RateLimiter: limiter,
		RateCleaner: cleaner,

		ConnectionLimiter: connectionLimiter,
	}, nil
}
