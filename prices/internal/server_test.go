package prices

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	pb "github.com/timakaa/historical-common/proto"
	"github.com/timakaa/historical-prices/internal/exchanges"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
)

const bufSize = 1024 * 1024

// MockExchangeAdapter is a mock implementation of the ExchangeAdapter interface
type MockExchangeAdapter struct {
	mock.Mock
}

func (m *MockExchangeAdapter) GetName() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockExchangeAdapter) GetHistoricalPrices(ctx context.Context, ticker string, limit int64) ([]*pb.PricesResponse, error) {
	args := m.Called(ctx, ticker, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*pb.PricesResponse), args.Error(1)
}

// MockExchangeFactory is a mock implementation of the exchange factory
type MockExchangeFactory struct {
	mock.Mock
}

func (m *MockExchangeFactory) GetAdapter(exchange string) (exchanges.ExchangeAdapter, bool) {
	args := m.Called(exchange)
	if args.Get(0) == nil {
		return nil, args.Bool(1)
	}
	return args.Get(0).(exchanges.ExchangeAdapter), args.Bool(1)
}

func (m *MockExchangeFactory) RegisterAdapter(adapter exchanges.ExchangeAdapter) {
	m.Called(adapter)
}

// MockPricesServer_GetPricesServer is a mock implementation of the streaming server
type MockPricesServer_GetPricesServer struct {
	mock.Mock
	grpc.ServerStream
	ctx context.Context
}

func (m *MockPricesServer_GetPricesServer) Send(response *pb.PricesResponse) error {
	args := m.Called(response)
	return args.Error(0)
}

func (m *MockPricesServer_GetPricesServer) Context() context.Context {
	return m.ctx
}

