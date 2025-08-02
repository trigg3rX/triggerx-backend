package actions

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/trigg3rX/triggerx-backend/cli/core/config"
	"github.com/urfave/cli"
)

type FaucetRequest struct {
	Address string `json:"address"`
}

type FaucetResponse struct {
	TxHash string `json:"txHash"`
	Error  string `json:"error,omitempty"`
}

func FundImuaAccount(ctx *cli.Context) error {
	log.Println("💰 Starting IMUA token funding process...")

	// Initialize config from environment variables
	err := config.Init()
	if err != nil {
		return fmt.Errorf("failed to initialize config: %v", err)
	}

	// Check required environment variables
	imuaHomeDir := os.Getenv("IMUA_HOME_DIR")
	if imuaHomeDir == "" {
		return fmt.Errorf("IMUA_HOME_DIR environment variable not set")
	}

	imuaAccountKeyName := os.Getenv("IMUA_ACCOUNT_KEY_NAME")
	if imuaAccountKeyName == "" {
		return fmt.Errorf("IMUA_ACCOUNT_KEY_NAME environment variable not set")
	}

	imuaCosGrpcURL := os.Getenv("IMUA_COS_GRPC_URL")
	if imuaCosGrpcURL == "" {
		return fmt.Errorf("IMUA_COS_GRPC_URL environment variable not set")
	}

	// Get keyring password from environment
	keyringPassword := os.Getenv("KEYRING_PASSWORD")
	if keyringPassword == "" {
		return fmt.Errorf("KEYRING_PASSWORD environment variable not set")
	}

	// Faucet URL
	faucetURL := os.Getenv("IMUA_FAUCET_URL")
	if faucetURL == "" {
		faucetURL = "https://241009-faucet.exocore-restaking.com/"
	}

	log.Printf("Funding configuration:")
	log.Printf("- Home Directory: %s", imuaHomeDir)
	log.Printf("- Account Key Name: %s", imuaAccountKeyName)
	log.Printf("- Faucet URL: %s", faucetURL)

	// Step 1: Get the validator address
	log.Println("\n🔍 Step 1: Getting validator address...")
	getAddrCmd := exec.Command("imuad",
		"--home", imuaHomeDir,
		"keys", "show", "-a", imuaAccountKeyName,
	)

	// Provide password for address query
	getAddrCmd.Stdin = strings.NewReader(keyringPassword + "\n")

	addrOutput, err := getAddrCmd.Output()
	if err != nil {
		return fmt.Errorf("failed to get validator address: %v", err)
	}

	validatorAddr := strings.TrimSpace(string(addrOutput))
	log.Printf("Validator address: %s", validatorAddr)

	// Step 2: Check current balance
	log.Println("\n💰 Step 2: Checking current balance...")
	err = checkImuaBalance(validatorAddr, imuaCosGrpcURL)
	if err != nil {
		log.Printf("Warning: Could not check balance: %v", err)
	}

	// Step 3: Request tokens from faucet
	log.Println("\n🚰 Step 3: Requesting IMUA tokens from faucet...")
	txHash, err := requestFaucetTokens(faucetURL, validatorAddr)
	if err != nil {
		return fmt.Errorf("failed to request tokens from faucet: %v", err)
	}

	log.Printf("✅ Faucet request successful!")
	log.Printf("Transaction Hash: %s", txHash)

	// Step 4: Wait and check balance again
	log.Println("\n⏳ Step 4: Waiting for transaction to be processed...")
	log.Println("Waiting 30 seconds for the transaction to be confirmed...")
	time.Sleep(30 * time.Second)

	log.Println("\n🔍 Step 5: Checking updated balance...")
	err = checkImuaBalance(validatorAddr, imuaCosGrpcURL)
	if err != nil {
		log.Printf("Warning: Could not check updated balance: %v", err)
	}

	log.Println("\n🎉 IMUA token funding completed successfully!")
	log.Printf("Validator Address: %s", validatorAddr)
	log.Printf("Transaction Hash: %s", txHash)

	// log.Println("\n📋 Next Steps:")
	// log.Println("1. 🏗️  Register as operator: ./triggerx register-imua-operator")
	// log.Println("2. 🚀 Complete registration: ./triggerx complete-imua-registration")

	return nil
}

