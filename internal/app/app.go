package app

import (
	"context"
	"sync"

	"github.com/Rowlyge/kuflow/internal/config"
	"github.com/Rowlyge/kuflow/internal/database"

	"github.com/jackc/pgx/v5/pgxpool"
)

// App — главный контейнер приложения.
type App struct {
	Config *config.Config

	DB *pgxpool.Pool

	Repositories   *Repositories
	Infrastructure *Infrastructure
	Services       *Services
	Handlers       *Handlers
	Middlewares    *Middlewares

	// Lifecycle фоновых workers.
	lifecycleCtx    context.Context
	lifecycleCancel context.CancelFunc
	lifecycleOnce   sync.Once
	lifecycleWG     sync.WaitGroup
}

// New создаёт приложение и инициализирует все зависимости.
//
// New не запускает background workers.
// Их запуском управляет App.Start.
func New(cfg *config.Config) (*App, error) {
	db, err := database.New(cfg.Database)
	if err != nil {
		return nil, err
	}

	repositories, err := NewRepositories(db)
	if err != nil {
		db.Close()

		return nil, err
	}

	infrastructure, err := NewInfrastructure(
		cfg.Proxy,
		cfg.Health,
	)
	if err != nil {
		db.Close()

		return nil, err
	}

	services, err := NewServices(
		cfg,
		repositories,
		infrastructure,
	)
	if err != nil {
		db.Close()

		return nil, err
	}

	handlers, err := NewHandlers(
		services,
		infrastructure,
	)
	if err != nil {
		db.Close()

		return nil, err
	}

	middlewares, err := NewMiddlewares(
		services,
		cfg,
	)
	if err != nil {
		db.Close()

		return nil, err
	}

	ctx, cancel := context.WithCancel(
		context.Background(),
	)

	return &App{
		Config: cfg,

		DB: db,

		Repositories:   repositories,
		Infrastructure: infrastructure,
		Services:       services,
		Handlers:       handlers,
		Middlewares:    middlewares,

		lifecycleCtx:    ctx,
		lifecycleCancel: cancel,
	}, nil
}

// Start запускает все background workers приложения.
//
// Start безопасно вызывается только один раз.
func (a *App) Start() {
	a.lifecycleOnce.Do(func() {

		a.lifecycleWG.Add(3)

		// Health Checker.
		go func() {
			defer a.lifecycleWG.Done()

			a.Infrastructure.HealthChecker.Start(
				a.lifecycleCtx,
			)

			a.Infrastructure.HealthChecker.Wait()
		}()

		// Runtime API Key Cache.
		go func() {
			defer a.lifecycleWG.Done()

			a.Services.AuthRefresher.Start(
				a.lifecycleCtx,
			)

			a.Services.AuthRefresher.Wait()
		}()

		// Runtime Rate Limiter cleanup.
		go func() {
			defer a.lifecycleWG.Done()

			a.Services.RateCleaner.Start(
				a.lifecycleCtx,
			)

			a.Services.RateCleaner.Wait()
		}()
	})
}

// Close корректно завершает работу приложения.
//
// Сначала останавливаются background workers,
// затем закрывается database connection pool.
func (a *App) Close() error {
	if a == nil {
		return nil
	}

	if a.lifecycleCancel != nil {
		a.lifecycleCancel()
	}

	a.lifecycleWG.Wait()

	if a.DB != nil {
		a.DB.Close()
	}

	return nil
}
