package integration

import (
	"context"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/Rowlyge/kuflow/internal/app"
	authcache "github.com/Rowlyge/kuflow/internal/auth/cache"
	"github.com/Rowlyge/kuflow/internal/balancer"
	"github.com/Rowlyge/kuflow/internal/config"
	"github.com/Rowlyge/kuflow/internal/connectionlimit"
	"github.com/Rowlyge/kuflow/internal/health"
	"github.com/Rowlyge/kuflow/internal/metrics"
	"github.com/Rowlyge/kuflow/internal/middleware"
	"github.com/Rowlyge/kuflow/internal/model"
	"github.com/Rowlyge/kuflow/internal/ratelimit"
	"github.com/Rowlyge/kuflow/internal/repository"
	"github.com/Rowlyge/kuflow/internal/router"
	"github.com/Rowlyge/kuflow/internal/service"
	authservice "github.com/Rowlyge/kuflow/internal/service/auth"
	"github.com/Rowlyge/kuflow/internal/upstream"
)

// Environment contains the reusable HTTP integration-test environment.
type Environment struct {
	server *httptest.Server
	client *http.Client

	authCache *authcache.Cache

	collector *metrics.Collector

	upstreams     *upstream.Manager
	healthChecker *health.Checker
	healthCancel  context.CancelFunc
}

// NewEnvironment creates a real KuFlow HTTP pipeline.
func NewEnvironment(
	upstreamTargets ...string,
) (*Environment, error) {

	cfg := &config.Config{}

	// Integration tests use the same API key header
	// as the production middleware.
	cfg.Auth.APIKeyHeader = "X-API-Key"

	collector := metrics.NewCollector()
	requests := &memoryRequestRepository{}
	stats := &memoryAPIKeyStatsRepository{}

	telemetry := service.NewTelemetryService(
		requests,
		stats,
		collector,
	)

	authCache := authcache.New()

	services := &app.Services{
		Health: service.NewHealthService(),

		Telemetry: telemetry,

		Auth: authservice.New(
			authservice.NewValidator(authCache),
		),

		// Default integration environment configuration.
		//
		// Tests that require a custom limit should use
		// NewEnvironmentWithLimits().
		RateLimiter: ratelimit.New(
			ratelimit.Config{
				Capacity:       100,
				RefillTokens:   100,
				RefillInterval: time.Minute,
			},
		),

		// Default connection limit.
		//
		// Tests that require a custom limit should use
		// NewEnvironmentWithLimits().
		ConnectionLimiter: connectionlimit.New(100),
	}

	infrastructure := &app.Infrastructure{
		Collector: collector,
	}

	var manager *upstream.Manager

	if len(upstreamTargets) > 0 {
		var err error

		manager, err = upstream.NewManager(upstreamTargets)
		if err != nil {
			return nil, err
		}

		infrastructure.Upstreams = manager

		proxyBalancer, err := balancer.New(
			"round_robin",
			manager,
		)
		if err != nil {
			return nil, err
		}

		proxyService, err := service.NewProxyService(
			proxyBalancer,
			config.ProxyConfig{
				DialTimeout:           time.Second,
				ResponseHeaderTimeout: time.Second,
			},
		)
		if err != nil {
			return nil, err
		}

		services.Proxy = proxyService
	}

	handlers, err := app.NewHandlers(
		services,
		infrastructure,
	)
	if err != nil {
		return nil, err
	}

	middlewares, err := app.NewMiddlewares(
		services,
		cfg,
	)
	if err != nil {
		return nil, err
	}

	application := &app.App{
		Config:         cfg,
		Services:       services,
		Handlers:       handlers,
		Middlewares:    middlewares,
		Infrastructure: infrastructure,
	}

	var handler http.Handler = router.New(application)

	if services.Proxy != nil {
		mux := http.NewServeMux()

		mux.Handle(
			"/health",
			middleware.Default(
				http.HandlerFunc(
					application.Handlers.Health.GetStatus,
				),
				application.Middlewares.RequestID,
				application.Middlewares.Logger,
				application.Middlewares.Telemetry,
			),
		)

		// Production proxy middleware order:
		//
		// RequestID
		//     ↓
		// Logger
		//     ↓
		// Auth
		//     ↓
		// ConnectionLimit
		//     ↓
		// RateLimit
		//     ↓
		// Telemetry
		//     ↓
		// Proxy / Balancer
		//
		// The order is important for integration tests:
		// rejected requests must never reach the proxy/upstream.
		mux.Handle(
			"/",
			middleware.Default(
				application.Services.Proxy,
				application.Middlewares.RequestID,
				application.Middlewares.Logger,
				application.Middlewares.Auth,
				application.Middlewares.ConnectionLimit,
				application.Middlewares.RateLimit,
				application.Middlewares.Telemetry,
			),
		)

		handler = mux
	}

	server := httptest.NewServer(handler)

	return &Environment{
		server:    server,
		client:    server.Client(),
		authCache: authCache,
		collector: collector,
		upstreams: manager,
	}, nil
}

