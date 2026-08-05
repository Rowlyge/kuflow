package auth

import (
	"context"
	"strings"
	"time"

	apikeyrepo "github.com/Rowlyge/kuflow/internal/repository/apikey"
)

// Validator выполняет проверку API-ключей.
type Validator struct {
	repository apikeyrepo.Repository
}

// NewValidator создаёт Validator.
func NewValidator(
	repository apikeyrepo.Repository,
) *Validator {

	return &Validator{
		repository: repository,
	}
}

// Validate проверяет корректность API Key.
func (v *Validator) Validate(
	ctx context.Context,
	apiKey string,
) error {

	apiKey = strings.TrimSpace(apiKey)

	if apiKey == "" {
		return ErrMissingAPIKey
	}

	key, err := v.repository.FindByKey(
		ctx,
		apiKey,
	)
	if err != nil {
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
