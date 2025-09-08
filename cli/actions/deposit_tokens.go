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

func DepositTokens(ctx *cli.Context) error {
	log.Println("💰 Starting token deposit process on Ethereum Sepolia...")

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
	tokenAddr := IM_ETH_ADDR // Default to imETH
	if tokenOverride := os.Getenv("TOKEN_ADDRESS"); tokenOverride != "" {
		tokenAddr = tokenOverride
	}

	// Get deposit amount from environment or use default
	depositAmountStr := os.Getenv("DEPOSIT_AMOUNT")
	if depositAmountStr == "" {
		depositAmountStr = "25" // Default 25 tokens as per documentation
	}

	depositAmount, err := strconv.ParseFloat(depositAmountStr, 64)
	if err != nil {
		return fmt.Errorf("invalid DEPOSIT_AMOUNT: %v", err)
	}

	// Imua configuration for verification
	imuaCosGrpcURL := os.Getenv("IMUA_COS_GRPC_URL")
	if imuaCosGrpcURL == "" {
		imuaCosGrpcURL = "https://api-cosmos-grpc.exocore-restaking.com:443"
	}

	log.Printf("📋 Deposit Configuration:")
	log.Printf("- Token Address: %s", tokenAddr)
	log.Printf("- Deposit Amount: %.2f", depositAmount)
	log.Printf("- Ethereum RPC: %s", ETHEREUM_RPC_URL)
	log.Printf("- Gateway Address: %s", GATEWAY_ADDR)

	// Step 1: Get vault address for the token
	log.Println("\n💰 Step 1: Getting vault address for token...")
	vaultCmd := exec.Command("cast", "call",
		"--rpc-url", ETHEREUM_RPC_URL,
		GATEWAY_ADDR,
		"tokenToVault(address) returns (address)",
		tokenAddr,
	)

	vaultOutput, err := vaultCmd.Output()
	if err != nil {
		return fmt.Errorf("failed to get vault address: %v", err)
	}

	vaultAddr := strings.TrimSpace(string(vaultOutput))
	log.Printf("Vault address: %s", vaultAddr)

	// Step 2: Approve token spending by vault
	log.Println("\n✅ Step 2: Approving token spending by vault...")

	// First get the maximum uint256 value
	maxUintCmd := exec.Command("cast", "maxu")
	maxUintOutput, err := maxUintCmd.Output()
	if err != nil {
		return fmt.Errorf("failed to get max uint256: %v", err)
	}
	maxUint := strings.TrimSpace(string(maxUintOutput))

	approveCmd := exec.Command("cast", "send",
		"--rpc-url", ETHEREUM_RPC_URL,
		tokenAddr,
		"approve(address,uint256)",
		vaultAddr,
		maxUint,
		"--private-key", ethPrivateKey,
	)

	approveOutput, err := approveCmd.CombinedOutput()
	if err != nil {
		log.Printf("Approve command output: %s", string(approveOutput))
		return fmt.Errorf("failed to approve token: %v", err)
	}

	log.Printf("Token approval successful: %s", string(approveOutput))

	// Step 3: Check if gateway is bootstrapped
	log.Println("\n🔍 Step 3: Checking bootstrap status...")
	bootstrapCmd := exec.Command("cast", "call",
		"--rpc-url", ETHEREUM_RPC_URL,
		GATEWAY_ADDR,
		"bootstrapped() returns (bool)",
	)

	bootstrapOutput, err := bootstrapCmd.Output()
	if err != nil {
		log.Printf("Warning: Could not check bootstrap status: %v", err)
		bootstrapOutput = []byte("false")
	}

	isBootstrapped := strings.TrimSpace(string(bootstrapOutput)) == "true"
	log.Printf("Bootstrap status: %t", isBootstrapped)

	// Step 4: Calculate LayerZero fees if bootstrapped
	depositValue := "0"
	if isBootstrapped {
		log.Println("\n💸 Step 4: Calculating LayerZero fees...")

		// Generate the LZ message to calculate the fees
		depositPrefix := "0x02"

		// Convert token address to bytes32
		tokenB32Cmd := exec.Command("cast", "2b", tokenAddr)
		tokenB32Output, err := tokenB32Cmd.Output()
		if err != nil {
			return fmt.Errorf("failed to convert token address: %v", err)
		}
		tokenB32 := strings.TrimSpace(string(tokenB32Output))

		// Convert ETH address to bytes32
		ethAddressB32Cmd := exec.Command("cast", "2b", ethAddress)
		ethAddressB32Output, err := ethAddressB32Cmd.Output()
		if err != nil {
			return fmt.Errorf("failed to convert ETH address: %v", err)
		}
		ethAddressB32 := strings.TrimSpace(string(ethAddressB32Output))

		// Convert amount to wei and then to hex
		amountWeiCmd := exec.Command("cast", "2w", fmt.Sprintf("%.0f", depositAmount))
		amountWeiOutput, err := amountWeiCmd.Output()
		if err != nil {
			return fmt.Errorf("failed to convert amount to wei: %v", err)
		}

		amountHexCmd := exec.Command("cast", "2h", strings.TrimSpace(string(amountWeiOutput)))
		amountHexOutput, err := amountHexCmd.Output()
		if err != nil {
			return fmt.Errorf("failed to convert amount to hex: %v", err)
		}

		// Pad the amount to 64 characters (32 bytes)
		amountHexStr := strings.TrimSpace(string(amountHexOutput))
		amountHexStr = strings.TrimPrefix(amountHexStr, "0x")
		quantityB32 := fmt.Sprintf("%064s", amountHexStr)
		quantityB32 = strings.ReplaceAll(quantityB32, " ", "0")

		// Concatenate the LZ message
		lzMessageCmd := exec.Command("cast", "ch", depositPrefix, tokenB32, ethAddressB32, quantityB32)
		lzMessageOutput, err := lzMessageCmd.Output()
		if err != nil {
			return fmt.Errorf("failed to create LZ message: %v", err)
		}
		lzMessage := strings.TrimSpace(string(lzMessageOutput))

		// Quote the LayerZero fees
		quoteCmd := exec.Command("cast", "call",
			"--rpc-url", ETHEREUM_RPC_URL,
			GATEWAY_ADDR,
			"quote(bytes)",
			lzMessage,
		)

		quoteOutput, err := quoteCmd.Output()
		if err != nil {
			log.Printf("Warning: Failed to quote LayerZero fees: %v", err)
		} else {
			depositValue = strings.TrimSpace(string(quoteOutput))
		}
	} else {
		log.Println("\n⚡ Step 4: Network not bootstrapped, no LayerZero fees required")
	}

	log.Printf("LayerZero value: %s", depositValue)

	// Step 5: Execute deposit transaction
	log.Println("\n📤 Step 5: Executing deposit transaction...")

	// Convert amount to wei for transaction
	amountWeiCmd := exec.Command("cast", "2w", fmt.Sprintf("%.0f", depositAmount))
	amountWeiOutput, err := amountWeiCmd.Output()
	if err != nil {
		return fmt.Errorf("failed to convert amount to wei: %v", err)
	}
	amountWei := strings.TrimSpace(string(amountWeiOutput))

	// Convert value to decimal if needed
	var finalValue string
	if depositValue != "0" {
		castDecimalCmd := exec.Command("cast", "2d", depositValue)
		castDecimalOutput, err := castDecimalCmd.Output()
		if err != nil {
			finalValue = depositValue // Use as is if conversion fails
		} else {
			finalValue = strings.TrimSpace(string(castDecimalOutput))
		}
	} else {
		finalValue = "0"
	}

	depositCmd := exec.Command("cast", "send",
		"--rpc-url", ETHEREUM_RPC_URL,
		GATEWAY_ADDR,
		"deposit(address,uint256)",
		tokenAddr,
		amountWei,
		"--private-key", ethPrivateKey,
		"--value", finalValue,
	)

	depositOutput, err := depositCmd.CombinedOutput()
	if err != nil {
		log.Printf("Deposit command output: %s", string(depositOutput))
		return fmt.Errorf("failed to execute deposit: %v", err)
	}

	log.Printf("✅ Deposit successful: %s", string(depositOutput))

	// Step 6: Provide verification instructions
	log.Println("\n🎉 Token deposit completed successfully!")
	log.Printf("📊 Summary:")
	log.Printf("  - Token: %s", tokenAddr)
	log.Printf("  - Amount: %.2f tokens", depositAmount)
	log.Printf("  - Deposited from: %s", ethAddress)
	log.Printf("  - Network: %s", "Ethereum Sepolia -> Imuachain")

	if isBootstrapped {
		log.Println("\n📋 Verification Instructions:")
		log.Println("1. 🔍 Monitor LayerZero message processing:")
		log.Println("   - Visit: https://testnet.layerzeroscan.com/")
		log.Println("   - Search for your transaction hash above")

		ethLZID := 40161 // Sepolia LayerZero ID
		stakerID := strings.ToLower(ethAddress) + "_" + fmt.Sprintf("0x%x", ethLZID)
		log.Println("2. 📊 Verify your deposit on Imua:")
		log.Printf("   Staker ID: %s", stakerID)
		log.Printf("   Query: imuad query assets staker-assets %s --node %s --output json | jq", stakerID, imuaCosGrpcURL)

		log.Println("3. 💰 Expected result should show:")
		log.Printf("   - total_deposit_amount: %s (%.2f tokens in wei)", amountWei, depositAmount)
		log.Printf("   - withdrawable_amount: %s (all deposited tokens)", amountWei)
	} else {
		log.Println("\n💡 Note: Network is in bootstrap mode")
		log.Println("- Your deposit will be processed when the network becomes active")
		log.Println("- No LayerZero fees were charged")
	}

	log.Println("\n📋 Next Steps:")
	log.Println("1. 🤝 Delegate tokens to your validator:")
	log.Println("   ./triggerx delegate-tokens")
	log.Println("2. 🚀 Or run complete process:")
	log.Println("   ./triggerx deposit-and-delegate")

	return nil
}
