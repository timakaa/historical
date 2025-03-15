package main

import (
	"log"

	auth "github.com/timakaa/historical-auth/internal"
)

func main() {
	if err := auth.Start(50052); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
