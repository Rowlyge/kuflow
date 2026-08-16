package ratelimit

import (
	"context"
	"time"
)

// Cleaner периодически очищает старые Bucket.
type Cleaner struct {
	store *Store

	interval time.Duration

	maxIdle time.Duration

	done chan struct{}
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

		done: make(chan struct{}),
	}
}

// Start запускает очистку.
func (c *Cleaner) Start(
	ctx context.Context,
) {

	go func() {
		defer close(c.done)

		ticker := time.NewTicker(
			c.interval,
		)
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

// Wait блокируется до завершения Cleaner.
func (c *Cleaner) Wait() {
	<-c.done
}
