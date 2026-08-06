package service

import (
	"context"

	"github.com/Rowlyge/kuflow/internal/metrics"
	"github.com/Rowlyge/kuflow/internal/model"
	"github.com/Rowlyge/kuflow/internal/repository"
)

// TelemetryService отвечает за сохранение телеметрии
// и обновление runtime-метрик.
type TelemetryService struct {
	requests repository.RequestRepository

	collector *metrics.Collector
}

// NewTelemetryService создаёт сервис телеметрии.
func NewTelemetryService(
	requests repository.RequestRepository,
	collector *metrics.Collector,
) *TelemetryService {

	return &TelemetryService{
		requests:  requests,
		collector: collector,
	}
}

// Save сохраняет информацию о запросе
// и обновляет runtime-метрики.
func (s *TelemetryService) Save(
	ctx context.Context,
	request *model.Request,
) error {

	err := s.requests.Create(ctx, request)

	// HTTP-метрики.
	s.collector.IncRequests()

	if request.StatusCode >= 400 {
		s.collector.IncFailed()
	} else {
		s.collector.IncSuccess()
	}

	s.collector.AddBytesOut(
		uint64(request.ResponseSize),
	)

	s.collector.ObserveDuration(
		request.Duration,
	)

	// Метрики самой телеметрии.
	if err != nil {
		s.collector.IncTelemetryFailed()
		return err
	}

	s.collector.IncTelemetrySaved()

	return nil
}

// RecordRateLimit сохраняет runtime-метрику
// превышения лимита запросов.
func (s *TelemetryService) RecordRateLimit() {

	s.collector.IncRateLimited()
}
