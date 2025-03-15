package main

import (
	"log"

	auth "github.com/timakaa/historical-auth/internal"
	"github.com/timakaa/historical-common/database"
)

func main() {
	// Initialize database connection
	_, err := database.InitDatabase()
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// Start the auth server
	if err := auth.Start(50052); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
