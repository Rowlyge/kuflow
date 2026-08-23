package apikey

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
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

	List(
		ctx context.Context,
	) ([]APIKey, error)

	ListEnabled(
		ctx context.Context,
	) ([]APIKey, error)

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

	if errors.Is(
		err,
		pgx.ErrNoRows,
	) {
		return nil, ErrNotFound
	}

	if err != nil {
		return nil, err
	}

	return &apiKey, nil
}

// List возвращает все API-ключи.
func (r *repository) List(
	ctx context.Context,
) ([]APIKey, error) {

	rows, err := r.db.Query(
		ctx,
		queryList,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	keys := make([]APIKey, 0)

	for rows.Next() {

		var key APIKey

		err := rows.Scan(
			&key.ID,
			&key.APIKey,
			&key.Owner,
			&key.Enabled,
			&key.CreatedAt,
			&key.ExpiresAt,
		)
		if err != nil {
			return nil, err
		}

		keys = append(
			keys,
			key,
		)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return keys, nil
}

// ListEnabled возвращает только активные API-ключи.
func (r *repository) ListEnabled(
	ctx context.Context,
) ([]APIKey, error) {

	rows, err := r.db.Query(
		ctx,
		queryListEnabled,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	keys := make([]APIKey, 0)

	for rows.Next() {

		var key APIKey

		err := rows.Scan(
			&key.ID,
			&key.APIKey,
			&key.Owner,
			&key.Enabled,
			&key.CreatedAt,
			&key.ExpiresAt,
		)
		if err != nil {
			return nil, err
		}

		keys = append(
			keys,
			key,
		)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return keys, nil
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
