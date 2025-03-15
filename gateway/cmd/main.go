package main

import (
	"log"

	"github.com/timakaa/historical-common/database"
	gateway "github.com/timakaa/historical-gateway/internal"
)

func main() {
	_, err := database.InitDatabase()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer database.CloseDatabase()

	server, err := gateway.NewServer("localhost:50051", "localhost:50052")
	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
	}

	if err := server.Start(50050); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
