package auth

import (
	"context"

	authcache "github.com/Rowlyge/kuflow/internal/auth/cache"
)

type Service struct {
	validator *Validator
}

func New(
	validator *Validator,
) *Service {

	return &Service{
		validator: validator,
	}
}

func (s *Service) Validate(
	ctx context.Context,
	apiKey string,
) (*authcache.APIKey, error) {

	return s.validator.Validate(
		ctx,
		apiKey,
	)
}
