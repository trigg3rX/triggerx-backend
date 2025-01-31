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
	"encoding/hex"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"os"
	"strconv"
	"strings"
	"crypto/rand"

	"github.com/urfave/cli/v2"
	"github.com/consensys/gnark-crypto/ecc/bn254"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/Layr-Labs/eigensdk-go/chainio/clients/wallet"
	"github.com/Layr-Labs/eigensdk-go/chainio/txmgr"
	"github.com/Layr-Labs/eigensdk-go/crypto/ecdsa"
	"github.com/Layr-Labs/eigensdk-go/crypto/bls"
	"github.com/Layr-Labs/eigensdk-go/logging"
	"github.com/Layr-Labs/eigensdk-go/signerv2"

	contractAVSDirectory "github.com/Layr-Labs/eigensdk-go/contracts/bindings/IAVSDirectory"
	contractERC20 "github.com/Layr-Labs/eigensdk-go/contracts/bindings/IERC20"
	contractStrategy "github.com/Layr-Labs/eigensdk-go/contracts/bindings/IStrategy"
	contractRegistryCoordinator "github.com/Layr-Labs/eigensdk-go/contracts/bindings/RegistryCoordinator"
	contractStrategyManager "github.com/Layr-Labs/eigensdk-go/contracts/bindings/StrategyManager"

	"github.com/trigg3rX/triggerx-backend/pkg/types"
	"github.com/trigg3rX/triggerx-backend/cli/utils"
)

func RegisterCommand() *cli.Command {
	return &cli.Command{
		Name:  "register",
		Usage: "Register a new Keeper",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "config",
				Usage:    "Path to the config file (triggerx_keeper.yaml)",
				Required: true,
			},
			&cli.StringFlag{
				Name:  "ecdsa-passphrase",
				Usage: "Passphrase for the ECDSA keystore file",
			},
			&cli.StringFlag{
				Name:  "bls-passphrase",
				Usage: "Passphrase for the BLS keystore file",
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
				Name:     "config",
				Usage:    "Path to the config file (triggerx_keeper.yaml)",
				Required: true,
			},
			&cli.StringFlag{
				Name:  "ecdsa-passphrase",
				Usage: "Passphrase for the ECDSA keystore file",
			},
		},
		Action: deregisterKeeper,
	}
}

