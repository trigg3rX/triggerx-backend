package chainio

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	sdkcommon "github.com/trigg3rX/triggerx-backend/pkg/common"

	"github.com/Layr-Labs/eigensdk-go/chainio/clients/avsregistry"
	"github.com/Layr-Labs/eigensdk-go/chainio/txmgr"
	"github.com/Layr-Labs/eigensdk-go/crypto/bls"
	logging "github.com/Layr-Labs/eigensdk-go/logging"
	eigenSdkTypes "github.com/Layr-Labs/eigensdk-go/types"

	txtaskmanager "github.com/trigg3rX/triggerx-backend/pkg/avsinterface/bindings/TriggerXTaskManager"
	"github.com/trigg3rX/triggerx-backend/pkg/core/config"
)

type SignatureWithSaltAndExpiry struct {
	Signature []byte
	Salt      [32]byte
	Expiry    *big.Int
}

type AvsWriterer interface {
	// Task Management
	CreateNewTask(
		ctx context.Context,
		jobId uint32,
		quorumNumbers []byte,
		quorumThreshold uint8,
	) ([8]byte, error)

	RespondToTask(
		ctx context.Context,
		task txtaskmanager.ITriggerXTaskManagerTask,
		taskResponse txtaskmanager.ITriggerXTaskManagerTaskResponse,
		nonSignerStakesAndSignature txtaskmanager.IBLSSignatureCheckerNonSignerStakesAndSignature,
	) (*types.Receipt, error)

	// Operator Management
	RegisterOperatorInQuorumWithAVSRegistryCoordinator(
		ctx context.Context,
		operatorEcdsaKeyPair *ecdsa.PrivateKey,
		registrationSigSalt [32]byte,
		registrationSigExpiry *big.Int,
		blsKeyPair *bls.KeyPair,
		quorumNumbers eigenSdkTypes.QuorumNums,
		socket string,
		shouldWaitForConfirmation bool,
	) (*types.Receipt, error)

	DeregisterOperatorFromAVS(
		ctx context.Context,
		operatorAddr common.Address,
	) (*types.Receipt, error)

	// Utility
	GetTxMgr() txmgr.TxManager
}

type AvsWriter struct {
	avsregistry.ChainWriter
	AvsContractBindings *AvsManagersBindings
	logger              logging.Logger
	TxMgr               txmgr.TxManager
}

var _ AvsWriterer = (*AvsWriter)(nil)

func BuildAvsWriterFromConfig(c *config.Config) (*AvsWriter, error) {
	return BuildAvsWriter(c.TxMgr, c.TriggerXServiceManagerAddr, c.OperatorStateRetrieverAddr, &c.EthHttpClient, c.Logger)
}

func BuildAvsWriter(txMgr txmgr.TxManager, registryCoordinatorAddr, operatorStateRetrieverAddr common.Address, ethHttpClient sdkcommon.EthClientInterface, logger logging.Logger) (*AvsWriter, error) {
	avsServiceBindings, err := NewAvsManagersBindings(registryCoordinatorAddr, operatorStateRetrieverAddr, ethHttpClient, logger)
	if err != nil {
		logger.Error("Failed to create contract bindings", "err", err)
		return nil, err
	}
	avsRegistryWriter, err := avsregistry.BuildAvsRegistryChainWriter(registryCoordinatorAddr, operatorStateRetrieverAddr, logger, ethHttpClient, txMgr)
	if err != nil {
		return nil, err
	}
	return NewAvsWriter(*avsRegistryWriter, avsServiceBindings, logger, txMgr), nil
}
func NewAvsWriter(avsRegistryWriter avsregistry.ChainWriter, avsServiceBindings *AvsManagersBindings, logger logging.Logger, txMgr txmgr.TxManager) *AvsWriter {
	return &AvsWriter{
		ChainWriter:         avsRegistryWriter,
		AvsContractBindings: avsServiceBindings,
		logger:              logger,
		TxMgr:               txMgr,
	}
}

func (w *AvsWriter) CreateNewTask(
	ctx context.Context,
	jobId uint32,
	quorumNumbers []byte,
	quorumThreshold uint8,
) ([8]byte, error) {
	txOpts, err := w.TxMgr.GetNoSendTxOpts()
	if err != nil {
		w.logger.Errorf("Error getting tx opts")
		return [8]byte{}, err
	}

	tx, err := w.AvsContractBindings.TaskManager.CreateNewTask(txOpts, jobId, quorumNumbers, quorumThreshold)
	if err != nil {
		w.logger.Errorf("Error assembling CreateNewTask tx")
		return [8]byte{}, err
	}

	receipt, err := w.TxMgr.Send(ctx, tx, true)
	if err != nil {
		w.logger.Errorf("Error submitting CreateNewTask tx")
		return [8]byte{}, err
	}

	taskCreatedEvent, err := w.AvsContractBindings.TaskManager.ParseTaskCreated(*receipt.Logs[0])
	if err != nil {
		w.logger.Error("Failed to parse new task created event", "err", err)
		return [8]byte{}, err
	}

	return taskCreatedEvent.TaskId, nil
}

