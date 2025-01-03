package chainio

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	sdkcommon "github.com/trigg3rX/triggerx-backend/pkg/common"

	"github.com/Layr-Labs/eigensdk-go/chainio/clients/avsregistry"
	"github.com/Layr-Labs/eigensdk-go/chainio/txmgr"
	logging "github.com/Layr-Labs/eigensdk-go/logging"

	registrycoordinator "github.com/trigg3rX/triggerx-backend/pkg/avsinterface/bindings/RegistryCoordinator"
	txservicemanager "github.com/trigg3rX/triggerx-backend/pkg/avsinterface/bindings/TriggerXServiceManager"
	txtaskmanager "github.com/trigg3rX/triggerx-backend/pkg/avsinterface/bindings/TriggerXTaskManager"
	"github.com/trigg3rX/triggerx-backend/pkg/core/config"
)

type SignatureWithSaltAndExpiry struct {
	Signature []byte
	Salt      [32]byte
	Expiry    *big.Int
}

type AvsWriterer interface {
	// TtiggerXTaskManager Functions
	CreateNewTask(
		ctx context.Context,
		jobId uint32,
		quorumNumbers []byte,
		quorumThreshold uint8,
	) (*types.Transaction, error)
	RespondToTask(
		ctx context.Context,
		task txtaskmanager.ITriggerXTaskManagerTask,
		taskResponse txtaskmanager.ITriggerXTaskManagerTaskResponse,
		nonSignerStakesAndSignature txtaskmanager.IBLSSignatureCheckerNonSignerStakesAndSignature,
	) (*types.Transaction, error)

	// ServiceManager - Keeper Management
	RegisterKeeperToTriggerX(
		ctx context.Context,
		operator common.Address,
		operatorSignature txservicemanager.ISignatureUtilsSignatureWithSaltAndExpiry,
	) (*types.Transaction, error)
	DeregisterKeeperFromTriggerX(ctx context.Context, operator common.Address) (*types.Transaction, error)
	BlacklistKeeper(ctx context.Context, operator common.Address) (*types.Transaction, error)
	UnblacklistKeeper(ctx context.Context, operator common.Address) (*types.Transaction, error)
	RegisterOperatorToAVS(
		ctx context.Context,
		operator common.Address,
		operatorSignature txservicemanager.ISignatureUtilsSignatureWithSaltAndExpiry,
	) (*types.Transaction, error)
	DeregisterOperatorFromAVS(ctx context.Context, operator common.Address) (*types.Transaction, error)
	CreateAVSRewardsSubmission(
		ctx context.Context,
		rewardsSubmissions []txservicemanager.IRewardsCoordinatorRewardsSubmission,
	) (*types.Transaction, error)
	CreateOperatorDirectedAVSRewardsSubmission(
		ctx context.Context,
		operatorDirectedRewardsSubmissions []txservicemanager.IRewardsCoordinatorOperatorDirectedRewardsSubmission,
	) (*types.Transaction, error)

	// StakeRegistry Functions
	Stake(ctx context.Context, amount *big.Int) (*types.Transaction, error)
	Unstake(ctx context.Context, amount *big.Int) (*types.Transaction, error)
	RemoveStake(ctx context.Context, user common.Address, amount *big.Int, reason string) (*types.Transaction, error)

	// RegistryCoordinator Functions
	RegisterOperator(
		ctx context.Context,
		quorumNumbers []byte,
		socket string,
		params registrycoordinator.IBLSApkRegistryPubkeyRegistrationParams,
		operatorSignature registrycoordinator.ISignatureUtilsSignatureWithSaltAndExpiry,
	) (*types.Transaction, error)
	DeregisterOperator(
		ctx context.Context,
		quorumNumbers []byte,
	) (*types.Transaction, error)
	RegisterOperatorWithChurn(
		ctx context.Context,
		quorumNumbers []byte,
		socket string,
		params registrycoordinator.IBLSApkRegistryPubkeyRegistrationParams,
		operatorKickParams []registrycoordinator.IRegistryCoordinatorOperatorKickParam,
		churnApproverSignature registrycoordinator.ISignatureUtilsSignatureWithSaltAndExpiry,
		operatorSignature registrycoordinator.ISignatureUtilsSignatureWithSaltAndExpiry,
	) (*types.Transaction, error)
	EjectOperator(
		ctx context.Context,
		operator common.Address,
		quorumNumbers []byte,
	) (*types.Transaction, error)
	UpdateOperatorSetParams(
		ctx context.Context,
		quorumNumber uint8,
		operatorSetParams registrycoordinator.IRegistryCoordinatorOperatorSetParam,
	) (*types.Transaction, error)
	UpdateQuorumOperatorSetParams(
		ctx context.Context,
		quorumNumbers uint8,
		operatorSetParams registrycoordinator.IRegistryCoordinatorOperatorSetParam,
	) (*types.Transaction, error)
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

func (w *AvsWriter) CreateNewTask(ctx context.Context, jobId uint32, quorumNumbers []byte, quorumThreshold uint8) (*types.Transaction, error) {
	w.logger.Info("Creating new task", "jobId", jobId, "quorumThreshold", quorumThreshold)

	txOpts, err := w.TxMgr.GetNoSendTxOpts()
	if err != nil {
		return nil, err
	}

	tx, err := w.AvsContractBindings.TaskManager.CreateNewTask(txOpts, jobId, quorumNumbers, quorumThreshold)
	if err != nil {
		return nil, fmt.Errorf("failed to create new task: %w", err)
	}

	_, err = w.TxMgr.Send(ctx, tx, true)
	return tx, err
}

func (w *AvsWriter) RespondToTask(ctx context.Context, task txtaskmanager.ITriggerXTaskManagerTask, taskResponse txtaskmanager.ITriggerXTaskManagerTaskResponse, nonSignerStakesAndSignature txtaskmanager.IBLSSignatureCheckerNonSignerStakesAndSignature) (*types.Transaction, error) {
	w.logger.Info("Responding to task", "taskId", taskResponse.TaskId)

	txOpts, err := w.TxMgr.GetNoSendTxOpts()
	if err != nil {
		return nil, err
	}

	tx, err := w.AvsContractBindings.TaskManager.RespondToTask(txOpts, task, taskResponse, nonSignerStakesAndSignature)
	if err != nil {
		return nil, fmt.Errorf("failed to respond to task: %w", err)
	}

	_, err = w.TxMgr.Send(ctx, tx, true)
	return tx, err
}

func (w *AvsWriter) RegisterKeeperToTriggerX(ctx context.Context, operator common.Address, operatorSignature txservicemanager.ISignatureUtilsSignatureWithSaltAndExpiry) (*types.Transaction, error) {
	w.logger.Info("Registering keeper to TriggerX", "operator", operator.Hex())

	txOpts, err := w.TxMgr.GetNoSendTxOpts()
	if err != nil {
		return nil, err
	}

	tx, err := w.AvsContractBindings.ServiceManager.RegisterKeeperToTriggerX(txOpts, operator, operatorSignature)
	if err != nil {
		return nil, fmt.Errorf("failed to register keeper: %w", err)
	}

	_, err = w.TxMgr.Send(ctx, tx, true)
	return tx, err
}

func (w *AvsWriter) DeregisterKeeperFromTriggerX(ctx context.Context, operator common.Address) (*types.Transaction, error) {
	w.logger.Info("Deregistering keeper from TriggerX", "operator", operator.Hex())

	txOpts, err := w.TxMgr.GetNoSendTxOpts()
	if err != nil {
		return nil, err
	}

	tx, err := w.AvsContractBindings.ServiceManager.DeregisterKeeperFromTriggerX(txOpts, operator)
	if err != nil {
		return nil, fmt.Errorf("failed to deregister keeper: %w", err)
	}

	_, err = w.TxMgr.Send(ctx, tx, true)
	return tx, err
}

func (w *AvsWriter) BlacklistKeeper(ctx context.Context, operator common.Address) (*types.Transaction, error) {
	w.logger.Info("Blacklisting keeper", "operator", operator.Hex())

	txOpts, err := w.TxMgr.GetNoSendTxOpts()
	if err != nil {
		return nil, err
	}

	tx, err := w.AvsContractBindings.ServiceManager.BlacklistKeeper(txOpts, operator)
	if err != nil {
		return nil, fmt.Errorf("failed to blacklist keeper: %w", err)
	}

	_, err = w.TxMgr.Send(ctx, tx, true)
	return tx, err
}

func (w *AvsWriter) UnblacklistKeeper(ctx context.Context, operator common.Address) (*types.Transaction, error) {
	w.logger.Info("Unblacklisting keeper", "operator", operator.Hex())

	txOpts, err := w.TxMgr.GetNoSendTxOpts()
	if err != nil {
		return nil, err
	}

	tx, err := w.AvsContractBindings.ServiceManager.UnblacklistKeeper(txOpts, operator)
	if err != nil {
		return nil, fmt.Errorf("failed to unblacklist keeper: %w", err)
	}

	_, err = w.TxMgr.Send(ctx, tx, true)
	return tx, err
}

func (w *AvsWriter) RegisterOperatorToAVS(ctx context.Context, operator common.Address, operatorSignature txservicemanager.ISignatureUtilsSignatureWithSaltAndExpiry) (*types.Transaction, error) {
	w.logger.Info("Registering operator to AVS", "operator", operator.Hex())

	txOpts, err := w.TxMgr.GetNoSendTxOpts()
	if err != nil {
		return nil, err
	}

	tx, err := w.AvsContractBindings.ServiceManager.RegisterOperatorToAVS(txOpts, operator, operatorSignature)
	if err != nil {
		return nil, fmt.Errorf("failed to register operator: %w", err)
	}

	_, err = w.TxMgr.Send(ctx, tx, true)
	return tx, err
}

func (w *AvsWriter) DeregisterOperatorFromAVS(ctx context.Context, operator common.Address) (*types.Transaction, error) {
	w.logger.Info("Deregistering operator from AVS", "operator", operator.Hex())

	txOpts, err := w.TxMgr.GetNoSendTxOpts()
	if err != nil {
		return nil, err
	}

	tx, err := w.AvsContractBindings.ServiceManager.DeregisterOperatorFromAVS(txOpts, operator)
	if err != nil {
		return nil, fmt.Errorf("failed to deregister operator: %w", err)
	}

	_, err = w.TxMgr.Send(ctx, tx, true)
	return tx, err
}

func (w *AvsWriter) CreateAVSRewardsSubmission(ctx context.Context, rewardsSubmissions []txservicemanager.IRewardsCoordinatorRewardsSubmission) (*types.Transaction, error) {
	w.logger.Info("Creating AVS rewards submission")

	txOpts, err := w.TxMgr.GetNoSendTxOpts()
	if err != nil {
		return nil, err
	}

	tx, err := w.AvsContractBindings.ServiceManager.CreateAVSRewardsSubmission(txOpts, rewardsSubmissions)
	if err != nil {
		return nil, fmt.Errorf("failed to create rewards submission: %w", err)
	}

	_, err = w.TxMgr.Send(ctx, tx, true)
	return tx, err
}

func (w *AvsWriter) CreateOperatorDirectedAVSRewardsSubmission(ctx context.Context, operatorDirectedRewardsSubmissions []txservicemanager.IRewardsCoordinatorOperatorDirectedRewardsSubmission) (*types.Transaction, error) {
	w.logger.Info("Creating operator directed AVS rewards submission")

	txOpts, err := w.TxMgr.GetNoSendTxOpts()
	if err != nil {
		return nil, err
	}

	tx, err := w.AvsContractBindings.ServiceManager.CreateOperatorDirectedAVSRewardsSubmission(txOpts, operatorDirectedRewardsSubmissions)
	if err != nil {
		return nil, fmt.Errorf("failed to create operator directed rewards submission: %w", err)
	}

	_, err = w.TxMgr.Send(ctx, tx, true)
	return tx, err
}

func (w *AvsWriter) Stake(ctx context.Context, amount *big.Int) (*types.Transaction, error) {
	w.logger.Info("Staking tokens", "amount", amount)

	txOpts, err := w.TxMgr.GetNoSendTxOpts()
	if err != nil {
		return nil, err
	}

	tx, err := w.AvsContractBindings.StakeRegistry.Stake(txOpts, amount)
	if err != nil {
		return nil, fmt.Errorf("failed to stake tokens: %w", err)
	}

	_, err = w.TxMgr.Send(ctx, tx, true)
	return tx, err
}

func (w *AvsWriter) Unstake(ctx context.Context, amount *big.Int) (*types.Transaction, error) {
	w.logger.Info("Unstaking tokens", "amount", amount)

	txOpts, err := w.TxMgr.GetNoSendTxOpts()
	if err != nil {
		return nil, err
	}

	tx, err := w.AvsContractBindings.StakeRegistry.Unstake(txOpts, amount)
	if err != nil {
		return nil, fmt.Errorf("failed to unstake tokens: %w", err)
	}

	_, err = w.TxMgr.Send(ctx, tx, true)
	return tx, err
}

func (w *AvsWriter) RemoveStake(ctx context.Context, user common.Address, amount *big.Int, reason string) (*types.Transaction, error) {
	w.logger.Info("Removing stake", "user", user.Hex(), "amount", amount, "reason", reason)

	txOpts, err := w.TxMgr.GetNoSendTxOpts()
	if err != nil {
		return nil, err
	}

	tx, err := w.AvsContractBindings.StakeRegistry.RemoveStake(txOpts, user, amount, reason)
	if err != nil {
		return nil, fmt.Errorf("failed to remove stake: %w", err)
	}

	_, err = w.TxMgr.Send(ctx, tx, true)
	return tx, err
}

func (w *AvsWriter) RegisterOperator(ctx context.Context, quorumNumbers []byte, socket string, params registrycoordinator.IBLSApkRegistryPubkeyRegistrationParams, operatorSignature registrycoordinator.ISignatureUtilsSignatureWithSaltAndExpiry) (*types.Transaction, error) {
	w.logger.Info("Registering operator", "quorumNumbers", quorumNumbers, "socket", socket)

	txOpts, err := w.TxMgr.GetNoSendTxOpts()
	if err != nil {
		return nil, err
	}

	tx, err := w.AvsContractBindings.RegistryCoordinator.RegisterOperator(txOpts, quorumNumbers, socket, params, operatorSignature)
	if err != nil {
		return nil, fmt.Errorf("failed to register operator: %w", err)
	}

	_, err = w.TxMgr.Send(ctx, tx, true)
	return tx, err
}

func (w *AvsWriter) DeregisterOperator(ctx context.Context, quorumNumbers []byte) (*types.Transaction, error) {
	w.logger.Info("Deregistering operator")

	txOpts, err := w.TxMgr.GetNoSendTxOpts()
	if err != nil {
		return nil, err
	}

	tx, err := w.AvsContractBindings.RegistryCoordinator.DeregisterOperator(txOpts, quorumNumbers)
	if err != nil {
		return nil, fmt.Errorf("failed to deregister operator: %w", err)
	}

	_, err = w.TxMgr.Send(ctx, tx, true)
	return tx, err
}

func (w *AvsWriter) UpdateSocket(ctx context.Context, socket string) (*types.Transaction, error) {
	w.logger.Info("Updating socket", "socket", socket)

	txOpts, err := w.TxMgr.GetNoSendTxOpts()
	if err != nil {
		return nil, err
	}

	tx, err := w.AvsContractBindings.RegistryCoordinator.UpdateSocket(txOpts, socket)
	if err != nil {
		return nil, fmt.Errorf("failed to update socket: %w", err)
	}

	_, err = w.TxMgr.Send(ctx, tx, true)
	return tx, err
}

func (w *AvsWriter) RegisterOperatorWithChurn(
	ctx context.Context,
	quorumNumbers []byte,
	socket string,
	params registrycoordinator.IBLSApkRegistryPubkeyRegistrationParams,
	operatorKickParams []registrycoordinator.IRegistryCoordinatorOperatorKickParam,
	churnApproverSignature registrycoordinator.ISignatureUtilsSignatureWithSaltAndExpiry,
	operatorSignature registrycoordinator.ISignatureUtilsSignatureWithSaltAndExpiry,
) (*types.Transaction, error) {
	w.logger.Info("Registering operator with churn", "quorumNumbers", quorumNumbers, "socket", socket)

	txOpts, err := w.TxMgr.GetNoSendTxOpts()
	if err != nil {
		return nil, err
	}

	tx, err := w.AvsContractBindings.RegistryCoordinator.RegisterOperatorWithChurn(
		txOpts,
		quorumNumbers,
		socket,
		params,
		operatorKickParams,
		churnApproverSignature,
		operatorSignature,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to register operator with churn: %w", err)
	}

	_, err = w.TxMgr.Send(ctx, tx, true)
	return tx, err
}

func (w *AvsWriter) EjectOperator(ctx context.Context, operator common.Address, quorumNumbers []byte) (*types.Transaction, error) {
	w.logger.Info("Ejecting operator", "operator", operator.Hex())

	txOpts, err := w.TxMgr.GetNoSendTxOpts()
	if err != nil {
		return nil, err
	}

	tx, err := w.AvsContractBindings.RegistryCoordinator.EjectOperator(txOpts, operator, quorumNumbers)
	if err != nil {
		return nil, fmt.Errorf("failed to eject operator: %w", err)
	}

	_, err = w.TxMgr.Send(ctx, tx, true)
	return tx, err
}

func (w *AvsWriter) UpdateOperatorSetParams(ctx context.Context, quorumNumber uint8, operatorSetParams registrycoordinator.IRegistryCoordinatorOperatorSetParam) (*types.Transaction, error) {
	w.logger.Info("Updating operator set params", "quorumNumber", quorumNumber)

	txOpts, err := w.TxMgr.GetNoSendTxOpts()
	if err != nil {
		return nil, err
	}

	tx, err := w.AvsContractBindings.RegistryCoordinator.SetOperatorSetParams(txOpts, quorumNumber, operatorSetParams)
	if err != nil {
		return nil, fmt.Errorf("failed to update operator set params: %w", err)
	}

	_, err = w.TxMgr.Send(ctx, tx, true)
	return tx, err
}

func (w *AvsWriter) UpdateQuorumOperatorSetParams(ctx context.Context, quorumNumbers uint8, operatorSetParams registrycoordinator.IRegistryCoordinatorOperatorSetParam) (*types.Transaction, error) {
	w.logger.Info("Updating quorum operator set params", "quorumNumbers", quorumNumbers)

	txOpts, err := w.TxMgr.GetNoSendTxOpts()
	if err != nil {
		return nil, err
	}

	tx, err := w.AvsContractBindings.RegistryCoordinator.SetOperatorSetParams(txOpts, quorumNumbers, operatorSetParams)
	if err != nil {
		return nil, fmt.Errorf("failed to update quorum operator set params: %w", err)
	}

	_, err = w.TxMgr.Send(ctx, tx, true)
	return tx, err
}
