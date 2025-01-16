package cmd

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"os"
	"path/filepath"
	// "strconv"
	"time"

	"github.com/urfave/cli/v2"
	"gopkg.in/yaml.v3"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/Layr-Labs/eigensdk-go/crypto/ecdsa"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
	"github.com/trigg3rX/triggerx-backend/pkg/types"

	// erc20 "github.com/Layr-Labs/eigensdk-go/contracts/bindings/IERC20"
	// strategy "github.com/Layr-Labs/eigensdk-go/contracts/bindings/IStrategy"
	// strategymanager "github.com/Layr-Labs/eigensdk-go/contracts/bindings/StrategyManager"
	contractAVSDirectory "github.com/trigg3rX/triggerx-contracts/bindings/contracts/AvsDirectory"
)

var (
	logger logging.Logger
)

var configPath = "config-files/triggerx_operator.yaml"

type BLSKeystore struct {
	PubKey string `json:"pubKey"`
	Crypto struct {
		Cipher       string `json:"cipher"`
		CipherText   string `json:"ciphertext"`
		CipherParams struct {
			IV string `json:"iv"`
		} `json:"cipherparams"`
	} `json:"crypto"`
}

func handleHomeDirPath(path string) string {
	if len(path) >= 2 && path[:2] == "~/" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return path
		}
		return filepath.Join(homeDir, path[2:])
	}
	return path
}

func RegisterCommand() *cli.Command {
	return &cli.Command{
		Name:  "register",
		Usage: "Register a new operator",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "passphrase",
				Usage:    "Passphrase for the ECDSA keystore file",
				Required: true,
			},
		// 	&cli.StringFlag{
		// 		Name:     "token-strategy-address",
		// 		Usage:    "Address of the token strategy",
		// 		Required: true,
		// 	},
		// 	&cli.StringFlag{
		// 		Name:     "stake-amount",
		// 		Usage:    "Amount of tokens to stake",
		// 		Required: true,
		// 	},
		},
		Action: registerOperator,
	}
}

func DeregisterCommand() *cli.Command {
	return &cli.Command{
		Name:   "deregister",
		Usage:  "Deregister an operator",
		Action: deregisterOperator,
	}
}

