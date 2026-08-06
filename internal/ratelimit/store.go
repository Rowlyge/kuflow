package ratelimit

import (
	"sync"
	"time"
)

// Store хранит все Bucket.
type Store struct {
	mu sync.RWMutex

	buckets map[string]*Bucket
}

// NewStore создаёт Store.
func NewStore() *Store {

	return &Store{
		buckets: make(map[string]*Bucket),
	}
}

// Get возвращает Bucket.
func (s *Store) Get(
	key string,
) (*Bucket, bool) {

	s.mu.RLock()
	defer s.mu.RUnlock()

	bucket, ok := s.buckets[key]

	return bucket, ok
}

// Set сохраняет Bucket.
func (s *Store) Set(
	key string,
	bucket *Bucket,
) {

	s.mu.Lock()
	defer s.mu.Unlock()

	s.buckets[key] = bucket
}

// Delete удаляет Bucket.
func (s *Store) Delete(
	key string,
) {

	s.mu.Lock()
	defer s.mu.Unlock()

	delete(
		s.buckets,
		key,
	)
}

// Cleanup удаляет давно неиспользуемые Bucket.
func (s *Store) Cleanup(
	maxIdle time.Duration,
) {

	now := time.Now()

	s.mu.Lock()
	defer s.mu.Unlock()

	for key, bucket := range s.buckets {

		if now.Sub(bucket.LastSeen()) > maxIdle {

			delete(
				s.buckets,
				key,
			)
		}
	}
}
