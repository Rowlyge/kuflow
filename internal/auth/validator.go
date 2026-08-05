package auth

import (
	"context"
	"errors"
	"strings"

	apikeyrepo "github.com/Rowlyge/kuflow/internal/repository/apikey"
)

var (
	ErrMissingAPIKey = errors.New("missing api key")
	ErrInvalidAPIKey = errors.New("invalid api key")
)

type Validator struct {
	repository apikeyrepo.Repository
}

func NewValidator(
	repository apikeyrepo.Repository,
) *Validator {

	return &Validator{
		repository: repository,
	}
}

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
		return ErrInvalidAPIKey
	}

	return nil
}
