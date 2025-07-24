package actions

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/trigg3rX/triggerx-backend/cli/core/config"
	"github.com/urfave/cli"
)

// isChainAVS determines if the given AVS address requires a validator public key
func isChainAVS(avsAddress string) bool {
	// List of known chain AVS addresses that require validator keys
	chainAVSAddresses := map[string]bool{
		// Add known chain AVS addresses here
		"0x72A5016ECb9EB01d7d54ae48bFFB62CA0B8e57a5": true,
	}

	// Check if this is a known chain AVS
	if _, exists := chainAVSAddresses[strings.ToLower(avsAddress)]; exists {
		return true
	}

	// Default to non-chain AVS (no validator key needed)
	return false
}

// OptInToAVS handles the AVS opt-in process
func OptInToAVS(ctx *cli.Context) error {
	log.Println("Starting AVS opt-in process...")

	// Initialize config from environment variables
	if err := config.Init(); err != nil {
		return fmt.Errorf("failed to initialize config: %v", err)
	}

	// Validate required environment variables
	requiredVars := map[string]string{
		"IMUA_HOME_DIR":         os.Getenv("IMUA_HOME_DIR"),
		"IMUA_ACCOUNT_KEY_NAME": os.Getenv("IMUA_ACCOUNT_KEY_NAME"),
		"IMUA_CHAIN_ID":         os.Getenv("IMUA_CHAIN_ID"),
		"IMUA_COS_GRPC_URL":     os.Getenv("IMUA_COS_GRPC_URL"),
		"KEYRING_PASSWORD":      os.Getenv("KEYRING_PASSWORD"),
		"AVS_ADDRESS":           os.Getenv("AVS_ADDRESS"),
	}

	for name, value := range requiredVars {
		if value == "" {
			return fmt.Errorf("%s environment variable not set", name)
		}
	}

	// Log configuration
	log.Printf("AVS opt-in configuration:")
	for name, value := range requiredVars {
		if name != "KEYRING_PASSWORD" {
			log.Printf("- %s: %s", name, value)
		} else {
			log.Printf("- %s: [redacted]", name)
		}
	}

	// Step 1: Prepare validator key if needed
	validatorKey := ""
	if isChainAVS(requiredVars["AVS_ADDRESS"]) {
		log.Println("\n🔑 Preparing validator key for chain AVS...")
		validatorCmd := exec.Command("imuad",
			"--home", requiredVars["IMUA_HOME_DIR"],
			"tendermint", "show-validator",
		)

		validatorOutput, err := validatorCmd.Output()
		if err != nil {
			return fmt.Errorf("failed to get validator consensus key: %v", err)
		}
		validatorKey = strings.TrimSpace(string(validatorOutput))
		log.Printf("Using validator key: %s", validatorKey)
	} else {
		log.Println("\nℹ️ Non-chain AVS detected - validator key not required")
	}

	// Step 2: Build and execute opt-in command
	log.Println("\n🚀 Executing opt-in to AVS transaction...")
	args := []string{
		"--home", requiredVars["IMUA_HOME_DIR"],
		"tx", "operator", "opt-into-avs",
		requiredVars["AVS_ADDRESS"],
		"--from", requiredVars["IMUA_ACCOUNT_KEY_NAME"],
		"--chain-id", requiredVars["IMUA_CHAIN_ID"],
		"--node", requiredVars["IMUA_COS_GRPC_URL"],
		"--gas", "200000", // Increased gas limit for safety
		"--gas-prices", "7hua",
		"--keyring-backend", "file",
		"-y",
	}

	// Add validator key only for chain AVS
	if validatorKey != "" {
		args = append(args, validatorKey)
	}

	optInCmd := exec.Command("imuad", args...)
	optInCmd.Stdin = strings.NewReader(requiredVars["KEYRING_PASSWORD"] + "\n")

	optInOutput, err := optInCmd.CombinedOutput()
	if err != nil {
		log.Printf("❌ Opt-in command output: %s", string(optInOutput))

		// Handle specific error cases
		if strings.Contains(string(optInOutput), "minimum self delegation") {
			return fmt.Errorf("insufficient self-delegation (minimum $1,000 USD required)")
		}
		if strings.Contains(string(optInOutput), "public key is not required") {
			return fmt.Errorf("validator key provided for non-chain AVS - please contact support")
		}

		return fmt.Errorf("failed to opt-in to AVS: %v", err)
	}

	log.Printf("✅ Opt-in submitted: %s", string(optInOutput))

	// Step 3: Transaction verification
	var txHash string
	if strings.Contains(string(optInOutput), "txhash:") {
		lines := strings.Split(string(optInOutput), "\n")
		for _, line := range lines {
			if strings.Contains(line, "txhash:") {
				txHash = strings.TrimSpace(strings.Split(line, "txhash:")[1])
				break
			}
		}
	}

	if txHash != "" {
		log.Printf("\n🔍 Transaction hash: %s", txHash)
		log.Println("Waiting for transaction confirmation...")
		time.Sleep(5 * time.Second) // Increased wait time

		// Query transaction status
		queryTxCmd := exec.Command("imuad",
			"--home", requiredVars["IMUA_HOME_DIR"],
			"query", "tx", txHash,
			"--node", requiredVars["IMUA_COS_GRPC_URL"],
			"--output", "json",
		)

		queryTxOutput, queryErr := queryTxCmd.Output()
		if queryErr != nil {
			log.Printf("⚠️ Could not verify transaction status: %v", queryErr)
			log.Println("Transaction was submitted but confirmation failed. Check chain explorer later.")
		} else {
			if strings.Contains(string(queryTxOutput), `"code":0`) {
				log.Println("🎉 Transaction confirmed successfully!")
			} else {
				log.Printf("Transaction details: %s", string(queryTxOutput))
				log.Println("⚠️ Transaction may have failed - check details above")
			}
		}
	}

	// Step 4: Final verification
	log.Println("\n🔎 Verifying AVS opt-in status...")
	getAddrCmd := exec.Command("imuad",
		"--home", requiredVars["IMUA_HOME_DIR"],
		"keys", "show", "-a", requiredVars["IMUA_ACCOUNT_KEY_NAME"],
		"--keyring-backend", "file",
	)
	getAddrCmd.Stdin = strings.NewReader(requiredVars["KEYRING_PASSWORD"] + "\n")

	addrOutput, err := getAddrCmd.Output()
	if err != nil {
		log.Printf("⚠️ Failed to get operator address: %v", err)
	} else {
		operatorAddr := strings.TrimSpace(string(addrOutput))
		log.Printf("Operator Address: %s", operatorAddr)

		// Check AVS list
		avsListCmd := exec.Command("imuad",
			"--home", requiredVars["IMUA_HOME_DIR"],
			"query", "operator", "get-avs-list",
			operatorAddr,
			"--node", requiredVars["IMUA_COS_GRPC_URL"],
			"--output", "json",
		)

		if avsListOutput, err := avsListCmd.Output(); err == nil {
			if strings.Contains(string(avsListOutput), requiredVars["AVS_ADDRESS"]) {
				log.Println("✅ Successfully opted into AVS!")
			} else {
				log.Println("⚠️ AVS not yet visible in operator's AVS list")
				log.Println("This may take some time to propagate. Check again later.")
			}
		} else {
			log.Printf("⚠️ Failed to query AVS list: %v", err)
		}
	}

	log.Println("\n✨ AVS opt-in process completed!")
	log.Println("Next steps:")
	log.Println("- Monitor your validator status")
	log.Println("- Ensure sufficient self-delegation")
	log.Println("- Check AVS dashboard for confirmation")

	return nil
}
