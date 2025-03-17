package main

import (
	"fmt"
)

// checker returns a payload
func checker() string {
	// This is a simple payload for demonstration
	payload := "Hello from the checker function!"
	return payload
}

func main() {
	// Call the checker function
	result := checker()
	
	// Display the payload
	fmt.Println("Payload received:", result)
}