package prices

import (
	"fmt"
	"log"
	"net"

	pb "github.com/timakaa/historical-common/proto"
	"github.com/timakaa/historical-prices/internal/exchanges"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Server struct {
	pb.UnimplementedPricesServer
	exchangeFactory *exchanges.ExchangeFactory
}

// NewServer creates a new server with the exchange factory
func NewServer() *Server {
	return &Server{
		exchangeFactory: exchanges.NewExchangeFactory(),
	}
}

func (s *Server) GetPrices(req *pb.PricesRequest, stream pb.Prices_GetPricesServer) error {
	log.Printf("Received request for ticker: %s from exchange: %s", req.GetTicker(), req.GetExchange())

	// Get adapter for the specified exchange
	adapter, exists := s.exchangeFactory.GetAdapter(req.GetExchange())
	if !exists {
		return status.Errorf(codes.InvalidArgument, "unsupported exchange: %s", req.GetExchange())
	}

	// Use limit from request or default
	limit := req.GetLimit()
	if limit <= 0 {
		limit = 100 // Default limit
	}

	// Get historical data from the exchange
	prices, err := adapter.GetHistoricalPrices(stream.Context(), req.GetTicker(), limit)
	if err != nil {
		log.Printf("Error getting prices from %s: %v", req.GetExchange(), err)
		return status.Errorf(codes.Internal, "failed to get prices: %v", err)
	}

	// Send data to the stream
	for _, price := range prices {
		if err := stream.Send(price); err != nil {
			return fmt.Errorf("error sending price data: %v", err)
		}
	}

	return nil
}

func Start(port int) error {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return fmt.Errorf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	pb.RegisterPricesServer(s, NewServer())

	log.Printf("Server listening on port %d", port)
	if err := s.Serve(lis); err != nil {
		return fmt.Errorf("failed to serve: %v", err)
	}

	return nil
}
