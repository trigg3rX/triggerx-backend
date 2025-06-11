package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type ConditionResult struct {
	Satisfied bool
	Timestamp time.Time
	Response  float64
	Price     float64
}

func condition() ConditionResult {
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
	defer func() {
		_ = resp.Body.Close()
	}()

	body, err := io.ReadAll(resp.Body)
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
			Price:     0,
		}
	}

	satisfied := ethPrice > 0

	return ConditionResult{
		Satisfied: satisfied,
		Timestamp: time.Now(),
		Response:  ethPrice,
		Price:     ethPrice,
	}
}

func main() {
	result := condition()

	fmt.Println("Condition satisfied:", result.Satisfied)
	fmt.Println("Timestamp:", result.Timestamp.Format(time.RFC3339))
	fmt.Println("Response:", result.Response)

	if result.Response != 0 {
		fmt.Println("Response:", result.Response)
	}

	if result.Satisfied {
		fmt.Println("Ethereum price is greater than 0")
	} else {
		fmt.Println("Ethereum price is 0 or could not be fetched")
	}
}