func CheckImuaBalance(ctx *cli.Context) error {
	log.Println("🔍 Checking IMUA account balance...")

	// Initialize config from environment variables
	err := config.Init()
	if err != nil {
		return fmt.Errorf("failed to initialize config: %v", err)
	}

	// Check required environment variables
	imuaHomeDir := os.Getenv("IMUA_HOME_DIR")
	if imuaHomeDir == "" {
		return fmt.Errorf("IMUA_HOME_DIR environment variable not set")
	}

	imuaAccountKeyName := os.Getenv("IMUA_ACCOUNT_KEY_NAME")
	if imuaAccountKeyName == "" {
		return fmt.Errorf("IMUA_ACCOUNT_KEY_NAME environment variable not set")
	}

	imuaCosGrpcURL := os.Getenv("IMUA_COS_GRPC_URL")
	if imuaCosGrpcURL == "" {
		return fmt.Errorf("IMUA_COS_GRPC_URL environment variable not set")
	}

	// Get keyring password from environment
	keyringPassword := os.Getenv("KEYRING_PASSWORD")
	if keyringPassword == "" {
		return fmt.Errorf("KEYRING_PASSWORD environment variable not set")
	}

	// Get the validator address
	log.Println("Getting validator address...")
	getAddrCmd := exec.Command("imuad",
		"--home", imuaHomeDir,
		"keys", "show", "-a", imuaAccountKeyName,
	)

	// Provide password for address query
	getAddrCmd.Stdin = strings.NewReader(keyringPassword + "\n")

	addrOutput, err := getAddrCmd.Output()
	if err != nil {
		return fmt.Errorf("failed to get validator address: %v", err)
	}

	validatorAddr := strings.TrimSpace(string(addrOutput))
	log.Printf("Validator address: %s", validatorAddr)

	// Check balance
	return checkImuaBalance(validatorAddr, imuaCosGrpcURL)
}

func checkImuaBalance(address, grpcURL string) error {
	log.Printf("Checking balance for address: %s", address)

	balanceCmd := exec.Command("imuad",
		"query", "bank", "balances", address,
		"--node", grpcURL,
		"--output", "json",
	)

	balanceOutput, err := balanceCmd.Output()
	if err != nil {
		return fmt.Errorf("failed to query balance: %v", err)
	}

	log.Printf("Balance query result: %s", string(balanceOutput))

	// Parse JSON to extract balance information
	var balanceData struct {
		Balances []struct {
			Denom  string `json:"denom"`
			Amount string `json:"amount"`
		} `json:"balances"`
	}

	err = json.Unmarshal(balanceOutput, &balanceData)
	if err != nil {
		log.Printf("Warning: Could not parse balance JSON: %v", err)
		log.Printf("Raw balance output: %s", string(balanceOutput))
		return nil
	}

	if len(balanceData.Balances) == 0 {
		log.Println("❌ Account has no balance (0 IMUA tokens)")
		log.Println("💡 Run './triggerx fund-imua-account' to get IMUA tokens from the faucet")
		return nil
	}

	log.Println("💰 Account Balances:")
	for _, balance := range balanceData.Balances {
		log.Printf("  - %s %s", balance.Amount, balance.Denom)
	}

	return nil
}

func requestFaucetTokens(faucetURL, address string) (string, error) {
	// Prepare request body
	requestBody := FaucetRequest{
		Address: address,
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %v", err)
	}

	log.Printf("Sending faucet request to: %s", faucetURL)
	log.Printf("Request body: %s", string(jsonData))

	// Make HTTP request
	resp, err := http.Post(faucetURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to make faucet request: %v", err)
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %v", err)
	}

	log.Printf("Faucet response status: %d", resp.StatusCode)
	log.Printf("Faucet response body: %s", string(body))

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("faucet request failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var faucetResp FaucetResponse
	err = json.Unmarshal(body, &faucetResp)
	if err != nil {
		return "", fmt.Errorf("failed to parse faucet response: %v", err)
	}

	if faucetResp.Error != "" {
		return "", fmt.Errorf("faucet error: %s", faucetResp.Error)
	}

	if faucetResp.TxHash == "" {
		return "", fmt.Errorf("no transaction hash received from faucet")
	}

	return faucetResp.TxHash, nil
}