func registerKeeper(c *cli.Context) error {
	logger, err := logging.NewZapLogger(logging.Development)
	if err != nil {
		return cli.Exit("Failed to initialize logger", 1)
	}

	logger.Info("Starting registration process...")

	nodeConfig, err := utils.GetConfig(c.String("config"))
	if err != nil {
		return cli.Exit("Failed to get node config", 1)
	}

	keystorePath := utils.HandleHomeDirPath(nodeConfig.EcdsaPrivateKeyStorePath)
	if keystorePath == "" {
		return cli.Exit("Fill in the ECDSA keystore path in the config file", 1)
	}

	blsKeystorePath := utils.HandleHomeDirPath(nodeConfig.BlsPrivateKeyStorePath)
	if blsKeystorePath == "" {
		return cli.Exit("Fill in the BLS keystore path in the config file", 1)
	}

	if _, err := os.Stat(keystorePath); os.IsNotExist(err) {
		return cli.Exit(fmt.Sprintf("Keystore file not found at path: %s", keystorePath), 1)
	}

	if _, err := os.Stat(blsKeystorePath); os.IsNotExist(err) {
		return cli.Exit(fmt.Sprintf("BLS keystore file not found at path: %s", blsKeystorePath), 1)
	}

	var ecdsaPassphrase string
	if c.String("ecdsa-passphrase") != "" {
		ecdsaPassphrase = c.String("ecdsa-passphrase")
	} else if nodeConfig.EcdsaPassphrase != "" {
		ecdsaPassphrase = nodeConfig.EcdsaPassphrase
	} else {
		return cli.Exit("ECDSA passphrase not provided in flag or config file", 1)
	}

	var blsPassphrase string
	if c.String("bls-passphrase") != "" {
		blsPassphrase = c.String("bls-passphrase")
	} else if nodeConfig.BlsPassphrase != "" {
		blsPassphrase = nodeConfig.BlsPassphrase
	} else {
		return cli.Exit("BLS passphrase not provided in flag or config file", 1)
	}

	blsKeyPair, err := bls.ReadPrivateKeyFromFile(blsKeystorePath, blsPassphrase)
	if err != nil {
		return cli.Exit("Failed to read BLS private key", 1)
	}

	ecdsaPrivKey, err := ecdsa.ReadKey(keystorePath, ecdsaPassphrase)
	if err != nil {
		return cli.Exit("Failed to read ECDSA keystore file", 1)
	}

	keeperAddress, err := ecdsa.GetAddressFromKeyStoreFile(keystorePath)
	if err != nil {
		return cli.Exit("Failed to get ECDSA public key", 1)
	}

	logger.Info("Keeper address", "address", keeperAddress.Hex())

	apiEndpoint := fmt.Sprintf("%s/keepers/address/%s", "https://data.triggerx.network/api", keeperAddress.Hex())
	logger.Info("Checking keeper registration ...")
	resp, err := http.Get(apiEndpoint)
	if err != nil {
		logger.Error("Failed to check keeper registration", "error", err)
		return cli.Exit("Failed to check keeper registration", 1)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		logger.Info("Keeper already registered", "address", keeperAddress.Hex())
		return cli.Exit("Keeper already registered", 0)
	}

	client, err := ethclient.Dial(nodeConfig.EthRpcUrl)
	if err != nil {
		return cli.Exit("Failed to connect to Ethereum client", 1)
	}

	chainID, err := client.ChainID(context.Background())
	if err != nil {
		return cli.Exit("Failed to get chain ID", 1)
	}

	signerV2, signerAddr, err := signerv2.SignerFromConfig(signerv2.Config{PrivateKey: ecdsaPrivKey}, chainID)
	if err != nil {
		return cli.Exit("Failed to create signer", 1)
	}

	txSender, err := wallet.NewPrivateKeyWallet(client, signerV2, signerAddr, logger)
	if err != nil {
		logger.Fatal("Failed to create transaction sender", "error", err)
	}

	txMgr := txmgr.NewSimpleTxManager(txSender, client, logger, signerAddr)

	noSendTxOpts, err := txMgr.GetNoSendTxOpts()
	if err != nil {
		return cli.Exit("Error creating transaction object", 1)
	}

	avsDirectoryAddr := common.HexToAddress(nodeConfig.AvsDirectoryAddress)
	avsDirectory, err := contractAVSDirectory.NewContractIAVSDirectory(avsDirectoryAddr, client)
	if err != nil {
		return cli.Exit("Failed to create AVSDirectory contract instance", 1)
	}

	registryCoordinatorContract, err := contractRegistryCoordinator.NewContractRegistryCoordinator(common.HexToAddress(nodeConfig.RegistryCoordinatorAddress), client)
	if err != nil {
		return cli.Exit("Failed to create RegistryCoordinator contract instance", 1)
	}

	logger.Info("Wallet and Contracts created")

	if c.String("strategy-address") != "" || c.String("amount") != "" {
		if c.String("strategy-address") == "" || c.String("amount") == "" {
			return cli.Exit("Both strategy-address and amount must be provided for staking", 1)
		}

		logger.Info("Depositing into strategy ...")
		if err := depositIntoStrategy(c, logger, client, txMgr, keeperAddress, nodeConfig); err != nil {
			return cli.Exit("Failed to deposit into strategy", 1)
		}
	}

	var saltBytes [32]byte
	if _, err := rand.Read(saltBytes[:]); err != nil {
		return cli.Exit("Failed to generate random salt", 1)
	}

	expiry, ok := big.NewInt(0).SetString("15792089237316195423570985008687907853269984665640564039457584007913129639935", 10)
	if !ok {
		return cli.Exit("Failed to set expiry", 1)
	}

	g1HashedMsgToSign, err := registryCoordinatorContract.PubkeyRegistrationMessageHash(
		&bind.CallOpts{},
		signerAddr,
	)
	if err != nil {
		return cli.Exit("Failed to get g1HashedMsgToSign", 1)
	}

	signedMsg := ConvertToBN254G1Point(
		blsKeyPair.SignHashedToCurveMessage(ConvertBn254GethToGnark(g1HashedMsgToSign)).G1Point,
	)

	G1pubkeyBN254 := ConvertToBN254G1Point(blsKeyPair.GetPubKeyG1())
	G2pubkeyBN254 := ConvertToBN254G2Point(blsKeyPair.GetPubKeyG2())

	pubkeyRegParams := contractRegistryCoordinator.IBLSApkRegistryPubkeyRegistrationParams{
		PubkeyRegistrationSignature: signedMsg,
		PubkeyG1:                    G1pubkeyBN254,
		PubkeyG2:                    G2pubkeyBN254,
	}

	digestHash, err := avsDirectory.CalculateOperatorAVSRegistrationDigestHash(
		&bind.CallOpts{},
		keeperAddress,
		common.HexToAddress(nodeConfig.ServiceManagerAddress),
		saltBytes,
		expiry,
	)
	if err != nil {
		return cli.Exit("Failed to calculate digest hash", 1)
	}

	operatorSignature, err := crypto.Sign(digestHash[:], ecdsaPrivKey)
	if err != nil {
		return cli.Exit("Failed to sign digest", 1)
	}
	operatorSignature[64] += 27

	operatorSignatureWithSaltAndExpiry := contractRegistryCoordinator.ISignatureUtilsSignatureWithSaltAndExpiry{
		Signature: operatorSignature,
		Salt:      saltBytes,
		Expiry:    expiry,
	}

	logger.Info("Parameters created.")
	logger.Info("Getting Quorum No ...")

	apiEndpoint = fmt.Sprintf("%s/quorums/registration", "https://data.triggerx.network/api")
	resp, err = http.Get(apiEndpoint)
	if err != nil {
		logger.Error("Failed to get quorum no", "error", err)
		return cli.Exit("Failed to get quorum no", 1)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Error("Failed to read response body", "error", err)
		return cli.Exit("Failed to read response body", 1)
	}

	quorumNo := "0" +strings.TrimSpace(string(body))
	quorumNoInt, err := strconv.Atoi(quorumNo)
	if err != nil {
		logger.Error("Failed to parse quorum number", "error", err)
		return cli.Exit("Failed to parse quorum number", 1)
	}
	quorumNoBytes, _ := hex.DecodeString(quorumNo)

	logger.Info("Creating registration transaction ...")
	tx, err := registryCoordinatorContract.RegisterOperator(
		noSendTxOpts,
		// []byte{0},
		quorumNoBytes,
		string(nodeConfig.ConnectionAddress),
		pubkeyRegParams,
		operatorSignatureWithSaltAndExpiry,
	)

	if err != nil {
		logger.Error("Registration failed", "error", err)
		return cli.Exit("Failed to create transaction", 1)
	}
	logger.Info("Created unsigned transaction", "txHash", tx.Hash().Hex())
	
	logger.Info("Sending transaction...")
	receipt, err := txMgr.Send(context.Background(), tx, true)
	if err != nil {
		logger.Error("Failed to send transaction", "error", err)
		return cli.Exit("Failed to send transaction", 1)
	}
	logger.Info("Transaction mined", "txHash", receipt.TxHash.Hex(), "blockNumber", receipt.BlockNumber)

	keeperData := types.KeeperData{
		WithdrawalAddress: keeperAddress.Hex(),
		RegisteredTx:      receipt.TxHash.Hex(),
		Status:            true,
		BlsSigningKeys:    []string{G1pubkeyBN254.X.String(), G1pubkeyBN254.Y.String()},
		ConnectionAddress: nodeConfig.ConnectionAddress,
		Verified:          false,
		CurrentQuorumNo:   quorumNoInt,
	}

	logger.Info("Registering keeper in database", "address", keeperData.WithdrawalAddress)

	jsonData, err := json.Marshal(keeperData)
	if err != nil {
		logger.Error("Failed to marshal keeper data", "error", err)
		return cli.Exit("Failed to marshal keeper data", 1)
	}

	apiEndpoint = fmt.Sprintf("%s/keepers", "https://data.triggerx.network/api")
	resp, err = http.Post(apiEndpoint, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		logger.Error("Failed to create keeper in database", "error", err)
		return cli.Exit("Failed to create keeper in database", 1)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		logger.Error("Failed to create keeper in database", "status", resp.StatusCode, "response", string(body))
		return cli.Exit("Failed to create keeper in database", 1)
	}

	logger.Info("Registration complete", "address", keeperAddress.Hex(), "txHash", receipt.TxHash.Hex())

	err = utils.DisplayBannerMessage(nodeConfig.KeeperName, keeperAddress.Hex())
	if err != nil {
		fmt.Printf("Error displaying banner: %v\n", err)
	}

	return nil
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

	stakeAmountFloat = stakeAmountFloat * 1e18
	stakeAmount := new(big.Int)
	stakeAmount.SetString(fmt.Sprintf("%.0f", stakeAmountFloat), 10)

	logger.Info("Stake amount", "wei", stakeAmount.String())

	noSendTxOpts, err := txMgr.GetNoSendTxOpts()
	if err != nil {
		return fmt.Errorf("error creating transaction object %v", err)
	}

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

	logger.Info("Using token", "address", underlyingTokenAddr.Hex())

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

	logger.Info("Token approved", "txHash", receipt1.TxHash.Hex())

	if receipt1.Status == 0 {
		return fmt.Errorf("approve transaction failed: %s", tx1.Hash().Hex())
	}

	logger.Info("Approval transaction mined", "txHash", receipt1.TxHash.Hex(), "blockNumber", receipt1.BlockNumber)

	logger.Info("Depositing into strategy", "amount", stakeAmount.String())

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

	logger.Info("Deposit transaction mined", "txHash", receipt2.TxHash.Hex(), "blockNumber", receipt2.BlockNumber)

	return nil
}

