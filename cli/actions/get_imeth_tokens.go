package actions

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/trigg3rX/triggerx-backend/cli/core/config"
	"github.com/urfave/cli"
)

func GetImethTokens(ctx *cli.Context) error {
	log.Println("💧 Starting imETH faucet request...")

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

	// Configuration
	const (
		IM_ETH_FAUCET_ADDR = "0xfCf79695f63FB9e540a1887A19dA94e10ADF13eB"
		IM_ETH_ADDR        = "0xF79F563571f7D8122611D0219A0d5449B5304F79"
		ETHEREUM_RPC_URL   = "https://eth-sepolia.g.alchemy.com/v2/U67yWPtGvZIz8FwnTcFEfERypsxYzfdR"
	)

	log.Printf("📋 Faucet Configuration:")
	log.Printf("- Faucet Address: %s", IM_ETH_FAUCET_ADDR)
	log.Printf("- imETH Token Address: %s", IM_ETH_ADDR)
	log.Printf("- Ethereum RPC: %s", ETHEREUM_RPC_URL)
	log.Printf("- Requesting address: %s", ethAddress)

	// Step 1: Check current imETH balance
	log.Println("\n💰 Step 1: Checking current imETH balance...")
	balanceCmd := exec.Command("cast", "call",
		"--rpc-url", ETHEREUM_RPC_URL,
		IM_ETH_ADDR,
		"balanceOf(address) returns (uint256)",
		ethAddress,
	)

	balanceOutput, err := balanceCmd.Output()
	if err != nil {
		log.Printf("Warning: Could not check balance: %v", err)
	} else {
		currentBalance := strings.TrimSpace(string(balanceOutput))
		log.Printf("Current imETH balance: %s wei", currentBalance)

		// Convert to decimal for readability
		if decimalCmd := exec.Command("cast", "2d", currentBalance); decimalCmd != nil {
			if decimalOutput, err := decimalCmd.Output(); err == nil {
				decimalBalance := strings.TrimSpace(string(decimalOutput))
				log.Printf("Current imETH balance: %s tokens", decimalBalance)
			}
		}
	}

	// Step 2: Request tokens from faucet
	log.Println("\n🚰 Step 2: Requesting 25 imETH tokens from faucet...")
	log.Println("Note: This faucet gives out 25 imETH every 24 hours per requesting address.")

	faucetCmd := exec.Command("cast", "send",
		"--rpc-url", ETHEREUM_RPC_URL,
		IM_ETH_FAUCET_ADDR,
		"requestTokens()",
		"--private-key", ethPrivateKey,
	)

	faucetOutput, err := faucetCmd.CombinedOutput()
	if err != nil {
		log.Printf("Faucet command output: %s", string(faucetOutput))
		return fmt.Errorf("failed to request tokens from faucet: %v", err)
	}

	log.Printf("✅ Faucet request successful: %s", string(faucetOutput))

	// Step 3: Check new balance
	log.Println("\n🔄 Step 3: Checking updated imETH balance...")
	newBalanceCmd := exec.Command("cast", "call",
		"--rpc-url", ETHEREUM_RPC_URL,
		IM_ETH_ADDR,
		"balanceOf(address) returns (uint256)",
		ethAddress,
	)

	newBalanceOutput, err := newBalanceCmd.Output()
	if err != nil {
		log.Printf("Warning: Could not check new balance: %v", err)
	} else {
		newBalance := strings.TrimSpace(string(newBalanceOutput))
		log.Printf("New imETH balance: %s wei", newBalance)

		// Convert to decimal for readability
		if decimalCmd := exec.Command("cast", "2d", newBalance); decimalCmd != nil {
			if decimalOutput, err := decimalCmd.Output(); err == nil {
				decimalBalance := strings.TrimSpace(string(decimalOutput))
				log.Printf("New imETH balance: %s tokens", decimalBalance)
			}
		}
	}

	log.Println("\n🎉 imETH tokens successfully acquired!")
	log.Println("📋 Next Steps:")
	log.Println("1. 💰 You can now deposit tokens:")
	log.Println("   ./triggerx deposit-tokens")
	log.Println("2. 🚀 Or run the complete process:")
	log.Println("   ./triggerx deposit-and-delegate")
	log.Println("")
	log.Println("💡 Note: The faucet has a 24-hour cooldown period per address.")

	return nil
}
