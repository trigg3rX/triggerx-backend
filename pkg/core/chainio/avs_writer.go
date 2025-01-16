package chainio

// import (
// 	"context"
// 	"fmt"
// 	"math/big"

// 	"crypto/ecdsa"

// 	gethtypes "github.com/ethereum/go-ethereum/core/types"
// 	gethcommon "github.com/ethereum/go-ethereum/common"

// 	sdkcommon "github.com/trigg3rX/triggerx-backend/pkg/common"
// 	"github.com/Layr-Labs/eigensdk-go/crypto/bls"

// 	"github.com/Layr-Labs/eigensdk-go/chainio/clients/avsregistry"
// 	"github.com/Layr-Labs/eigensdk-go/chainio/txmgr"
// 	logging "github.com/Layr-Labs/eigensdk-go/logging"
// 	"github.com/Layr-Labs/eigensdk-go/types"

// 	registrycoordinator "github.com/Layr-Labs/eigensdk-go/contracts/bindings/RegistryCoordinator"
// 	txservicemanager "github.com/trigg3rX/triggerx-contracts/bindings/contracts/TriggerXServiceManager"
// 	txtaskmanager "github.com/trigg3rX/triggerx-contracts/bindings/contracts/TriggerXTaskManager"
// 	"github.com/trigg3rX/triggerx-backend/pkg/core/config"
// )

// type SignatureWithSaltAndExpiry struct {
// 	Signature []byte
// 	Salt      [32]byte
// 	Expiry    *big.Int
// }

// type AvsWriterer interface {
// 	// TtiggerXTaskManager Functions
// 	CreateNewTask(						// createNewTask, 0x6566ba20
// 		ctx context.Context,
// 		jobId uint32,
// 		quorumNumbers []byte,
// 		quorumThreshold uint8,
// 	) ([8]byte, error)
// 	RespondToTask(						// respondToTask, 0xbdf31991	
// 		ctx context.Context,
// 		task txtaskmanager.ITriggerXTaskManagerTask,
// 		taskResponse txtaskmanager.ITriggerXTaskManagerTaskResponse,
// 		nonSignerStakesAndSignature txtaskmanager.IBLSSignatureCheckerNonSignerStakesAndSignature,
// 	) ([8]byte, error)

// 	// ServiceManager - Keeper Management
// 	RegisterKeeperToTriggerX(			// registerKeeperToTriggerX, 0x763ac957
// 		ctx context.Context,
// 		operator gethcommon.Address,
// 		operatorSignature txservicemanager.ISignatureUtilsSignatureWithSaltAndExpiry,
// 	) (gethcommon.Address, error)
// 	DeregisterKeeperFromTriggerX(		// deregisterKeeperFromTriggerX, 0x6cf5ca21
// 		ctx context.Context, 
// 		operator gethcommon.Address,
// 	) (gethcommon.Address, error)
// 	BlacklistKeeper(					// blacklistKeeper, 0x26a965f0
// 		ctx context.Context, 
// 		operator gethcommon.Address,
// 	) (gethcommon.Address, error)
// 	UnblacklistKeeper(					// unblacklistKeeper, 0xa41d3f94
// 		ctx context.Context, 
// 		operator gethcommon.Address,
// 	) (gethcommon.Address, error)
// 	CreateAVSRewardsSubmission(			// createAVSRewardsSubmission, 0xfce36c7d
// 		ctx context.Context,
// 		rewardsSubmissions []txservicemanager.IRewardsCoordinatorRewardsSubmission,
// 	) (*gethtypes.Receipt, error)
// 	CreateOperatorDirectedAVSRewardsSubmission( // createOperatorDirectedAVSRewardsSubmission, 0xa20b99bf
// 		ctx context.Context,
// 		operatorDirectedRewardsSubmissions []txservicemanager.IRewardsCoordinatorOperatorDirectedRewardsSubmission,
// 	) (*gethtypes.Receipt, error)

