package cache

import "time"

// APIKey представляет запись,
// которая хранится в Runtime Cache.
type APIKey struct {
	ID        int64
	Key       string
	Owner     string
	Enabled   bool
	CreatedAt time.Time
	ExpiresAt *time.Time
}
