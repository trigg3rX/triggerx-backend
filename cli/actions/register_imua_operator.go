package actions

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/trigg3rX/triggerx-backend/cli/core/config"
	"github.com/urfave/cli"
)

func RegisterImuaOperator(ctx *cli.Context) error {
	log.Println("Starting Imuachain operator registration...")

	// Initialize config from environment variables
	err := config.Init()
	if err != nil {
		return fmt.Errorf("failed to initialize config: %v", err)
	}

	// Step 0: Ensure validator key exists
	log.Println("🔑 Checking validator key setup...")
	if err := EnsureValidatorKeyExists(); err != nil {
		return fmt.Errorf("validator key setup failed: %v", err)
	}
	log.Println("✅ Validator key is ready")

	// Check required environment variables for Imua registration
	operatorName := os.Getenv("OPERATOR_NAME")
	if operatorName == "" {
		return fmt.Errorf("OPERATOR_NAME environment variable not set")
	}

	imuaHomeDir := os.Getenv("IMUA_HOME_DIR")
	if imuaHomeDir == "" {
		return fmt.Errorf("IMUA_HOME_DIR environment variable not set")
	}

	imuaAccountKeyName := os.Getenv("IMUA_ACCOUNT_KEY_NAME")
	if imuaAccountKeyName == "" {
		return fmt.Errorf("IMUA_ACCOUNT_KEY_NAME environment variable not set")
	}

	imuaChainID := os.Getenv("IMUA_CHAIN_ID")
	if imuaChainID == "" {
		return fmt.Errorf("IMUA_CHAIN_ID environment variable not set")
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

	log.Printf("Registering operator with the following configuration:")
	log.Printf("- Operator Name: %s", operatorName)
	log.Printf("- Home Directory: %s", imuaHomeDir)
	log.Printf("- Account Key Name: %s", imuaAccountKeyName)
	log.Printf("- Chain ID: %s", imuaChainID)
	log.Printf("- GRPC URL: %s", imuaCosGrpcURL)

	// Step 1.5: Check account balance before attempting registration
	log.Println("💰 Checking account balance before registration...")
	hasBalance, err := checkAccountHasBalance(imuaHomeDir, imuaAccountKeyName, imuaCosGrpcURL, keyringPassword)
	if err != nil {
		log.Printf("Warning: Could not check account balance: %v", err)
	} else if !hasBalance {
		log.Println("❌ Account has insufficient IMUA tokens for gas fees")
		log.Println("")
		log.Println("🚰 To fund your account, run:")
		log.Println("   ./triggerx fund-imua-account")
		log.Println("")
		log.Println("💡 Or check your current balance:")
		log.Println("   ./triggerx check-imua-balance")
		log.Println("")
		return fmt.Errorf("account not funded with IMUA tokens - registration requires gas fees")
	}
	log.Println("✅ Account has sufficient balance for registration")

	// Step 2: Register the operator
	log.Println("Step 1: Registering operator on Imuachain...")
	cmd := exec.Command("imuad",
		"tx", "operator", "register-operator",
		"--from", imuaAccountKeyName,
		"--chain-id", imuaChainID,
		"--node", imuaCosGrpcURL,
		"--gas", "150000",
		"--gas-prices", "7hua",
		"--meta-info", operatorName,
		"--commission-rate", "0.10",
		"--commission-max-rate", "0.20",
		"--commission-max-change-rate", "0.01",
		"-y", // Non-interactive flag
	)

	// Set up stdin to provide the password
	cmd.Stdin = strings.NewReader(keyringPassword + "\n")

	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("Operator registration command output: %s", string(output))

		// Check if it's a funding issue and provide helpful message
		if strings.Contains(string(output), "account") && strings.Contains(string(output), "not found") {
			log.Println("")
			log.Println("💡 This error usually means your account is not funded with IMUA tokens.")
			log.Println("🚰 To fund your account, run:")
			log.Println("   ./triggerx fund-imua-account")
			log.Println("")
			return fmt.Errorf("account funding issue - please fund your account with IMUA tokens")
		}

		return fmt.Errorf("failed to register operator: %v", err)
	}

	log.Printf("Operator registration successful: %s", string(output))

	// Step 3: Wait a moment for the transaction to be processed
	log.Println("Waiting for transaction to be processed...")

	// Step 4: Verify operator registration
	log.Println("Step 2: Verifying operator registration...")

	// Get the operator address
	getAddrCmd := exec.Command("imuad",
		"--home", imuaHomeDir,
		"keys", "show", "-a", imuaAccountKeyName,
	)

	// Provide password for address query too
	getAddrCmd.Stdin = strings.NewReader(keyringPassword + "\n")

	addrOutput, err := getAddrCmd.Output()
	if err != nil {
		return fmt.Errorf("failed to get operator address: %v", err)
	}

	operatorAddr := string(addrOutput)
	operatorAddr = strings.TrimSpace(operatorAddr) // Remove trailing newline

	log.Printf("Operator address: %s", operatorAddr)

	// Query operator info
	queryCmd := exec.Command("imuad",
		"--home", imuaHomeDir,
		"query", "operator", "get-operator-info",
		operatorAddr,
		"--node", imuaCosGrpcURL,
		"--output", "json",
	)

	queryOutput, err := queryCmd.Output()
	if err != nil {
		log.Printf("Query output: %s", string(queryOutput))
		log.Printf("Warning: Could not verify operator registration immediately: %v", err)
		log.Println("This is often normal as transactions may take a moment to be processed and indexed.")
		log.Println("You can manually verify using:")
		log.Printf("  imuad query operator get-operator-info %s --node %s --output json", operatorAddr, imuaCosGrpcURL)
	} else {
		log.Printf("Operator registration verified: %s", string(queryOutput))
	}

	log.Println("\n✅ Imuachain operator registration completed successfully!")
	log.Printf("Operator Name: %s", operatorName)
	log.Printf("Operator Address: %s", operatorAddr)

	return nil
}

func checkAccountHasBalance(homeDir, keyName, grpcURL, keyringPassword string) (bool, error) {
	// Get the validator address
	getAddrCmd := exec.Command("imuad",
		"--home", homeDir,
		"keys", "show", "-a", keyName,
	)

	// Provide password for address query
	getAddrCmd.Stdin = strings.NewReader(keyringPassword + "\n")

	addrOutput, err := getAddrCmd.Output()
	if err != nil {
		return false, fmt.Errorf("failed to get validator address: %v", err)
	}

	validatorAddr := strings.TrimSpace(string(addrOutput))

	// Check balance (balance queries don't require password)
	balanceCmd := exec.Command("imuad",
		"query", "bank", "balances", validatorAddr,
		"--node", grpcURL,
		"--output", "json",
	)

	balanceOutput, err := balanceCmd.Output()
	if err != nil {
		return false, fmt.Errorf("failed to query balance: %v", err)
	}

	// Parse JSON to check if there are any balances
	var balanceData struct {
		Balances []struct {
			Denom  string `json:"denom"`
			Amount string `json:"amount"`
		} `json:"balances"`
	}

	err = json.Unmarshal(balanceOutput, &balanceData)
	if err != nil {
		return false, fmt.Errorf("failed to parse balance JSON: %v", err)
	}

	// Return true if there are any balances (specifically looking for IMUA or hua tokens)
	return len(balanceData.Balances) > 0, nil
}