// 	// StakeRegistry Functions
// 	Stake(								// stake, 0xa694fc3a
// 		ctx context.Context, 
// 		amount *big.Int,
// 	) (*gethtypes.Receipt, error)
// 	Unstake(							// unstake, 0x2e17de78
// 		ctx context.Context, 
// 		amount *big.Int,
// 	) (*gethtypes.Receipt, error)
// 	RemoveStake(						// removeStake, 0x1238bf4e
// 		ctx context.Context, 
// 		user gethcommon.Address,
// 		amount *big.Int,
// 		reason string,
// 	) (gethcommon.Address, error)

// 	// RegistryCoordinator Functions
// 	RegisterOperator(
// 		ctx context.Context,
// 		operatorEcdsaPrivKey *ecdsa.PrivateKey,
// 		blskeypair *bls.KeyPair,
// 		quorumNumbers types.QuorumNums,
// 		socket string,
// 		waitForReceipt bool,
// 	) (*gethtypes.Receipt, error)
// 	UpdateStakesOfEntireOperatorSetForQuorums(
// 		ctx context.Context,
// 		operatorsPerQuorum [][]gethcommon.Address,
// 		quorumNumbers types.QuorumNums,
// 		waitForReceipt bool,
// 	) (*gethtypes.Receipt, error)
// 	UpdateStakesOfOperatorSubsetForAllQuorums(
// 		ctx context.Context,
// 		operators []gethcommon.Address,
// 		waitForReceipt bool,
// 	) (*gethtypes.Receipt, error)
// 	DeregisterOperator(
// 		ctx context.Context,
// 		quorumNumbers types.QuorumNums,
// 		pubkey registrycoordinator.BN254G1Point,
// 		waitForReceipt bool,
// 	) (*gethtypes.Receipt, error)
// 	EjectOperator(
// 		ctx context.Context,
// 		operator gethcommon.Address,
// 		quorumNumbers []byte,
// 	) (*gethtypes.Receipt, error)
// 	UpdateSocket(
// 		ctx context.Context,
// 		socket types.Socket,
// 		waitForReceipt bool,
// 	) (*gethtypes.Receipt, error)

// 	// Utility Functions
// 	GetTxMgr() txmgr.TxManager
// }

// type AvsWriter struct {
// 	avsregistry.ChainWriter
// 	AvsContractBindings *AvsManagersBindings
// 	logger              logging.Logger
// 	TxMgr               txmgr.TxManager
// }

// var _ AvsWriterer = (*AvsWriter)(nil)

// func BuildAvsWriterFromConfig(c *config.Config) (*AvsWriter, error) {
// 	return BuildAvsWriter(c.TxMgr, c.TriggerXServiceManagerAddr, c.OperatorStateRetrieverAddr, &c.EthHttpClient, c.Logger)
// }

// func BuildAvsWriter(txMgr txmgr.TxManager, registryCoordinatorAddr, operatorStateRetrieverAddr gethcommon.Address, ethHttpClient sdkcommon.EthClientInterface, logger logging.Logger) (*AvsWriter, error) {
// 	avsServiceBindings, err := NewAvsManagersBindings(registryCoordinatorAddr, operatorStateRetrieverAddr, ethHttpClient, logger)
// 	if err != nil {
// 		logger.Error("Failed to create contract bindings", "err", err)
// 		return nil, err
// 	}
// 	avsRegistryWriter, err := avsregistry.BuildAvsRegistryChainWriter(registryCoordinatorAddr, operatorStateRetrieverAddr, logger, ethHttpClient, txMgr)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return NewAvsWriter(*avsRegistryWriter, avsServiceBindings, logger, txMgr), nil
// }
// func NewAvsWriter(avsRegistryWriter avsregistry.ChainWriter, avsServiceBindings *AvsManagersBindings, logger logging.Logger, txMgr txmgr.TxManager) *AvsWriter {
// 	return &AvsWriter{
// 		ChainWriter:         avsRegistryWriter,
// 		AvsContractBindings: avsServiceBindings,
// 		logger:              logger,
// 		TxMgr:               txMgr,
// 	}
// }

