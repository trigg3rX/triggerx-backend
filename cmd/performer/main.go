package main

import (
	"log"
	"github.com/gin-gonic/gin"

	"github.com/trigg3rX/triggerx-backend/internal/performer"
)

func main() {
	router := gin.Default()
	router.POST("/task/execute", performer.ExecuteTask)
	log.Println("Server starting on :4003")

	if err := router.Run(":4003"); err != nil {
		log.Fatal(err)
	}
}