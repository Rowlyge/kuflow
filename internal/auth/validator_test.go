package auth

import (
	"context"
	"testing"
	"time"

	authcache "github.com/Rowlyge/kuflow/internal/auth/cache"
)

func TestValidator_Validate(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name      string
		apiKey    string
		keys      map[string]authcache.APIKey
		wantError error
	}{
		{
			name:      "missing api key",
			apiKey:    "",
			keys:      map[string]authcache.APIKey{},
			wantError: ErrMissingAPIKey,
		},
		{
			name:      "unknown api key",
			apiKey:    "unknown-key",
			keys:      map[string]authcache.APIKey{},
			wantError: ErrInvalidAPIKey,
		},
		{
			name:   "valid enabled api key",
			apiKey: "valid-key",
			keys: map[string]authcache.APIKey{
				"valid-key": {
					ID:        1,
					Key:       "valid-key",
					Owner:     "test-user",
					Enabled:   true,
					CreatedAt: now,
				},
			},
			wantError: nil,
		},
		{
			name:   "disabled api key",
			apiKey: "disabled-key",
			keys: map[string]authcache.APIKey{
				"disabled-key": {
					ID:        2,
					Key:       "disabled-key",
					Owner:     "test-user",
					Enabled:   false,
					CreatedAt: now,
				},
			},
			wantError: ErrInvalidAPIKey,
		},
		{
			name:   "api key with whitespace",
			apiKey: "  valid-key  ",
			keys: map[string]authcache.APIKey{
				"valid-key": {
					ID:        3,
					Key:       "valid-key",
					Owner:     "test-user",
					Enabled:   true,
					CreatedAt: now,
				},
			},
			wantError: nil,
		},
		{
			name:   "api key without expiration",
			apiKey: "no-expiration-key",
			keys: map[string]authcache.APIKey{
				"no-expiration-key": {
					ID:        4,
					Key:       "no-expiration-key",
					Owner:     "test-user",
					Enabled:   true,
					CreatedAt: now,
					ExpiresAt: nil,
				},
			},
			wantError: nil,
		},
		{
			name:   "api key with future expiration",
			apiKey: "future-key",
			keys: map[string]authcache.APIKey{
				"future-key": {
					ID:        5,
					Key:       "future-key",
					Owner:     "test-user",
					Enabled:   true,
					CreatedAt: now,
					ExpiresAt: func() *time.Time {
						t := now.Add(1 * time.Hour)
						return &t
					}(),
				},
			},
			wantError: nil,
		},
		{
			name:   "expired api key",
			apiKey: "expired-key",
			keys: map[string]authcache.APIKey{
				"expired-key": {
					ID:        6,
					Key:       "expired-key",
					Owner:     "test-user",
					Enabled:   true,
					CreatedAt: now.Add(-2 * time.Hour),
					ExpiresAt: func() *time.Time {
						t := now.Add(-1 * time.Minute)
						return &t
					}(),
				},
			},
			wantError: ErrInvalidAPIKey,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cache := authcache.New()
			cache.Replace(tt.keys)

			validator := NewValidator(cache)

			err := validator.Validate(
				context.Background(),
				tt.apiKey,
			)

			if err != tt.wantError {
				t.Fatalf(
					"expected error %v, got %v",
					tt.wantError,
					err,
				)
			}
		})
	}
}
