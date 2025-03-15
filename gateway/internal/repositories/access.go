package repositories

import (
	"context"

	pb "github.com/timakaa/historical-common/proto"
)

type AccessRepository interface {
	ValidateToken(ctx context.Context, token, service string) (*pb.ValidateResponse, error)
	CheckHealth() bool
}

type GrpcAccessRepository struct {
	client pb.AccessManagerClient
}

func NewAccessRepository(client pb.AccessManagerClient) *GrpcAccessRepository {
	return &GrpcAccessRepository{
		client: client,
	}
}

func (r *GrpcAccessRepository) ValidateToken(ctx context.Context, token, service string) (*pb.ValidateResponse, error) {
	return r.client.ValidateToken(ctx, &pb.ValidateRequest{
		Token:   token,
		Service: service,
	})
}

func (r *GrpcAccessRepository) CheckHealth() bool {
	// Implementation of health check
	return true
}
