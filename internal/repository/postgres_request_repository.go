package repository

import (
	"context"

	"github.com/Rowlyge/kuflow/internal/model"

	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresRequestRepository struct {
	db *pgxpool.Pool
}

func NewPostgresRequestRepository(
	db *pgxpool.Pool,
) *PostgresRequestRepository {

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
		INSERT INTO requests (
			method,
			path,
			status_code,
			duration_ms,
			response_size,
			client_ip,
			user_agent,
			upstream,
			created_at
		)
		VALUES (
			$1,$2,$3,$4,$5,$6,$7,$8,$9
		)
		`,
		request.Method,
		request.Path,
		request.StatusCode,
		request.Duration.Milliseconds(),
		request.ResponseSize,
		request.ClientIP,
		request.UserAgent,
		request.Upstream,
		request.CreatedAt,
	)

	return err
}
