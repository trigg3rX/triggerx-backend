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

// OptInToAVS handles the AVS opt-in process using cast send
func OptInToAVS(ctx *cli.Context) error {
	log.Println("Starting AVS opt-in process...")

	// Initialize config from environment variables
	if err := config.Init(); err != nil {
		return fmt.Errorf("failed to initialize config: %v", err)
	}

	// Validate required environment variables
	requiredVars := map[string]string{
		"AVS_ADDRESS":      os.Getenv("AVS_ADDRESS"),
		"IMUA_PRIVATE_KEY": os.Getenv("IMUA_PRIVATE_KEY"),
		"RPC_URL":          os.Getenv("RPC_URL"),
	}

	// Set default values if not provided
	if requiredVars["AVS_ADDRESS"] == "" {
		requiredVars["AVS_ADDRESS"] = "0x72A5016ECb9EB01d7d54ae48bFFB62CA0B8e57a5"
	}
	if requiredVars["RPC_URL"] == "" {
		requiredVars["RPC_URL"] = "https://api-eth.exocore-restaking.com"
	}

	// Check required variables
	for name, value := range requiredVars {
		if value == "" {
			return fmt.Errorf("%s environment variable not set", name)
		}
	}

	// Log configuration (redact private key)
	log.Printf("AVS opt-in configuration:")
	for name, value := range requiredVars {
		if name != "IMUA_PRIVATE_KEY" {
			log.Printf("- %s: %s", name, value)
		} else {
			log.Printf("- %s: [redacted]", name)
		}
	}

	// Check if cast is available
	if err := checkCastAvailable(); err != nil {
		return fmt.Errorf("cast command not available: %v", err)
	}

	// Execute the cast send command
	log.Println("\n🚀 Executing registerOperatorToAVS transaction...")

	castCmd := exec.Command("cast", "send",
		requiredVars["AVS_ADDRESS"],
		"registerOperatorToAVS()",
		"--rpc-url", requiredVars["RPC_URL"],
		"--private-key", requiredVars["IMUA_PRIVATE_KEY"],
		"--gas-limit", "1000000",
	)

	// Execute the command and capture output
	castOutput, err := castCmd.CombinedOutput()
	if err != nil {
		log.Printf("❌ Cast send command output: %s", string(castOutput))

		// Handle specific error cases
		if strings.Contains(string(castOutput), "insufficient funds") {
			return fmt.Errorf("insufficient ETH balance for gas fees")
		}
		if strings.Contains(string(castOutput), "execution reverted") {
			return fmt.Errorf("transaction reverted - check if already registered or meet requirements")
		}
		if strings.Contains(string(castOutput), "invalid private key") {
			return fmt.Errorf("invalid private key format")
		}

		return fmt.Errorf("failed to execute registerOperatorToAVS: %v", err)
	}

	log.Printf("✅ Transaction submitted: %s", string(castOutput))

	// Extract transaction hash from output
	var txHash string
	outputStr := string(castOutput)

	// Cast typically outputs just the transaction hash
	lines := strings.Split(strings.TrimSpace(outputStr), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		// Look for a hex string that looks like a transaction hash (0x followed by 64 hex chars)
		if strings.HasPrefix(line, "0x") && len(line) == 66 {
			txHash = line
			break
		}
	}

	if txHash != "" {
		log.Printf("\n🔍 Transaction hash: %s", txHash)
		log.Println("Waiting for transaction confirmation...")
		time.Sleep(10 * time.Second) // Wait for block confirmation

		// Verify transaction status using cast
		if err := verifyTransaction(txHash, requiredVars["RPC_URL"]); err != nil {
			log.Printf("⚠️ Could not verify transaction status: %v", err)
			log.Println("Transaction was submitted but confirmation failed. Check block explorer.")
		}
	} else {
		log.Println("⚠️ Could not extract transaction hash from output")
		log.Println("Transaction may have been submitted. Check your wallet or block explorer.")
	}

	log.Println("\n✨ AVS opt-in process completed!")
	log.Println("Next steps:")
	log.Println("- Check transaction on block explorer")
	log.Println("- Monitor your operator status on AVS dashboard")
	log.Printf("- Transaction hash: %s", txHash)

	return nil
}

// checkCastAvailable verifies that the cast command is available
func checkCastAvailable() error {
	cmd := exec.Command("cast", "--help")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("cast command not found. Please install Foundry toolkit: https://getfoundry.sh/")
	}
	return nil
}

// verifyTransaction checks if the transaction was successful
func verifyTransaction(txHash, rpcURL string) error {
	log.Println("🔍 Verifying transaction status...")

	// Use cast to get transaction receipt
	receiptCmd := exec.Command("cast", "receipt", txHash, "--rpc-url", rpcURL)
	receiptOutput, err := receiptCmd.Output()
	if err != nil {
		return fmt.Errorf("failed to get transaction receipt: %v", err)
	}

	receiptStr := string(receiptOutput)

	// Check if transaction was successful (status: 0x1 means success)
	if strings.Contains(receiptStr, "status") && strings.Contains(receiptStr, "0x1") {
		log.Println("🎉 Transaction confirmed successfully!")
		return nil
	} else if strings.Contains(receiptStr, "status") && strings.Contains(receiptStr, "0x0") {
		log.Println("❌ Transaction failed!")
		log.Printf("Receipt: %s", receiptStr)
		return fmt.Errorf("transaction failed with status 0x0")
	}

	log.Println("⚠️ Transaction status unclear")
	log.Printf("Receipt: %s", receiptStr)
	return nil
}
