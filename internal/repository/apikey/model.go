package apikey

import "time"

// APIKey описывает запись из таблицы api_keys.
type APIKey struct {
	ID        int64
	APIKey    string
	Owner     string
	Enabled   bool
	CreatedAt time.Time
	ExpiresAt *time.Time
}
