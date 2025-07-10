package operator

import (
	"context"
	"fmt"
	"math/big"
	"os"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/imua-xyz/imua-avs-sdk/client/txmgr"
	"github.com/imua-xyz/imua-avs-sdk/crypto/bls"
	sdklogging "github.com/imua-xyz/imua-avs-sdk/logging"
	"github.com/imua-xyz/imua-avs-sdk/signer"

	// chain "github.com/trigg3rX/triggerx-backend/cli/core/chainio"
	// "github.com/trigg3rX/triggerx-backend/cli/core/chainio/eth"
	blscommon "github.com/prysmaticlabs/prysm/v5/crypto/bls/common"
	"github.com/trigg3rX/triggerx-backend/cli/types"
)

const (
	maxRetries = 80
	retryDelay = 1 * time.Second
)

type Operator struct {
	config       types.NodeConfig
	logger       sdklogging.Logger
	ethClient    eth.EthClient
	avsWriter    chain.AvsWriter
	avsReader    chain.ChainReader
	blsKeypair   blscommon.SecretKey
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

	blsKeyPassword, ok := os.LookupEnv("OPERATOR_BLS_KEY_PASSWORD")
	if !ok {
		logger.Info("OPERATOR_BLS_KEY_PASSWORD env var not set. using empty string")
	}
	blsKeyPair, err := bls.ReadPrivateKeyFromFile(c.BlsPrivateKeyStorePath, blsKeyPassword)
	if err != nil {
		logger.Error("Cannot parse bls private key", "err", err)
		return nil, err
	}

	chainId, err := ethRpcClient.ChainID(context.Background())
	if err != nil {
		logger.Error("Cannot get chainId", "err", err)
		return nil, err
	}

	ecdsaKeyPassword, ok := os.LookupEnv("OPERATOR_ECDSA_KEY_PASSWORD")
	if !ok {
		logger.Info("OPERATOR_ECDSA_KEY_PASSWORD env var not set. using empty string")
	}

	signer, operatorSender, err := signer.SignerFromConfig(signer.Config{
		KeystorePath: c.OperatorEcdsaPrivateKeyStorePath,
		Password:     ecdsaKeyPassword,
	}, chainId)
	if err != nil {
		panic(err)
	}
	logger.Info("operatorSender:", "operatorSender", operatorSender.String())

	balance, err := ethRpcClient.BalanceAt(context.Background(), operatorSender, nil)
	if err != nil {
		logger.Error("Cannot get Balance", "err", err)
	}
	if balance.Cmp(big.NewInt(0)) != 1 {
		logger.Error("operatorSender has not enough Balance")
	}
	if c.OperatorAddress != operatorSender.String() {
		logger.Error("operatorSender is not equal OperatorAddress")
	}
	txMgr := txmgr.NewSimpleTxManager(ethRpcClient, logger, signer, common.HexToAddress(c.OperatorAddress))

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

	logger.Info("Operator info",
		"operatorAddr", c.OperatorAddress,
		"operatorKey", operator.blsKeypair.PublicKey().Marshal(),
	)

	return operator, nil
}

// func (o *Operator) registerOperatorOnStartup() {
// 	// Register operator to chain
// 	operatorAddress, err := core.SwitchEthAddressToImAddress(o.operatorAddr.String())
// 	if err != nil {
// 		o.logger.Error("Cannot switch eth address to im address", "err", err)
// 		panic(err)
// 	}

// 	// Check if operator is already registered
// 	flag, err := o.avsReader.IsOperator(&bind.CallOpts{}, o.operatorAddr.String())
// 	if err != nil {
// 		o.logger.Error("Cannot exec IsOperator", "err", err)
// 		panic(err)
// 	}
// 	if !flag {
// 		o.logger.Error("Operator is not registered.", "err", err)
// 		panic(fmt.Sprintf("Operator is not registered: %s", operatorAddress))
// 	}

// 	// Register BLS Public Key if not already registered
// 	pubKey, err := o.avsReader.GetRegisteredPubkey(&bind.CallOpts{}, o.operatorAddr.String(), o.avsAddr.String())
// 	if err != nil {
// 		o.logger.Error("Cannot exec GetRegisteredPubKey", "err", err)
// 		panic(err)
// 	}

// 	if len(pubKey) == 0 {
// 		// Register BLS Public Key via EVM transaction
// 		msg := fmt.Sprintf(core.BLSMessageToSign,
// 			core.ChainIDWithoutRevision("imuachainlocalnet_232"), operatorAddress)
// 		hashedMsg := crypto.Keccak256Hash([]byte(msg))
// 		sig := o.blsKeypair.Sign(hashedMsg.Bytes())

