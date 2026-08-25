package apikey

import (
	"context"

	apikeyrepo "github.com/Rowlyge/kuflow/internal/repository/apikey"
)

type Service struct {
	repository apikeyrepo.Repository
}

func New(
	repository apikeyrepo.Repository,
) *Service {

	return &Service{
		repository: repository,
	}
}

func (s *Service) List(
	ctx context.Context,
) ([]apikeyrepo.APIKey, error) {

	return s.repository.List(ctx)
}

func (s *Service) ListEnabled(
	ctx context.Context,
) ([]apikeyrepo.APIKey, error) {

	return s.repository.ListEnabled(ctx)
}

func (s *Service) FindByKey(
	ctx context.Context,
	key string,
) (*apikeyrepo.APIKey, error) {

	return s.repository.FindByKey(
		ctx,
		key,
	)
}

func (s *Service) Disable(
	ctx context.Context,
	id int64,
) error {

	return s.repository.Disable(
		ctx,
		id,
	)
}

func (s *Service) Delete(
	ctx context.Context,
	id int64,
) error {

	return s.repository.Delete(
		ctx,
		id,
	)
}

func (s *Service) Create(
	ctx context.Context,
	key *apikeyrepo.APIKey,
) error {

	return s.repository.Create(
		ctx,
		key,
	)
}
