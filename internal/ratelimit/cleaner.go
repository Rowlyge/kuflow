package ratelimit

import (
	"context"
	"sync"
	"time"
)

// Cleaner периодически очищает
// старые Bucket.
type Cleaner struct {
	store *Store

	interval time.Duration
	maxIdle  time.Duration

	wg sync.WaitGroup
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
		maxIdle:  maxIdle,
	}
}

// Start запускает периодическую очистку.
//
// Start не блокирует вызывающий поток.
//
// Жизненным циклом контекста владеет вызывающий код,
// в частности App. Cleaner только наблюдает за ctx.Done().
func (c *Cleaner) Start(
	ctx context.Context,
) {

	c.wg.Add(1)

	go func() {
		defer c.wg.Done()

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

// Wait блокируется до завершения
// goroutine Cleaner.
func (c *Cleaner) Wait() {
	c.wg.Wait()
}
