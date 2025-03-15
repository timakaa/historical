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

type Client struct {
	conn   *grpc.ClientConn
	client pb.GatewayClient
}

func NewClient(address string) (*Client, error) {
	conn, err := grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect: %v", err)
	}

	client := pb.NewGatewayClient(conn)
	return &Client{
		conn:   conn,
		client: client,
	}, nil
}

func (c *Client) GetHistoricalPrices(exchange, ticker string) ([]*pb.HistoricalPricesResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	stream, err := c.client.GetHistoricalPrices(ctx, &pb.HistoricalPricesRequest{
		Exchange: exchange,
		Ticker:   ticker,
	})
	if err != nil {
		return nil, fmt.Errorf("error getting stream: %v", err)
	}

	var prices []*pb.HistoricalPricesResponse
	for {
		price, err := stream.Recv()
		if err != nil {
			if err == io.EOF {
				return prices, nil
			}
			return nil, fmt.Errorf("error receiving price: %v", err)
		}
		prices = append(prices, price)
	}
}

func (c *Client) Health() (*pb.HealthResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	return c.client.Health(ctx, &pb.HealthRequest{})
}

func (c *Client) Close() error {
	return c.conn.Close()
} 