func (w *AvsWriter) RespondToTask(
	ctx context.Context,
	task txtaskmanager.ITriggerXTaskManagerTask,
	taskResponse txtaskmanager.ITriggerXTaskManagerTaskResponse,
	nonSignerStakesAndSignature txtaskmanager.IBLSSignatureCheckerNonSignerStakesAndSignature,
) (*types.Receipt, error) {
	txOpts, err := w.TxMgr.GetNoSendTxOpts()
	if err != nil {
		w.logger.Errorf("Error getting tx opts")
		return nil, err
	}

	tx, err := w.AvsContractBindings.TaskManager.RespondToTask(txOpts, task, taskResponse, nonSignerStakesAndSignature)
	if err != nil {
		w.logger.Error("Error assembling RespondToTask tx", "err", err)
		return nil, err
	}

	receipt, err := w.TxMgr.Send(ctx, tx, true)
	if err != nil {
		w.logger.Errorf("Error submitting RespondToTask tx")
		return nil, err
	}

	return receipt, nil
}

func (w *AvsWriter) GetTxMgr() txmgr.TxManager {
	return w.TxMgr
}

func (w *AvsWriter) RegisterOperatorInQuorumWithAVSRegistryCoordinator(
	ctx context.Context,
	operatorEcdsaKeyPair *ecdsa.PrivateKey,
	registrationSigSalt [32]byte,
	registrationSigExpiry *big.Int,
	blsKeyPair *bls.KeyPair,
	quorumNumbers eigenSdkTypes.QuorumNums,
	socket string,
	shouldWaitForConfirmation bool,
) (*types.Receipt, error) {
	operatorAddr := crypto.PubkeyToAddress(operatorEcdsaKeyPair.PublicKey)

	// Create the registration message hash
	// This should match the contract's signing scheme
	registrationData := []byte("triggerx-registration")
	messageHash := crypto.Keccak256(
		registrationData,
		operatorAddr.Bytes(),
		registrationSigSalt[:],
		common.LeftPadBytes(registrationSigExpiry.Bytes(), 32),
	)

	// Sign the registration message
	signature, err := crypto.Sign(messageHash, operatorEcdsaKeyPair)
	if err != nil {
		return nil, fmt.Errorf("failed to sign registration message: %w", err)
	}

	// Create the signature struct
	operatorSignature := SignatureWithSaltAndExpiry{
		Signature: signature,
		Salt:      registrationSigSalt,
		Expiry:    registrationSigExpiry,
	}

	// Get transaction options
	txOpts, err := w.GetTxMgr().GetNoSendTxOpts()
	if err != nil {
		return nil, fmt.Errorf("failed to get tx opts: %w", err)
	}

	// Call the contract
	tx, err := w.AvsContractBindings.ServiceManager.RegisterOperatorToAVS(
		txOpts,
		operatorAddr,
		struct {
			Signature []byte
			Salt      [32]byte
			Expiry    *big.Int
		}{
			Signature: operatorSignature.Signature,
			Salt:      operatorSignature.Salt,
			Expiry:    operatorSignature.Expiry,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to register operator: %w", err)
	}

	// Send the transaction
	receipt, err := w.GetTxMgr().Send(ctx, tx, shouldWaitForConfirmation)
	if err != nil {
		return nil, fmt.Errorf("failed to send registration tx: %w", err)
	}

	return receipt, nil
}

func (w *AvsWriter) DeregisterOperatorFromAVS(
	ctx context.Context,
	operatorAddr common.Address,
) (*types.Receipt, error) {
	txOpts, err := w.TxMgr.GetNoSendTxOpts()
	if err != nil {
		w.logger.Error("Error getting tx opts", "err", err)
		return nil, err
	}

	tx, err := w.AvsContractBindings.ServiceManager.DeregisterOperatorFromAVS(txOpts, operatorAddr)
	if err != nil {
		w.logger.Error("Error assembling DeregisterOperatorFromAVS tx", "err", err)
		return nil, err
	}

	receipt, err := w.TxMgr.Send(ctx, tx, true)
	if err != nil {
		w.logger.Error("Error submitting DeregisterOperatorFromAVS tx", "err", err)
		return nil, err
	}

	return receipt, nil
}
