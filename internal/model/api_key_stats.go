package model

import "time"

type APIKeyStats struct {
	APIKeyID      int64
	TotalRequests int64
	LastSeenIP    string
	LastSeenAt    time.Time
}
