package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
)

// Response structure to match the CoinGecko API response
type CoinGeckoResponse struct {
	Ethereum struct {
		USD float64 `json:"usd"`
	} `json:"ethereum"`
}

// checker fetches the current Ethereum price and returns it as a string
func checker() string {
	// CoinGecko API URL for Ethereum price in USD
	url := "https://api.coingecko.com/api/v3/simple/price?ids=ethereum&vs_currencies=usd"

	// Make HTTP GET request
	resp, err := http.Get(url)
	if err != nil {
		return "Error fetching data: " + err.Error()
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "Error reading response: " + err.Error()
	}

	// Parse JSON response
	var response CoinGeckoResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		return "Error parsing JSON: " + err.Error()
	}

	// Extract the Ethereum price in USD and convert to string
	price := response.Ethereum.USD
	return strconv.FormatFloat(price, 'f', 2, 64)
}

func main() {
	// Call the checker function
	result := checker()

	jsonValue, _ := json.Marshal(result)
	fmt.Println(string(jsonValue))
}
