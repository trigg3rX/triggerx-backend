package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

// ConditionResult represents the result of condition evaluation
type ConditionResult struct {
	Satisfied bool
	Timestamp time.Time
	Response  float64
	Price     float64
}

// condition evaluates user-defined conditions and returns the result
func condition() ConditionResult {
	// Fetch Ethereum price from CoinGecko API
	url := "https://api.coingecko.com/api/v3/simple/price?ids=ethereum&vs_currencies=usd"
	resp, err := http.Get(url)
	if err != nil {
		return ConditionResult{
			Satisfied: false,
			Timestamp: time.Now(),
			Response:  0,
			Price:     0,
		}
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return ConditionResult{
			Satisfied: false,
			Timestamp: time.Now(),
			Response:  0,
			Price:     0,
		}
	}

	var result map[string]map[string]float64
	if err := json.Unmarshal(body, &result); err != nil {
		return ConditionResult{
			Satisfied: false,
			Timestamp: time.Now(),
			Response:  0,
			Price:     0,
		}
	}

	ethPrice, ok := result["ethereum"]["usd"]
	if !ok {
		return ConditionResult{
			Satisfied: false,
			Timestamp: time.Now(),
			Response:  0,
		}
	}

	// Check if the price is between 1835 and 1838
	satisfied := ethPrice >= 1835 && ethPrice <= 1838

	var response float64
	if satisfied {
		response = ethPrice
	} else {
		response = ethPrice
	}

	// Return the condition result
	return ConditionResult{
		Satisfied: satisfied,
		Timestamp: time.Now(),
		Response:  response,
		Price:     ethPrice,
	}
}

func main() {
	// Call the condition function
	result := condition()

	// Print the results
	fmt.Println("Condition satisfied:", result.Satisfied)
	fmt.Println("Timestamp:", result.Timestamp.Format(time.RFC3339))
	fmt.Println("Ethereum Price:", result.Price)

	if result.Response != 0 {
		fmt.Println("Response:", result.Response)
	}

	// Take action based on condition result
	if result.Satisfied {
		fmt.Println("Executing actions for satisfied condition...")
		// Add your actions here
	} else {
		fmt.Println("Condition not satisfied, no action taken.")
	}
}
