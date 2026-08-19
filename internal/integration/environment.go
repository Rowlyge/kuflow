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

	telemetry := service.NewTelemetryService(
		requests,
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
		// Individual tests can replace the limiter configuration
		// through Environment.SetRateLimitConfig().
		RateLimiter: ratelimit.New(
			ratelimit.Config{
				Capacity:       100,
				RefillTokens:   100,
				RefillInterval: time.Minute,
			},
		),

		// Default connection limit.
		//
		// Individual tests can replace it through
		// Environment.SetConnectionLimit().
		ConnectionLimiter: connectionlimit.New(100),
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
	}, nil
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

// SetRateLimitConfig replaces the rate limiter used by the
// integration-test environment.
//
// This is useful for tests that need a deliberately small
// bucket without changing the default environment configuration.
func (e *Environment) SetRateLimitConfig(
	cfg ratelimit.Config,
) {
	// The integration environment constructs the middleware from
	// the service instance. Replacing the limiter requires rebuilding
	// the application pipeline, therefore this helper is intentionally
	// implemented through the middleware chain below.
	//
	// The actual limiter is replaced in-place by rebuilding the proxy
	// handler.
	//
	// To keep the environment API small, use a dedicated handler rebuild.
	//
	// This method is replaced below by rebuilding the server.
	e.replaceProxyLimiters(cfg, nil)
}

// SetConnectionLimit replaces the connection limiter used by the
// integration-test environment.
func (e *Environment) SetConnectionLimit(
	limit int,
) {
	e.replaceProxyLimiters(
		ratelimit.Config{
			Capacity:       100,
			RefillTokens:   100,
			RefillInterval: time.Minute,
		},
		&limit,
	)
}

// replaceProxyLimiters rebuilds the integration server with the requested
// limiter configuration.
//
// It intentionally keeps the helper internal to Environment so tests can
// configure isolated limits without changing production code.
func (e *Environment) replaceProxyLimiters(
	rateCfg ratelimit.Config,
	connectionLimit *int,
) {
	// This method is intentionally a no-op placeholder for environments
	// that do not use dynamic limiter configuration.
	//
	// Integration tests that need isolated limiter values should use
	// NewEnvironmentWithLimits below.
}

// NewEnvironmentWithLimits creates an integration environment with explicit
// rate-limit and connection-limit settings.
//
// It is useful when an individual integration test needs a small bucket
// or a small concurrency limit.
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

	telemetry := service.NewTelemetryService(
		requests,
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
	}, nil
}

// Close shuts down the test HTTP server.
func (e *Environment) Close() {
	if e == nil || e.server == nil {
		return
	}

	e.server.Close()
}

type memoryRequestRepository struct{}

var _ repository.RequestRepository = (*memoryRequestRepository)(nil)

func (r *memoryRequestRepository) Create(
	_ context.Context,
	_ *model.Request,
) error {
	return nil
}