// func (w *AvsWriter) GetTxMgr() txmgr.TxManager {
// 	return w.TxMgr
// }

// func (w *AvsWriter) CreateNewTask(ctx context.Context, jobId uint32, quorumNumbers []byte, quorumThreshold uint8) ([8]byte, error) {
// 	w.logger.Info("Creating new task", "jobId", jobId, "quorumThreshold", quorumThreshold)

// 	txOpts, err := w.TxMgr.GetNoSendTxOpts()
// 	if err != nil {
// 		return [8]byte{}, err
// 	}

// 	tx, err := w.AvsContractBindings.TaskManager.CreateNewTask(txOpts, jobId, quorumNumbers, quorumThreshold)
// 	if err != nil {
// 		return [8]byte{}, fmt.Errorf("failed to create new task: %w", err)
// 	}

// 	receipt, err := w.TxMgr.Send(ctx, tx, true)
// 	if err != nil {
// 		w.logger.Errorf("Error submitting CreateNewTask tx")
// 		return [8]byte{}, err
// 	}

// 	taskCreatedEvent, err := w.AvsContractBindings.TaskManager.ParseTaskCreated(*receipt.Logs[0])
// 	if err != nil {
// 		w.logger.Error("Failed to parse new task created event", "err", err)
// 		return [8]byte{}, err
// 	}

// 	return taskCreatedEvent.TaskId, nil
// }

// func (w *AvsWriter) RespondToTask(
// 	ctx context.Context, 
// 	task txtaskmanager.ITriggerXTaskManagerTask, 
// 	taskResponse txtaskmanager.ITriggerXTaskManagerTaskResponse, 
// 	nonSignerStakesAndSignature txtaskmanager.IBLSSignatureCheckerNonSignerStakesAndSignature,
// ) ([8]byte, error) {
// 	w.logger.Info("Responding to task", "taskId", taskResponse.TaskId)

// 	txOpts, err := w.TxMgr.GetNoSendTxOpts()
// 	if err != nil {
// 		w.logger.Errorf("Error getting tx opts")
// 		return [8]byte{}, err
// 	}

// 	tx, err := w.AvsContractBindings.TaskManager.RespondToTask(txOpts, task, taskResponse, nonSignerStakesAndSignature)
// 	if err != nil {
// 		w.logger.Error("Error assembling RespondToTask tx", "err", err)
// 		return [8]byte{}, err
// 	}

// 	receipt, err := w.TxMgr.Send(ctx, tx, true)
// 	if err != nil {
// 		w.logger.Errorf("Error submitting RespondToTask tx")
// 		return [8]byte{}, err
// 	}

// 	taskResponseEvent, err := w.AvsContractBindings.TaskManager.ParseTaskResponded(*receipt.Logs[0])
// 	if err != nil {
// 		w.logger.Error("Failed to parse task responded event", "err", err)
// 		return [8]byte{}, err
// 	}

// 	return taskResponseEvent.TaskId, nil
// }

// func (w *AvsWriter) RegisterKeeperToTriggerX(
// 	ctx context.Context, 
// 	operator gethcommon.Address, 
// 	operatorSignature txservicemanager.ISignatureUtilsSignatureWithSaltAndExpiry,
// ) (gethcommon.Address, error) {
// 	w.logger.Info("Registering keeper to TriggerX", "operator", operator.Hex())

// 	txOpts, err := w.TxMgr.GetNoSendTxOpts()
// 	if err != nil {
// 		return gethcommon.Address{}, err
// 	}

// 	tx, err := w.AvsContractBindings.ServiceManager.RegisterKeeperToTriggerX(txOpts, operator, operatorSignature)
// 	if err != nil {
// 		return gethcommon.Address{}, fmt.Errorf("failed to register keeper: %w", err)
// 	}

