# gRPC Microservices Example

This project demonstrates a basic implementation of a microservices architecture using gRPC in Go.

## Contents

- [Overview](#overview)
- [Project Structure](#project-structure)
- [Protocol Buffers](#protocol-buffers)
- [Services Implementation](#services-implementation)
- [Running the Application](#running-the-application)
- [Error Handling in gRPC](#error-handling-in-grpc)
- [Additional Features](#additional-features)

## Overview

gRPC is a high-performance RPC framework developed by Google. It uses Protocol Buffers for service interface definition and message format, and HTTP/2 for transport.

Key advantages of gRPC:

1. **High Performance** - uses HTTP/2 and binary protocol
2. **Strong Typing** - API contracts defined using Protocol Buffers
3. **Multi-language Support** - automatic client and server code generation
4. **Bidirectional Streaming** - support for streaming data in both directions

## Project Structure

```
crypto_prices/
├── cmd/
│   ├── app/
│   │   └── main.go         # Основное приложение (клиент + сервер)
│   ├── client/
│   │   └── main.go         # Отдельный клиент
│   └── server/
│       └── main.go         # Отдельный сервер
├── internal/
│   ├── client/
│   │   └── client.go       # Реализация gRPC клиента
│   └── server/
│       └── server.go       # Реализация gRPC сервера
└── proto/
    ├── hello.proto         # Определение Protocol Buffers
    ├── hello.pb.go         # Сгенерированный код для сообщений
    └── hello_grpc.pb.go    # Сгенерированный код для сервиса
```

## Protocol Buffers

Protocol Buffers (protobuf) is a mechanism for serializing structured data developed by Google. In gRPC, it's used to define:

1. Message structures
2. Service methods

Our main proto files:

### prices.proto

```protobuf
syntax = "proto3";

package prices;

option go_package = "crypto_prices/proto";

service HistoricalPrices {
  rpc GetHistoricalPrices (HistoricalPricesRequest) returns (stream HistoricalPricesResponse) {}
}

message HistoricalPricesRequest {
  string ticker = 1;
  string exchange = 2;
}

message HistoricalPricesResponse {
  string Date = 1;
  double Open = 2;
  double High = 3;
  double Low = 4;
  double Close = 5;
  double Volume = 6;
}
```

### access.proto

```protobuf
syntax = "proto3";

package access;

option go_package = "crypto_prices/proto";

service AccessManager {
  rpc ValidateToken (ValidateRequest) returns (ValidateResponse) {}
  rpc CreateToken (CreateTokenRequest) returns (CreateTokenResponse) {}
  rpc RevokeToken (RevokeTokenRequest) returns (RevokeTokenResponse) {}
}

message ValidateRequest {
  string token = 1;
  string service = 2;
}

message ValidateResponse {
  bool is_valid = 1;
  string user_id = 2;
  repeated string permissions = 3;
}

message CreateTokenRequest {
  string user_id = 1;
  repeated string permissions = 2;
  int64 expires_in = 3; // in seconds
}

message CreateTokenResponse {
  string token = 1;
  int64 expires_at = 2;
}

message RevokeTokenRequest {
  string token = 1;
}

message RevokeTokenResponse {
  bool success = 1;
}
```

### Code Generation

To generate Go code from proto files, we use:

1. `protoc` - Protocol Buffers compiler
2. `protoc-gen-go` - Plugin for generating Go code for messages
3. `protoc-gen-go-grpc` - Plugin for generating Go code for gRPC services

Command for generation:

```bash
protoc --go_out=. --go_opt=paths=source_relative \
    --go-grpc_out=. --go-grpc_opt=paths=source_relative \
    common/proto/*.proto
```

## Services Implementation

### Historical Prices Service

The Historical Prices service provides cryptocurrency price history data:

```go
func (s *Server) GetHistoricalPrices(req *pb.HistoricalPricesRequest, stream pb.HistoricalPrices_GetHistoricalPricesServer) error {
    log.Printf("Received request for ticker: %s from exchange: %s", req.GetTicker(), req.GetExchange())

    // Example implementation - in production, this would fetch data from a database
    for i := range [3]int{} {
        price := &pb.HistoricalPricesResponse{
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
```

### Access Manager Service

The Access Manager service handles authentication and authorization:

```go
func (s *Server) ValidateToken(ctx context.Context, req *pb.ValidateRequest) (*pb.ValidateResponse, error) {
    // In production, this would validate the token against a database
    return &pb.ValidateResponse{
        IsValid: true,
        UserId: "test-user",
        Permissions: []string{"read:prices"},
    }, nil
}

func (s *Server) CreateToken(ctx context.Context, req *pb.CreateTokenRequest) (*pb.CreateTokenResponse, error) {
    // In production, this would create and store a token
    return &pb.CreateTokenResponse{
        Token: "test-token",
        ExpiresAt: time.Now().Add(time.Duration(req.ExpiresIn) * time.Second).Unix(),
    }, nil
}
```

### API Gateway

The API Gateway provides a REST API interface using Gin and communicates with other services via gRPC:

```go
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

    // Collect all prices
    var prices []*pb.HistoricalPricesResponse
    for {
        price, err := stream.Recv()
        if err != nil {
            if err == io.EOF {
                break
            }
            c.JSON(http.StatusInternalServerError, gin.H{"error": "error receiving prices"})
            return
        }
        prices = append(prices, price)
    }

    c.JSON(http.StatusOK, prices)
}
```

## Running the Application

The project includes a Makefile with commands to build and run the services:

```bash
# Run a specific service
make run SERVICE=prices
make run SERVICE=access
make run SERVICE=gateway

# Run all services
make run-all

# Build services
make build-all

# Generate proto code
make gen

# Stop all services
make stop
```

For development with hot-reload, you can use:

```bash
make dev SERVICE=prices
```

## Error Handling in gRPC

gRPC has a built-in error handling system based on status codes. Here's how to use it:

### Server-side error generation:

```go
if token == "" {
    return nil, status.Errorf(codes.InvalidArgument, "token cannot be empty")
}

if !isValidToken(token) {
    return nil, status.Errorf(codes.Unauthenticated, "invalid token")
}
```

### Client-side error handling:

```go
resp, err := client.ValidateToken(ctx, req)
if err != nil {
    st, ok := status.FromError(err)
    if !ok {
        // Not a gRPC error
        return fmt.Errorf("unknown error: %v", err)
    }

    switch st.Code() {
    case codes.InvalidArgument:
        return fmt.Errorf("validation error: %s", st.Message())
    case codes.Unauthenticated:
        return fmt.Errorf("authentication error: %s", st.Message())
    default:
        return fmt.Errorf("error (code=%s): %s", st.Code(), st.Message())
    }
}
```

## Additional Features

### 1. Secure Connection (TLS)

For secure connections, you can use TLS:

```go
// Server
creds, err := credentials.NewServerTLSFromFile("server.crt", "server.key")
s := grpc.NewServer(grpc.Creds(creds))

// Client
creds, err := credentials.NewClientTLSFromFile("server.crt", "")
conn, err := grpc.Dial(address, grpc.WithTransportCredentials(creds))
```

### 2. Middleware (Interceptors)

For adding common functionality (logging, authentication, metrics), you can use interceptors:

```go
// Server interceptor
func loggingInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
    log.Printf("Request: %v", req)
    resp, err := handler(ctx, req)
    log.Printf("Response: %v", resp)
    return resp, err
}

s := grpc.NewServer(grpc.UnaryInterceptor(loggingInterceptor))
```

### 3. Load Balancing

gRPC supports load balancing between multiple servers:

```go
conn, err := grpc.Dial(
    "service-name",
    grpc.WithDefaultServiceConfig(`{"loadBalancingPolicy":"round_robin"}`),
    grpc.WithResolvers(resolver),
)
```

### 4. Fault Tolerance

For improved fault tolerance, you can use retries and timeouts:

```go
// Timeout
ctx, cancel := context
```
