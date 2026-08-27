package auth

import (
	"context"
	"errors"
	"strings"
	"time"

	authcache "github.com/Rowlyge/kuflow/internal/auth/cache"
)

var (
	ErrMissingAPIKey = errors.New("missing api key")
	ErrInvalidAPIKey = errors.New("invalid api key")
)

type Validator struct {
	cache *authcache.Cache
}

func NewValidator(
	cache *authcache.Cache,
) *Validator {

	return &Validator{
		cache: cache,
	}
}

func (v *Validator) Validate(
	ctx context.Context,
	apiKey string,
) (*authcache.APIKey, error) {
	_ = ctx

	apiKey = strings.TrimSpace(apiKey)

	if apiKey == "" {
		return nil, ErrMissingAPIKey
	}

	key, ok := v.cache.Get(apiKey)
	if !ok {
		return nil, ErrInvalidAPIKey
	}

	if !key.Enabled {
		return nil, ErrInvalidAPIKey
	}

	// Если срок действия задан и уже истёк,
	// ключ считается недействительным.
	if key.ExpiresAt != nil &&
		!time.Now().Before(*key.ExpiresAt) {

		return nil, ErrInvalidAPIKey
	}

	return &key, nil
}