// 	receipt, err := w.TxMgr.Send(ctx, tx, true)
// 	if err != nil {
// 		w.logger.Errorf("Error submitting RegisterKeeperToTriggerX tx")
// 		return gethcommon.Address{}, err
// 	}

// 	keeperRegisteredEvent, err := w.AvsContractBindings.ServiceManager.ParseKeeperAdded(*receipt.Logs[0])
// 	if err != nil {
// 		w.logger.Error("Failed to parse keeper registered event", "err", err)
// 		return gethcommon.Address{}, err
// 	}

// 	return keeperRegisteredEvent.Operator, nil
// }

// func (w *AvsWriter) DeregisterKeeperFromTriggerX(ctx context.Context, operator gethcommon.Address) (gethcommon.Address, error) {
// 	w.logger.Info("Deregistering keeper from TriggerX", "operator", operator.Hex())

// 	txOpts, err := w.TxMgr.GetNoSendTxOpts()
// 	if err != nil {
// 		return gethcommon.Address{}, err
// 	}

// 	tx, err := w.AvsContractBindings.ServiceManager.DeregisterKeeperFromTriggerX(txOpts, operator)
// 	if err != nil {
// 		return gethcommon.Address{}, fmt.Errorf("failed to deregister keeper: %w", err)
// 	}

// 	receipt, err := w.TxMgr.Send(ctx, tx, true)
// 	if err != nil {
// 		w.logger.Errorf("Error submitting DeregisterKeeperFromTriggerX tx")
// 		return gethcommon.Address{}, err
// 	}

// 	keeperDeregisteredEvent, err := w.AvsContractBindings.ServiceManager.ParseKeeperRemoved(*receipt.Logs[0])
// 	if err != nil {
// 		w.logger.Error("Failed to parse keeper deregistered event", "err", err)
// 		return gethcommon.Address{}, err
// 	}

// 	return keeperDeregisteredEvent.Operator, nil
// }

// func (w *AvsWriter) BlacklistKeeper(ctx context.Context, operator gethcommon.Address) (gethcommon.Address, error) {
// 	w.logger.Info("Blacklisting keeper", "operator", operator.Hex())

// 	txOpts, err := w.TxMgr.GetNoSendTxOpts()
// 	if err != nil {
// 		return gethcommon.Address{}, err
// 	}

// 	tx, err := w.AvsContractBindings.ServiceManager.BlacklistKeeper(txOpts, operator)
// 	if err != nil {package types

// 		type RegisterKeeperRequest struct {
// 			KeeperAddress     string   `json:"keeper_address"`
// 			Signature         string   `json:"signature"`
// 			Salt              string   `json:"salt"`
// 			Expiry            string   `json:"expiry"`
// 			BlsPublicKey      string   `json:"bls_public_key"`
// 			// TokenStrategyAddr string   `json:"token_strategy_addr"`
// 			// StakeAmount       string   `json:"stake_amount"`
// 		}
		
// 		type RegisterKeeperResponse struct {
// 			KeeperID          int64  `json:"keeper_id"`
// 			RegisteredTx     string `json:"registered_tx"`
// 			PeerID           string `json:"peer_id"`
// 		}
		
// 		type DeregisterKeeperRequest struct {
// 			KeeperID int64 `json:"keeper_id"`
// 		}
		
// 		type DeregisterKeeperResponse struct {
// 			DeregisteredTx string `json:"deregistered_tx"`
// 		}
		
// 		return gethcommon.Address{}, fmt.Errorf("failed to blacklist keeper: %w", err)
// 	}

// 	receipt, err := w.TxMgr.Send(ctx, tx, true)
// 	if err != nil {
// 		w.logger.Errorf("Error submitting BlacklistKeeper tx")
// 		return gethcommon.Address{}, err
// 	}

// 	keeperBlacklistedEvent, err := w.AvsContractBindings.ServiceManager.ParseKeeperBlacklisted(*receipt.Logs[0])
// 	if err != nil {
// 		w.logger.Error("Failed to parse keeper blacklisted event", "err", err)
// 		return gethcommon.Address{}, err
// 	}

