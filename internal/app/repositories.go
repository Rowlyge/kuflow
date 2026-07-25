package app

import (
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Rowlyge/kuflow/internal/repository"
)

// Repositories объединяет все репозитории приложения.
type Repositories struct {
	Request repository.RequestRepository
}

// NewRepositories создаёт все репозитории приложения.
func NewRepositories(
	db *pgxpool.Pool,
) (*Repositories, error) {

	return &Repositories{

		Request: repository.NewPostgresRequestRepository(db),
	}, nil
}
