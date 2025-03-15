package repositories

import (
	"context"

	pb "github.com/timakaa/historical-common/proto"
)

type HistoricalRepository interface {
	GetHistoricalPrices(ctx context.Context, exchange, ticker string) (pb.HistoricalPrices_GetHistoricalPricesClient, error)
	CheckHealth() bool
}

type GrpcHistoricalRepository struct {
	client pb.HistoricalPricesClient
}

func NewHistoricalRepository(client pb.HistoricalPricesClient) *GrpcHistoricalRepository {
	return &GrpcHistoricalRepository{
		client: client,
	}
}

func (r *GrpcHistoricalRepository) GetHistoricalPrices(ctx context.Context, exchange, ticker string) (pb.HistoricalPrices_GetHistoricalPricesClient, error) {
	return r.client.GetHistoricalPrices(ctx, &pb.HistoricalPricesRequest{
		Exchange: exchange,
		Ticker:   ticker,
	})
}

func (r *GrpcHistoricalRepository) CheckHealth() bool {
	// Implementation of health check
	return true
}
