package repository

import (
	"context"

	"github.com/Rowlyge/kuflow/internal/model"

	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresRequestRepository struct {
	db *pgxpool.Pool
}

func NewPostgresRequestRepository(db *pgxpool.Pool) *PostgresRequestRepository {
	return &PostgresRequestRepository{
		db: db,
	}
}

func (r *PostgresRequestRepository) Create(
	ctx context.Context,
	request *model.Request,
) error {

	_, err := r.db.Exec(
		ctx,
		`
		INSERT INTO requests
		(method, url, client_ip, user_agent)
		VALUES ($1, $2, $3, $4)
		`,
		request.Method,
		request.URL,
		request.ClientIP,
		request.UserAgent,
	)

	return err
}
