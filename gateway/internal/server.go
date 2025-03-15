package gateway

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/timakaa/historical-common/proto"
	"github.com/timakaa/historical-gateway/internal/handlers"
	"github.com/timakaa/historical-gateway/internal/middleware"
)

type Server struct {
	router *gin.Engine
}

func NewServer(historicalAddr, accessAddr string) (*Server, error) {
	// Connect to other services
	pricesConn, err := grpc.NewClient(historicalAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to historical service: %v", err)
	}

	authConn, err := grpc.NewClient(accessAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to access service: %v", err)
	}

	pricesClient := proto.NewPricesClient(pricesConn)
	authClient := proto.NewAuthClient(authConn)

	// Create handlers
	healthHandler := handlers.NewHealthHandler(pricesClient, authClient)
	pricesHandler := handlers.NewPricesHandler(pricesClient, authClient)
	authHandler := handlers.NewAuthHandler(authClient)

	// Create middleware
	authMiddleware := middleware.NewAuthMiddleware(authClient)

	router := gin.Default()
	router.SetTrustedProxies([]string{"127.0.0.1", "::1"})

	server := &Server{
		router: router,
	}

	// Setup routes
	server.setupRoutes(healthHandler, pricesHandler, authHandler, authMiddleware)

	return server, nil
}

func (s *Server) setupRoutes(
	healthHandler *handlers.HealthHandler,
	pricesHandler *handlers.PricesHandler,
	authHandler *handlers.AuthHandler,
	authMiddleware *middleware.AuthMiddleware,
) {
	// Health check endpoint
	healthHandler.RegisterRoutes(s.router)

	// API endpoints
	api := s.router.Group("/api/v1")
	{
		pricesHandler.RegisterRoutes(api, authMiddleware.Authenticate())
		authHandler.RegisterRoutes(api)
	}
}

func (s *Server) Start(port int) error {
	return s.router.Run(fmt.Sprintf(":%d", port))
}
