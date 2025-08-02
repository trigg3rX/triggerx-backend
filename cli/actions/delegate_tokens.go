package actions

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/trigg3rX/triggerx-backend/cli/core/config"
	"github.com/urfave/cli"
)

func DelegateTokens(ctx *cli.Context) error {
	log.Println("🤝 Starting token delegation process...")

	// Initialize config from environment variables
	err := config.Init()
	if err != nil {
		return fmt.Errorf("failed to initialize config: %v", err)
	}

	// Check required environment variables
	ethPrivateKey := os.Getenv("OPERATOR_PRIVATE_KEY")
	if ethPrivateKey == "" {
		return fmt.Errorf("OPERATOR_PRIVATE_KEY environment variable not set")
	}

	// Add 0x prefix if not present
	if !strings.HasPrefix(ethPrivateKey, "0x") {
		ethPrivateKey = "0x" + ethPrivateKey
	}

	ethAddress := config.GetOperatorAddress().Hex()
	log.Printf("Using Ethereum address: %s", ethAddress)

	// Configuration with official contract addresses
	const (
		// Official contract addresses on Ethereum Sepolia
		IM_ETH_ADDR      = "0xF79F563571f7D8122611D0219A0d5449B5304F79"
		WST_ETH_ADDR     = "0xB82381A3fBD3FaFA77B3a7bE693342618240067b"
		GATEWAY_ADDR     = "0x64B5B5A618072C1E4D137f91Af780e3B17A81f3f"
		ETHEREUM_RPC_URL = "https://eth-sepolia.g.alchemy.com/v2/U67yWPtGvZIz8FwnTcFEfERypsxYzfdR"
	)

	// User configuration
	tokenAddr := IM_ETH_ADDR // Default to exoETH
	if tokenOverride := os.Getenv("TOKEN_ADDRESS"); tokenOverride != "" {
		tokenAddr = tokenOverride
	}

	// Get delegation amount from environment or use default
	delegateAmountStr := os.Getenv("DELEGATE_AMOUNT")
	if delegateAmountStr == "" {
		delegateAmountStr = os.Getenv("DEPOSIT_AMOUNT") // Use deposit amount if delegate amount not specified
		if delegateAmountStr == "" {
			delegateAmountStr = "25" // Default 25 tokens as per documentation
		}
	}

	delegateAmount, err := strconv.ParseFloat(delegateAmountStr, 64)
	if err != nil {
		return fmt.Errorf("invalid DELEGATE_AMOUNT: %v", err)
	}

	// Imua configuration
	imuaCosGrpcURL := os.Getenv("IMUA_COS_GRPC_URL")
	if imuaCosGrpcURL == "" {
		imuaCosGrpcURL = "https://api-cosmos-grpc.exocore-restaking.com:443" // Default fallback
	}

	imuaHomeDir := os.Getenv("IMUA_HOME_DIR")
	imuaAccountKeyName := os.Getenv("IMUA_ACCOUNT_KEY_NAME")

	if imuaHomeDir == "" || imuaAccountKeyName == "" {
		return fmt.Errorf("IMUA_HOME_DIR and IMUA_ACCOUNT_KEY_NAME environment variables must be set")
	}

	// Get keyring password from environment
	keyringPassword := os.Getenv("KEYRING_PASSWORD")
	if keyringPassword == "" {
		return fmt.Errorf("KEYRING_PASSWORD environment variable not set")
	}

	log.Printf("📋 Delegation Configuration:")
	log.Printf("- Token Address: %s", tokenAddr)
	log.Printf("- Delegation Amount: %.2f", delegateAmount)
	log.Printf("- Ethereum RPC: %s", ETHEREUM_RPC_URL)
	log.Printf("- Gateway Address: %s", GATEWAY_ADDR)

	// Step 1: Get Imua operator address for delegation
	log.Println("\n🏠 Step 1: Getting Imua operator address...")
	getAddrCmd := exec.Command("imuad",
		"--home", imuaHomeDir,
		"keys", "show", "-a", imuaAccountKeyName,
	)

	// Provide password for address query
	getAddrCmd.Stdin = strings.NewReader(keyringPassword + "\n")

	addrOutput, err := getAddrCmd.Output()
	if err != nil {
		return fmt.Errorf("failed to get Imua operator address: %v", err)
	}

	imAddress := strings.TrimSpace(string(addrOutput))
	if imAddress == "" {
		return fmt.Errorf("empty Imua operator address returned")
	}
	log.Printf("Delegating to Imua address: %s", imAddress)

	// Step 2: Check if gateway is bootstrapped
	log.Println("\n🔍 Step 2: Checking bootstrap status...")
	bootstrapCmd := exec.Command("cast", "call",
		"--rpc-url", ETHEREUM_RPC_URL,
		GATEWAY_ADDR,
		"bootstrapped() returns (bool)",
	)

	bootstrapOutput, err := bootstrapCmd.Output()
	isBootstrapped := false
	if err != nil {
		log.Printf("Warning: Could not check bootstrap status: %v", err)
		log.Println("Assuming network is not bootstrapped...")
	} else {
		bootstrapResult := strings.TrimSpace(string(bootstrapOutput))
		isBootstrapped = bootstrapResult == "true"
	}
	log.Printf("Bootstrap status: %t", isBootstrapped)

	// Step 3: Calculate delegation LayerZero fees if bootstrapped
	delegateValue := "0"
	if isBootstrapped {
		log.Println("\n💸 Step 3: Calculating delegation LayerZero fees...")

		// Convert amount to wei first for proper hex conversion
		amountWeiCmd := exec.Command("cast", "2w", fmt.Sprintf("%.0f", delegateAmount))
		amountWeiOutput, err := amountWeiCmd.Output()
		if err != nil {
			return fmt.Errorf("failed to convert amount to wei: %v", err)
		}
		amountWei := strings.TrimSpace(string(amountWeiOutput))

		// Convert wei to hex
		amountHexCmd := exec.Command("cast", "2h", amountWei)
		amountHexOutput, err := amountHexCmd.Output()
		if err != nil {
			return fmt.Errorf("failed to convert amount to hex: %v", err)
		}
		amountHex := strings.TrimSpace(string(amountHexOutput))

		// Remove 0x prefix and pad to 64 characters (32 bytes)
		amountHex = strings.TrimPrefix(amountHex, "0x")
		// Pad with leading zeros to make it 64 characters
		for len(amountHex) < 64 {
			amountHex = "0" + amountHex
		}
		quantityB32 := "0x" + amountHex

		// Generate the LZ message for delegation
		delegatePrefix := "0x08"

		// Convert token address to bytes32
		tokenB32Cmd := exec.Command("cast", "2b", tokenAddr)
		tokenB32Output, err := tokenB32Cmd.Output()
		if err != nil {
			return fmt.Errorf("failed to convert token address to bytes32: %v", err)
		}
		tokenB32 := strings.TrimSpace(string(tokenB32Output))

		// Convert ETH address to bytes32
		ethAddressB32Cmd := exec.Command("cast", "2b", ethAddress)
		ethAddressB32Output, err := ethAddressB32Cmd.Output()
		if err != nil {
			return fmt.Errorf("failed to convert ETH address to bytes32: %v", err)
		}
		ethAddressB32 := strings.TrimSpace(string(ethAddressB32Output))

		// Convert Imua address to bytes
		imAddressBytesCmd := exec.Command("cast", "fu", imAddress)
		imAddressBytesOutput, err := imAddressBytesCmd.Output()
		if err != nil {
			return fmt.Errorf("failed to convert Imua address to bytes: %v", err)
		}
		imAddressBytes := strings.TrimSpace(string(imAddressBytesOutput))

		// Create delegation LZ message
		lzDelegateMessageCmd := exec.Command("cast", "ch",
			delegatePrefix,
			tokenB32,
			ethAddressB32,
			imAddressBytes,
			quantityB32,
		)
		lzDelegateMessageOutput, err := lzDelegateMessageCmd.Output()
		if err != nil {
			return fmt.Errorf("failed to create delegation LZ message: %v", err)
		}
		lzDelegateMessage := strings.TrimSpace(string(lzDelegateMessageOutput))
		log.Printf("Generated LZ message: %s", lzDelegateMessage)

		// Quote delegation LayerZero fees
		quoteDelegateCmd := exec.Command("cast", "call",
			"--rpc-url", ETHEREUM_RPC_URL,
			GATEWAY_ADDR,
			"quote(bytes)",
			lzDelegateMessage,
		)

		quoteDelegateOutput, err := quoteDelegateCmd.Output()
		if err != nil {
			log.Printf("Warning: Failed to quote delegation LayerZero fees: %v", err)
			log.Println("This might be due to network issues or incorrect message format.")
			return fmt.Errorf("LayerZero fee calculation failed for delegation: %v", err)
		}

		delegateValue = strings.TrimSpace(string(quoteDelegateOutput))
		log.Printf("Delegation LayerZero fee calculated: %s", delegateValue)

		// Validate that we got a proper fee value
		if delegateValue == "" || delegateValue == "0x" {
			log.Printf("Warning: Received empty or invalid LayerZero fee: '%s'", delegateValue)
			log.Println("Proceeding with 0 value, but this might cause transaction failure")
			delegateValue = "0"
		}
	} else {
		log.Println("\n⚡ Step 3: Network not bootstrapped, no LayerZero fees required")
	}

	log.Printf("Delegation LayerZero value: %s", delegateValue)

	// Step 4: Execute delegation transaction
	log.Println("\n📤 Step 4: Executing delegation transaction...")

	// Convert amount to wei for transaction
	amountWeiCmd := exec.Command("cast", "2w", fmt.Sprintf("%.0f", delegateAmount))
	amountWeiOutput, err := amountWeiCmd.Output()
	if err != nil {
		return fmt.Errorf("failed to convert amount to wei: %v", err)
	}
	amountWei := strings.TrimSpace(string(amountWeiOutput))

	// Convert delegation value to decimal if needed
	var finalDelegateValue string
	if delegateValue != "0" && delegateValue != "" {
		if strings.HasPrefix(delegateValue, "0x") {
			// Convert hex to decimal
			castDecimalCmd := exec.Command("cast", "2d", delegateValue)
			castDecimalOutput, err := castDecimalCmd.Output()
			if err != nil {
				log.Printf("Warning: Failed to convert LayerZero fee to decimal: %v", err)
				log.Printf("Using raw value: %s", delegateValue)
				finalDelegateValue = delegateValue
			} else {
				finalDelegateValue = strings.TrimSpace(string(castDecimalOutput))
			}
		} else {
			// Already in decimal format
			finalDelegateValue = delegateValue
		}
	} else {
		finalDelegateValue = "0"
	}

	log.Printf("Final delegation value for transaction: %s", finalDelegateValue)

	// Execute the delegation transaction
	delegateCmd := exec.Command("cast", "send",
		"--rpc-url", ETHEREUM_RPC_URL,
		GATEWAY_ADDR,
		"delegateTo(string,address,uint256)",
		imAddress,
		tokenAddr,
		amountWei,
		"--private-key", ethPrivateKey,
		"--value", finalDelegateValue,
	)

	log.Printf("Executing delegation command: %s", strings.Join(delegateCmd.Args, " "))
	delegateOutput, err := delegateCmd.CombinedOutput()
	if err != nil {
		log.Printf("Delegation command output: %s", string(delegateOutput))
		return fmt.Errorf("failed to execute delegation transaction: %v\nOutput: %s", err, string(delegateOutput))
	}

	transactionOutput := strings.TrimSpace(string(delegateOutput))
	log.Printf("✅ Delegation transaction successful!")
	log.Printf("Transaction details: %s", transactionOutput)

	// Extract transaction hash if available
	var txHash string
	lines := strings.Split(transactionOutput, "\n")
	for _, line := range lines {
		if strings.Contains(line, "transactionHash") || strings.Contains(line, "blockHash") {
			log.Printf("📝 %s", line)
		}
		// Try to extract transaction hash
		if strings.HasPrefix(strings.TrimSpace(line), "0x") && len(strings.TrimSpace(line)) == 66 {
			txHash = strings.TrimSpace(line)
		}
	}

	// Step 5: Provide verification instructions
	log.Println("\n🎉 Token delegation completed successfully!")
	log.Printf("📊 Summary:")
	log.Printf("  - Token: %s", tokenAddr)
	log.Printf("  - Amount: %.2f tokens (%s wei)", delegateAmount, amountWei)
	log.Printf("  - Delegated from: %s", ethAddress)
	log.Printf("  - Delegated to: %s", imAddress)
	log.Printf("  - Network: %s", "Ethereum Sepolia -> Imuachain")
	log.Printf("  - LayerZero Fee: %s", finalDelegateValue)

	if isBootstrapped {
		log.Println("\n📋 Verification Instructions:")
		log.Println("1. 🔍 Monitor LayerZero message processing:")
		log.Println("   - Visit: https://testnet.layerzeroscan.com/")
		if txHash != "" {
			log.Printf("   - Search for transaction hash: %s", txHash)
		} else {
			log.Println("   - Search for your transaction hash from the output above")
		}

		ethLZID := 40161 // Sepolia LayerZero ID
		stakerID := strings.ToLower(ethAddress) + "_" + fmt.Sprintf("0x%x", ethLZID)
		log.Println("\n2. 📊 Verify your delegation on Imua:")
		log.Printf("   Staker ID: %s", stakerID)
		log.Printf("   Query command:")
		log.Printf("   imuad query assets staker-assets %s --node %s --output json | jq", stakerID, imuaCosGrpcURL)

		log.Println("\n3. 🤝 Expected result should show:")
		log.Printf("   - total_deposit_amount: (unchanged)")
		log.Printf("   - withdrawable_amount: (reduced by %s)", amountWei)
		log.Printf("   - pending_undelegation_amount: 0")

		log.Println("\n⏱️  Note: LayerZero message processing may take a few minutes")
		log.Println("   Please wait before checking the delegation status on Imua")
	} else {
		log.Println("\n💡 Note: Network is in bootstrap mode")
		log.Println("- Your delegation will be processed when the network becomes active")
		log.Println("- No LayerZero fees were charged")
		log.Println("- The delegation is immediately effective in bootstrap mode")
	}

	log.Println("\n📋 Next Steps:")
	log.Println("1. 🔗 Associate operator for post-bootstrap phase:")
	log.Println("   ./triggerx associate-operator")
	log.Println("2. 🚀 Proceed with AVS opt-in:")
	log.Println("   ./triggerx opt-in-to-avs")
	log.Println("3. ⚡ Complete TriggerX registration:")
	log.Println("   ./triggerx complete-registration")

	log.Println("\n💡 Important Notes:")
	log.Println("- Ensure you have at least 1,000 USD in self-delegated tokens")
	log.Println("- Validator inclusion happens at the beginning of the next epoch (max 1 hour)")
	log.Println("- Only top 50 validators by total stake are included in the active set")

	return nil
}