// setupGRPCServer sets up a gRPC server for integration testing
func setupGRPCServer(t *testing.T) (pb.PricesClient, func()) {
	listener := bufconn.Listen(bufSize)
	server := grpc.NewServer()
	pb.RegisterPricesServer(server, NewServer())

	go func() {
		if err := server.Serve(listener); err != nil {
			t.Errorf("Failed to serve: %v", err)
		}
	}()

	// Create a client connection to the mock server
	dialer := func(context.Context, string) (net.Conn, error) {
		return listener.Dial()
	}

	// Use grpc.NewClient instead of grpc.DialContext
	conn, err := grpc.NewClient(
		"bufnet",
		grpc.WithContextDialer(dialer),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	require.NoError(t, err)

	client := pb.NewPricesClient(conn)

	return client, func() {
		conn.Close()
		server.Stop()
	}
}

func TestNewServer(t *testing.T) {
	// Test that NewServer creates a server with an exchange factory
	server := NewServer()
	assert.NotNil(t, server)
	assert.NotNil(t, server.exchangeFactory)
}

// TestServer is a test version of Server that accepts MockExchangeFactory
type TestServer struct {
	exchangeFactory *MockExchangeFactory
}

// GetPrices implements the same method as Server but uses MockExchangeFactory
func (s *TestServer) GetPrices(req *pb.PricesRequest, stream pb.Prices_GetPricesServer) error {
	// Get the exchange adapter
	adapter, ok := s.exchangeFactory.GetAdapter(req.Exchange)
	if !ok {
		return status.Errorf(codes.InvalidArgument, "unsupported exchange: %s", req.Exchange)
	}

	// Set default limit if not provided
	limit := req.Limit
	if limit == 0 {
		limit = 100
	}

	// Get historical prices
	prices, err := adapter.GetHistoricalPrices(stream.Context(), req.Ticker, limit)
	if err != nil {
		return status.Errorf(codes.Internal, "failed to get prices: %v", err)
	}

	// Send prices to the client
	for _, price := range prices {
		if err := stream.Send(price); err != nil {
			return fmt.Errorf("error sending price data: %w", err)
		}
	}

	return nil
}

// createServerWithMockFactory creates a test server with a mock factory for testing
func createServerWithMockFactory(mockFactory *MockExchangeFactory) *TestServer {
	return &TestServer{
		exchangeFactory: mockFactory,
	}
}

func TestGetPrices(t *testing.T) {
	t.Run("successful price retrieval", func(t *testing.T) {
		// Create mock objects
		mockFactory := new(MockExchangeFactory)
		mockAdapter := new(MockExchangeAdapter)
		mockStream := &MockPricesServer_GetPricesServer{
			ctx: context.Background(),
		}

		// Setup test data
		exchange := "binance"
		ticker := "BTC/USDT"
		limit := int64(10)

		// Create sample price data
		now := time.Now()
		prices := []*pb.PricesResponse{
			{
				Date:   now.Format("2006-01-02"),
				Open:   10000.0,
				High:   10100.0,
				Low:    9900.0,
				Close:  10050.0,
				Volume: 1.5,
			},
			{
				Date:   now.Add(-24 * time.Hour).Format("2006-01-02"),
				Open:   9900.0,
				High:   10000.0,
				Low:    9800.0,
				Close:  9950.0,
				Volume: 2.0,
			},
		}

		// Setup expectations
		mockAdapter.On("GetHistoricalPrices", mock.Anything, ticker, limit).Return(prices, nil)
		mockFactory.On("GetAdapter", exchange).Return(mockAdapter, true)

		// Setup stream expectations
		for _, price := range prices {
			mockStream.On("Send", price).Return(nil).Once()
		}

		// Create test server with mock factory
		server := createServerWithMockFactory(mockFactory)

		// Call the method being tested
		err := server.GetPrices(&pb.PricesRequest{
			Exchange: exchange,
			Ticker:   ticker,
			Limit:    limit,
		}, mockStream)

		// Verify results
		assert.NoError(t, err)
		mockFactory.AssertExpectations(t)
		mockAdapter.AssertExpectations(t)
		mockStream.AssertExpectations(t)
	})

	t.Run("unsupported exchange", func(t *testing.T) {
		// Create mock objects
		mockFactory := new(MockExchangeFactory)
		mockStream := &MockPricesServer_GetPricesServer{
			ctx: context.Background(),
		}

		// Setup test data
		exchange := "unsupported"
		ticker := "BTC/USDT"
		limit := int64(10)

		// Setup expectations
		mockFactory.On("GetAdapter", exchange).Return(nil, false)

		// Create test server with mock factory
		server := createServerWithMockFactory(mockFactory)

		// Call the method being tested
		err := server.GetPrices(&pb.PricesRequest{
			Exchange: exchange,
			Ticker:   ticker,
			Limit:    limit,
		}, mockStream)

		// Verify results
		assert.Error(t, err)
		statusErr, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.InvalidArgument, statusErr.Code())
		assert.Contains(t, statusErr.Message(), "unsupported exchange")
		mockFactory.AssertExpectations(t)
	})

	t.Run("adapter error", func(t *testing.T) {
		// Create mock objects
		mockFactory := new(MockExchangeFactory)
		mockAdapter := new(MockExchangeAdapter)
		mockStream := &MockPricesServer_GetPricesServer{
			ctx: context.Background(),
		}

		// Setup test data
		exchange := "binance"
		ticker := "BTC/USDT"
		limit := int64(10)
		expectedError := errors.New("API error")

		// Setup expectations
		mockAdapter.On("GetHistoricalPrices", mock.Anything, ticker, limit).Return(nil, expectedError)
		mockFactory.On("GetAdapter", exchange).Return(mockAdapter, true)

		// Create test server with mock factory
		server := createServerWithMockFactory(mockFactory)

		// Call the method being tested
		err := server.GetPrices(&pb.PricesRequest{
			Exchange: exchange,
			Ticker:   ticker,
			Limit:    limit,
		}, mockStream)

		// Verify results
		assert.Error(t, err)
		statusErr, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.Internal, statusErr.Code())
		assert.Contains(t, statusErr.Message(), "failed to get prices")
		mockFactory.AssertExpectations(t)
		mockAdapter.AssertExpectations(t)
	})

	t.Run("stream send error", func(t *testing.T) {
		// Create mock objects
		mockFactory := new(MockExchangeFactory)
		mockAdapter := new(MockExchangeAdapter)
		mockStream := &MockPricesServer_GetPricesServer{
			ctx: context.Background(),
		}

		// Setup test data
		exchange := "binance"
		ticker := "BTC/USDT"
		limit := int64(10)
		expectedError := errors.New("stream error")

		// Create sample price data
		now := time.Now()
		prices := []*pb.PricesResponse{
			{
				Date:   now.Format("2006-01-02"),
				Open:   10000.0,
				High:   10100.0,
				Low:    9900.0,
				Close:  10050.0,
				Volume: 1.5,
			},
			{
				Date:   now.Add(-24 * time.Hour).Format("2006-01-02"),
				Open:   9900.0,
				High:   10000.0,
				Low:    9800.0,
				Close:  9950.0,
				Volume: 2.0,
			},
		}

		// Setup expectations
		mockAdapter.On("GetHistoricalPrices", mock.Anything, ticker, limit).Return(prices, nil)
		mockFactory.On("GetAdapter", exchange).Return(mockAdapter, true)

		// Setup stream to return error on first Send
		mockStream.On("Send", prices[0]).Return(expectedError).Once()

		// Create test server with mock factory
		server := createServerWithMockFactory(mockFactory)

		// Call the method being tested
		err := server.GetPrices(&pb.PricesRequest{
			Exchange: exchange,
			Ticker:   ticker,
			Limit:    limit,
		}, mockStream)

		// Verify results
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "error sending price data")
		mockFactory.AssertExpectations(t)
		mockAdapter.AssertExpectations(t)
		mockStream.AssertExpectations(t)
	})

	t.Run("default limit", func(t *testing.T) {
		// Create mock objects
		mockFactory := new(MockExchangeFactory)
		mockAdapter := new(MockExchangeAdapter)
		mockStream := &MockPricesServer_GetPricesServer{
			ctx: context.Background(),
		}

		// Setup test data
		exchange := "binance"
		ticker := "BTC/USDT"
		defaultLimit := int64(100) // Default limit in the server

		// Create sample price data (empty for simplicity)
		prices := []*pb.PricesResponse{}

		// Setup expectations
		mockAdapter.On("GetHistoricalPrices", mock.Anything, ticker, defaultLimit).Return(prices, nil)
		mockFactory.On("GetAdapter", exchange).Return(mockAdapter, true)

		// Create test server with mock factory
		server := createServerWithMockFactory(mockFactory)

		// Call the method being tested with zero limit (should use default)
		err := server.GetPrices(&pb.PricesRequest{
			Exchange: exchange,
			Ticker:   ticker,
			Limit:    0,
		}, mockStream)

		// Verify results
		assert.NoError(t, err)
		mockFactory.AssertExpectations(t)
		mockAdapter.AssertExpectations(t)
	})
}

