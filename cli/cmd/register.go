package cmd

// TODO: 
// - Use (eigensdk-go/chainio/elcontracts).IsOperatorRegistered() to check for EigenLayerRegistration
// - Use (eigensdk-go/chainio/elcontracts).GetStrategyAndUnderlyingERC20Token() to get strategy and underlying ERC20 token
// - Use (eigensdk-go/chainio/elcontracts).DepositERC20IntoStrategy() to deposit into strategy

// - Use (eigensdk-go/chainio/avsregistry).GetQuorumCount() to get quorum count
// - Use (eigensdk-go/chainio/elcontracts).GetOperatorId() to get operator ID and save in yaml

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/consensys/gnark-crypto/ecc/bn254"
	"github.com/urfave/cli/v2"
	"gopkg.in/yaml.v3"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/Layr-Labs/eigensdk-go/chainio/clients/wallet"
	"github.com/Layr-Labs/eigensdk-go/chainio/txmgr"

	"github.com/Layr-Labs/eigensdk-go/crypto/ecdsa"
	"github.com/Layr-Labs/eigensdk-go/logging"
	"github.com/Layr-Labs/eigensdk-go/signerv2"
	"github.com/trigg3rX/triggerx-backend/pkg/types"

	"crypto/rand"

	eigensdkbls "github.com/Layr-Labs/eigensdk-go/crypto/bls"

	contractAVSDirectory "github.com/Layr-Labs/eigensdk-go/contracts/bindings/IAVSDirectory"
	contractERC20 "github.com/Layr-Labs/eigensdk-go/contracts/bindings/IERC20"
	contractStrategy "github.com/Layr-Labs/eigensdk-go/contracts/bindings/IStrategy"
	contractRegistryCoordinator "github.com/Layr-Labs/eigensdk-go/contracts/bindings/RegistryCoordinator"
	contractStrategyManager "github.com/Layr-Labs/eigensdk-go/contracts/bindings/StrategyManager"
)

var configPath = "config-files/triggerx_keeper.yaml"

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
		Usage: "Register a new Keeper",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "ecdsa-passphrase",
				Usage:    "Passphrase for the ECDSA keystore file",
				Required: true,
			},
			&cli.StringFlag{
				Name:     "bls-passphrase",
				Usage:    "Passphrase for the BLS keystore file",
				Required: true,
			},
			&cli.StringFlag{
				Name:  "strategy-address",
				Usage: "Address of the strategies provided here: https://github.com/trigg3rX/triggerx-contracts",
			},
			&cli.StringFlag{
				Name:  "amount",
				Usage: "Amount of token to stake in strategy in ETH",
			},
		},
		Action: registerKeeper,
	}
}

func DeregisterCommand() *cli.Command {
	return &cli.Command{
		Name:  "deregister",
		Usage: "Deregister an keeper",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "ecdsa-passphrase",
				Usage:    "Passphrase for the ECDSA keystore file",
				Required: true,
			},
		},
		Action: deregisterKeeper,
	}
}

