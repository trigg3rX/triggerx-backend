package fees

// import (
// 	"encoding/json"
// 	"fmt"
// 	"io/ioutil"
// 	"net/http"
// )

// // FeeCalculationRequest represents the request structure for fee calculation
// type FeeCalculationRequest struct {
// 	CodeURL           string  `json:"code_url"` // URL of the code to fetch
// 	DynamicComplexity float64 `json:"dynamic_complexity"`
// 	ExecTime          float64 `json:"exec_time"`
// 	MemUsed           float64 `json:"mem_used"`
// 	BandwidthTransfer float64 `json:"bandwidth_transfer"`
// 	GasUnits          float64 `json:"gas_units"`
// 	Wstatic           float64 `json:"w_static"`  // Weight for static complexity
// 	Wdynamic          float64 `json:"w_dynamic"` // Weight for dynamic complexity
// }

// // FeeCalculationResponse represents the response structure for fee calculation
// type FeeCalculationResponse struct {
// 	TotalFee float64 `json:"total_fee"`
// }

// // CalculateFees calculates the fees based on the provided parameters
// func CalculateFees(w http.ResponseWriter, r *http.Request) {
// 	var req FeeCalculationRequest
// 	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
// 		http.Error(w, err.Error(), http.StatusBadRequest)
// 		return
// 	}

// 	// Fetch the code from the provided URL
// 	code, err := fetchCode(req.CodeURL)
// 	if err != nil {
// 		http.Error(w, err.Error(), http.StatusInternalServerError)
// 		return
// 	}

// 	// Calculate Cstatic based on the fetched code
// 	Cstatic := calculateStaticComplexity(code)

// 	// Calculate Cdynamic
// 	Cdynamic := (req.ExecTime * req.Wdynamic) + (req.MemUsed * req.Wdynamic) + (req.BandwidthTransfer * req.Wdynamic)

// 	// Calculate Cindex
// 	Cindex := (req.Wstatic * Cstatic) + (req.Wdynamic * Cdynamic)

// 	// Constants
// 	const Pcomplexity = 0.01 // Dollar value per unit complexity
// 	const Gbase = 0.01       // Fixed operational overhead

// 	// Calculate gas fees
// 	Gfees := req.GasUnits * Pcomplexity

// 	// Calculate total fee
// 	totalFee := (Cindex * Pcomplexity) + Gfees + Gbase

// 	// Prepare response
// 	response := FeeCalculationResponse{TotalFee: totalFee}
// 	w.Header().Set("Content-Type", "application/json")
// 	json.NewEncoder(w).Encode(response)
// }

// // fetchCode fetches the code from the provided URL
// func fetchCode(url string) (string, error) {
// 	resp, err := http.Get(url)
// 	if err != nil {
// 		return "", err
// 	}
// 	defer resp.Body.Close()

// 	if resp.StatusCode != http.StatusOK {
// 		return "", fmt.Errorf("failed to fetch code: %s", resp.Status)
// 	}

// 	body, err := ioutil.ReadAll(resp.Body)
// 	if err != nil {
// 		return "", err
// 	}

// 	return string(body), nil
// }

// // calculateStaticComplexity analyzes the code and returns its static complexity
// func calculateStaticComplexity(code string) float64 {
// 	// Placeholder for actual complexity calculation logic
// 	// This could involve analyzing the code structure, number of functions, lines of code, etc.
// 	// For now, let's return a dummy value
// 	return float64(len(code)) * 0.001 // Example: complexity based on code length
// }