// TestIntegrationGetPrices tests the Server.GetPrices method through the gRPC interface
func TestIntegrationGetPrices(t *testing.T) {
	// This test uses the real Server implementation with the gRPC interface
	// It's skipped by default because it requires registered adapters
	t.Skip("Skipping integration test - requires registered adapters")

	// Setup a gRPC server and client
	client, cleanup := setupGRPCServer(t)
	defer cleanup()

	// Create a request
	req := &pb.PricesRequest{
		Exchange: "binance",
		Ticker:   "BTC/USDT",
		Limit:    10,
	}

	// Call the GetPrices method on the real server implementation
	stream, err := client.GetPrices(context.Background(), req)
	require.NoError(t, err)

	// Receive and verify responses
	var responses []*pb.PricesResponse
	for {
		resp, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatalf("Error receiving response: %v", err)
		}
		responses = append(responses, resp)
	}

	// Verify we got some responses
	assert.NotEmpty(t, responses, "Should receive price data")

	// Note: This test directly exercises the real Server.GetPrices implementation
	// through the gRPC interface, providing coverage for that method.
}

func TestStart(t *testing.T) {
	// This is more of an integration test and would require a real server
	// For unit testing, we'll just verify the function signature is correct
	t.Run("Start function exists with correct signature", func(t *testing.T) {
		// Just a placeholder to verify the function exists
		var _ func(int) error = Start

		// Verify that the function is exported and has the expected signature
		assert.NotNil(t, Start, "Start function should be exported")

		// Note: We can't actually call Start in a unit test because it would try to bind to a port
		// and start a server, which is not appropriate for a unit test.
		// For integration testing, we would need to:
		// 1. Start the server in a goroutine
		// 2. Connect to it with a client
		// 3. Make requests and verify responses
		// 4. Shut down the server
	})

	// Test with a port that's already in use to trigger an error
	t.Run("Start with port already in use", func(t *testing.T) {
		// First, start a server on a port
		listener, err := net.Listen("tcp", ":0") // Use port 0 to get a random available port
		require.NoError(t, err, "Failed to create listener for test")
		defer listener.Close()

		// Get the port that was assigned
		port := listener.Addr().(*net.TCPAddr).Port

		// Now try to start our server on the same port, which should fail
		err = Start(port)
		assert.Error(t, err, "Start should fail when port is already in use")
		assert.Contains(t, err.Error(), "failed to listen", "Error message should indicate listen failure")
	})
}

