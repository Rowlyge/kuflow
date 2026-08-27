package app

import (
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Rowlyge/kuflow/internal/repository"
	apikeyrepo "github.com/Rowlyge/kuflow/internal/repository/apikey"
)

type Repositories struct {
	Request     repository.RequestRepository
	APIKey      apikeyrepo.Repository
	APIKeyStats repository.APIKeyStatsRepository
}

func NewRepositories(
	db *pgxpool.Pool,
) (*Repositories, error) {

	return &Repositories{
		Request:     repository.NewPostgresRequestRepository(db),
		APIKey:      apikeyrepo.New(db),
		APIKeyStats: repository.NewPostgresAPIKeyStatsRepository(db),
	}, nil
}
