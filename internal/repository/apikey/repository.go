package apikey

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrNotFound = errors.New("api key not found")
)

// Repository описывает работу с API-ключами.
type Repository interface {
	FindByKey(
		ctx context.Context,
		key string,
	) (*APIKey, error)

	Create(
		ctx context.Context,
		key *APIKey,
	) error

	Disable(
		ctx context.Context,
		id int64,
	) error

	Delete(
		ctx context.Context,
		id int64,
	) error
}

type repository struct {
	db *pgxpool.Pool
}

// New создаёт Repository.
func New(
	db *pgxpool.Pool,
) Repository {

	return &repository{
		db: db,
	}
}

// FindByKey ищет API-ключ.
func (r *repository) FindByKey(
	ctx context.Context,
	key string,
) (*APIKey, error) {

	var apiKey APIKey

	err := r.db.QueryRow(
		ctx,
		queryFindByKey,
		key,
	).Scan(
		&apiKey.ID,
		&apiKey.APIKey,
		&apiKey.Owner,
		&apiKey.Enabled,
		&apiKey.CreatedAt,
		&apiKey.ExpiresAt,
	)

	if err != nil {
		return nil, ErrNotFound
	}

	return &apiKey, nil
}

// Create создаёт новый ключ.
func (r *repository) Create(
	ctx context.Context,
	key *APIKey,
) error {

	return r.db.QueryRow(
		ctx,
		queryCreate,
		key.APIKey,
		key.Owner,
		key.Enabled,
		key.ExpiresAt,
	).Scan(
		&key.ID,
	)
}

// Disable отключает ключ.
func (r *repository) Disable(
	ctx context.Context,
	id int64,
) error {

	_, err := r.db.Exec(
		ctx,
		queryDisable,
		id,
	)

	return err
}

// Delete удаляет ключ.
func (r *repository) Delete(
	ctx context.Context,
	id int64,
) error {

	_, err := r.db.Exec(
		ctx,
		queryDelete,
		id,
	)

	return err
}
