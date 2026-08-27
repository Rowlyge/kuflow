package repository

import (
	"context"

	"github.com/Rowlyge/kuflow/internal/model"
)

type APIKeyStatsRepository interface {
	UpdateUsage(
		ctx context.Context,
		apiKeyID int64,
		ip string,
	) error

	GetByAPIKeyID(
		ctx context.Context,
		apiKeyID int64,
	) (*model.APIKeyStats, error)

	List(
		ctx context.Context,
	) ([]model.APIKeyStats, error)
}
