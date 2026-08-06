package auth

import (
	"context"
	"strings"
	"time"

	authcache "github.com/Rowlyge/kuflow/internal/auth/cache"
)

// Validator выполняет проверку API-ключей.
type Validator struct {
	cache *authcache.Cache
}

// NewValidator создаёт Validator.
func NewValidator(
	cache *authcache.Cache,
) *Validator {

	return &Validator{
		cache: cache,
	}
}

// Validate проверяет корректность API Key.
func (v *Validator) Validate(
	ctx context.Context,
	apiKey string,
) error {

	_ = ctx

	apiKey = strings.TrimSpace(apiKey)

	if apiKey == "" {
		return ErrMissingAPIKey
	}

	key, ok := v.cache.Get(apiKey)
	if !ok {
		return ErrInvalidAPIKey
	}

	if !key.Enabled {
		return ErrDisabledAPIKey
	}

	if key.ExpiresAt != nil &&
		time.Now().After(*key.ExpiresAt) {

		return ErrExpiredAPIKey
	}

	return nil
}
