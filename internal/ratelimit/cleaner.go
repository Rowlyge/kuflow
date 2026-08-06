package ratelimit

import (
	"context"
	"time"
)

// Cleaner периодически очищает
// старые Bucket.
type Cleaner struct {
	store *Store

	interval time.Duration

	maxIdle time.Duration
}

// NewCleaner создаёт Cleaner.
func NewCleaner(
	store *Store,
	interval time.Duration,
	maxIdle time.Duration,
) *Cleaner {

	return &Cleaner{

		store: store,

		interval: interval,

		maxIdle: maxIdle,
	}
}

// Start запускает очистку.
func (c *Cleaner) Start(
	ctx context.Context,
) {

	ticker := time.NewTicker(
		c.interval,
	)

	go func() {

		defer ticker.Stop()

		for {

			select {

			case <-ctx.Done():
				return

			case <-ticker.C:

				c.store.Cleanup(
					c.maxIdle,
				)
			}
		}
	}()
}
