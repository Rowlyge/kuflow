package app

import (
	"github.com/Rowlyge/kuflow/internal/config"
	"github.com/Rowlyge/kuflow/internal/database"
	"github.com/Rowlyge/kuflow/internal/service"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Services struct {
	Health *service.HealthService
}

type App struct {
	Config   *config.Config
	DB       *pgxpool.Pool
	Services *Services
}

func New(cfg *config.Config) (*App, error) {

	db, err := database.New(cfg.Database)
	if err != nil {
		return nil, err
	}

	services := &Services{
		Health: service.NewHealthService(),
	}

	return &App{
		Config:   cfg,
		DB:       db,
		Services: services,
	}, nil
}

func (a *App) Close() {
	if a.DB != nil {
		a.DB.Close()
	}
}
