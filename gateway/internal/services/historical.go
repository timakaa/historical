package services

import (
	"context"

	pb "github.com/timakaa/historical-common/proto"
	"github.com/timakaa/historical-gateway/internal/repositories"
)

type HistoricalService interface {
	GetHistoricalPrices(ctx context.Context, exchange, ticker string) (pb.HistoricalPrices_GetHistoricalPricesClient, error)
	CheckHealth() map[string]interface{}
}

type DefaultHistoricalService struct {
	repo repositories.HistoricalRepository
}

func NewHistoricalService(repo repositories.HistoricalRepository) *DefaultHistoricalService {
	return &DefaultHistoricalService{
		repo: repo,
	}
}

func (s *DefaultHistoricalService) GetHistoricalPrices(ctx context.Context, exchange, ticker string) (pb.HistoricalPrices_GetHistoricalPricesClient, error) {
	return s.repo.GetHistoricalPrices(ctx, exchange, ticker)
}

func (s *DefaultHistoricalService) CheckHealth() map[string]interface{} {
	status := "up"
	message := "OK"

	if !s.repo.CheckHealth() {
		status = "down"
		message = "Service unavailable"
	}

	return map[string]interface{}{
		"status":  status,
		"message": message,
	}
}
