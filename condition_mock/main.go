package main

import (
	"fmt"
	"log"
	"net/http"
	"time"
)

func main() {
	log.Println("Starting condition mock server on port 8080")
	http.HandleFunc("/price", func(w http.ResponseWriter, r *http.Request) {
		price := getPrice()
		log.Printf("Price: %s", price)
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte(price))
		if err != nil {
			log.Printf("Error writing price: %v", err)
		}
	})
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Printf("Error starting server: %v", err)
	}
}

func getPrice() string {
	now := time.Now().Unix()
	// Use even/odd seconds to determine direction
	base := 90.0
	rangeMax := 100.0
	step := 0.5
	steps := (now % int64((rangeMax-base)/step+1))
	val := base + float64(steps)*step

	// Alternate increment/decrement every second
	if (now/1)%2 == 1 {
		val = rangeMax - (val - base)
	}

	// Clamp to bounds just in case
	if val < base {
		val = base
	}
	if val > rangeMax {
		val = rangeMax
	}

	return fmt.Sprintf("%.2f", val)
}
