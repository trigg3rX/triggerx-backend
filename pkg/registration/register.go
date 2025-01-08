package registration

import (
	"context"
	"crypto/ecdsa"
	"encoding/hex"
	"fmt"
	"math/big"

	"github.com/Layr-Labs/eigensdk-go/crypto/bls"
	sdklogging "github.com/Layr-Labs/eigensdk-go/logging"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"

	// eigenSdkTypes "github.com/Layr-Labs/eigensdk-go/types"
	sdkcommon "github.com/trigg3rX/triggerx-backend/pkg/common"

	txservicemanager "github.com/trigg3rX/triggerx-backend/pkg/avsinterface/bindings/TriggerXServiceManager"

	sdkelcontracts "github.com/Layr-Labs/eigensdk-go/chainio/clients/elcontracts"
	"github.com/trigg3rX/triggerx-backend/pkg/core/chainio"
)

// Registration handles the registration process for keepers
type Registration struct {
	logger           sdklogging.Logger
	ethClient        sdkcommon.EthClientInterface
	avsReader        chainio.AvsReaderer
	avsWriter        chainio.AvsWriterer
	eigenlayerReader sdkelcontracts.ChainReader
	eigenlayerWriter sdkelcontracts.ChainWriter
	keeperAddr       common.Address
	blsKeypair       *bls.KeyPair
}

func NewRegistration(
	logger sdklogging.Logger,
	ethClient sdkcommon.EthClientInterface,
	avsReader chainio.AvsReaderer,
	avsWriter chainio.AvsWriterer,
	eigenlayerReader sdkelcontracts.ChainReader,
	keeperAddr common.Address,
	blsKeypair *bls.KeyPair,
) *Registration {
	if avsReader == nil {
		panic("avsReader cannot be nil")
	}

	return &Registration{
		logger:           logger,
		ethClient:        ethClient,
		avsReader:        avsReader,
		avsWriter:        avsWriter,
		eigenlayerReader: eigenlayerReader,
		keeperAddr:       keeperAddr,
		blsKeypair:       blsKeypair,
	}
}

func (r *Registration) DepositIntoStrategy(strategyAddr common.Address, amount *big.Int) error {
	_, tokenAddr, err := r.eigenlayerReader.GetStrategyAndUnderlyingToken(context.Background(), strategyAddr)
	if err != nil {
		r.logger.Error("Failed to fetch strategy contract", "err", err)
		return err
	}

	contractErc20Mock, err := r.avsReader.GetErc20Mock(&bind.CallOpts{}, tokenAddr)
	if err != nil {
		r.logger.Error("Failed to fetch ERC20Mock contract", "err", err)
		return err
	}

	txOpts, err := r.avsWriter.GetTxMgr().GetNoSendTxOpts()
	if err != nil {
		r.logger.Error("Error getting no send tx opts")
		return err
	}

	tx, err := contractErc20Mock.Mint(txOpts, r.keeperAddr, amount)
	if err != nil {
		r.logger.Error("Error assembling Mint tx")
		return err
	}

	_, err = r.avsWriter.GetTxMgr().Send(context.Background(), tx, true)
	if err != nil {
		r.logger.Error("Error submitting Mint tx")
		return err
	}

	_, err = r.eigenlayerWriter.DepositERC20IntoStrategy(context.Background(), strategyAddr, amount, true)
	if err != nil {
		r.logger.Error("Error depositing into strategy", "err", err)
		return err
	}
	return nil
}

func (r *Registration) RegisterOperatorWithAvs(operatorEcdsaKeyPair *ecdsa.PrivateKey) (string, string, error) {
	// 1. First check if operator is already registered
	operatorId, err := r.avsReader.GetOperatorId(&bind.CallOpts{}, r.keeperAddr)
	if err != nil {
		return "", "", fmt.Errorf("failed to check operator registration: %w", err)
	}
	if operatorId != [32]byte{} {
		r.logger.Info("Operator already registered with AVS")
		return hex.EncodeToString(operatorId[:]), "", nil
	}

	// Generate a random salt for registration
	operatorToAvsRegistrationSigSalt := [32]byte{123}

	// 3. Get current block for expiry calculation
	curBlockNum, err := r.ethClient.BlockNumber(context.Background())
	if err != nil {
		return "", "", fmt.Errorf("unable to get current block number: %w", err)
	}

	curBlock, err := r.ethClient.HeaderByNumber(context.Background(), big.NewInt(int64(curBlockNum)))
	if err != nil {
		return "", "", fmt.Errorf("unable to get current block: %w", err)
	}

	// Set signature validity period (about 11.5 days)
	sigValidForSeconds := int64(1_000_000)
	operatorToAvsRegistrationSigExpiry := big.NewInt(int64(curBlock.Time) + sigValidForSeconds)

	// 4. Register operator with AVS
	receipt, err := r.avsWriter.RegisterKeeperToTriggerX(
		context.Background(),
		r.keeperAddr,
		txservicemanager.ISignatureUtilsSignatureWithSaltAndExpiry{
			Salt:   operatorToAvsRegistrationSigSalt,
			Expiry: operatorToAvsRegistrationSigExpiry,
		},
	)
	if err != nil {
		return "", "", fmt.Errorf("failed to register operator with AVS registry coordinator: %w", err)
	}

	// 5. Verify registration
	newOperatorId, err := r.avsReader.GetOperatorId(&bind.CallOpts{}, r.keeperAddr)
	if err != nil {
		return "", "", fmt.Errorf("failed to verify operator registration: %w", err)
	}
	if newOperatorId == [32]byte{} {
		return "", "", fmt.Errorf("operator registration failed: operator ID is still empty after registration")
	}

	r.logger.Info("Registration verified successfully",
		"operatorId", hex.EncodeToString(newOperatorId[:]),
		"address", r.keeperAddr.Hex(),
	)

	return hex.EncodeToString(newOperatorId[:]), receipt.Hex(), nil
}