func registerOperator(c *cli.Context) error {
	// Initialize logger with error checking
	err := logging.InitLogger(logging.Development, "register")
	if err != nil {
		return cli.Exit(fmt.Sprintf("Failed to initialize logger: %v", err), 1)
	}
	logger = logging.GetLogger()

	// tokenStrategyAddr := c.String("token-strategy-address")
	// stakeAmountFloat, err := strconv.ParseFloat(c.String("stake-amount"), 64)
	// if err != nil {
	// 	return cli.Exit(fmt.Sprintf("Failed to parse stake amount: %v", err), 1)
	// }

	// // Convert to Wei (1 ETH = 10^18 Wei)
	// stakeAmountFloat = stakeAmountFloat * 1e18
	// stakeAmount := new(big.Int)
	// stakeAmount.SetString(fmt.Sprintf("%.0f", stakeAmountFloat), 10)

	// // Add logging to verify the amount
	// logger.Infof("Stake amount in Wei: %s", stakeAmount.String())

	nodeConfig, err := getConfig()
	if err != nil {
		return cli.Exit(fmt.Sprintf("Failed to get node config: %v", err), 1)
	}

	keystorePath := handleHomeDirPath(nodeConfig.EcdsaPrivateKeyStorePath)
	if keystorePath == "" {
		return cli.Exit("Fill in the ECDSA keystore path in the config file", 1)
	}

	blsKeystorePath := handleHomeDirPath(nodeConfig.BlsPrivateKeyStorePath)
	if blsKeystorePath == "" {
		return cli.Exit("Fill in the BLS keystore path in the config file", 1)
	}

	// Check if keystore file exists
	if _, err := os.Stat(keystorePath); os.IsNotExist(err) {
		return cli.Exit(fmt.Sprintf("keystore file not found at path: %s", keystorePath), 1)
	}

	if _, err := os.Stat(blsKeystorePath); os.IsNotExist(err) {
		return cli.Exit(fmt.Sprintf("bls keystore file not found at path: %s", blsKeystorePath), 1)
	}

	// Read and parse BLS keystore file
	blsKeystoreBytes, err := os.ReadFile(blsKeystorePath)
	if err != nil {
		return cli.Exit(fmt.Sprintf("Failed to read BLS keystore file: %v", err), 1)
	}

	var blsKeystore BLSKeystore
	if err := json.Unmarshal(blsKeystoreBytes, &blsKeystore); err != nil {
		return cli.Exit(fmt.Sprintf("Failed to parse BLS keystore JSON: %v", err), 1)
	}

	if blsKeystore.PubKey == "" {
		return cli.Exit("BLS public key not found in keystore", 1)
	}

	// Read ECDSA private key
	ecdsaPrivKey, err := ecdsa.ReadKey(keystorePath, c.String("passphrase"))
	if err != nil {
		return cli.Exit(fmt.Sprintf("Failed to read ECDSA keystore file: %v", err), 1)
	}

	keeperAddress, err := ecdsa.GetAddressFromKeyStoreFile(keystorePath)
	if err != nil {
		return cli.Exit(fmt.Sprintf("Failed to get ECDSA public key: %v", err), 1)
	}

	// Connect to Ethereum client
	client, err := ethclient.Dial(nodeConfig.EthRpcUrl)
	if err != nil {
		return cli.Exit(fmt.Sprintf("Failed to connect to Ethereum client: %v", err), 1)
	}

	// Get chain ID
	chainID, err := client.ChainID(context.Background())
	if err != nil {
		return cli.Exit(fmt.Sprintf("Failed to get chain ID: %v", err), 1)
	}
	logger.Infof("Connected to chain ID: %s", chainID.String())

	logger.Infof("Using address: %s", keeperAddress.Hex())

	// Create transaction options with proper parameters
	auth, err := bind.NewKeyedTransactorWithChainID(ecdsaPrivKey, chainID)
	if err != nil {
		return cli.Exit(fmt.Sprintf("Failed to create transaction auth: %v", err), 1)
	}

	// Get the current nonce
	nonce, err := client.PendingNonceAt(context.Background(), keeperAddress)
	if err != nil {
		return cli.Exit(fmt.Sprintf("Failed to get nonce: %v", err), 1)
	}

	// Get gas price from network
	gasPrice, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		return cli.Exit(fmt.Sprintf("Failed to suggest gas price: %v", err), 1)
	}

	// Set transaction parameters
	auth.Nonce = big.NewInt(int64(nonce))
	auth.Value = big.NewInt(0) // in wei
	// auth.GasLimit = uint64(300000) // in units
	auth.GasPrice = gasPrice

	logger.Infof("Transaction parameters set - Nonce: %d, Gas Price: %s", nonce, gasPrice.String())

	// Generate random salt
	salt := make([]byte, 32)
	if _, err := rand.Read(salt); err != nil {
		return cli.Exit(fmt.Sprintf("Failed to generate random salt: %v", err), 1)
	}

	// Set expiry to 1 hour from now
	expiry := big.NewInt(time.Now().Add(1 * time.Hour).Unix())

	// Create AVSDirectory contract instance
	avsDirectoryAddr := common.HexToAddress(nodeConfig.AvsDirectoryAddress)
	avsDirectory, err := contractAVSDirectory.NewContractAvsDirectory(avsDirectoryAddr, client)
	if err != nil {
		return cli.Exit(fmt.Sprintf("Failed to create AVSDirectory contract instance: %v", err), 1)
	}

	// strategyContract, err := strategy.NewContractIStrategy(
	// 	common.HexToAddress(tokenStrategyAddr),
	// 	client,
	// )
	// if err != nil {
	// 	return cli.Exit(fmt.Sprintf("Failed to create strategy contract instance: %v", err), 1)
	// }

	// strategyManagerContract, err := strategymanager.NewContractStrategyManager(
	// 	common.HexToAddress(nodeConfig.StrategyManagerAddress),
	// 	client,
	// )
	// if err != nil {
	// 	return cli.Exit(fmt.Sprintf("Failed to create strategy manager contract instance: %v", err), 1)
	// }

	// logger.Infof("Contracts loaded successfully")

	// underlyingTokenAddress, err := strategyContract.UnderlyingToken(&bind.CallOpts{})
	// if err != nil {
	// 	return cli.Exit(fmt.Sprintf("Failed to get underlying token: %v", err), 1)
	// }

	// tokenContract, err := erc20.NewContractIERC20(
	// 	underlyingTokenAddress,
	// 	client,
	// )
	// if err != nil {
	// 	return cli.Exit(fmt.Sprintf("Failed to create underlying token contract instance: %v", err), 1)
	// }

	// logger.Infof("Underlying token address: %s", underlyingTokenAddress.Hex())

	// // Before approving, check token balance
	// tokenBalance, err := tokenContract.BalanceOf(&bind.CallOpts{}, keeperAddress)
	// if err != nil {
	// 	return cli.Exit(fmt.Sprintf("Failed to get token balance: %v", err), 1)
	// }
	// logger.Infof("Token balance: %s", tokenBalance.String())

	// if tokenBalance.Cmp(stakeAmount) < 0 {
	// 	return cli.Exit(fmt.Sprintf("Insufficient token balance. Required: %s, Available: %s", stakeAmount.String(), tokenBalance.String()), 1)
	// }

	// tx1, err := tokenContract.Approve(auth, common.HexToAddress(nodeConfig.StrategyManagerAddress), stakeAmount)
	// if err != nil {
	// 	return cli.Exit(fmt.Sprintf("Failed to approve token: %v", err), 1)
	// }

	// logger.Infof("Approved token: %s", tx1.Hash().Hex())

	// // Wait for the approve transaction to be mined before proceeding
	// receipt1, err := bind.WaitMined(context.Background(), client, tx1)
	// if err != nil {
	// 	return cli.Exit(fmt.Sprintf("Failed to wait for approve transaction: %v", err), 1)
	// }

	// if receipt1.Status == 0 {
	// 	return cli.Exit(fmt.Sprintf("Approve transaction failed: %s", tx1.Hash().Hex()), 1)
	// }

	// // Get fresh nonce for the deposit transaction
	// nonce, err = client.PendingNonceAt(context.Background(), keeperAddress)
	// if err != nil {
	// 	return cli.Exit(fmt.Sprintf("Failed to get nonce: %v", err), 1)
	// }
	// auth.Nonce = big.NewInt(int64(nonce))

	// // Get fresh gas price
	// gasPrice, err = client.SuggestGasPrice(context.Background())
	// if err != nil {
	// 	return cli.Exit(fmt.Sprintf("Failed to suggest gas price: %v", err), 1)
	// }
	// auth.GasPrice = gasPrice

	// logger.Infof("Transaction parameters set - Nonce: %d, Gas Price: %s", nonce, gasPrice.String())

	// // Now proceed with deposit
	// tx2, err := strategyManagerContract.DepositIntoStrategy(
	// 	auth,
	// 	common.HexToAddress(tokenStrategyAddr), // strategy address
	// 	underlyingTokenAddress,                 // WETH address
	// 	stakeAmount,                            // 0.1 WETH in Wei
	// )
	// if err != nil {
	// 	return cli.Exit(fmt.Sprintf("Failed to deposit into strategy: %v", err), 1)
	// }

	// logger.Infof("Deposited into strategy: %s", tx2.Hash().Hex())

	// Calculate registration digest hash
	digestHash, err := avsDirectory.CalculateOperatorAVSRegistrationDigestHash(
		&bind.CallOpts{},
		keeperAddress,
		common.HexToAddress(nodeConfig.ServiceManagerAddress),
		[32]byte(salt),
		expiry,
	)
	if err != nil {
		return cli.Exit(fmt.Sprintf("Failed to calculate digest hash: %v", err), 1)
	}

	// Sign the digest hash
	signature, err := crypto.Sign(digestHash[:], ecdsaPrivKey)
	if err != nil {
		return cli.Exit(fmt.Sprintf("Failed to sign digest: %v", err), 1)
	}

	// Prepare registration request
	registerRequest := types.RegisterKeeperRequest{
		KeeperAddress:     keeperAddress.Hex(),
		Signature:         hexutil.Encode(signature),
		Salt:              hexutil.Encode(salt),
		Expiry:            expiry.String(),
		BlsPublicKey:      blsKeystore.PubKey,
		// TokenStrategyAddr: tokenStrategyAddr,
		// StakeAmount:       stakeAmount.String(),
	}

	// Convert request to JSON
	jsonData, err := json.Marshal(registerRequest)
	if err != nil {
		return cli.Exit(fmt.Sprintf("Failed to marshal request data: %v", err), 1)
	}
	logger.Infof("Registering keeper with request: %+v", registerRequest)

	// Create HTTP request
	req, err := http.NewRequest("POST", "http://localhost:8081/api/cli/register", bytes.NewBuffer(jsonData))
	if err != nil {
		return cli.Exit(fmt.Sprintf("Failed to create HTTP request: %v", err), 1)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Host", "localhost")

	// Make the request
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return cli.Exit(fmt.Sprintf("Failed to send registration request: %v", err), 1)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		return cli.Exit(fmt.Sprintf("Registration request failed with status: %d", resp.StatusCode), 1)
	}

	// Parse response
	var registerResponse types.RegisterKeeperResponse
	if err := json.NewDecoder(resp.Body).Decode(&registerResponse); err != nil {
		return cli.Exit(fmt.Sprintf("Failed to decode response: %v", err), 1)
	}

	logger.Infof("Keeper registered to TriggerX AVS: %+v", registerResponse)
	logger.Info("Successfully registered keeper to TriggerX AVS")
	return nil
}

func deregisterOperator(c *cli.Context) error {
	return nil
}

func getConfig() (config types.NodeConfig, err error) {
	yamlFile, err := os.ReadFile(configPath)
	if err != nil {
		return config, err
	}
	err = yaml.Unmarshal(yamlFile, &config)
	return config, err
}