func deregisterKeeper(c *cli.Context) error {
	logger, err := logging.NewZapLogger(logging.Development)
	if err != nil {
		return cli.Exit("Failed to initialize logger", 1)
	}

	nodeConfig, err := utils.GetConfig(c.String("config"))
	if err != nil {
		return cli.Exit("Failed to get node config", 1)
	}

	keystorePath := utils.HandleHomeDirPath(nodeConfig.EcdsaPrivateKeyStorePath)
	if keystorePath == "" {
		return cli.Exit("Fill in the ECDSA keystore path in the config file", 1)
	}

	if _, err := os.Stat(keystorePath); os.IsNotExist(err) {
		return cli.Exit(fmt.Sprintf("Keystore file not found at path: %s", keystorePath), 1)
	}

	ecdsaPrivKey, err := ecdsa.ReadKey(keystorePath, c.String("ecdsa-passphrase"))
	if err != nil {
		return cli.Exit("Failed to read ECDSA keystore file", 1)
	}

	keeperAddress, err := ecdsa.GetAddressFromKeyStoreFile(keystorePath)
	if err != nil {
		return cli.Exit("Failed to get ECDSA public key", 1)
	}

	client, err := ethclient.Dial(nodeConfig.EthRpcUrl)
	if err != nil {
		return cli.Exit("Failed to connect to Ethereum client", 1)
	}

	chainID, err := client.ChainID(context.Background())
	if err != nil {
		return cli.Exit("Failed to get chain ID", 1)
	}
	logger.Info("Connected to chain", "chainID", chainID.String())

	logger.Info("Using keeper address", "address", keeperAddress.Hex())

	signerV2, signerAddr, err := signerv2.SignerFromConfig(signerv2.Config{PrivateKey: ecdsaPrivKey}, chainID)
	if err != nil {
		return cli.Exit("Failed to create signer", 1)
	}

	txSender, err := wallet.NewPrivateKeyWallet(client, signerV2, signerAddr, logger)
	if err != nil {
		logger.Fatal("Failed to create transaction sender", "error", err)
	}

	txMgr := txmgr.NewSimpleTxManager(txSender, client, logger, signerAddr)

	noSendTxOpts, err := txMgr.GetNoSendTxOpts()
	if err != nil {
		return cli.Exit("Error creating transaction object", 1)
	}

	registryCoordinatorContract, err := contractRegistryCoordinator.NewContractRegistryCoordinator(common.HexToAddress(nodeConfig.RegistryCoordinatorAddress), client)
	if err != nil {
		return cli.Exit("Failed to create RegistryCoordinator contract instance", 1)
	}

	logger.Info("Deregistering keeper", "address", keeperAddress.Hex())

	tx, err := registryCoordinatorContract.DeregisterOperator(noSendTxOpts, []byte{0})
	if err != nil {
		return cli.Exit("Failed to deregister operator", 1)
	}

	receipt, err := txMgr.Send(context.Background(), tx, true)
	if err != nil {
		return cli.Exit("Failed to send transaction", 1)
	}

	keeperData := types.KeeperData{
		WithdrawalAddress: keeperAddress.Hex(),
		RegisteredTx:      receipt.TxHash.Hex(),
		Status:            false,
		BlsSigningKeys:    []string{},
		ConnectionAddress: nodeConfig.ConnectionAddress,
		Verified:          false,
		CurrentQuorumNo:   int(0),
	}

	logger.Info("Updating keeper in database", "address", keeperData.WithdrawalAddress)

	jsonData, err := json.Marshal(keeperData)
	if err != nil {
		logger.Error("Failed to marshal keeper data", "error", err)
		return cli.Exit("Failed to marshal keeper data", 1)
	}

	apiEndpoint := fmt.Sprintf("%s/keepers", "https://data.triggerx.network/api")
	resp, err := http.Post(apiEndpoint, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		logger.Error("Failed to update keeper in database", "error", err)
		return cli.Exit("Failed to update keeper in database", 1)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		logger.Error("Failed to update keeper in database", "status", resp.StatusCode, "response", string(body))
		return cli.Exit("Failed to update keeper in database", 1)
	}

	logger.Info("Deregistration complete", "txHash", receipt.TxHash.Hex(), "blockNumber", receipt.BlockNumber)

	return nil
}

func ConvertBn254GethToGnark(input contractRegistryCoordinator.BN254G1Point) *bn254.G1Affine {
	return bls.NewG1Point(input.X, input.Y).G1Affine
}

func ConvertToBN254G1Point(input *bls.G1Point) contractRegistryCoordinator.BN254G1Point {
	output := contractRegistryCoordinator.BN254G1Point{
		X: input.X.BigInt(big.NewInt(0)),
		Y: input.Y.BigInt(big.NewInt(0)),
	}
	return output
}

func ConvertToBN254G2Point(input *bls.G2Point) contractRegistryCoordinator.BN254G2Point {
	output := contractRegistryCoordinator.BN254G2Point{
		X: [2]*big.Int{input.X.A1.BigInt(big.NewInt(0)), input.X.A0.BigInt(big.NewInt(0))},
		Y: [2]*big.Int{input.Y.A1.BigInt(big.NewInt(0)), input.Y.A0.BigInt(big.NewInt(0))},
	}
	return output
}