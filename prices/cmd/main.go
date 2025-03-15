package main

import (
	"log"

	prices "github.com/timakaa/historical-prices/internal"
)

func main() {
	if err := prices.Start(50051); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
} 