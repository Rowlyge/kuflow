package cache

import (
	"context"

	apikeyrepo "github.com/Rowlyge/kuflow/internal/repository/apikey"
)

type Loader struct {
	repository apikeyrepo.Repository
	cache      *Cache
}

func NewLoader(
	repository apikeyrepo.Repository,
	cache *Cache,
) *Loader {

	return &Loader{
		repository: repository,
		cache:      cache,
	}
}

// Load читает активные ключи из PostgreSQL
// и полностью заменяет Runtime Cache.
func (l *Loader) Load(
	ctx context.Context,
) error {

	keys, err := l.repository.ListEnabled(
		ctx,
	)
	if err != nil {
		return err
	}

	data := make(
		map[string]APIKey,
		len(keys),
	)

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

	l.cache.Replace(
		data,
	)

	return nil
}
