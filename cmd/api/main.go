package main

import (
	// "log"
	"net/http"
	"github.com/trigg3rX/go-backend/pkg/api"
	"github.com/trigg3rX/go-backend/pkg/database"
)

func Handler(w http.ResponseWriter, r *http.Request) {
	// Initialize database connection
	config := database.NewConfig()
	conn, err := database.NewConnection(config)
	if err != nil {
		http.Error(w, "Failed to connect to database", http.StatusInternalServerError)
		return
	}
	defer conn.Close()

	// Initialize and start server
	server := api.NewServer(conn)
	server.ServeHTTP(w, r)
} 