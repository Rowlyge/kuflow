package cache

import (
	"context"
	"errors"
	"testing"
	"time"

	apikeyrepo "github.com/Rowlyge/kuflow/internal/repository/apikey"
)

type mockAPIKeyRepository struct {
	keys []apikeyrepo.APIKey
	err  error
}

func (m *mockAPIKeyRepository) FindByKey(
	ctx context.Context,
	key string,
) (*apikeyrepo.APIKey, error) {
	return nil, errors.New("not implemented")
}

func (m *mockAPIKeyRepository) List(
	ctx context.Context,
) ([]apikeyrepo.APIKey, error) {
	if m.err != nil {
		return nil, m.err
	}

	return m.keys, nil
}

func (m *mockAPIKeyRepository) Create(
	ctx context.Context,
	key *apikeyrepo.APIKey,
) error {
	return errors.New("not implemented")
}

func (m *mockAPIKeyRepository) Disable(
	ctx context.Context,
	id int64,
) error {
	return errors.New("not implemented")
}

func (m *mockAPIKeyRepository) Delete(
	ctx context.Context,
	id int64,
) error {
	return errors.New("not implemented")
}

func TestRefresherReload(t *testing.T) {
	ctx := context.Background()

	repository := &mockAPIKeyRepository{
		keys: []apikeyrepo.APIKey{
			{
				ID:      1,
				APIKey:  "old-key",
				Owner:   "test-user",
				Enabled: true,
			},
		},
	}

	runtimeCache := New()

	loader := NewLoader(
		repository,
		runtimeCache,
	)

	refresher := NewRefresher(
		loader,
		10*time.Second,
	)

	// Первоначальная загрузка.
	if err := refresher.Reload(ctx); err != nil {
		t.Fatalf("initial Reload() failed: %v", err)
	}

	if _, ok := runtimeCache.Get("old-key"); !ok {
		t.Fatal("old-key was not loaded into runtime cache")
	}

	// Имитируем изменение API-ключа в PostgreSQL.
	repository.keys = []apikeyrepo.APIKey{
		{
			ID:      2,
			APIKey:  "new-key",
			Owner:   "test-user",
			Enabled: true,
		},
	}

	// Dynamic Cache Reload должен применить изменение
	// немедленно, без ожидания interval.
	if err := refresher.Reload(ctx); err != nil {
		t.Fatalf("dynamic Reload() failed: %v", err)
	}

	if _, ok := runtimeCache.Get("new-key"); !ok {
		t.Fatal("new-key was not loaded into runtime cache")
	}

	// Старый cache должен быть полностью заменён.
	if _, ok := runtimeCache.Get("old-key"); ok {
		t.Fatal("old-key is still present after cache reload")
	}
}

func TestRefresherReloadErrorKeepsCache(t *testing.T) {
	ctx := context.Background()

	repository := &mockAPIKeyRepository{
		keys: []apikeyrepo.APIKey{
			{
				ID:      1,
				APIKey:  "stable-key",
				Owner:   "test-user",
				Enabled: true,
			},
		},
	}

	runtimeCache := New()

	loader := NewLoader(
		repository,
		runtimeCache,
	)

	refresher := NewRefresher(
		loader,
		10*time.Second,
	)

	// Загружаем первоначальное состояние.
	if err := refresher.Reload(ctx); err != nil {
		t.Fatalf("initial Reload() failed: %v", err)
	}

	if _, ok := runtimeCache.Get("stable-key"); !ok {
		t.Fatal("stable-key was not loaded into runtime cache")
	}

	// Имитируем ошибку PostgreSQL.
	expectedErr := errors.New("database unavailable")
	repository.err = expectedErr

	// Reload должен вернуть ошибку.
	err := refresher.Reload(ctx)
	if !errors.Is(err, expectedErr) {
		t.Fatalf(
			"expected Reload() error %v, got %v",
			expectedErr,
			err,
		)
	}

	// Старый cache должен остаться рабочим.
	if _, ok := runtimeCache.Get("stable-key"); !ok {
		t.Fatal("stable-key was removed after failed reload")
	}

	// Проверяем, что cache не был очищен.
	if got := runtimeCache.Size(); got != 1 {
		t.Fatalf(
			"expected cache size 1 after failed reload, got %d",
			got,
		)
	}
}