// 	return keeperBlacklistedEvent.Operator, nil
// }

// func (w *AvsWriter) UnblacklistKeeper(ctx context.Context, operator gethcommon.Address) (gethcommon.Address, error) {
// 	w.logger.Info("Unblacklisting keeper", "operator", operator.Hex())

// 	txOpts, err := w.TxMgr.GetNoSendTxOpts()
// 	if err != nil {
// 		return gethcommon.Address{}, err
// 	}

// 	tx, err := w.AvsContractBindings.ServiceManager.UnblacklistKeeper(txOpts, operator)
// 	if err != nil {
// 		return gethcommon.Address{}, fmt.Errorf("failed to unblacklist keeper: %w", err)
// 	}

// 	receipt, err := w.TxMgr.Send(ctx, tx, true)
// 	if err != nil {
// 		w.logger.Errorf("Error submitting UnblacklistKeeper tx")
// 		return gethcommon.Address{}, err
// 	}

// 	keeperUnblacklistedEvent, err := w.AvsContractBindings.ServiceManager.ParseKeeperUnblacklisted(*receipt.Logs[0])
// 	if err != nil {
// 		w.logger.Error("Failed to parse keeper unblacklisted event", "err", err)
// 		return gethcommon.Address{}, err
// 	}

// 	return keeperUnblacklistedEvent.Operator, nil
// }

// func (w *AvsWriter) CreateAVSRewardsSubmission(ctx context.Context, rewardsSubmissions []txservicemanager.IRewardsCoordinatorRewardsSubmission) (*gethtypes.Receipt, error) {
// 	w.logger.Info("Creating AVS rewards submission")

// 	txOpts, err := w.TxMgr.GetNoSendTxOpts()
// 	if err != nil {
// 		return nil, err
// 	}

// 	tx, err := w.AvsContractBindings.ServiceManager.CreateAVSRewardsSubmission(txOpts, rewardsSubmissions)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to create rewards submission: %w", err)
// 	}

// 	receipt, err := w.TxMgr.Send(ctx, tx, true)
// 	return receipt, err
// }

// func (w *AvsWriter) CreateOperatorDirectedAVSRewardsSubmission(ctx context.Context, operatorDirectedRewardsSubmissions []txservicemanager.IRewardsCoordinatorOperatorDirectedRewardsSubmission) (*gethtypes.Receipt, error) {
// 	w.logger.Info("Creating operator directed AVS rewards submission")

// 	txOpts, err := w.TxMgr.GetNoSendTxOpts()
// 	if err != nil {
// 		return nil, err
// 	}

// 	tx, err := w.AvsContractBindings.ServiceManager.CreateOperatorDirectedAVSRewardsSubmission(txOpts, operatorDirectedRewardsSubmissions)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to create operator directed rewards submission: %w", err)
// 	}

// 	receipt, err := w.TxMgr.Send(ctx, tx, true)
// 	return receipt, err
// }

// func (w *AvsWriter) Stake(ctx context.Context, amount *big.Int) (*gethtypes.Receipt, error) {
// 	w.logger.Info("Staking tokens", "amount", amount)

// 	txOpts, err := w.TxMgr.GetNoSendTxOpts()
// 	if err != nil {
// 		return nil, err
// 	}

// 	tx, err := w.AvsContractBindings.StakeRegistry.Stake(txOpts, amount)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to stake tokens: %w", err)
// 	}

// 	receipt, err := w.TxMgr.Send(ctx, tx, true)
// 	return receipt, err
// }

// func (w *AvsWriter) Unstake(ctx context.Context, amount *big.Int) (*gethtypes.Receipt, error) {
// 	w.logger.Info("Unstaking tokens", "amount", amount)

// 	txOpts, err := w.TxMgr.GetNoSendTxOpts()
// 	if err != nil {
// 		return nil, err
// 	}

// 	tx, err := w.AvsContractBindings.StakeRegistry.Unstake(txOpts, amount)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to unstake tokens: %w", err)
// 	}

