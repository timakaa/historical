package client

import (
	"context"
	"fmt"
	"io"
	"time"

	pb "github.com/timakaa/historical-common/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Client is a gRPC client for the Greeter service
type Client struct {
	conn   *grpc.ClientConn
	client pb.PricesClient
}

// NewClient creates a new gRPC client connected to the specified address
func NewClient(address string) (*Client, error) {
	// Set up a connection to the server with insecure credentials (no TLS)
	conn, err := grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect: %v", err)
	}

	client := pb.NewPricesClient(conn)
	return &Client{
		conn:   conn,
		client: client,
	}, nil
}

// SayHello sends a greeting to the server
func (c *Client) GetPrices(exchange string, ticker string) ([]*pb.PricesResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	// Create request
	req := &pb.PricesRequest{
		Exchange: exchange,
		Ticker:   ticker,
	}

	// Get stream
	stream, err := c.client.GetPrices(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("error getting stream: %v", err)
	}

	// Slice to store all prices
	var prices []*pb.PricesResponse

	// Read from stream
	for {
		price, err := stream.Recv()
		if err != nil {
			if err == io.EOF {
				// End of stream, return collected prices
				return prices, nil
			}
			return nil, fmt.Errorf("error receiving price: %v", err)
		}

		// Append each price to our slice
		prices = append(prices, price)
	}
}

// Close closes the client connection
func (c *Client) Close() error {
	return c.conn.Close()
}
