package auth

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"

	"github.com/timakaa/historical-common/database"
	"github.com/timakaa/historical-common/database/models"
	pb "github.com/timakaa/historical-common/proto"
	"gorm.io/gorm"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Server struct {
	pb.UnimplementedAuthServer
	db *gorm.DB
}

// NewServer creates a new Auth server
func NewServer(db *gorm.DB) *Server {
	return &Server{
		db: db,
	}
}

func (s *Server) ValidateToken(ctx context.Context, req *pb.ValidateRequest) (*pb.ValidateResponse, error) {
	if req.Token == "" {
		return nil, status.Error(codes.InvalidArgument, "token is required")
	}

	// Check if database connection is valid
	if s.db == nil {
		log.Printf("Database connection is nil")
		return nil, status.Error(codes.Internal, "database connection not available")
	}

	// Find token in database
	var token models.Token
	result := s.db.Where("token_string = ?", req.Token).First(&token)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			log.Printf("Token not found: %s", req.Token)
			return &pb.ValidateResponse{
				IsValid: false,
			}, nil
		}
		log.Printf("Error finding token: %v", result.Error)
		return nil, status.Error(codes.Internal, "failed to validate token")
	}

	// Check if token is expired
	if token.IsExpired() {
		log.Printf("Token expired: %s", req.Token)
		return &pb.ValidateResponse{
			IsValid: false,
		}, nil
	}

	log.Printf("Token validated successfully: %s", req.Token)

	// Token is valid
	return &pb.ValidateResponse{
		IsValid:     true,
		UserId:      fmt.Sprintf("user-%d", token.ID), // Using token ID as user ID for simplicity
		Permissions: token.Permissions,
	}, nil
}

func (s *Server) CreateToken(ctx context.Context, req *pb.CreateTokenRequest) (*pb.CreateTokenResponse, error) {
	// Validate request
	if req.ExpiresIn <= 0 {
		return nil, status.Error(codes.InvalidArgument, "expires_in must be positive")
	}

	// Check if database connection is valid
	if s.db == nil {
		log.Printf("Database connection is nil")
		return nil, status.Error(codes.Internal, "database connection not available")
	}

	// Create new token
	token := models.NewToken(req.Permissions, req.ExpiresIn)

	// Ensure permissions are properly serialized
	if err := token.BeforeSave(); err != nil {
		log.Printf("Error serializing permissions: %v", err)
		return nil, status.Error(codes.Internal, "failed to process token data")
	}

	// Save token to database with error handling
	result := s.db.Create(token)
	if result.Error != nil {
		log.Printf("Error creating token: %v", result.Error)
		return nil, status.Error(codes.Internal, "failed to create token")
	}

	log.Printf("Token created successfully: %s", token.TokenString)

	// Return response
	return &pb.CreateTokenResponse{
		Token:     token.TokenString,
		ExpiresAt: token.ExpiresAt.Unix(),
	}, nil
}

func (s *Server) RevokeToken(ctx context.Context, req *pb.RevokeTokenRequest) (*pb.RevokeTokenResponse, error) {
	if req.Token == "" {
		return nil, status.Error(codes.InvalidArgument, "token is required")
	}

	// Check if database connection is valid
	if s.db == nil {
		log.Printf("Database connection is nil")
		return nil, status.Error(codes.Internal, "database connection not available")
	}

	// Delete token from database
	result := s.db.Where("token_string = ?", req.Token).Delete(&models.Token{})
	if result.Error != nil {
		log.Printf("Error revoking token: %v", result.Error)
		return nil, status.Error(codes.Internal, "failed to revoke token")
	}

	// Check if token was found and deleted
	if result.RowsAffected == 0 {
		log.Printf("Token not found for revocation: %s", req.Token)
		return &pb.RevokeTokenResponse{
			Success: false,
		}, nil
	}

	log.Printf("Token revoked successfully: %s", req.Token)
	return &pb.RevokeTokenResponse{
		Success: true,
	}, nil
}