// NewEnvironmentWithHealth creates an integration environment
// with two upstreams and a running Health Checker.
func NewEnvironmentWithHealth(
	upstreamA string,
	upstreamB string,
) (*Environment, error) {

	env, err := NewEnvironment(
		upstreamA,
		upstreamB,
	)
	if err != nil {
		return nil, err
	}

	healthCtx, healthCancel := context.WithCancel(
		context.Background(),
	)

	healthChecker := health.NewChecker(
		env.upstreams,
		config.HealthConfig{
			Enabled:          true,
			Interval:         20 * time.Millisecond,
			Timeout:          50 * time.Millisecond,
			Path:             "/health",
			FailureThreshold: 1,
			SuccessThreshold: 1,
		},
		nil,
	)

	env.healthChecker = healthChecker
	env.healthCancel = healthCancel

	healthChecker.Start(healthCtx)

	return env, nil
}

// URL returns the base URL of the KuFlow test server.
func (e *Environment) URL() string {
	return e.server.URL
}

// Client returns the HTTP client configured for the test server.
func (e *Environment) Client() *http.Client {
	return e.client
}

// SetAPIKeys replaces the runtime authentication cache.
func (e *Environment) SetAPIKeys(
	keys ...authcache.APIKey,
) {
	data := make(map[string]authcache.APIKey, len(keys))

	for _, key := range keys {
		data[key.Key] = key
	}

	e.authCache.Replace(data)
}

// Default integration environment configuration.
// Tests that require custom limits should use
// NewEnvironmentWithLimits().
func NewEnvironmentWithLimits(
	rateCfg ratelimit.Config,
	connectionLimit int,
	upstreamTargets ...string,
) (*Environment, error) {

	cfg := &config.Config{}

	// Match production API-key configuration.
	cfg.Auth.APIKeyHeader = "X-API-Key"

	collector := metrics.NewCollector()
	requests := &memoryRequestRepository{}
	stats := &memoryAPIKeyStatsRepository{}

	telemetry := service.NewTelemetryService(
		requests,
		stats,
		collector,
	)

	authCache := authcache.New()

	services := &app.Services{
		Health: service.NewHealthService(),

		Telemetry: telemetry,

		Auth: authservice.New(
			authservice.NewValidator(authCache),
		),

		RateLimiter: ratelimit.New(rateCfg),

		ConnectionLimiter: connectionlimit.New(connectionLimit),
	}

	infrastructure := &app.Infrastructure{
		Collector: collector,
	}

	if len(upstreamTargets) > 0 {
		manager, err := upstream.NewManager(upstreamTargets)
		if err != nil {
			return nil, err
		}

		infrastructure.Upstreams = manager

		proxyBalancer, err := balancer.New(
			"round_robin",
			manager,
		)
		if err != nil {
			return nil, err
		}

		proxyService, err := service.NewProxyService(
			proxyBalancer,
			config.ProxyConfig{
				DialTimeout:           time.Second,
				ResponseHeaderTimeout: time.Second,
			},
		)
		if err != nil {
			return nil, err
		}

		services.Proxy = proxyService
	}

	handlers, err := app.NewHandlers(
		services,
		infrastructure,
	)
	if err != nil {
		return nil, err
	}

	middlewares, err := app.NewMiddlewares(
		services,
		cfg,
	)
	if err != nil {
		return nil, err
	}

	application := &app.App{
		Config:         cfg,
		Services:       services,
		Handlers:       handlers,
		Middlewares:    middlewares,
		Infrastructure: infrastructure,
	}

	var handler http.Handler = router.New(application)

	if services.Proxy != nil {
		mux := http.NewServeMux()

		mux.Handle(
			"/health",
			middleware.Default(
				http.HandlerFunc(
					application.Handlers.Health.GetStatus,
				),
				application.Middlewares.RequestID,
				application.Middlewares.Logger,
				application.Middlewares.Telemetry,
			),
		)

		mux.Handle(
			"/",
			middleware.Default(
				application.Services.Proxy,
				application.Middlewares.RequestID,
				application.Middlewares.Logger,
				application.Middlewares.Auth,
				application.Middlewares.ConnectionLimit,
				application.Middlewares.RateLimit,
				application.Middlewares.Telemetry,
			),
		)

		handler = mux
	}

	server := httptest.NewServer(handler)

	return &Environment{
		server:    server,
		client:    server.Client(),
		authCache: authCache,
		collector: collector,
	}, nil
}

// Close shuts down the test HTTP server.
func (e *Environment) Close() {
	if e == nil {
		return
	}

	if e.healthCancel != nil {
		e.healthCancel()
	}

	if e.healthChecker != nil {
		e.healthChecker.Wait()
	}

	if e.server != nil {
		e.server.Close()
	}
}

type memoryRequestRepository struct{}

var _ repository.RequestRepository = (*memoryRequestRepository)(nil)

func (r *memoryRequestRepository) Create(
	_ context.Context,
	_ *model.Request,
) error {
	return nil
}

func (e *Environment) Collector() *metrics.Collector {
	return e.collector
}

type memoryAPIKeyStatsRepository struct{}

var _ repository.APIKeyStatsRepository = (*memoryAPIKeyStatsRepository)(nil)

func (r *memoryAPIKeyStatsRepository) UpdateUsage(
	ctx context.Context,
	apiKeyID int64,
	ip string,
) error {
	return nil
}

func (r *memoryAPIKeyStatsRepository) GetByAPIKeyID(
	ctx context.Context,
	apiKeyID int64,
) (*model.APIKeyStats, error) {
	return &model.APIKeyStats{}, nil
}

func (r *memoryAPIKeyStatsRepository) List(
	ctx context.Context,
) ([]model.APIKeyStats, error) {
	return []model.APIKeyStats{}, nil
}
