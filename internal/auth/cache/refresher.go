package cache

import (
	"context"
	"log"
	"time"
)

// Refresher периодически обновляет Runtime Cache
// из PostgreSQL.
type Refresher struct {
	loader *Loader

	interval time.Duration
}

// NewRefresher создаёт Refresher.
func NewRefresher(
	loader *Loader,
	interval time.Duration,
) *Refresher {

	if interval <= 0 {
		interval = 10 * time.Second
	}

	return &Refresher{
		loader:   loader,
		interval: interval,
	}
}

// Reload немедленно обновляет Runtime Cache
// из PostgreSQL.
//
// Метод блокирует вызывающий поток до завершения
// загрузки или отмены контекста.
func (r *Refresher) Reload(
	ctx context.Context,
) error {

	if err := r.loader.Load(ctx); err != nil {
		return err
	}

	return nil
}

// Start запускает цикл обновления.
//
// Метод не блокирует вызывающий поток.
func (r *Refresher) Start(
	ctx context.Context,
) {

	go func() {

		// Первичная загрузка сразу после запуска.
		if err := r.Reload(ctx); err != nil {

			log.Printf(
				"[AuthCache] initial load failed: %v",
				err,
			)

		} else {

			log.Printf(
				"[AuthCache] initial load completed",
			)
		}

		ticker := time.NewTicker(r.interval)
		defer ticker.Stop()

		for {

			select {

			case <-ctx.Done():

				log.Printf(
					"[AuthCache] refresher stopped",
				)

				return

			case <-ticker.C:

				if err := r.Reload(ctx); err != nil {

					log.Printf(
						"[AuthCache] refresh failed: %v",
						err,
					)

					continue
				}

				log.Printf(
					"[AuthCache] cache refreshed",
				)
			}
		}
	}()
}
