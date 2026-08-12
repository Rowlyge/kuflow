package app

import (
	"time"

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
		10*time.Second,
	)

	auth := authservice.New(
		authservice.NewValidator(cache),
	)

	// =========================
	// Runtime Rate Limiter
	// =========================

	limiter := ratelimit.New(
		ratelimit.Config{
			Capacity:       100,
			RefillTokens:   100,
			RefillInterval: time.Minute,
		},
	)

	cleaner := ratelimit.NewCleaner(
		limiter.Store(),
		time.Minute,
		10*time.Minute,
	)

	// =========================
	// Runtime Connection Limiter
	// =========================

	connectionLimiter := connectionlimit.New(100)

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
