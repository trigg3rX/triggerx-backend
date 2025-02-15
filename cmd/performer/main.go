package main

import (
	"log"
	"github.com/gin-gonic/gin"
	"github.com/trigg3rX/triggerx-backend/execute/performer"

	// "Execution_Service/services"  
)

func main() {
	// services.Init()
	router := gin.Default()
	router.POST("/task/execute", performer.ExecuteTask)
	log.Println("Server starting on :4003")

	if err := router.Run(":4003"); err != nil {
		log.Fatal(err)
	}
}