func (s *Server) UpdateTokenCandlesLeft(ctx context.Context, req *pb.UpdateTokenCandlesLeftRequest) (*pb.UpdateTokenCandlesLeftResponse, error) {
	revokeTokenReq := &pb.RevokeTokenRequest{
		Token: req.Token,
	}

	if req.DecreaseCandles < 0 {
		s.RevokeToken(ctx, revokeTokenReq)
		return &pb.UpdateTokenCandlesLeftResponse{CandlesLeft: req.DecreaseCandles}, nil
	}

	// Check if database connection is valid
	if s.db == nil {
		log.Printf("Database connection is nil")
		return nil, status.Error(codes.Internal, "database connection not available")
	}

	// Find token in database
	var token models.Token
	result := s.db.Where("token_string = ?", req.Token).First(&token)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			log.Printf("Token not found: %s", req.Token)
			return nil, status.Error(codes.NotFound, "token not found")
		}
		log.Printf("Error finding token: %v", result.Error)
		return nil, status.Error(codes.Internal, "failed to find token")
	}

	var currentCandlesLeft int64
	err := s.db.Model(&token).Select("candles_left").Scan(&currentCandlesLeft).Error
	if err != nil {
		log.Printf("Error scanning candles_left: %v", err)
		currentCandlesLeft = 0
	}

	newCandlesLeft := max(0, currentCandlesLeft-req.DecreaseCandles)

	if newCandlesLeft <= 0 {
		s.db.Delete(&token)
		return &pb.UpdateTokenCandlesLeftResponse{
			CandlesLeft: newCandlesLeft,
		}, nil
	}

	// Update the value in the database
	result = s.db.Model(&token).Update("candles_left", newCandlesLeft)
	if result.Error != nil {
		log.Printf("Error updating candles_left: %v", result.Error)
		return nil, status.Error(codes.Internal, "failed to update token")
	}

	log.Printf("Token candles_left updated successfully: %s, new value: %d (decreased by %d)",
		req.Token, newCandlesLeft, req.DecreaseCandles)
	return &pb.UpdateTokenCandlesLeftResponse{
		CandlesLeft: newCandlesLeft,
	}, nil
}

// GetTokenInfo retrieves information about a token
func (s *Server) GetTokenInfo(ctx context.Context, req *pb.GetTokenInfoRequest) (*pb.GetTokenInfoResponse, error) {
	if req.Token == "" {
		return nil, status.Error(codes.InvalidArgument, "token is required")
	}

	// Check if database connection is valid
	if s.db == nil {
		log.Printf("Database connection is nil")
		return nil, status.Error(codes.Internal, "database connection not available")
	}

	// Find token in database
	var token models.Token
	result := s.db.Where("token_string = ?", req.Token).First(&token)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			log.Printf("Token not found: %s", req.Token)
			return nil, status.Error(codes.NotFound, "token not found")
		}
		log.Printf("Error finding token: %v", result.Error)
		return nil, status.Error(codes.Internal, "failed to find token")
	}

	// Get the candles_left value
	var candlesLeft int64
	err := s.db.Model(&token).Select("candles_left").Scan(&candlesLeft).Error
	if err != nil {
		log.Printf("Error scanning candles_left: %v", err)
		candlesLeft = 0
	}

	return &pb.GetTokenInfoResponse{
		Token:       token.TokenString,
		CandlesLeft: candlesLeft,
		ExpiresAt:   token.ExpiresAt.Unix(),
		Permissions: token.Permissions,
	}, nil
}

func Start(port int) error {
	// Get database connection using the provider
	db := database.Provider.GetDB()
	if db == nil {
		return fmt.Errorf("failed to get database connection")
	}

	// Auto migrate the tokens table
	if err := db.AutoMigrate(&models.Token{}); err != nil {
		return fmt.Errorf("failed to migrate database: %v", err)
	}
	log.Println("Database migration completed successfully")

	// Set up gRPC server
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return fmt.Errorf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	pb.RegisterAuthServer(s, NewServer(db))

	log.Printf("Auth listening on port %d", port)
	if err := s.Serve(lis); err != nil {
		return fmt.Errorf("failed to serve: %v", err)
	}

	return nil
}
