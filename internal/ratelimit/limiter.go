package ratelimit

// Limiter управляет всеми Token Bucket.
type Limiter struct {
	store *Store

	config Config
}

// New создаёт новый Rate Limiter.
func New(
	cfg Config,
) *Limiter {

	return &Limiter{

		store: NewStore(),

		config: cfg,
	}
}

// Allow проверяет,
// разрешён ли запрос.
func (l *Limiter) Allow(
	key string,
) Decision {

	bucket, ok := l.store.Get(key)

	if !ok {

		bucket = NewBucket(
			l.config,
		)

		l.store.Set(
			key,
			bucket,
		)
	}

	return bucket.Allow()
}

// Store возвращает внутренний Store.
func (l *Limiter) Store() *Store {
	return l.store
}
