package operator

import (
	"context"
	"crypto/sha256"
	"fmt"
	"math/big"
	"os"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/imua-xyz/imua-avs-sdk/client/txmgr"
	sdklogging "github.com/imua-xyz/imua-avs-sdk/logging"

	"github.com/Layr-Labs/eigensdk-go/crypto/bls"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/trigg3rX/triggerx-backend/cli/core"
	chain "github.com/trigg3rX/triggerx-backend/cli/core/chainio"
	"github.com/trigg3rX/triggerx-backend/cli/core/chainio/eth"
	"github.com/trigg3rX/triggerx-backend/cli/types"
)

// getKeyPair creates a BLS keypair from a hex private key (same as aggregator implementation)
func getKeyPair(privateKey string) (*bls.KeyPair, error) {
	// Add 0x prefix if not present
	var prefixedPrivateKey string
	if len(privateKey) >= 2 && privateKey[:2] == "0x" {
		prefixedPrivateKey = privateKey
	} else {
		prefixedPrivateKey = "0x" + privateKey
	}

	// Create Fr element from hashed private key
	frElement := new(fr.Element)
	hasher := sha256.New()
	hasher.Write([]byte(prefixedPrivateKey))
	frElement.SetBytes(hasher.Sum(nil))

	// Create KeyPair from Fr element
	keyPair := bls.NewKeyPair(frElement)

	return keyPair, nil
}

type Operator struct {
	config       types.NodeConfig
	logger       sdklogging.Logger
	ethClient    eth.EthClient
	avsWriter    chain.AvsWriter
	avsReader    chain.ChainReader
	blsKeypair   *bls.KeyPair
	operatorAddr common.Address
	avsAddr      common.Address
}

func NewOperatorFromConfig(c types.NodeConfig) (*Operator, error) {
	var logLevel sdklogging.LogLevel
	if c.Production {
		logLevel = sdklogging.Production
	} else {
		logLevel = sdklogging.Development
	}
	logger, err := sdklogging.NewZapLogger(logLevel)
	if err != nil {
		return nil, err
	}

	var ethRpcClient eth.EthClient
	ethRpcClient, err = eth.NewClient(c.EthRpcUrl)
	if err != nil {
		logger.Error("can not create http eth client", "err", err)
		return nil, err
	}

	// Load BLS private key from environment variable
	blsPrivateKeyHex := os.Getenv("BLS_PRIVATE_KEY")
	var blsKeyPair *bls.KeyPair
	if blsPrivateKeyHex != "" {
		logger.Info("Loading BLS private key from environment")
		// Remove 0x prefix if present
		if len(blsPrivateKeyHex) >= 2 && blsPrivateKeyHex[:2] == "0x" {
			blsPrivateKeyHex = blsPrivateKeyHex[2:]
		}

		// Create a new BLS secret key and set from hex string
		blsKeyPair, err = getKeyPair(blsPrivateKeyHex)
		if err != nil {
			logger.Error("Failed to create BLS secret key from hex", "err", err)
			return nil, err
		}
		logger.Info("Successfully loaded BLS private key from environment")
	} else {
		logger.Error("BLS_PRIVATE_KEY environment variable not set")
		return nil, fmt.Errorf("BLS_PRIVATE_KEY environment variable not set")
	}

	chainId, err := ethRpcClient.ChainID(context.Background())
	if err != nil {
		logger.Error("Cannot get chainId", "err", err)
		return nil, err
	}

	// Load ECDSA private key from environment variable instead of keystore file
	operatorPrivateKeyHex := os.Getenv("OPERATOR_PRIVATE_KEY")
	if operatorPrivateKeyHex == "" {
		logger.Error("OPERATOR_PRIVATE_KEY environment variable not set")
		return nil, fmt.Errorf("OPERATOR_PRIVATE_KEY environment variable not set")
	}

	// Remove 0x prefix if present
	if len(operatorPrivateKeyHex) >= 2 && operatorPrivateKeyHex[:2] == "0x" {
		operatorPrivateKeyHex = operatorPrivateKeyHex[2:]
	}

	// Convert hex private key to ECDSA private key
	ecdsaPrivateKey, err := crypto.HexToECDSA(operatorPrivateKeyHex)
	if err != nil {
		logger.Error("Failed to convert private key to ECDSA", "err", err)
		return nil, fmt.Errorf("failed to convert private key to ECDSA: %w", err)
	}

	// Derive operator address from private key
	operatorSender := crypto.PubkeyToAddress(ecdsaPrivateKey.PublicKey)
	logger.Info("operatorSender:", "operatorSender", operatorSender.String())

	// Create a signer function from the ECDSA private key
	signerFn := func(ctx context.Context, address common.Address) (bind.SignerFn, error) {
		if address != operatorSender {
			return nil, fmt.Errorf("signer address mismatch: expected %s, got %s", operatorSender.Hex(), address.Hex())
		}

		// Return a bind.SignerFn that signs transactions using legacy format like the faucet
		return func(signer common.Address, tx *ethtypes.Transaction) (*ethtypes.Transaction, error) {
			if signer != operatorSender {
				return nil, fmt.Errorf("signer address mismatch: expected %s, got %s", operatorSender.Hex(), signer.Hex())
			}

			// Get transaction details
			to := tx.To()
			value := tx.Value()
			data := tx.Data()
			gasLimit := tx.Gas()

			// Get current nonce and gas price for creating a new legacy transaction
			nonce, err := ethRpcClient.PendingNonceAt(ctx, operatorSender)
			if err != nil {
				return nil, fmt.Errorf("failed to get nonce: %w", err)
			}

			gasPrice, err := ethRpcClient.SuggestGasPrice(ctx)
			if err != nil {
				return nil, fmt.Errorf("failed to get gas price: %w", err)
			}

			// Create a new legacy transaction (like the faucet does)
			var newTx *ethtypes.Transaction
			if to != nil {
				newTx = ethtypes.NewTransaction(nonce, *to, value, gasLimit, gasPrice, data)
			} else {
				// Contract creation transaction
				newTx = ethtypes.NewContractCreation(nonce, value, gasLimit, gasPrice, data)
			}

			// Sign the new transaction with EIP155Signer (like the faucet does)
			signedTx, err := ethtypes.SignTx(newTx, ethtypes.NewEIP155Signer(chainId), ecdsaPrivateKey)
			if err != nil {
				return nil, fmt.Errorf("failed to sign transaction: %w", err)
			}
			return signedTx, nil
		}, nil
	}

	balance, err := ethRpcClient.BalanceAt(context.Background(), operatorSender, nil)
	if err != nil {
		logger.Error("Cannot get Balance", "err", err)
	}
	if balance.Cmp(big.NewInt(0)) != 1 {
		logger.Warn("Operator has low or zero balance - you may need testnet funds for transactions", "balance", balance.String())
	} else {
		logger.Info("Operator balance check passed", "balance", balance.String())
	}

	// Check if addresses match (case-insensitive)
	if strings.ToLower(c.OperatorAddress) != strings.ToLower(operatorSender.String()) {
		logger.Warn("Configured operator address differs from keystore address",
			"configured", c.OperatorAddress,
			"keystore", operatorSender.String())
		logger.Info("Using keystore address as operator address", "address", operatorSender.String())
		// Update the config to use the keystore address
		c.OperatorAddress = operatorSender.String()
	}
	txMgr := txmgr.NewSimpleTxManager(ethRpcClient, logger, signerFn, common.HexToAddress(c.OperatorAddress))

	avsReader, _ := chain.BuildChainReader(
		common.HexToAddress(c.AVSAddress),
		ethRpcClient,
		logger)

	avsWriter, _ := chain.BuildChainWriter(
		common.HexToAddress(c.AVSAddress),
		ethRpcClient,
		logger,
		txMgr)

	operator := &Operator{
		config:       c,
		logger:       logger,
		ethClient:    ethRpcClient,
		avsWriter:    avsWriter,
		avsReader:    *avsReader,
		blsKeypair:   blsKeyPair,
		operatorAddr: common.HexToAddress(c.OperatorAddress),
		avsAddr:      common.HexToAddress(c.AVSAddress),
	}

	if c.RegisterOperatorOnStartup {
		operator.registerOperatorOnStartup()
	}

	// Check if BLS keypair is available before trying to use it
	if operator.blsKeypair != nil {
		logger.Info("Operator info",
			"operatorAddr", c.OperatorAddress,
			"operatorKey", operator.blsKeypair.GetPubKeyG2().Marshal(),
		)
	} else {
		logger.Info("Operator info",
			"operatorAddr", c.OperatorAddress,
			"operatorKey", "BLS keypair not loaded",
		)
	}

	return operator, nil
}