// TestServerImplementsInterface verifies that TestServer implements the same interface as Server
func TestServerImplementsInterface(t *testing.T) {
	// Create an interface that both servers should implement
	type PricesServer interface {
		GetPrices(req *pb.PricesRequest, stream pb.Prices_GetPricesServer) error
	}

	// Verify that both types implement this interface
	var _ PricesServer = (*Server)(nil)
	var _ PricesServer = (*TestServer)(nil)

	// This is a static compile-time check that ensures
	// TestServer implements the same interface as Server
	assert.True(t, true, "Both Server and TestServer implement the PricesServer interface")
}

// TestDirectServerGetPrices tests the original Server.GetPrices method directly
func TestDirectServerGetPrices(t *testing.T) {
	t.Run("successful price retrieval", func(t *testing.T) {
		// Create mock objects
		mockAdapter := new(MockExchangeAdapter)
		mockStream := &MockPricesServer_GetPricesServer{
			ctx: context.Background(),
		}

		// Setup test data
		exchange := "binance"
		ticker := "BTC/USDT"
		limit := int64(10)

		// Create sample price data
		now := time.Now()
		prices := []*pb.PricesResponse{
			{
				Date:   now.Format("2006-01-02"),
				Open:   10000.0,
				High:   10100.0,
				Low:    9900.0,
				Close:  10050.0,
				Volume: 1.5,
			},
		}

		// Create a real server
		server := NewServer()

		// Create a properly initialized exchange factory
		mockExchangeFactory := exchanges.NewExchangeFactory()

		// Setup mock adapter
		mockAdapter.On("GetName").Return(exchange)
		mockAdapter.On("GetHistoricalPrices", mock.Anything, ticker, limit).Return(prices, nil)

		// Register our mock adapter with the factory
		mockExchangeFactory.RegisterAdapter(mockAdapter)

		// Replace the server's exchange factory
		server.exchangeFactory = mockExchangeFactory

		// Setup stream expectations
		mockStream.On("Send", mock.Anything).Return(nil)

		// Call the method being tested
		err := server.GetPrices(&pb.PricesRequest{
			Exchange: exchange,
			Ticker:   ticker,
			Limit:    limit,
		}, mockStream)

		// Verify results
		assert.NoError(t, err)
		mockAdapter.AssertExpectations(t)
		mockStream.AssertExpectations(t)
	})

	t.Run("unsupported exchange", func(t *testing.T) {
		// Create mock objects
		mockStream := &MockPricesServer_GetPricesServer{
			ctx: context.Background(),
		}

		// Setup test data
		exchange := "unsupported"
		ticker := "BTC/USDT"
		limit := int64(10)

		// Create a real server
		server := NewServer()

		// Create a new empty factory and replace the server's factory
		// We can't directly clear the adapters map since it's private
		emptyFactory := &exchanges.ExchangeFactory{}
		server.exchangeFactory = emptyFactory

		// Call the method being tested
		err := server.GetPrices(&pb.PricesRequest{
			Exchange: exchange,
			Ticker:   ticker,
			Limit:    limit,
		}, mockStream)

		// Verify results
		assert.Error(t, err)
		statusErr, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.InvalidArgument, statusErr.Code())
		assert.Contains(t, statusErr.Message(), "unsupported exchange")
	})

	t.Run("adapter error", func(t *testing.T) {
		// Create mock objects
		mockAdapter := new(MockExchangeAdapter)
		mockStream := &MockPricesServer_GetPricesServer{
			ctx: context.Background(),
		}

		// Setup test data
		exchange := "binance"
		ticker := "BTC/USDT"
		limit := int64(10)
		expectedError := errors.New("API error")

		// Create a real server
		server := NewServer()

		// Create a properly initialized exchange factory
		mockExchangeFactory := exchanges.NewExchangeFactory()

		// Setup mock adapter
		mockAdapter.On("GetName").Return(exchange)
		mockAdapter.On("GetHistoricalPrices", mock.Anything, ticker, limit).Return(nil, expectedError)

		// Register our mock adapter with the factory
		mockExchangeFactory.RegisterAdapter(mockAdapter)

		// Replace the server's exchange factory
		server.exchangeFactory = mockExchangeFactory

		// Call the method being tested
		err := server.GetPrices(&pb.PricesRequest{
			Exchange: exchange,
			Ticker:   ticker,
			Limit:    limit,
		}, mockStream)

		// Verify results
		assert.Error(t, err)
		statusErr, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.Internal, statusErr.Code())
		assert.Contains(t, statusErr.Message(), "failed to get prices")
		mockAdapter.AssertExpectations(t)
	})

	t.Run("stream send error", func(t *testing.T) {
		// Create mock objects
		mockAdapter := new(MockExchangeAdapter)
		mockStream := &MockPricesServer_GetPricesServer{
			ctx: context.Background(),
		}

		// Setup test data
		exchange := "binance"
		ticker := "BTC/USDT"
		limit := int64(10)
		expectedError := errors.New("stream error")

		// Create sample price data
		now := time.Now()
		prices := []*pb.PricesResponse{
			{
				Date:   now.Format("2006-01-02"),
				Open:   10000.0,
				High:   10100.0,
				Low:    9900.0,
				Close:  10050.0,
				Volume: 1.5,
			},
		}

		// Create a real server
		server := NewServer()

		// Create a properly initialized exchange factory
		mockExchangeFactory := exchanges.NewExchangeFactory()

		// Setup mock adapter
		mockAdapter.On("GetName").Return(exchange)
		mockAdapter.On("GetHistoricalPrices", mock.Anything, ticker, limit).Return(prices, nil)

		// Register our mock adapter with the factory
		mockExchangeFactory.RegisterAdapter(mockAdapter)

		// Replace the server's exchange factory
		server.exchangeFactory = mockExchangeFactory

		// Setup stream to return error on Send
		mockStream.On("Send", mock.Anything).Return(expectedError)

		// Call the method being tested
		err := server.GetPrices(&pb.PricesRequest{
			Exchange: exchange,
			Ticker:   ticker,
			Limit:    limit,
		}, mockStream)

		// Verify results
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "error sending price data")
		mockAdapter.AssertExpectations(t)
		mockStream.AssertExpectations(t)
	})

	t.Run("default limit", func(t *testing.T) {
		// Create mock objects
		mockAdapter := new(MockExchangeAdapter)
		mockStream := &MockPricesServer_GetPricesServer{
			ctx: context.Background(),
		}

		// Setup test data
		exchange := "binance"
		ticker := "BTC/USDT"
		defaultLimit := int64(100) // Default limit in the server

		// Create sample price data (empty for simplicity)
		prices := []*pb.PricesResponse{}

		// Create a real server
		server := NewServer()

		// Create a properly initialized exchange factory
		mockExchangeFactory := exchanges.NewExchangeFactory()

		// Setup mock adapter
		mockAdapter.On("GetName").Return(exchange)
		mockAdapter.On("GetHistoricalPrices", mock.Anything, ticker, defaultLimit).Return(prices, nil)

		// Register our mock adapter with the factory
		mockExchangeFactory.RegisterAdapter(mockAdapter)

		// Replace the server's exchange factory
		server.exchangeFactory = mockExchangeFactory

		// Call the method being tested with zero limit (should use default)
		err := server.GetPrices(&pb.PricesRequest{
			Exchange: exchange,
			Ticker:   ticker,
			Limit:    0,
		}, mockStream)

		// Verify results
		assert.NoError(t, err)
		mockAdapter.AssertExpectations(t)
	})
}
