package ratelimit

import (
	"context"
	"time"
)

// Cleanup запускает периодическую очистку
// неактивных Bucket.
func (l *Limiter) Cleanup(
	ctx context.Context,
	interval time.Duration,
	maxIdle time.Duration,
) {

	ticker := time.NewTicker(interval)

	go func() {

		defer ticker.Stop()

		for {

			select {

			case <-ctx.Done():
				return

			case <-ticker.C:

				l.store.Cleanup(
					maxIdle,
				)
			}
		}
	}()
}
