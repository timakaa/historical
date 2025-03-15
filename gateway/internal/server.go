package gateway

import (
	"fmt"
	"log"
	"net/http"

	pb "github.com/timakaa/historical-common/proto"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Server struct {
	router           *gin.Engine
	historicalClient pb.HistoricalPricesClient
	accessClient     pb.AccessManagerClient
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

	router := gin.Default()

	router.SetTrustedProxies([]string{"127.0.0.1", "::1"})

	server := &Server{
		router:           router,
		historicalClient: pb.NewHistoricalPricesClient(historicalConn),
		accessClient:     pb.NewAccessManagerClient(accessConn),
	}

	// Setup routes
	server.setupRoutes()

	return server, nil
}

func (s *Server) setupRoutes() {
	// Health check endpoint
	s.router.GET("/health", s.handleHealth)

	// API endpoints
	api := s.router.Group("/api/v1")
	api.Use(s.authMiddleware())
	{
		api.GET("/prices/:exchange/:ticker", s.handleGetHistoricalPrices)
	}
}

func (s *Server) authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("Authorization")
		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "missing authorization header"})
			c.Abort()
			return
		}

		// Validate token with Access Manager
		resp, err := s.accessClient.ValidateToken(c.Request.Context(), &pb.ValidateRequest{
			Token:   token,
			Service: "gateway",
		})
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			c.Abort()
			return
		}

		// Store user info in context
		c.Set("userID", resp.UserId)
		c.Set("permissions", resp.Permissions)
		c.Next()
	}
}

func (s *Server) handleHealth(c *gin.Context) {
	// Check health of all services
	historicalHealth := s.checkHistoricalHealth()
	accessHealth := s.checkAccessHealth()

	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
		"services": gin.H{
			"historical": historicalHealth,
			"access":     accessHealth,
		},
	})
}

func (s *Server) handleGetHistoricalPrices(c *gin.Context) {
	exchange := c.Param("exchange")
	ticker := c.Param("ticker")

	// Get stream from historical service
	stream, err := s.historicalClient.GetHistoricalPrices(c.Request.Context(), &pb.HistoricalPricesRequest{
		Exchange: exchange,
		Ticker:   ticker,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get prices"})
		return
	}

	// Set up SSE headers
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("Transfer-Encoding", "chunked")

	// Stream prices to client
	for {
		price, err := stream.Recv()
		if err != nil {
			if err.Error() == "EOF" {
				break
			}
			log.Printf("Error receiving price: %v", err)
			break
		}

		// Send price as SSE
		c.SSEvent("price", price)
		c.Writer.Flush()
	}
}

func (s *Server) checkHistoricalHealth() gin.H {
	// TODO: Implement actual health check
	return gin.H{
		"status":  "up",
		"message": "OK",
	}
}

func (s *Server) checkAccessHealth() gin.H {
	// TODO: Implement actual health check
	return gin.H{
		"status":  "up",
		"message": "OK",
	}
}

func (s *Server) Start(port int) error {
	return s.router.Run(fmt.Sprintf(":%d", port))
}
