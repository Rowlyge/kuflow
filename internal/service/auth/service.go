package auth

import (
	"context"

	apikeyrepo "github.com/Rowlyge/kuflow/internal/repository/apikey"
)

// Service представляет сервис авторизации.
type Service struct {
	validator *Validator
}

// New создаёт AuthService.
func New(
	repository apikeyrepo.Repository,
) *Service {

	return &Service{
		validator: NewValidator(repository),
	}
}

// Validate проверяет API Key.
func (s *Service) Validate(
	ctx context.Context,
	apiKey string,
) error {

	return s.validator.Validate(
		ctx,
		apiKey,
	)
}
