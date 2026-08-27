package apikey

import (
	"context"

	"github.com/Rowlyge/kuflow/internal/model"
	"github.com/Rowlyge/kuflow/internal/repository"
)

type StatsService struct {
	stats repository.APIKeyStatsRepository
}

func NewStatsService(
	stats repository.APIKeyStatsRepository,
) *StatsService {
	return &StatsService{
		stats: stats,
	}
}

func (s *StatsService) GetByAPIKeyID(
	ctx context.Context,
	apiKeyID int64,
) (*model.APIKeyStats, error) {
	return s.stats.GetByAPIKeyID(
		ctx,
		apiKeyID,
	)
}

func (s *StatsService) List(
	ctx context.Context,
) ([]model.APIKeyStats, error) {
	return s.stats.List(ctx)
}
