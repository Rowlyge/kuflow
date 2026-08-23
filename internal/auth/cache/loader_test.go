package cache

import (
	"context"
	"errors"
	"testing"

	apikeyrepo "github.com/Rowlyge/kuflow/internal/repository/apikey"
)

type mockLoaderRepository struct {
	keys []apikeyrepo.APIKey
	err  error
}

func (m *mockLoaderRepository) FindByKey(
	ctx context.Context,
	key string,
) (*apikeyrepo.APIKey, error) {
	return nil, errors.New("not implemented")
}

func (m *mockLoaderRepository) List(
	ctx context.Context,
) ([]apikeyrepo.APIKey, error) {
	return nil, errors.New("not implemented")
}

func (m *mockLoaderRepository) ListEnabled(
	ctx context.Context,
) ([]apikeyrepo.APIKey, error) {

	if m.err != nil {
		return nil, m.err
	}

	return m.keys, nil
}

func (m *mockLoaderRepository) Create(
	ctx context.Context,
	key *apikeyrepo.APIKey,
) error {
	return errors.New("not implemented")
}

func (m *mockLoaderRepository) Disable(
	ctx context.Context,
	id int64,
) error {
	return errors.New("not implemented")
}

func (m *mockLoaderRepository) Delete(
	ctx context.Context,
	id int64,
) error {
	return errors.New("not implemented")
}

func TestLoaderLoad(t *testing.T) {

	repository := &mockLoaderRepository{
		keys: []apikeyrepo.APIKey{
			{
				APIKey:  "key-1",
				Owner:   "user-1",
				Enabled: true,
			},
			{
				APIKey:  "key-2",
				Owner:   "user-2",
				Enabled: true,
			},
		},
	}

	cache := New()

	loader := NewLoader(
		repository,
		cache,
	)

	err := loader.Load(
		context.Background(),
	)
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}

	if cache.Size() != 2 {
		t.Fatalf(
			"expected cache size 2, got %d",
			cache.Size(),
		)
	}
}
