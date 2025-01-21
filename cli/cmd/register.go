package cmd

import (
	"context"
	"fmt"
	"math/big"

	"os"
	"path/filepath"

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
	contractAVSDirectory "github.com/trigg3rX/triggerx-contracts/bindings/contracts/AvsDirectory"
	regcoord "github.com/trigg3rX/triggerx-contracts/bindings/contracts/RegistryCoordinator"
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

	blsKeyPair, err := eigensdkbls.ReadPrivateKeyFromFile(blsKeystorePath, c.String("passphrase"))
	if err != nil {
		return cli.Exit(fmt.Sprintf("Failed to read BLS private key: %v", err), 1)
	}

	// logger.Infof("blsKeyPair: %+v", blsKeyPair)

	// Read ECDSA private key
	ecdsaPrivKey, err := ecdsa.ReadKey(keystorePath, c.String("passphrase"))
	if err != nil {
		return cli.Exit(fmt.Sprintf("Failed to read ECDSA keystore file: %v", err), 1)
	}

	// logger.Infof("ecdsaPrivKey: %+v", ecdsaPrivKey)

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
	avsDirectory, err := contractAVSDirectory.NewContractAvsDirectory(avsDirectoryAddr, client)
	if err != nil {
		return cli.Exit(fmt.Sprintf("Failed to create AVSDirectory contract instance: %v", err), 1)
	}

	registryCoordinatorContract, err := regcoord.NewContractRegistryCoordinator(common.HexToAddress(nodeConfig.RegistryCoordinatorAddress), client)
	if err != nil {
		return cli.Exit(fmt.Sprintf("Failed to create RegistryCoordinator contract instance: %v", err), 1)
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

	pubkeyRegParams := regcoord.IBLSApkRegistryPubkeyRegistrationParams{
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

	operatorSignatureWithSaltAndExpiry := regcoord.ISignatureUtilsSignatureWithSaltAndExpiry{
		Signature: operatorSignature,
		Salt:      saltBytes,
		Expiry:    expiry,
	}

	logger.Info("Creating unsigned transaction")
	tx, err := registryCoordinatorContract.RegisterOperator(
		noSendTxOpts,
		[]byte{0},
		"",
		pubkeyRegParams,
		operatorSignatureWithSaltAndExpiry,
	)
	if err != nil {
		logger.Errorf("Failed to create transaction: %v", err)
		return cli.Exit(fmt.Sprintf("Failed to create transaction: %v", err), 1)
	}
	logger.Info("Successfully created unsigned transaction", "txHash", tx.Hash().Hex())

	logger.Info("Sending transaction")
	receipt, err := txMgr.Send(context.Background(), tx, true)
	if err != nil {
		logger.Errorf("Failed to send transaction: %v", err)
		return cli.Exit(fmt.Sprintf("Failed to send transaction: %v", err), 1)
	}
	logger.Info("Transaction successfully sent and mined",
		"txHash", receipt.TxHash.Hex(),
		"blockNumber", receipt.BlockNumber,
		"gasUsed", receipt.GasUsed)

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

func ConvertBn254GethToGnark(input regcoord.BN254G1Point) *bn254.G1Affine {
	return eigensdkbls.NewG1Point(input.X, input.Y).G1Affine
}

func ConvertToBN254G1Point(input *eigensdkbls.G1Point) regcoord.BN254G1Point {
	output := regcoord.BN254G1Point{
		X: input.X.BigInt(big.NewInt(0)),
		Y: input.Y.BigInt(big.NewInt(0)),
	}
	return output
}

func ConvertToBN254G2Point(input *eigensdkbls.G2Point) regcoord.BN254G2Point {
	output := regcoord.BN254G2Point{
		X: [2]*big.Int{input.X.A1.BigInt(big.NewInt(0)), input.X.A0.BigInt(big.NewInt(0))},
		Y: [2]*big.Int{input.Y.A1.BigInt(big.NewInt(0)), input.Y.A0.BigInt(big.NewInt(0))},
	}
	return output
}
