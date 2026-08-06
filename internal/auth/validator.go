package auth

import (
	"context"
	"errors"
	"strings"

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
		return ErrInvalidAPIKey
	}

	return nil
}
