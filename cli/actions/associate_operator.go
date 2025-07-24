package actions

import (
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/trigg3rX/triggerx-backend/cli/core/config"
	"github.com/urfave/cli"
)

type AssociationResponse struct {
	Operator string `json:"operator"`
}

func AssociateOperator(ctx *cli.Context) error {
	log.Println("🚀 Starting operator association process (Post-Bootstrap Phase)")
	log.Println("⏳ This may take a few moments...")

	// Initialize configuration
	if err := config.Init(); err != nil {
		return fmt.Errorf("configuration initialization failed: %v", err)
	}

	// Validate environment variables
	ethPrivateKey := os.Getenv("OPERATOR_PRIVATE_KEY")
	if ethPrivateKey == "" {
		return fmt.Errorf("OPERATOR_PRIVATE_KEY must be set in .env file")
	}

	imuaPrivateKey := os.Getenv("IMUA_PRIVATE_KEY")
	if imuaPrivateKey == "" {
		return fmt.Errorf("IMUA_PRIVATE_KEY must be set in .env file\n" +
			"Get it by running: 'imuad keys unsafe-export-eth-key YOUR_ACCOUNT_NAME'")
	}

	imuaHomeDir := os.Getenv("IMUA_HOME_DIR")
	if imuaHomeDir == "" {
		imuaHomeDir = filepath.Join(os.Getenv("HOME"), ".imuad")
	}

	imuaAccountKeyName := os.Getenv("IMUA_ACCOUNT_KEY_NAME")
	if imuaAccountKeyName == "" {
		return fmt.Errorf("IMUA_ACCOUNT_KEY_NAME must be set in .env file")
	}

	// Set default URLs if not provided
	imuaEthRPCURL := os.Getenv("IMUA_ETH_RPC_URL")
	if imuaEthRPCURL == "" {
		imuaEthRPCURL = "https://api-eth.exocore-restaking.com"
	}

	imuaCosGrpcURL := os.Getenv("IMUA_COS_GRPC_URL")
	if imuaCosGrpcURL == "" {
		imuaCosGrpcURL = "https://api-cosmos-grpc.exocore-restaking.com:443"
	}

	// Constants
	const (
		ethLZID         = 40161 // Sepolia LayerZero ID
		imuaGatewayAddr = "0xdDf5218Dbff297ADdF17fB7977E2469D774545ED"
	)

	// Get Ethereum address from private key
	ethAddress := config.GetOperatorAddress().Hex()
	log.Printf("🔑 Ethereum Address: %s", ethAddress)

	// 1. Check balance
	log.Println("\n🔍 Step 1: Checking Ethereum address balance...")
	balance, err := getBalance(ethAddress, imuaEthRPCURL)
	if err != nil {
		return fmt.Errorf("failed to check balance: %v", err)
	}
	log.Printf("💰 Current Balance: %s hua", balance.String())

	// Calculate required transfer amount (0.5 IMUA in wei)
	transferAmount := new(big.Int).Mul(big.NewInt(5), big.NewInt(1e17)) // 0.5 * 10^18

	// 2. Transfer funds if needed
	if balance.Cmp(transferAmount) < 0 {
		log.Printf("\n🔄 Step 2: Transferring 0.5 IMUA to %s...", ethAddress)

		if err := transferFunds(imuaEthRPCURL, ethAddress, transferAmount.String(), imuaPrivateKey); err != nil {
			return fmt.Errorf("funds transfer failed: %v", err)
		}

		// Wait for transfer to complete
		time.Sleep(5 * time.Second)
	} else {
		log.Println("✅ Sufficient balance detected")
	}

	// 3. Get Imua operator address
	log.Println("\n📝 Step 3: Retrieving Imua operator address...")
	imOperatorAddr, err := getImuaOperatorAddress(imuaHomeDir, imuaAccountKeyName)
	if err != nil {
		return fmt.Errorf("failed to get operator address: %v", err)
	}
	log.Printf("🏷️ Operator Address: %s", imOperatorAddr)

	// 4. Check existing association
	log.Println("\n🔎 Step 4: Checking for existing association...")
	stakerID := fmt.Sprintf("%s_0x%x", strings.ToLower(ethAddress), ethLZID)

	if existingOp, err := checkExistingAssociation(stakerID, imuaCosGrpcURL); err == nil && existingOp != "" {
		if existingOp == imOperatorAddr {
			log.Println("✅ Association already exists!")
			printSuccess(ethAddress, imOperatorAddr, stakerID)
			return nil
		}
		return fmt.Errorf("existing association with different operator: %s", existingOp)
	}

	// 5. Create association
	log.Println("\n⚡ Step 5: Creating operator association...")
	if err := createAssociation(imuaEthRPCURL, imuaGatewayAddr, ethLZID, imOperatorAddr, ethPrivateKey); err != nil {
		return fmt.Errorf("association failed: %v", err)
	}

	// 6. Verify association
	log.Println("\n🔍 Step 6: Verifying association...")
	time.Sleep(3 * time.Second) // Wait for state update

	if verifiedOp, err := checkExistingAssociation(stakerID, imuaCosGrpcURL); err != nil {
		log.Printf("⚠️ Verification failed (check later with): imuad query delegation associated-operator-by-staker %s --node %s",
			stakerID, imuaCosGrpcURL)
	} else if verifiedOp == imOperatorAddr {
		log.Println("✅ Association verified!")
		printSuccess(ethAddress, imOperatorAddr, stakerID)
	} else {
		return fmt.Errorf("verification failed - expected %s got %s", imOperatorAddr, verifiedOp)
	}

	return nil
}

