package integration

import (
	"context"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/Rowlyge/kuflow/internal/app"
	"github.com/Rowlyge/kuflow/internal/auth/cache"
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
//
// It intentionally does not create a PostgreSQL connection. Integration tests
// use an in-memory request repository and httptest upstreams instead.
type Environment struct {
	server *httptest.Server
	client *http.Client
}

// NewEnvironment creates a real KuFlow HTTP pipeline.
//
// With no upstream targets it exposes the normal health/metrics routes used
// by infrastructure smoke tests. When at least one upstream target is passed,
// the proxy route is also wired to a real ProxyService and production
// middleware. Authentication is intentionally omitted from that route until
// integration stage 6.2, where API-key handling becomes the subject under test.
func NewEnvironment(upstreamTargets ...string) (*Environment, error) {
	cfg := &config.Config{}

	collector := metrics.NewCollector()
	requests := &memoryRequestRepository{}

	telemetry := service.NewTelemetryService(
		requests,
		collector,
	)

	cache := cache.New()

	services := &app.Services{
		Health: service.NewHealthService(),

		Telemetry: telemetry,

		Auth: authservice.New(
			authservice.NewValidator(cache),
		),

		RateLimiter: ratelimit.New(
			ratelimit.Config{
				Capacity:     100,
				RefillTokens: 100,
			},
		),

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

		// Keep the production proxy middleware order, except Auth.
		// Authentication becomes part of the dedicated 6.2 integration stage.
		mux.Handle(
			"/",
			middleware.Default(
				application.Services.Proxy,
				application.Middlewares.RequestID,
				application.Middlewares.Logger,
				application.Middlewares.ConnectionLimit,
				application.Middlewares.RateLimit,
				application.Middlewares.Telemetry,
			),
		)

		handler = mux
	}

	server := httptest.NewServer(handler)

	return &Environment{
		server: server,
		client: server.Client(),
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

// Close shuts down the test HTTP server and releases its client resources.
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
