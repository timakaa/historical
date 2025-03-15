package main

import (
	"log"

	gateway "github.com/timakaa/historical-gateway/internal"
)

func main() {
	server, err := gateway.NewServer("localhost:50051", "localhost:50052")
	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
	}

	if err := server.Start(50050); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
} 