// 		_, err = o.avsWriter.RegisterBLSPublicKey(
// 			context.Background(),
// 			o.avsAddr.String(),
// 			o.blsKeypair.PublicKey().Marshal(),
// 			sig.Marshal())

// 		if err != nil {
// 			o.logger.Error("operator failed to registerBLSPublicKey", "err", err)
// 			panic(err)
// 		}
// 		o.logger.Info("BLS Public Key registered successfully")
// 	}
// }

func (o *Operator) RegisterOperator() error {
	operatorAddress, err := core.SwitchEthAddressToImAddress(o.operatorAddr.String())
	if err != nil {
		o.logger.Error("Cannot switch eth address to im address", "err", err)
		return err
	}

	// Check if operator is registered
	flag, err := o.avsReader.IsOperator(&bind.CallOpts{}, o.operatorAddr.String())
	if err != nil {
		o.logger.Error("Cannot exec IsOperator", "err", err)
		return err
	}
	if !flag {
		return fmt.Errorf("operator is not registered: %s", operatorAddress)
	}

	// Register BLS Public Key
	pubKey, err := o.avsReader.GetRegisteredPubkey(&bind.CallOpts{}, o.operatorAddr.String(), o.avsAddr.String())
	if err != nil {
		o.logger.Error("Cannot exec GetRegisteredPubKey", "err", err)
		return err
	}

	if len(pubKey) == 0 {
		msg := fmt.Sprintf(core.BLSMessageToSign,
			core.ChainIDWithoutRevision("imuachainlocalnet_232"), operatorAddress)
		hashedMsg := crypto.Keccak256Hash([]byte(msg))
		sig := o.blsKeypair.Sign(hashedMsg.Bytes())

		_, err = o.avsWriter.RegisterBLSPublicKey(
			context.Background(),
			o.avsAddr.String(),
			o.blsKeypair.PublicKey().Marshal(),
			sig.Marshal())

		if err != nil {
			o.logger.Error("operator failed to registerBLSPublicKey", "err", err)
			return err
		}
	}

	// Ensure operator has sufficient delegation
	return o.ensureDelegation()
}

func (o *Operator) ensureDelegation() error {
	for attempt := 1; attempt <= maxRetries; attempt++ {
		// Check operator delegation USD amount
		amount, err := o.avsReader.GetOperatorOptedUSDValue(&bind.CallOpts{}, o.avsAddr.String(), o.operatorAddr.String())
		if err != nil {
			o.logger.Error("Cannot exec GetOperatorOptedUSDValue", "err", err)
			return err
		}

		if !amount.IsZero() && !amount.IsNegative() {
			o.logger.Info("Operator has sufficient delegation", "amount", amount)
			break
		}

		// Perform deposit and delegation if amount is zero
		if amount.IsZero() {
			err := o.Deposit()
			if err != nil {
				return fmt.Errorf("cannot deposit: %w", err)
			}
			err = o.Delegate()
			if err != nil {
				return fmt.Errorf("cannot delegate: %w", err)
			}
			err = o.SelfDelegate()
			if err != nil {
				return fmt.Errorf("cannot self delegate: %w", err)
			}
		}

		// Wait for delegation to be processed
		for waitAttempt := 1; waitAttempt <= maxRetries; waitAttempt++ {
			amount, err := o.avsReader.GetOperatorOptedUSDValue(&bind.CallOpts{}, o.avsAddr.String(), o.operatorAddr.String())
			if err == nil && !amount.IsZero() && !amount.IsNegative() {
				return nil
			}

			if err != nil {
				o.logger.Error("Cannot GetOperatorOptedUSDValue",
					"err", err,
					"attempt", waitAttempt,
					"max_attempts", maxRetries)
			} else {
				o.logger.Info("OperatorOptedUSDValue is zero or negative",
					"operator_usd_value", amount,
					"attempt", waitAttempt,
					"max_attempts", maxRetries)
			}
			time.Sleep(retryDelay)
		}
	}

	return fmt.Errorf("failed to ensure delegation after %d attempts", maxRetries)
}

// Placeholder functions - implement based on your requirements
func (o *Operator) Deposit() error {
	// Implement deposit logic
	o.logger.Info("Executing deposit...")
	return nil
}

func (o *Operator) Delegate() error {
	// Implement delegation logic
	o.logger.Info("Executing delegation...")
	return nil
}

func (o *Operator) SelfDelegate() error {
	// Implement self-delegation logic
	o.logger.Info("Executing self-delegation...")
	return nil
}
