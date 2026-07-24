package model

import "time"

type Request struct {
	ID        int64
	Method    string
	URL       string
	ClientIP  string
	UserAgent string
	CreatedAt time.Time
}