func (o *Operator) registerOperatorOnStartup() {
	err := o.RegisterOperatorWithChain()
	if err != nil {
		// This error might only be that the operator was already registered with chain, so we don't want to fatal
		o.logger.Error("Error registering operator with chain", "err", err)
	} else {
		o.logger.Infof("Registered operator with chain")
	}

	err = o.RegisterOperatorWithAvs()
	if err != nil {
		o.logger.Fatal("Error registering operator with avs", "err", err)
	}

	// Register BLS Public Key if BLS keypair is available
	if o.blsKeypair != nil {
		err = o.RegisterBLSPublicKey()
		if err != nil {
			o.logger.Error("Error registering BLS public key", "err", err)
		}
	}
}

func (o *Operator) RegisterBLSPublicKey() error {
	if o.blsKeypair == nil {
		return fmt.Errorf("BLS keypair not available")
	}

	// Check if BLS Public Key is already registered
	pubKey, err := o.avsReader.GetRegisteredPubkey(&bind.CallOpts{}, o.operatorAddr.String(), o.avsAddr.String())
	if err != nil {
		o.logger.Error("Cannot exec GetRegisteredPubKey", "err", err)
		return err
	}

	if len(pubKey) == 0 {
		o.logger.Info("Registering BLS Public Key...")

		// Convert Ethereum address to IM address
		operatorAddress, err := core.SwitchEthAddressToImAddress(o.operatorAddr.String())
		if err != nil {
			o.logger.Error("Cannot switch eth address to im address", "err", err)
			return err
		}

		// Create BLS message to sign
		msg := fmt.Sprintf(core.BLSMessageToSign,
			core.ChainIDWithoutRevision("imuachainlocalnet_232"), operatorAddress)
		hashedMsg := crypto.Keccak256Hash([]byte(msg))

		// Sign the message with BLS private key (convert to [32]byte array)
		var messageArray [32]byte
		copy(messageArray[:], hashedMsg.Bytes())
		sig := o.blsKeypair.SignMessage(messageArray)

		// Register BLS Public Key via EVM transaction
		_, err = o.avsWriter.RegisterBLSPublicKey(
			context.Background(),
			o.avsAddr.String(),
			o.blsKeypair.GetPubKeyG2().Marshal(),
			sig.Marshal())

		if err != nil {
			o.logger.Error("operator failed to registerBLSPublicKey", "err", err)
			return err
		}
		o.logger.Info("BLS Public Key registered successfully")
	} else {
		o.logger.Info("BLS Public Key already registered")
	}

	return nil
}
