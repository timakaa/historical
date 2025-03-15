package prices

import (
	"fmt"
	"log"
	"net"

	pb "github.com/timakaa/historical-common/proto"

	"google.golang.org/grpc"
)

type Server struct {
	pb.UnimplementedPricesServer
}

func (s *Server) GetPrices(req *pb.PricesRequest, stream pb.Prices_GetPricesServer) error {
	log.Printf("Received request for ticker: %s from exchange: %s", req.GetTicker(), req.GetExchange())

	// TODO: Create a real business logic
	for i := range [3]int{} {
		price := &pb.PricesResponse{
			Date:   fmt.Sprintf("2024-03-%02d", i+1),
			Open:   100.0 + float64(i),
			High:   101.0 + float64(i),
			Low:    99.0 + float64(i),
			Close:  100.5 + float64(i),
			Volume: 1000.0 + float64(i),
		}

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
	pb.RegisterPricesServer(s, &Server{})

	log.Printf("Server listening on port %d", port)
	if err := s.Serve(lis); err != nil {
		return fmt.Errorf("failed to serve: %v", err)
	}

	return nil
}