func registerKeeper(c *cli.Context) error {
	logger, err := logging.NewZapLogger(logging.Development)
	if err != nil {
		return cli.Exit(fmt.Sprintf("Failed to initialize logger: %v", err), 1)
	}

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

	blsKeyPair, err := eigensdkbls.ReadPrivateKeyFromFile(blsKeystorePath, c.String("bls-passphrase"))
	if err != nil {
		return cli.Exit(fmt.Sprintf("Failed to read BLS private key: %v", err), 1)
	}

	// logger.Infof("blsKeyPair: %+v", blsKeyPair)

	// Read ECDSA private key
	ecdsaPrivKey, err := ecdsa.ReadKey(keystorePath, c.String("ecdsa-passphrase"))
	if err != nil {
		return cli.Exit(fmt.Sprintf("Failed to read ECDSA keystore file: %v", err), 1)
	}

	// logger.Infof("ecdsaPrivKey: %+v", ecdsaPrivKey)

	keeperAddress, err := ecdsa.GetAddressFromKeyStoreFile(keystorePath)
	if err != nil {
		return cli.Exit(fmt.Sprintf("Failed to get ECDSA public key: %v", err), 1)
	}

	apiEndpoint := fmt.Sprintf("%s/keepers/address/%s", "https://data.triggerx.network/api", keeperAddress.Hex())
	logger.Infof("Checking keeper registration at: %s", apiEndpoint)
	resp, err := http.Get(apiEndpoint)
	if err != nil {
		logger.Error("Failed to check keeper registration", "error", err, "endpoint", apiEndpoint)
		return cli.Exit(fmt.Sprintf("Failed to check keeper registration: %v", err), 1)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		logger.Info("Keeper is already registered", "address", keeperAddress.Hex())
		return cli.Exit(fmt.Sprintf("Keeper is already registered: %s", keeperAddress.Hex()), 0)
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

	signerV2, signerAddr, err := signerv2.SignerFromConfig(signerv2.Config{PrivateKey: ecdsaPrivKey}, chainID)
	if err != nil {
		return cli.Exit(fmt.Sprintf("Failed to create signer: %v", err), 1)
	}

	txSender, err := wallet.NewPrivateKeyWallet(client, signerV2, signerAddr, logger)
	if err != nil {
		logger.Fatalf("Failed to create transaction sender: %v", err)
	}

	txMgr := txmgr.NewSimpleTxManager(txSender, client, logger, signerAddr)

	noSendTxOpts, err := txMgr.GetNoSendTxOpts()
	if err != nil {
		return cli.Exit(fmt.Sprintf("error creating transaction object %v", err), 1)
	}

	// Create AVSDirectory contract instance
	avsDirectoryAddr := common.HexToAddress(nodeConfig.AvsDirectoryAddress)
	avsDirectory, err := contractAVSDirectory.NewContractIAVSDirectory(avsDirectoryAddr, client)
	if err != nil {
		return cli.Exit(fmt.Sprintf("Failed to create AVSDirectory contract instance: %v", err), 1)
	}

	registryCoordinatorContract, err := contractRegistryCoordinator.NewContractRegistryCoordinator(common.HexToAddress(nodeConfig.RegistryCoordinatorAddress), client)
	if err != nil {
		return cli.Exit(fmt.Sprintf("Failed to create RegistryCoordinator contract instance: %v", err), 1)
	}

	// If both strategy-address and amount flags are present, perform staking
	if c.String("strategy-address") != "" || c.String("amount") != "" {
		if c.String("strategy-address") == "" || c.String("amount") == "" {
			return cli.Exit("Both strategy-address and amount must be provided for staking", 1)
		}

		if err := depositIntoStrategy(c, logger, client, txMgr, keeperAddress, nodeConfig); err != nil {
			return cli.Exit(fmt.Sprintf("Failed to deposit into strategy: %v", err), 1)
		}
	}

	// Generate random salt
	var saltBytes [32]byte
	if _, err := rand.Read(saltBytes[:]); err != nil {
		return cli.Exit(fmt.Sprintf("Failed to generate random salt: %v", err), 1)
	}

	// logger.Infof("saltBytes: %+v", saltBytes)

	// Set expiry to 1 hour from now
	expiry, ok := big.NewInt(0).SetString("15792089237316195423570985008687907853269984665640564039457584007913129639935", 10)
	if !ok {
		return cli.Exit("Failed to set expiry", 1)
	}

	// logger.Infof("expiry: %+v", expiry)

	g1HashedMsgToSign, err := registryCoordinatorContract.PubkeyRegistrationMessageHash(
		&bind.CallOpts{},
		signerAddr,
	)
	if err != nil {
		return cli.Exit(fmt.Sprintf("Failed to get g1HashedMsgToSign: %v", err), 1)
	}

	// logger.Infof("g1HashedMsgToSign: %+v", g1HashedMsgToSign)

	signedMsg := ConvertToBN254G1Point(
		blsKeyPair.SignHashedToCurveMessage(ConvertBn254GethToGnark(g1HashedMsgToSign)).G1Point,
	)

	// logger.Infof("signedMsg: %+v", signedMsg)

	G1pubkeyBN254 := ConvertToBN254G1Point(blsKeyPair.GetPubKeyG1())
	G2pubkeyBN254 := ConvertToBN254G2Point(blsKeyPair.GetPubKeyG2())

	pubkeyRegParams := contractRegistryCoordinator.IBLSApkRegistryPubkeyRegistrationParams{
		PubkeyRegistrationSignature: signedMsg,
		PubkeyG1:                    G1pubkeyBN254,
		PubkeyG2:                    G2pubkeyBN254,
	}

	// Calculate registration digest hash
	digestHash, err := avsDirectory.CalculateOperatorAVSRegistrationDigestHash(
		&bind.CallOpts{},
		keeperAddress,
		common.HexToAddress(nodeConfig.ServiceManagerAddress),
		saltBytes,
		expiry,
	)
	if err != nil {
		return cli.Exit(fmt.Sprintf("Failed to calculate digest hash: %v", err), 1)
	}

	// Sign the digest hash
	operatorSignature, err := crypto.Sign(digestHash[:], ecdsaPrivKey)
	if err != nil {
		return cli.Exit(fmt.Sprintf("Failed to sign digest: %v", err), 1)
	}

	operatorSignature[64] += 27

	operatorSignatureWithSaltAndExpiry := contractRegistryCoordinator.ISignatureUtilsSignatureWithSaltAndExpiry{
		Signature: operatorSignature,
		Salt:      saltBytes,
		Expiry:    expiry,
	}

	// // Get quorum number from API
	// apiEndpoint = fmt.Sprintf("%s/quorums/registration", "https://data.triggerx.network/api")
	// logger.Info("Fetching quorum number from API", "endpoint", apiEndpoint)
	// resp, err = http.Get(apiEndpoint)
	// if err != nil {
	// 	logger.Error("Failed to get quorum number from API",
	// 		"error", err,
	// 		"endpoint", apiEndpoint)
	// 	return cli.Exit(fmt.Sprintf("Failed to get quorum number from API: %v", err), 1)
	// }
	// defer resp.Body.Close()

	// if resp.StatusCode != http.StatusOK {
	// 	body, _ := io.ReadAll(resp.Body)
	// 	logger.Error("API returned non-200 status code",
	// 		"statusCode", resp.StatusCode,
	// 		"response", string(body),
	// 		"endpoint", apiEndpoint)
	// 	return cli.Exit(fmt.Sprintf("Failed to get quorum number. Status: %d", resp.StatusCode), 1)
	// }

	// var quorumNumber uint8
	// if err := json.NewDecoder(resp.Body).Decode(&quorumNumber); err != nil {
	// 	logger.Error("Failed to decode quorum number from response",
	// 		"error", err,
	// 		"endpoint", apiEndpoint)
	// 	return cli.Exit(fmt.Sprintf("Failed to decode quorum number: %v", err), 1)
	// }

	// logger.Info("Creating registration transaction",
	// 	"operatorAddress", keeperAddress.Hex(),
	// 	"quorumNumber", 0,
	// 	"socket", nodeConfig.ConnectionAddress,
	// 	"pubkeyRegParams", pubkeyRegParams,
	// 	"operatorSignatureWithSaltAndExpiry", operatorSignatureWithSaltAndExpiry,
	// )

	logger.Info("Creating unsigned transaction")
	tx, err := registryCoordinatorContract.RegisterOperator(
		noSendTxOpts,
		[]byte{0},
		string(nodeConfig.ConnectionAddress),
		pubkeyRegParams,
		operatorSignatureWithSaltAndExpiry,
	)

	if err != nil {
		logger.Error("Registration failed",
			"error", err)
		return cli.Exit(fmt.Sprintf("Failed to create transaction: %v", err), 1)
	}
	logger.Info("Successfully created unsigned transaction",
		"txHash", tx.Hash().Hex())

	logger.Info("Sending transaction to network")
	receipt, err := txMgr.Send(context.Background(), tx, true)
	if err != nil {
		logger.Error("Failed to send transaction",
			"error", err,
			"txHash", tx.Hash().Hex())
		return cli.Exit(fmt.Sprintf("Failed to send transaction: %v", err), 1)
	}
	logger.Info("Transaction successfully sent and mined",
		"txHash", receipt.TxHash.Hex(),
		"blockNumber", receipt.BlockNumber,
		"gasUsed", receipt.GasUsed)

	// Create keeper data in database
	keeperData := types.KeeperData{
		WithdrawalAddress: keeperAddress.Hex(),
		RegisteredTx:      receipt.TxHash.Hex(),
		Status:            true,
		BlsSigningKeys:    []string{G1pubkeyBN254.X.String(), G1pubkeyBN254.Y.String()},
		ConnectionAddress: nodeConfig.ConnectionAddress,
		Verified:          false,
		CurrentQuorumNo:   int(0),
	}

	logger.Info("Preparing to create keeper in database",
		"withdrawalAddress", keeperData.WithdrawalAddress,
		"registeredTx", keeperData.RegisteredTx,
		"quorumNumber", keeperData.CurrentQuorumNo)

	jsonData, err := json.Marshal(keeperData)
	if err != nil {
		logger.Error("Failed to marshal keeper data",
			"error", err,
			"keeperData", fmt.Sprintf("%+v", keeperData))
		return cli.Exit(fmt.Sprintf("Failed to marshal keeper data: %v", err), 1)
	}

	// Make POST request to create keeper
	apiEndpoint = fmt.Sprintf("%s/keepers", "https://data.triggerx.network/api")
	resp, err = http.Post(apiEndpoint, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		logger.Error("Failed to create keeper in database",
			"error", err,
			"endpoint", apiEndpoint)
		return cli.Exit(fmt.Sprintf("Failed to create keeper in database: %v", err), 1)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		logger.Error("Failed to create keeper in database",
			"statusCode", resp.StatusCode,
			"response", string(body),
			"endpoint", apiEndpoint)
		return cli.Exit(fmt.Sprintf("Failed to create keeper in database. Status: %d", resp.StatusCode), 1)
	}

	logger.Info("Successfully completed registration process",
		"operatorAddress", keeperAddress.Hex(),
		"txHash", receipt.TxHash.Hex())
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

func ConvertBn254GethToGnark(input contractRegistryCoordinator.BN254G1Point) *bn254.G1Affine {
	return eigensdkbls.NewG1Point(input.X, input.Y).G1Affine
}

func ConvertToBN254G1Point(input *eigensdkbls.G1Point) contractRegistryCoordinator.BN254G1Point {
	output := contractRegistryCoordinator.BN254G1Point{
		X: input.X.BigInt(big.NewInt(0)),
		Y: input.Y.BigInt(big.NewInt(0)),
	}
	return output
}

func ConvertToBN254G2Point(input *eigensdkbls.G2Point) contractRegistryCoordinator.BN254G2Point {
	output := contractRegistryCoordinator.BN254G2Point{
		X: [2]*big.Int{input.X.A1.BigInt(big.NewInt(0)), input.X.A0.BigInt(big.NewInt(0))},
		Y: [2]*big.Int{input.Y.A1.BigInt(big.NewInt(0)), input.Y.A0.BigInt(big.NewInt(0))},
	}
	return output
}

func depositIntoStrategy(c *cli.Context, logger logging.Logger, client *ethclient.Client, txMgr *txmgr.SimpleTxManager, keeperAddress common.Address, nodeConfig types.NodeConfig) error {
	strategyAddr := c.String("strategy-address")
	if strategyAddr == "" {
		return fmt.Errorf("strategy-address is required for staking")
	}

	stakeAmountFloat, err := strconv.ParseFloat(c.String("amount"), 64)
	if err != nil {
		return fmt.Errorf("failed to parse stake amount: %v", err)
	}

	// Convert to Wei (1 ETH = 10^18 Wei)
	stakeAmountFloat = stakeAmountFloat * 1e18
	stakeAmount := new(big.Int)
	stakeAmount.SetString(fmt.Sprintf("%.0f", stakeAmountFloat), 10)

	logger.Infof("Stake amount in Wei: %s", stakeAmount.String())

	noSendTxOpts, err := txMgr.GetNoSendTxOpts()
	if err != nil {
		return fmt.Errorf("error creating transaction object %v", err)
	}

	// Create contract instances and perform staking
	strategyManager, err := contractStrategyManager.NewContractStrategyManager(common.HexToAddress(nodeConfig.StrategyManagerAddress), client)
	if err != nil {
		return fmt.Errorf("failed to create StrategyManager contract instance: %v", err)
	}

	strategyContract, err := contractStrategy.NewContractIStrategy(common.HexToAddress(strategyAddr), client)
	if err != nil {
		return fmt.Errorf("failed to create Strategy contract instance: %v", err)
	}

	underlyingTokenAddr, err := strategyContract.UnderlyingToken(&bind.CallOpts{})
	if err != nil {
		return fmt.Errorf("failed to get underlying token address: %v", err)
	}

	logger.Info("Underlying token address", "address", underlyingTokenAddr.Hex())

	tokenContract, err := contractERC20.NewContractIERC20(underlyingTokenAddr, client)
	if err != nil {
		return fmt.Errorf("failed to create ERC20 contract instance: %v", err)
	}

	tokenBalance, err := tokenContract.BalanceOf(nil, keeperAddress)
	if err != nil {
		return fmt.Errorf("failed to get token balance: %v", err)
	}

	if tokenBalance.Cmp(stakeAmount) < 0 {
		return fmt.Errorf("insufficient token balance. Required: %s, Available: %s",
			stakeAmount.String(), tokenBalance.String())
	}

	tx1, err := tokenContract.Approve(noSendTxOpts, common.HexToAddress(nodeConfig.StrategyManagerAddress), stakeAmount)
	if err != nil {
		return fmt.Errorf("failed to approve token: %v", err)
	}

	receipt1, err := txMgr.Send(context.Background(), tx1, true)
	if err != nil {
		return fmt.Errorf("failed to send transaction: %v", err)
	}

	logger.Info("Approved token", "txHash", receipt1.TxHash.Hex())

	if receipt1.Status == 0 {
		return fmt.Errorf("approve transaction failed: %s", tx1.Hash().Hex())
	}

	logger.Info("Approval transaction successfully mined",
		"txHash", receipt1.TxHash.Hex(),
		"blockNumber", receipt1.BlockNumber,
		"gasUsed", receipt1.GasUsed)

	logger.Info("Depositing into strategy",
		"stakeAmount", stakeAmount.String(),
		"balance", tokenBalance.String())

	tx2, err := strategyManager.DepositIntoStrategy(
		noSendTxOpts,
		common.HexToAddress(strategyAddr),
		underlyingTokenAddr,
		stakeAmount,
	)
	if err != nil {
		return fmt.Errorf("failed to deposit into strategy: %v", err)
	}

	receipt2, err := txMgr.Send(context.Background(), tx2, true)
	if err != nil {
		return fmt.Errorf("failed to send transaction: %v", err)
	}

	logger.Infof("Transaction successfully sent and mined",
		"txHash", receipt2.TxHash.Hex(),
		"blockNumber", receipt2.BlockNumber,
		"gasUsed", receipt2.GasUsed)

	return nil
}

func deregisterKeeper(c *cli.Context) error {
	// logger, err := logging.NewZapLogger(logging.Development)
	// if err != nil {
	// 	return cli.Exit(fmt.Sprintf("Failed to initialize logger: %v", err), 1)
	// }

	// nodeConfig, err := getConfig()
	// if err != nil {
	// 	return cli.Exit(fmt.Sprintf("Failed to get node config: %v", err), 1)
	// }

	// keystorePath := handleHomeDirPath(nodeConfig.EcdsaPrivateKeyStorePath)
	// if keystorePath == "" {
	// 	return cli.Exit("Fill in the ECDSA keystore path in the config file", 1)
	// }

	// // Check if keystore file exists
	// if _, err := os.Stat(keystorePath); os.IsNotExist(err) {
	// 	return cli.Exit(fmt.Sprintf("keystore file not found at path: %s", keystorePath), 1)
	// }

	// // Read ECDSA private key
	// ecdsaPrivKey, err := ecdsa.ReadKey(keystorePath, c.String("ecdsa-passphrase"))
	// if err != nil {
	// 	return cli.Exit(fmt.Sprintf("Failed to read ECDSA keystore file: %v", err), 1)
	// }

	// // logger.Infof("ecdsaPrivKey: %+v", ecdsaPrivKey)

	// keeperAddress, err := ecdsa.GetAddressFromKeyStoreFile(keystorePath)
	// if err != nil {
	// 	return cli.Exit(fmt.Sprintf("Failed to get ECDSA public key: %v", err), 1)
	// }

	// // Connect to Ethereum client
	// client, err := ethclient.Dial(nodeConfig.EthRpcUrl)
	// if err != nil {
	// 	return cli.Exit(fmt.Sprintf("Failed to connect to Ethereum client: %v", err), 1)
	// }

	// // Get chain ID
	// chainID, err := client.ChainID(context.Background())
	// if err != nil {
	// 	return cli.Exit(fmt.Sprintf("Failed to get chain ID: %v", err), 1)
	// }
	// logger.Infof("Connected to chain ID: %s", chainID.String())

	// logger.Infof("Using address: %s", keeperAddress.Hex())

	// signerV2, signerAddr, err := signerv2.SignerFromConfig(signerv2.Config{PrivateKey: ecdsaPrivKey}, chainID)
	// if err != nil {
	// 	return cli.Exit(fmt.Sprintf("Failed to create signer: %v", err), 1)
	// }

	// txSender, err := wallet.NewPrivateKeyWallet(client, signerV2, signerAddr, logger)
	// if err != nil {
	// 	logger.Fatalf("Failed to create transaction sender: %v", err)
	// }

	// txMgr := txmgr.NewSimpleTxManager(txSender, client, logger, signerAddr)

	// noSendTxOpts, err := txMgr.GetNoSendTxOpts()
	// if err != nil {
	// 	return cli.Exit(fmt.Sprintf("error creating transaction object %v", err), 1)
	// }

	// registryCoordinatorContract, err := contractRegistryCoordinator.NewContractRegistryCoordinator(common.HexToAddress(nodeConfig.RegistryCoordinatorAddress), client)
	// if err != nil {
	// 	return cli.Exit(fmt.Sprintf("Failed to create RegistryCoordinator contract instance: %v", err), 1)
	// }

	// // TODO: Get the Quorum No of Keeper Registration

	// logger.Info("Deregistering keeper", "operatorAddress", keeperAddress.Hex())

	// tx, err := registryCoordinatorContract.DeregisterOperator(noSendTxOpts, keeperAddress.Hex(), []byte{0})
	// if err != nil {
	// 	return cli.Exit(fmt.Sprintf("Failed to deregister operator: %v", err), 1)
	// }

	// receipt, err := txMgr.Send(context.Background(), tx, true)
	// if err != nil {
	// 	return cli.Exit(fmt.Sprintf("Failed to send transaction: %v", err), 1)
	// }

	// // Create keeper data in database
	// keeperData := types.KeeperData{
	// 	WithdrawalAddress: keeperAddress.Hex(),
	// 	RegisteredTx:      receipt.TxHash.Hex(),
	// 	Status:            false,
	// 	BlsSigningKeys:    []string{},
	// 	ConnectionAddress: nodeConfig.ConnectionAddress,
	// 	Verified:          false,
	// 	CurrentQuorumNo:   int(0),
	// }

	// logger.Info("Preparing to create keeper in database",
	// 	"withdrawalAddress", keeperData.WithdrawalAddress,
	// 	"registeredTx", keeperData.RegisteredTx,
	// 	"quorumNumber", keeperData.CurrentQuorumNo)

	// jsonData, err := json.Marshal(keeperData)
	// if err != nil {
	// 	logger.Error("Failed to marshal keeper data",
	// 		"error", err,
	// 		"keeperData", fmt.Sprintf("%+v", keeperData))
	// 	return cli.Exit(fmt.Sprintf("Failed to marshal keeper data: %v", err), 1)
	// }

	// // Make POST request to create keeper
	// apiEndpoint := fmt.Sprintf("%s/keepers", "https://data.triggerx.network/api")
	// resp, err := http.Post(apiEndpoint, "application/json", bytes.NewBuffer(jsonData))
	// if err != nil {
	// 	logger.Error("Failed to create keeper in database",
	// 		"error", err,
	// 		"endpoint", apiEndpoint)
	// 	return cli.Exit(fmt.Sprintf("Failed to create keeper in database: %v", err), 1)
	// }
	// defer resp.Body.Close()

	// if resp.StatusCode != http.StatusCreated {
	// 	body, _ := io.ReadAll(resp.Body)
	// 	logger.Error("Failed to create keeper in database",
	// 		"statusCode", resp.StatusCode,
	// 		"response", string(body),
	// 		"endpoint", apiEndpoint)
	// 	return cli.Exit(fmt.Sprintf("Failed to create keeper in database. Status: %d", resp.StatusCode), 1)
	// }

	// logger.Info("Transaction successfully sent and mined",
	// 	"txHash", receipt.TxHash.Hex(),
	// 	"blockNumber", receipt.BlockNumber,
	// 	"gasUsed", receipt.GasUsed)

	return nil
}
