package main

import (
	"log"

	access "github.com/timakaa/historical-access/internal"
)

func main() {
	if err := access.Start(50052); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
} 