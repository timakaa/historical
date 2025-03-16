package client

import (
	"context"
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	pb "github.com/timakaa/historical-common/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
)

const bufSize = 1024 * 1024

// mockAuthServer implements the AuthServer interface for testing
type mockAuthServer struct {
	pb.UnimplementedAuthServer
	validateTokenFunc func(context.Context, *pb.ValidateRequest) (*pb.ValidateResponse, error)
	createTokenFunc   func(context.Context, *pb.CreateTokenRequest) (*pb.CreateTokenResponse, error)
	revokeTokenFunc   func(context.Context, *pb.RevokeTokenRequest) (*pb.RevokeTokenResponse, error)
}

func (m *mockAuthServer) ValidateToken(ctx context.Context, req *pb.ValidateRequest) (*pb.ValidateResponse, error) {
	if m.validateTokenFunc != nil {
		return m.validateTokenFunc(ctx, req)
	}
	return &pb.ValidateResponse{IsValid: true, UserId: "test-user", Permissions: []string{"read"}}, nil
}

func (m *mockAuthServer) CreateToken(ctx context.Context, req *pb.CreateTokenRequest) (*pb.CreateTokenResponse, error) {
	if m.createTokenFunc != nil {
		return m.createTokenFunc(ctx, req)
	}
	return &pb.CreateTokenResponse{Token: "test-token", ExpiresAt: time.Now().Add(time.Hour).Unix()}, nil
}

func (m *mockAuthServer) RevokeToken(ctx context.Context, req *pb.RevokeTokenRequest) (*pb.RevokeTokenResponse, error) {
	if m.revokeTokenFunc != nil {
		return m.revokeTokenFunc(ctx, req)
	}
	return &pb.RevokeTokenResponse{Success: true}, nil
}

// setupTest sets up a mock gRPC server and client for testing
func setupTest(t *testing.T, mockServer *mockAuthServer) (*Client, func()) {
	listener := bufconn.Listen(bufSize)
	server := grpc.NewServer()
	pb.RegisterAuthServer(server, mockServer)

	go func() {
		if err := server.Serve(listener); err != nil {
			t.Errorf("Failed to serve: %v", err)
		}
	}()

	// Create a client connection to the mock server
	dialer := func(context.Context, string) (net.Conn, error) {
		return listener.Dial()
	}

	conn, err := grpc.DialContext(
		context.Background(),
		"bufnet",
		grpc.WithContextDialer(dialer),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	require.NoError(t, err)

	client := &Client{
		conn:   conn,
		client: pb.NewAuthClient(conn),
	}

	return client, func() {
		conn.Close()
		server.Stop()
	}
}

func TestNewClient(t *testing.T) {
	// This test is more of an integration test and would require a real server
	// For unit testing, we'll just verify that the function signature is correct
	t.Run("NewClient function exists with correct signature", func(t *testing.T) {
		// Just a placeholder to verify the function exists
		var _ func(string) (*Client, error) = NewClient
	})

	// Test error handling when grpc.NewClient returns an error
	t.Run("error handling when connection fails", func(t *testing.T) {
		// Save the original function to restore it later
		originalNewClient := grpcNewClient
		defer func() { grpcNewClient = originalNewClient }()

		// Replace the function with one that always returns an error
		expectedErr := fmt.Errorf("simulated connection error")
		grpcNewClient = func(target string, opts ...grpc.DialOption) (*grpc.ClientConn, error) {
			return nil, expectedErr
		}

		// Now test NewClient with our mocked function
		client, err := NewClient("test-address")

		// Verify that the error is properly handled and returned
		require.Error(t, err)
		require.Nil(t, client)
		require.Contains(t, err.Error(), "failed to connect")
		require.Contains(t, err.Error(), expectedErr.Error())
	})

	// Note: Full testing of NewClient requires a running gRPC server
	// This is more of an integration test than a unit test
	t.Run("integration test would require", func(t *testing.T) {
		t.Skip("Skipping integration test for NewClient - requires running gRPC server")

		// Example of integration test (not executed):
		// client, err := NewClient("localhost:50051")
		// require.NoError(t, err)
		// require.NotNil(t, client)
		// defer client.Close()

		// Check client functionality
		// resp, err := client.ValidateToken("test-token", "test-service")
		// require.NoError(t, err)
		// require.NotNil(t, resp)
	})

	// Test for unavailable server
	// This test can be executed as it only checks client creation
	// with an unavailable address, which doesn't require a real server
	t.Run("connection to unavailable server", func(t *testing.T) {
		// Use a known unavailable address
		client, err := NewClient("localhost:12345")

		// Client should be created, but will error when used
		// In gRPC, connections are established lazily
		require.NoError(t, err)
		require.NotNil(t, client)

		// Cleanup resources
		defer client.Close()

		// Attempting to use the client would result in an error
		// But we won't check that in this test as it would require
		// waiting for a connection timeout
	})
}

func TestValidateToken(t *testing.T) {
	t.Run("successful validation", func(t *testing.T) {
		mockServer := &mockAuthServer{
			validateTokenFunc: func(ctx context.Context, req *pb.ValidateRequest) (*pb.ValidateResponse, error) {
				assert.Equal(t, "test-token", req.Token)
				assert.Equal(t, "test-service", req.Service)
				return &pb.ValidateResponse{
					IsValid:     true,
					UserId:      "test-user",
					Permissions: []string{"read", "write"},
				}, nil
			},
		}

		client, cleanup := setupTest(t, mockServer)
		defer cleanup()

		resp, err := client.ValidateToken("test-token", "test-service")
		require.NoError(t, err)
		assert.True(t, resp.IsValid)
		assert.Equal(t, "test-user", resp.UserId)
		assert.Equal(t, []string{"read", "write"}, resp.Permissions)
	})

	t.Run("validation error", func(t *testing.T) {
		mockServer := &mockAuthServer{
			validateTokenFunc: func(ctx context.Context, req *pb.ValidateRequest) (*pb.ValidateResponse, error) {
				return nil, status.Error(codes.InvalidArgument, "invalid token")
			},
		}

		client, cleanup := setupTest(t, mockServer)
		defer cleanup()

		resp, err := client.ValidateToken("invalid-token", "test-service")
		require.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "invalid token")
	})
}

