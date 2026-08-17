package app

import (
	"context"
	"fmt"
	"sync"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Rowlyge/kuflow/internal/config"
	"github.com/Rowlyge/kuflow/internal/database"
)

// lifecycleWorker определяет контракт background worker.
//
// App управляет запуском и завершением worker,
// а сам worker отвечает за собственный внутренний lifecycle.
type lifecycleWorker interface {
	Start(context.Context)
	Wait()
}

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
	lifecycleCtx     context.Context
	lifecycleCancel  context.CancelFunc
	lifecycleOnce    sync.Once
	lifecycleWG      sync.WaitGroup
	lifecycleWorkers []lifecycleWorker
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

	lifecycleWorkers := []lifecycleWorker{
		infrastructure.HealthChecker,
		services.AuthRefresher,
		services.RateCleaner,
	}

	return &App{
		Config: cfg,

		DB: db,

		Repositories:   repositories,
		Infrastructure: infrastructure,
		Services:       services,
		Handlers:       handlers,
		Middlewares:    middlewares,

		lifecycleCtx:     ctx,
		lifecycleCancel:  cancel,
		lifecycleWorkers: lifecycleWorkers,
	}, nil
}

// Start запускает все background workers приложения.
//
// Start безопасно вызывается только один раз.
func (a *App) Start() {
	a.lifecycleOnce.Do(func() {
		a.lifecycleWG.Add(
			len(a.lifecycleWorkers),
		)

		for _, worker := range a.lifecycleWorkers {
			go func(w lifecycleWorker) {
				defer a.lifecycleWG.Done()

				w.Start(a.lifecycleCtx)
				w.Wait()
			}(worker)
		}
	})
}

// Close корректно завершает работу приложения.
//
// Сначала отменяется lifecycle context,
// затем App ожидает завершения всех background workers.
// Если workers не завершились до истечения ctx,
// возвращается ошибка и database pool не закрывается.
//
// Database pool закрывается только после успешного
// завершения всех background workers.
func (a *App) Close(
	ctx context.Context,
) error {

	if a == nil {
		return nil
	}

	if ctx == nil {
		return fmt.Errorf(
			"application shutdown context is nil",
		)
	}

	if a.lifecycleCancel != nil {
		a.lifecycleCancel()
	}

	workersDone := make(chan struct{})

	go func() {
		a.lifecycleWG.Wait()

		close(workersDone)
	}()

	select {

	case <-workersDone:

		if a.DB != nil {
			a.DB.Close()
		}

		return nil

	case <-ctx.Done():

		return fmt.Errorf(
			"application shutdown timeout: %w",
			ctx.Err(),
		)
	}
}
