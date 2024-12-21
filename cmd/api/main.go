package main

import (
	"github.com/trigg3rX/triggerx-backend/pkg/api"
	"github.com/trigg3rX/triggerx-backend/pkg/database"
	"log"
)

func main() {
	log.Println("Starting API server...")

	// Initialize database connection
	config := database.NewConfig()
	conn, err := database.NewConnection(config)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer conn.Close()

	// Initialize and start server
	server := api.NewServer(conn)
	log.Println("Server initialized, starting on port 8080...")
	if err := server.Start("8080"); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
