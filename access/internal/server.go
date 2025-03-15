package access

import (
	"context"
	"fmt"
	"log"
	"net"

	pb "github.com/timakaa/historical-common/proto"

	"google.golang.org/grpc"
)

type Server struct {
	pb.UnimplementedAccessManagerServer
}

func (s *Server) ValidateToken(ctx context.Context, req *pb.ValidateRequest) (*pb.ValidateResponse, error) {
	// TODO: Implement token validation
	return &pb.ValidateResponse{
		IsValid: true,
		UserId: "test-user",
		Permissions: []string{"read:prices"},
	}, nil
}

func (s *Server) CreateToken(ctx context.Context, req *pb.CreateTokenRequest) (*pb.CreateTokenResponse, error) {
	// TODO: Implement token creation
	return &pb.CreateTokenResponse{
		Token: "test-token",
		ExpiresAt: 1234567890,
	}, nil
}

func (s *Server) RevokeToken(ctx context.Context, req *pb.RevokeTokenRequest) (*pb.RevokeTokenResponse, error) {
	// TODO: Implement token revocation
	return &pb.RevokeTokenResponse{
		Success: true,
	}, nil
}

func Start(port int) error {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return fmt.Errorf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	pb.RegisterAccessManagerServer(s, &Server{})

	log.Printf("Access Manager listening on port %d", port)
	if err := s.Serve(lis); err != nil {
		return fmt.Errorf("failed to serve: %v", err)
	}

	return nil
} 