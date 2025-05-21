package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
)

type CoinGeckoResponse struct {
	Ethereum struct {
		USD float64 `json:"usd"`
	} `json:"ethereum"`
}

func checker() string {
	url := "https://api.coingecko.com/api/v3/simple/price?ids=ethereum&vs_currencies=usd"

	resp, err := http.Get(url)
	if err != nil {
		return "Error fetching data: " + err.Error()
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "Error reading response: " + err.Error()
	}

	var response CoinGeckoResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		return "Error parsing JSON: " + err.Error()
	}

	price := response.Ethereum.USD
	return strconv.FormatFloat(price, 'f', 2, 64)
}

func main() {
	result := checker()

	jsonValue, _ := json.Marshal(result)
	fmt.Println(string(jsonValue))
}
