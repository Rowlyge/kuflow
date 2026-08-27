package repository

import (
	"context"

	"github.com/Rowlyge/kuflow/internal/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresAPIKeyStatsRepository struct {
	db *pgxpool.Pool
}

func NewPostgresAPIKeyStatsRepository(
	db *pgxpool.Pool,
) *PostgresAPIKeyStatsRepository {

	return &PostgresAPIKeyStatsRepository{
		db: db,
	}
}

func (r *PostgresAPIKeyStatsRepository) UpdateUsage(
	ctx context.Context,
	apiKeyID int64,
	ip string,
) error {

	const query = `
INSERT INTO api_key_stats (
	api_key_id,
	requests_total,
	last_used_at,
	last_seen_ip
)
VALUES (
	$1,
	1,
	NOW(),
	$2
)
ON CONFLICT (api_key_id)
DO UPDATE SET
	requests_total = api_key_stats.requests_total + 1,
	last_used_at = NOW(),
	last_seen_ip = EXCLUDED.last_seen_ip;
`

	_, err := r.db.Exec(
		ctx,
		query,
		apiKeyID,
		ip,
	)

	return err
}

func (r *PostgresAPIKeyStatsRepository) GetByAPIKeyID(
	ctx context.Context,
	apiKeyID int64,
) (*model.APIKeyStats, error) {

	const query = `
SELECT
    api_key_id,
    requests_total,
    last_seen_ip,
    last_used_at
FROM api_key_stats
WHERE api_key_id = $1;
`

	stats := &model.APIKeyStats{}

	err := r.db.QueryRow(
		ctx,
		query,
		apiKeyID,
	).Scan(
		&stats.APIKeyID,
		&stats.TotalRequests,
		&stats.LastSeenIP,
		&stats.LastSeenAt,
	)

	if err != nil {

		if err == pgx.ErrNoRows {
			return nil, nil
		}

		return nil, err
	}

	return stats, nil
}

func (r *PostgresAPIKeyStatsRepository) List(
	ctx context.Context,
) ([]model.APIKeyStats, error) {

	const query = `
SELECT
    api_key_id,
    requests_total,
    last_seen_ip,
    last_used_at
FROM api_key_stats
ORDER BY requests_total DESC;
`

	rows, err := r.db.Query(
		ctx,
		query,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []model.APIKeyStats

	for rows.Next() {

		var stats model.APIKeyStats

		err := rows.Scan(
			&stats.APIKeyID,
			&stats.TotalRequests,
			&stats.LastSeenIP,
			&stats.LastSeenAt,
		)
		if err != nil {
			return nil, err
		}

		result = append(
			result,
			stats,
		)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return result, nil
}