// Helper functions

func getBalance(address, rpcURL string) (*big.Int, error) {
	cmd := exec.Command("cast", "balance", address, "--rpc-url", rpcURL)
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("balance check failed: %v", err)
	}
	balanceStr := strings.TrimSpace(string(output))
	balance, ok := new(big.Int).SetString(balanceStr, 10)
	if !ok {
		return nil, fmt.Errorf("failed to parse balance: %s", balanceStr)
	}
	return balance, nil
}

func transferFunds(rpcURL, toAddress, amount, privateKey string) error {
	cmd := exec.Command("cast", "send",
		"--rpc-url", rpcURL,
		toAddress,
		"--value", amount,
		"--private-key", privateKey,
	)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("transfer failed: %v\nOutput: %s", err, string(output))
	}
	log.Printf("💸 Transfer successful: %s", string(output))
	return nil
}

func getImuaOperatorAddress(homeDir, accountName string) (string, error) {
	keyringPassword := os.Getenv("KEYRING_PASSWORD")
	if keyringPassword == "" {
		return "", fmt.Errorf("KEYRING_PASSWORD environment variable not set")
	}
	getAddrCmd := exec.Command("imuad",
		"--home", homeDir,
		"keys", "show", "-a", accountName,
	)

	// Provide password for address query
	getAddrCmd.Stdin = strings.NewReader(keyringPassword + "\n")

	addrOutput, err := getAddrCmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get validator address: %v", err)
	}

	validatorAddr := strings.TrimSpace(string(addrOutput))
	return validatorAddr, nil
}

func checkExistingAssociation(stakerID, grpcURL string) (string, error) {
	cmd := exec.Command("imuad", "query", "delegation", "associated-operator-by-staker",
		stakerID,
		"--node", grpcURL,
		"--output", "json",
	)
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("query failed: %v", err)
	}

	var resp AssociationResponse
	if err := json.Unmarshal(output, &resp); err != nil {
		return "", fmt.Errorf("failed to parse response: %v", err)
	}
	return resp.Operator, nil
}

func createAssociation(rpcURL, gatewayAddr string, lzID uint32, operatorAddr, privateKey string) error {
	cmd := exec.Command("cast", "send",
		"--rpc-url", rpcURL,
		gatewayAddr,
		"associateOperatorWithEVMStaker(uint32,string)",
		fmt.Sprintf("%d", lzID),
		operatorAddr,
		"--private-key", privateKey,
	)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("transaction failed: %v\nOutput: %s", err, string(output))
	}
	log.Printf("📝 Transaction successful: %s", string(output))
	return nil
}

func printSuccess(ethAddr, imuaAddr, stakerID string) {
	log.Println("\n🎉 Successfully Associated Operator!")
	log.Printf("├─ Ethereum Address: %s", ethAddr)
	log.Printf("├─ Imua Operator:    %s", imuaAddr)
	log.Printf("└─ Staker ID:        %s", stakerID)

	log.Println("\n💡 Next Steps:")
	log.Println("- Ensure you have at least 1,000 USD in self-delegation")
	log.Println("- Validator will be eligible at next epoch (within 1 hour)")
	log.Println("- Top 50 validators by stake will be active")
}
