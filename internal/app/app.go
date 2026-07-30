package app

import (
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
}

// New создаёт приложение и инициализирует все зависимости.
func New(cfg *config.Config) (*App, error) {

	db, err := database.New(cfg.Database)
	if err != nil {
		return nil, err
	}

	repositories, err := NewRepositories(db)
	if err != nil {
		return nil, err
	}

	infrastructure, err := NewInfrastructure(
		cfg.Proxy.Upstreams,
	)
	if err != nil {
		return nil, err
	}

	services, err := NewServices(
		cfg,
		repositories,
		infrastructure,
	)
	if err != nil {
		return nil, err
	}

	handlers, err := NewHandlers(services)
	if err != nil {
		return nil, err
	}

	middlewares, err := NewMiddlewares(services)
	if err != nil {
		return nil, err
	}

	// Запускаем Health Checker.
	infrastructure.HealthChecker.Start()

	return &App{
		Config:         cfg,
		DB:             db,
		Repositories:   repositories,
		Infrastructure: infrastructure,
		Services:       services,
		Handlers:       handlers,
		Middlewares:    middlewares,
	}, nil
}

// Close освобождает все ресурсы приложения.
func (a *App) Close() {

	// Останавливаем фоновые задачи.
	a.Infrastructure.HealthChecker.Stop()

	a.DB.Close()
}
