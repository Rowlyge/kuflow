package auth

import "context"

// Service предоставляет сервис авторизации.
type Service struct {
	validator *Validator
}

// New создаёт сервис авторизации.
func New(
	validator *Validator,
) *Service {

	return &Service{
		validator: validator,
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
