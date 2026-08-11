//go:build integration

package cache

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/Rowlyge/kuflow/internal/config"
	"github.com/Rowlyge/kuflow/internal/database"
	apikeyrepo "github.com/Rowlyge/kuflow/internal/repository/apikey"
)

func TestRefresherReloadIntegration(t *testing.T) {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		t.Fatal("DATABASE_URL is not set")
	}

	ctx := context.Background()

	db, err := database.New(
		config.DatabaseConfig{
			URL: databaseURL,
		},
	)
	if err != nil {
		t.Fatalf("connect to PostgreSQL: %v", err)
	}
	defer db.Close()

	repository := apikeyrepo.New(db)

	runtimeCache := New()

	loader := NewLoader(
		repository,
		runtimeCache,
	)

	refresher := NewRefresher(
		loader,
		10*time.Second,
	)

	const (
		keyV1 = "integration-key-v1"
		keyV2 = "integration-key-v2"
		owner = "refresher-integration-test"
	)

	// Очистка перед тестом.
	_, err = db.Exec(
		ctx,
		`DELETE FROM api_keys WHERE owner = $1`,
		owner,
	)
	if err != nil {
		t.Fatalf("cleanup before test: %v", err)
	}

	// Очистка после теста.
	t.Cleanup(func() {
		_, _ = db.Exec(
			context.Background(),
			`DELETE FROM api_keys WHERE owner = $1`,
			owner,
		)
	})

	// 1. Создаём первоначальный ключ в PostgreSQL.
	_, err = db.Exec(
		ctx,
		`
		INSERT INTO api_keys (
			api_key,
			owner,
			enabled
		)
		VALUES ($1, $2, TRUE)
		`,
		keyV1,
		owner,
	)
	if err != nil {
		t.Fatalf("insert key v1: %v", err)
	}

	// 2. Загружаем его в Runtime Cache.
	if err := refresher.Reload(ctx); err != nil {
		t.Fatalf("initial reload: %v", err)
	}

	if _, ok := runtimeCache.Get(keyV1); !ok {
		t.Fatalf("key v1 was not loaded into runtime cache")
	}

	// 3. Меняем состояние PostgreSQL.
	_, err = db.Exec(
		ctx,
		`DELETE FROM api_keys WHERE api_key = $1`,
		keyV1,
	)
	if err != nil {
		t.Fatalf("delete key v1: %v", err)
	}

	_, err = db.Exec(
		ctx,
		`
		INSERT INTO api_keys (
			api_key,
			owner,
			enabled
		)
		VALUES ($1, $2, TRUE)
		`,
		keyV2,
		owner,
	)
	if err != nil {
		t.Fatalf("insert key v2: %v", err)
	}

	// 4. Немедленный Reload.
	//
	// Главное требование Dynamic Cache Reload:
	// не ждём periodic interval в 10 секунд.
	start := time.Now()

	if err := refresher.Reload(ctx); err != nil {
		t.Fatalf("reload after database change: %v", err)
	}

	elapsed := time.Since(start)

	if elapsed >= 10*time.Second {
		t.Fatalf(
			"Reload took too long: %v; expected immediate reload",
			elapsed,
		)
	}

	// 5. Старый ключ должен исчезнуть.
	if _, ok := runtimeCache.Get(keyV1); ok {
		t.Fatalf(
			"old key v1 is still present in runtime cache",
		)
	}

	// 6. Новый ключ должен появиться.
	if _, ok := runtimeCache.Get(keyV2); !ok {
		t.Fatalf(
			"new key v2 was not loaded into runtime cache",
		)
	}
}