// 	receipt, err := w.TxMgr.Send(ctx, tx, true)
// 	return receipt, err
// }

// func (w *AvsWriter) RemoveStake(ctx context.Context, user gethcommon.Address, amount *big.Int, reason string) (gethcommon.Address, error) {
// 	w.logger.Info("Removing stake", "user", user.Hex(), "amount", amount, "reason", reason)

// 	txOpts, err := w.TxMgr.GetNoSendTxOpts()
// 	if err != nil {
// 		return gethcommon.Address{}, err
// 	}

// 	tx, err := w.AvsContractBindings.StakeRegistry.RemoveStake(txOpts, user, amount, reason)
// 	if err != nil {
// 		return gethcommon.Address{}, fmt.Errorf("failed to remove stake: %w", err)
// 	}

// 	receipt, err := w.TxMgr.Send(ctx, tx, true)
// 	if err != nil {
// 		w.logger.Errorf("Error submitting RemoveStake tx")
// 		return gethcommon.Address{}, err
// 	}

// 	stakeRemovedEvent, err := w.AvsContractBindings.StakeRegistry.ParseStakeRemoved(*receipt.Logs[0])
// 	if err != nil {
// 		w.logger.Error("Failed to parse stake removed event", "err", err)
// 		return gethcommon.Address{}, err
// 	}

// 	return stakeRemovedEvent.User, nil
// }

// func (w *AvsWriter) RegisterOperator(ctx context.Context, operatorEcdsaPrivKey *ecdsa.PrivateKey, blskeypair *bls.KeyPair, quorumNumbers types.QuorumNums, socket string, waitForReceipt bool) (*gethtypes.Receipt, error) {
// 	return w.ChainWriter.RegisterOperator(ctx, operatorEcdsaPrivKey, blskeypair, quorumNumbers, socket, waitForReceipt)
// }

// func (w *AvsWriter) UpdateStakesOfEntireOperatorSetForQuorums(ctx context.Context, operatorsPerQuorum [][]gethcommon.Address, quorumNumbers types.QuorumNums, waitForReceipt bool) (*gethtypes.Receipt, error) {
// 	return w.ChainWriter.UpdateStakesOfEntireOperatorSetForQuorums(ctx, operatorsPerQuorum, quorumNumbers, waitForReceipt)
// }

// func (w *AvsWriter) UpdateStakesOfOperatorSubsetForAllQuorums(ctx context.Context, operators []gethcommon.Address, waitForReceipt bool) (*gethtypes.Receipt, error) {
// 	return w.ChainWriter.UpdateStakesOfOperatorSubsetForAllQuorums(ctx, operators, waitForReceipt)
// }

// func (w *AvsWriter) DeregisterOperator(ctx context.Context, quorumNumbers types.QuorumNums, pubkey registrycoordinator.BN254G1Point, waitForReceipt bool) (*gethtypes.Receipt, error) {
// 	return w.ChainWriter.DeregisterOperator(ctx, quorumNumbers, pubkey, waitForReceipt)
// }

// func (w *AvsWriter) UpdateSocket(ctx context.Context, socket types.Socket, waitForReceipt bool) (*gethtypes.Receipt, error) {
// 	return w.ChainWriter.UpdateSocket(ctx, socket, waitForReceipt)
// }

// func (w *AvsWriter) EjectOperator(ctx context.Context, operator gethcommon.Address, quorumNumbers []byte) (*gethtypes.Receipt, error) {
// 	w.logger.Info("Ejecting operator", "operator", operator.Hex())

// 	txOpts, err := w.TxMgr.GetNoSendTxOpts()
// 	if err != nil {
// 		return nil, err
// 	}

// 	tx, err := w.AvsContractBindings.RegistryCoordinator.EjectOperator(txOpts, operator, quorumNumbers)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to eject operator: %w", err)
// 	}

// 	receipt, err := w.TxMgr.Send(ctx, tx, true)
// 	return receipt, err
// }