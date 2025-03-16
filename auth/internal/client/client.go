package client

import (
	"context"
	"fmt"
	"time"

	pb "github.com/timakaa/historical-common/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// For testing purposes
var grpcNewClient = grpc.NewClient

type Client struct {
	conn   *grpc.ClientConn
	client pb.AuthClient
}

func NewClient(address string) (*Client, error) {
	conn, err := grpcNewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect: %v", err)
	}

	client := pb.NewAuthClient(conn)
	return &Client{
		conn:   conn,
		client: client,
	}, nil
}

func (c *Client) ValidateToken(token, service string) (*pb.ValidateResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	return c.client.ValidateToken(ctx, &pb.ValidateRequest{
		Token:   token,
		Service: service,
	})
}

func (c *Client) CreateToken(userID string, permissions []string, expiresIn int64) (*pb.CreateTokenResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	return c.client.CreateToken(ctx, &pb.CreateTokenRequest{
		Permissions: permissions,
		ExpiresIn:   expiresIn,
	})
}

func (c *Client) RevokeToken(token string) (*pb.RevokeTokenResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	return c.client.RevokeToken(ctx, &pb.RevokeTokenRequest{
		Token: token,
	})
}

func (c *Client) Close() error {
	return c.conn.Close()
}
