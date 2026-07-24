package database

import (
	"context"
	"fmt"

	"github.com/Rowlyge/kuflow/internal/config"

	"github.com/jackc/pgx/v5/pgxpool"
)

func New(cfg config.DatabaseConfig) (*pgxpool.Pool, error) {

	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host,
		cfg.Port,
		cfg.User,
		cfg.Password,
		cfg.Name,
		cfg.SSLMode,
	)

	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		return nil, err
	}

	if err := pool.Ping(context.Background()); err != nil {
		pool.Close()
		return nil, err
	}

	return pool, nil
}