func TestCreateToken(t *testing.T) {
	t.Run("successful token creation", func(t *testing.T) {
		mockServer := &mockAuthServer{
			createTokenFunc: func(ctx context.Context, req *pb.CreateTokenRequest) (*pb.CreateTokenResponse, error) {
				assert.Equal(t, []string{"read", "write"}, req.Permissions)
				assert.Equal(t, int64(3600), req.ExpiresIn)
				return &pb.CreateTokenResponse{
					Token:     "new-token",
					ExpiresAt: time.Now().Add(time.Hour).Unix(),
				}, nil
			},
		}

		client, cleanup := setupTest(t, mockServer)
		defer cleanup()

		resp, err := client.CreateToken("user-123", []string{"read", "write"}, 3600)
		require.NoError(t, err)
		assert.Equal(t, "new-token", resp.Token)
		assert.Greater(t, resp.ExpiresAt, time.Now().Unix())
	})

	t.Run("token creation error", func(t *testing.T) {
		mockServer := &mockAuthServer{
			createTokenFunc: func(ctx context.Context, req *pb.CreateTokenRequest) (*pb.CreateTokenResponse, error) {
				return nil, status.Error(codes.Internal, "failed to create token")
			},
		}

		client, cleanup := setupTest(t, mockServer)
		defer cleanup()

		resp, err := client.CreateToken("user-123", []string{"read"}, 3600)
		require.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "failed to create token")
	})
}

func TestRevokeToken(t *testing.T) {
	t.Run("successful token revocation", func(t *testing.T) {
		mockServer := &mockAuthServer{
			revokeTokenFunc: func(ctx context.Context, req *pb.RevokeTokenRequest) (*pb.RevokeTokenResponse, error) {
				assert.Equal(t, "test-token", req.Token)
				return &pb.RevokeTokenResponse{
					Success: true,
				}, nil
			},
		}

		client, cleanup := setupTest(t, mockServer)
		defer cleanup()

		resp, err := client.RevokeToken("test-token")
		require.NoError(t, err)
		assert.True(t, resp.Success)
	})

	t.Run("token revocation error", func(t *testing.T) {
		mockServer := &mockAuthServer{
			revokeTokenFunc: func(ctx context.Context, req *pb.RevokeTokenRequest) (*pb.RevokeTokenResponse, error) {
				return nil, status.Error(codes.NotFound, "token not found")
			},
		}

		client, cleanup := setupTest(t, mockServer)
		defer cleanup()

		resp, err := client.RevokeToken("non-existent-token")
		require.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "token not found")
	})
}

func TestClose(t *testing.T) {
	mockServer := &mockAuthServer{}
	client, cleanup := setupTest(t, mockServer)

	err := client.Close()
	require.NoError(t, err)

	// Cleanup should still work even after we've closed the client
	cleanup()
}

func TestContextCancellation(t *testing.T) {
	t.Run("context cancellation in ValidateToken", func(t *testing.T) {
		// Create channels for synchronization
		blockCh := make(chan struct{})
		doneCh := make(chan struct{})

		// Create a mock server that will block until receiving a signal
		mockServer := &mockAuthServer{
			validateTokenFunc: func(ctx context.Context, req *pb.ValidateRequest) (*pb.ValidateResponse, error) {
				// Check if context was cancelled
				select {
				case <-ctx.Done():
					// Context was cancelled
					close(doneCh)
					return nil, ctx.Err()
				case <-blockCh:
					// Received unblock signal
					close(doneCh)
					return &pb.ValidateResponse{IsValid: true}, nil
				}
			},
		}

		// Setup client
		client, cleanup := setupTest(t, mockServer)
		defer cleanup()

		// Run method call in a separate goroutine
		errCh := make(chan error)
		respCh := make(chan *pb.ValidateResponse)
		go func() {
			resp, err := client.ValidateToken("test-token", "test-service")
			respCh <- resp
			errCh <- err
		}()

		// Give some time for the request to start
		time.Sleep(10 * time.Millisecond)

		// Close the client, which should lead to context cancellation
		client.Close()

		// Check the result
		select {
		case <-doneCh:
			// Server finished processing the request
			resp := <-respCh
			err := <-errCh
			require.Error(t, err)
			require.Nil(t, resp)
		case <-time.After(100 * time.Millisecond):
			// Test timeout
			t.Fatal("Test timed out waiting for context cancellation")
			// Unblock the server anyway
			close(blockCh)
		}
	})
}
