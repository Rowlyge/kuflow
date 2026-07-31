package database

import (
	"context"

	"github.com/Rowlyge/kuflow/internal/config"

	"github.com/jackc/pgx/v5/pgxpool"
)

// New создаёт подключение к PostgreSQL.
func New(
	cfg config.DatabaseConfig,
) (*pgxpool.Pool, error) {

	pool, err := pgxpool.New(
		context.Background(),
		cfg.URL,
	)
	if err != nil {
		return nil, err
	}

	if err := pool.Ping(
		context.Background(),
	); err != nil {

		pool.Close()
		return nil, err
	}

	return pool, nil
}
