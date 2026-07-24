package service

import (
	"context"

	"github.com/Rowlyge/kuflow/internal/model"
	"github.com/Rowlyge/kuflow/internal/repository"
)

type TelemetryService struct {
	requests repository.RequestRepository
}

func NewTelemetryService(
	requests repository.RequestRepository,
) *TelemetryService {

	return &TelemetryService{
		requests: requests,
	}
}

func (s *TelemetryService) Save(
	ctx context.Context,
	request *model.Request,
) error {

	return s.requests.Create(ctx, request)
}
