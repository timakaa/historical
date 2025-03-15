package gateway

import (
	"fmt"

	pb "github.com/timakaa/historical-common/proto"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/timakaa/historical-gateway/internal/handlers"
	"github.com/timakaa/historical-gateway/internal/middleware"
	"github.com/timakaa/historical-gateway/internal/repositories"
	"github.com/timakaa/historical-gateway/internal/services"
)

type Server struct {
	router *gin.Engine
}

func NewServer(historicalAddr, accessAddr string) (*Server, error) {
	// Connect to other services
	historicalConn, err := grpc.NewClient(historicalAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to historical service: %v", err)
	}

	accessConn, err := grpc.NewClient(accessAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to access service: %v", err)
	}

	// Create repositories
	historicalRepo := repositories.NewHistoricalRepository(pb.NewHistoricalPricesClient(historicalConn))
	accessRepo := repositories.NewAccessRepository(pb.NewAccessManagerClient(accessConn))

	// Create services
	historicalService := services.NewHistoricalService(historicalRepo)
	authService := services.NewAuthService(accessRepo)

	// Create handlers
	healthHandler := handlers.NewHealthHandler(historicalService, authService)
	pricesHandler := handlers.NewPricesHandler(historicalService)

	// Create middleware
	authMiddleware := middleware.NewAuthMiddleware(authService)

	router := gin.Default()
	router.SetTrustedProxies([]string{"127.0.0.1", "::1"})

	server := &Server{
		router: router,
	}

	// Setup routes
	server.setupRoutes(healthHandler, pricesHandler, authMiddleware)

	return server, nil
}

func (s *Server) setupRoutes(
	healthHandler *handlers.HealthHandler,
	pricesHandler *handlers.PricesHandler,
	authMiddleware *middleware.AuthMiddleware,
) {
	// Health check endpoint
	s.router.GET("/health", healthHandler.Handle)

	// API endpoints
	api := s.router.Group("/api/v1")
	api.Use(authMiddleware.Authenticate())
	{
		api.GET("/prices/:exchange/:ticker", pricesHandler.HandleGetHistoricalPrices)
	}
}

func (s *Server) Start(port int) error {
	return s.router.Run(fmt.Sprintf(":%d", port))
}
