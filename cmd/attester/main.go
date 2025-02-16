package main

import (
	"log"
	"net/http"

	"github.com/trigg3rX/triggerx-backend/internal/attester"
)

func main() {
	http.HandleFunc("/task/validate", attester.ValidateTask)
	log.Println("Validation Server starting on :4002")
	log.Fatal(http.ListenAndServe(":4002", nil)) 
}