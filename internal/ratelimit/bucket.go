package ratelimit

import (
	"math"
	"sync"
	"time"
)

// Decision содержит результат проверки Rate Limiter.
type Decision struct {
	Allowed bool

	Limit int

	Remaining int

	RetryAfter time.Duration
}

// Bucket реализует алгоритм Token Bucket.
type Bucket struct {
	mu sync.Mutex

	capacity int

	tokens float64

	refillTokens int

	refillInterval time.Duration

	lastRefill time.Time

	lastSeen time.Time
}

// NewBucket создаёт новый Bucket.
func NewBucket(
	cfg Config,
) *Bucket {

	now := time.Now()

	return &Bucket{

		capacity: cfg.Capacity,

		tokens: float64(cfg.Capacity),

		refillTokens: cfg.RefillTokens,

		refillInterval: cfg.RefillInterval,

		lastRefill: now,

		lastSeen: now,
	}
}

// Allow пытается выдать токен.
func (b *Bucket) Allow() Decision {

	b.mu.Lock()
	defer b.mu.Unlock()

	now := time.Now()

	b.refill(now)

	b.lastSeen = now

	decision := Decision{

		Limit: b.capacity,
	}

	if b.tokens < 1 {

		decision.Allowed = false

		decision.Remaining = 0

		missing := 1 - b.tokens

		seconds := missing *
			b.refillInterval.Seconds() /
			float64(b.refillTokens)

		decision.RetryAfter =
			time.Duration(
				math.Ceil(seconds),
			) * time.Second

		return decision
	}

	b.tokens--

	decision.Allowed = true

	decision.Remaining = int(math.Floor(b.tokens))

	return decision
}

// LastSeen возвращает время последнего обращения.
func (b *Bucket) LastSeen() time.Time {

	b.mu.Lock()
	defer b.mu.Unlock()

	return b.lastSeen
}

// refill восстанавливает токены.
func (b *Bucket) refill(
	now time.Time,
) {

	elapsed := now.Sub(b.lastRefill)

	if elapsed <= 0 {
		return
	}

	intervals :=
		float64(elapsed) /
			float64(b.refillInterval)

	if intervals <= 0 {
		return
	}

	b.tokens +=
		intervals *
			float64(b.refillTokens)

	if b.tokens > float64(b.capacity) {
		b.tokens = float64(b.capacity)
	}

	b.lastRefill = now
}
