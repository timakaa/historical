package services

import (
	"context"

	"github.com/timakaa/historical-gateway/internal/repositories"
)

type AuthService interface {
	ValidateToken(ctx context.Context, token string) (string, []string, error)
	CheckHealth() map[string]interface{}
}

type DefaultAuthService struct {
	repo repositories.AccessRepository
}

func NewAuthService(repo repositories.AccessRepository) *DefaultAuthService {
	return &DefaultAuthService{
		repo: repo,
	}
}

func (s *DefaultAuthService) ValidateToken(ctx context.Context, token string) (string, []string, error) {
	resp, err := s.repo.ValidateToken(ctx, token, "gateway")
	if err != nil {
		return "", nil, err
	}

	return resp.UserId, resp.Permissions, nil
}

func (s *DefaultAuthService) CheckHealth() map[string]interface{} {
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
