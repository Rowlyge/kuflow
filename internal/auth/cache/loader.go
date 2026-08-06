package cache

import (
	"context"

	apikeyrepo "github.com/Rowlyge/kuflow/internal/repository/apikey"
)

// Loader отвечает за загрузку
// API-ключей из PostgreSQL
// в Runtime Cache.
type Loader struct {
	repository apikeyrepo.Repository
	cache      *Cache
}

// NewLoader создаёт Loader.
func NewLoader(
	repository apikeyrepo.Repository,
	cache *Cache,
) *Loader {

	return &Loader{
		repository: repository,
		cache:      cache,
	}
}

// Load перечитывает все API-ключи
// из базы данных и полностью
// обновляет Runtime Cache.
func (l *Loader) Load(
	ctx context.Context,
) error {

	keys, err := l.repository.List(ctx)
	if err != nil {
		return err
	}

	data := make(map[string]APIKey, len(keys))

	for _, key := range keys {

		data[key.APIKey] = APIKey{
			ID:        key.ID,
			Key:       key.APIKey,
			Owner:     key.Owner,
			Enabled:   key.Enabled,
			CreatedAt: key.CreatedAt,
			ExpiresAt: key.ExpiresAt,
		}
	}

	l.cache.Replace(data)

	return nil
}
