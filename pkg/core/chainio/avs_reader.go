package chainio

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"

	sdkavsregistry "github.com/Layr-Labs/eigensdk-go/chainio/clients/avsregistry"
	logging "github.com/Layr-Labs/eigensdk-go/logging"

	erc20mock "github.com/trigg3rX/triggerx-backend/pkg/avsinterface/bindings/ERC20Mock"
	txregistrycoordinator "github.com/trigg3rX/triggerx-backend/pkg/avsinterface/bindings/RegistryCoordinator"
	txtaskmanager "github.com/trigg3rX/triggerx-backend/pkg/avsinterface/bindings/TriggerXTaskManager"
	sdkcommon "github.com/trigg3rX/triggerx-backend/pkg/common"
	"github.com/trigg3rX/triggerx-backend/pkg/core/config"
)

type AvsReaderer interface {
	// TriggerXTaskManager methods
	GetTaskHash(						// taskHashes, 0x304feba2
		ctx context.Context,
		taskId [8]byte,
	) ([32]byte, error)
	GetTaskResponseHash(				// taskResponseHashes, 0xd82c7b5c	
		ctx context.Context,
		taskId [8]byte,
	) ([32]byte, error)
	GetJobToTaskCounter(				// jobToTaskCounter, 0x9f2d70df
		ctx context.Context,
		jobId uint32,
	) (uint32, error)
	GetTaskResponseWindowBlock(			// TASK_RESPONSE_WINDOW_BLOCK, 0x1ad43189
		ctx context.Context,
	) (uint32, error)
	GenerateTaskId(						// generateTaskId, 0x8e91269d
		ctx context.Context,
		jobId uint32,
		taskNum uint32,
	) ([8]byte, error)
	CheckSignatures(					// checkSignatures, 0x6efb4636
		ctx context.Context,
		msgHash [32]byte,
		quorumNumbers []byte,
		referenceBlockNumber uint32,
		nonSignerStakesAndSignature txtaskmanager.IBLSSignatureCheckerNonSignerStakesAndSignature,
	) (txtaskmanager.IBLSSignatureCheckerQuorumStakeTotals, error)
	TrySignatureAndApkVerification(		// trySignatureAndApkVerification, 0x171f1d5b
		ctx context.Context,
		msgHash [32]byte,
		apk txtaskmanager.BN254G1Point,
		apkG2 txtaskmanager.BN254G2Point,
		sigma txtaskmanager.BN254G1Point,
	) (bool, bool, error)

	// TriggerXServiceManager methods
	GetOperatorRestakedStrategies(		// getOperatorRestakedStrategies, 0x33cfb7b7
		ctx context.Context,
		operator common.Address,
	) ([]common.Address, error)
	GetRestakeableStrategies(			// getRestakeableStrategies, 0xe481af9d
		ctx context.Context,
	) ([]common.Address, error)
	IsBlackListed(						// isBlackListed, 0xe47d6060
		ctx context.Context,
		operator common.Address,
	) (bool, error)
	GetQuorumManager(					// quorumManager, 0xb5f7eb6b
		ctx context.Context,
	) (common.Address, error)
	GetTaskManager(						// taskManager, 0xa50a640e
		ctx context.Context,
	) (common.Address, error)
	GetTaskValidator(					// taskValidator, 0xfd38ec8c
		ctx context.Context,
	) (common.Address, error)

	// Misc
	GetErc20Mock(
		ctx context.Context,
		tokenAddr common.Address,
	) (*erc20mock.ContractERC20Mock, error)

	// TriggerXStakeRegistry read methods
	GetStake(							// getStake, 0x7a766460
		ctx context.Context, 
		user common.Address,
	) (struct {
		Amount *big.Int
		Exists bool
	}, error)

	// RegistryCoordinator read methods
	GetOperatorChurnApprovalDigestHash(
		ctx context.Context,
		registeringOperator common.Address,
		registeringOperatorId [32]byte,
		operatorKickParams []txregistrycoordinator.IRegistryCoordinatorOperatorKickParam,
		salt [32]byte,
		expiry *big.Int,
	) ([32]byte, error)
	GetCurrentQuorumBitmap(
		ctx context.Context,
		operatorId [32]byte,
	) (*big.Int, error)
	GetOperator(
		ctx context.Context,
		operator common.Address,
	) (txregistrycoordinator.IRegistryCoordinatorOperatorInfo, error)
	GetOperatorFromId(
		ctx context.Context,
		operatorId [32]byte,
	) (common.Address, error)
	GetOperatorId(
		ctx context.Context,
		operator common.Address,
	) ([32]byte, error)
	GetOperatorSetParams(
		ctx context.Context,
		quorumNumber uint8,
	) (txregistrycoordinator.IRegistryCoordinatorOperatorSetParam, error)
	GetOperatorStatus(
		ctx context.Context, 
		operator common.Address,
	) (uint8, error)
	GetQuorumBitmapAtBlockNumberByIndex(
		ctx context.Context,
		operatorId [32]byte,
		blockNumber uint32,
		index *big.Int,
	) (*big.Int, error)
	GetQuorumBitmapHistoryLength(
		ctx context.Context,
		operatorId [32]byte,
	) (*big.Int, error)
	GetQuorumBitmapIndicesAtBlockNumber(
		ctx context.Context,
		blockNumber uint32,
		operatorIds [][32]byte,
	) ([]uint32, error)
	GetQuorumBitmapUpdateByIndex(
		ctx context.Context,
		operatorId [32]byte,
		index *big.Int,
	) (txregistrycoordinator.IRegistryCoordinatorQuorumBitmapUpdate, error)
	IsChurnApproverSaltUsed(
		ctx context.Context,
		salt [32]byte,
	) (bool, error)
	GetLastEjectionTimestamp(
		ctx context.Context,
		operator common.Address,
	) (*big.Int, error)
	GetPubkeyRegistrationMessageHash(
		ctx context.Context,
		operator common.Address,
	) (txregistrycoordinator.BN254G1Point, error)
	GetQuorumCount(
		ctx context.Context,
	) (uint8, error)
	GetQuorumUpdateBlockNumber(
		ctx context.Context,
		quorumNumber uint8,
	) (*big.Int, error)
}

type AvsReader struct {
	sdkavsregistry.ChainReader
	AvsServiceBindings *AvsManagersBindings
	logger             logging.Logger
}

var _ AvsReaderer = (*AvsReader)(nil)

func BuildAvsReaderFromConfig(c *config.Config) (*AvsReader, error) {
	return BuildAvsReader(c.TriggerXServiceManagerAddr, c.OperatorStateRetrieverAddr, &c.EthHttpClient, c.Logger)
}

func BuildAvsReader(serviceManagerAddr, operatorStateRetrieverAddr common.Address, ethHttpClient sdkcommon.EthClientInterface, logger logging.Logger) (*AvsReader, error) {
	avsManagersBindings, err := NewAvsManagersBindings(serviceManagerAddr, operatorStateRetrieverAddr, ethHttpClient, logger)
	if err != nil {
		return nil, err
	}
	avsRegistryReader, err := sdkavsregistry.BuildAvsRegistryChainReader(serviceManagerAddr, operatorStateRetrieverAddr, ethHttpClient, logger)
	if err != nil {
		return nil, err
	}
	return NewAvsReader(*avsRegistryReader, avsManagersBindings, logger)
}
func NewAvsReader(avsRegistryReader sdkavsregistry.ChainReader, avsServiceBindings *AvsManagersBindings, logger logging.Logger) (*AvsReader, error) {
	return &AvsReader{
		ChainReader:        avsRegistryReader,
		AvsServiceBindings: avsServiceBindings,
		logger:             logger,
	}, nil
}

func (r *AvsReader) CheckSignatures(
	ctx context.Context, msgHash [32]byte, quorumNumbers []byte, referenceBlockNumber uint32, nonSignerStakesAndSignature txtaskmanager.IBLSSignatureCheckerNonSignerStakesAndSignature,
) (txtaskmanager.IBLSSignatureCheckerQuorumStakeTotals, error) {
	stakeTotalsPerQuorum, _, err := r.AvsServiceBindings.TaskManager.CheckSignatures(
		&bind.CallOpts{}, msgHash, quorumNumbers, referenceBlockNumber, nonSignerStakesAndSignature,
	)
	if err != nil {
		return txtaskmanager.IBLSSignatureCheckerQuorumStakeTotals{}, err
	}
	return stakeTotalsPerQuorum, nil
}

func (r *AvsReader) GetErc20Mock(ctx context.Context, tokenAddr common.Address) (*erc20mock.ContractERC20Mock, error) {
	erc20Mock, err := r.AvsServiceBindings.GetErc20Mock(tokenAddr)
	if err != nil {
		r.logger.Error("Failed to fetch ERC20Mock contract", "err", err)
		return nil, err
	}
	return erc20Mock, nil
}

func (r *AvsReader) GetTaskHash(ctx context.Context, taskId [8]byte) ([32]byte, error) {
	return r.AvsServiceBindings.TaskManager.TaskHashes(&bind.CallOpts{}, taskId)
}

func (r *AvsReader) GetTaskResponseHash(ctx context.Context, taskId [8]byte) ([32]byte, error) {
	return r.AvsServiceBindings.TaskManager.TaskResponseHashes(&bind.CallOpts{}, taskId)
}

func (r *AvsReader) GetJobToTaskCounter(ctx context.Context, jobId uint32) (uint32, error) {
	return r.AvsServiceBindings.TaskManager.JobToTaskCounter(&bind.CallOpts{}, jobId)
}

func (r *AvsReader) GenerateTaskId(ctx context.Context, jobId uint32, taskNum uint32) ([8]byte, error) {
	return r.AvsServiceBindings.TaskManager.GenerateTaskId(&bind.CallOpts{}, jobId, taskNum)
}

func (r *AvsReader) IsOperatorBlacklisted(ctx context.Context, operator common.Address) (bool, error) {
	return r.AvsServiceBindings.ServiceManager.IsBlackListed(&bind.CallOpts{}, operator)
}

func (r *AvsReader) GetTaskManager(ctx context.Context) (common.Address, error) {
	return r.AvsServiceBindings.ServiceManager.TaskManager(&bind.CallOpts{})
}

func (r *AvsReader) GetTaskValidator(ctx context.Context) (common.Address, error) {
	return r.AvsServiceBindings.ServiceManager.TaskValidator(&bind.CallOpts{})
}

func (r *AvsReader) GetQuorumManager(ctx context.Context) (common.Address, error) {
	return r.AvsServiceBindings.ServiceManager.QuorumManager(&bind.CallOpts{})
}

func (r *AvsReader) GetTaskResponseWindowBlock(ctx context.Context) (uint32, error) {
	return r.AvsServiceBindings.TaskManager.TASKRESPONSEWINDOWBLOCK(&bind.CallOpts{})
}

func (r *AvsReader) TrySignatureAndApkVerification(
	ctx context.Context,
	msgHash [32]byte,
	apk txtaskmanager.BN254G1Point,
	apkG2 txtaskmanager.BN254G2Point,
	sigma txtaskmanager.BN254G1Point,
) (bool, bool, error) {
	result, err := r.AvsServiceBindings.TaskManager.TrySignatureAndApkVerification(
		&bind.CallOpts{},
		msgHash,
		apk,
		apkG2,
		sigma,
	)
	if err != nil {
		return false, false, err
	}
	return result.PairingSuccessful, result.SiganatureIsValid, nil
}

func (r *AvsReader) GetOperatorRestakedStrategies(ctx context.Context, operator common.Address) ([]common.Address, error) {
	return r.AvsServiceBindings.ServiceManager.GetOperatorRestakedStrategies(&bind.CallOpts{}, operator)
}

func (r *AvsReader) GetRestakeableStrategies(ctx context.Context) ([]common.Address, error) {
	return r.AvsServiceBindings.ServiceManager.GetRestakeableStrategies(&bind.CallOpts{})
}

func (r *AvsReader) IsBlackListed(ctx context.Context, operator common.Address) (bool, error) {
	return r.AvsServiceBindings.ServiceManager.IsBlackListed(&bind.CallOpts{}, operator)
}

func (r *AvsReader) GetStake(ctx context.Context, user common.Address) (struct {
	Amount *big.Int
	Exists bool
}, error) {
	return r.AvsServiceBindings.StakeRegistry.GetStake(&bind.CallOpts{}, user)
}

func (r *AvsReader) GetOperatorChurnApprovalDigestHash(
	ctx context.Context,
	registeringOperator common.Address,
	registeringOperatorId [32]byte,
	operatorKickParams []txregistrycoordinator.IRegistryCoordinatorOperatorKickParam,
	salt [32]byte,
	expiry *big.Int,
) ([32]byte, error) {
	return r.AvsServiceBindings.RegistryCoordinator.CalculateOperatorChurnApprovalDigestHash(
		&bind.CallOpts{},
		registeringOperator,
		registeringOperatorId,
		operatorKickParams,
		salt,
		expiry,
	)
}

func (r *AvsReader) GetCurrentQuorumBitmap(ctx context.Context, operatorId [32]byte) (*big.Int, error) {
	return r.AvsServiceBindings.RegistryCoordinator.GetCurrentQuorumBitmap(&bind.CallOpts{}, operatorId)
}

func (r *AvsReader) GetOperator(ctx context.Context, operator common.Address) (txregistrycoordinator.IRegistryCoordinatorOperatorInfo, error) {
	return r.AvsServiceBindings.RegistryCoordinator.GetOperator(&bind.CallOpts{}, operator)
}

func (r *AvsReader) GetOperatorFromId(ctx context.Context, operatorId [32]byte) (common.Address, error) {
	return r.AvsServiceBindings.RegistryCoordinator.GetOperatorFromId(&bind.CallOpts{}, operatorId)
}

func (r *AvsReader) GetOperatorId(ctx context.Context, operator common.Address) ([32]byte, error) {
	return r.AvsServiceBindings.RegistryCoordinator.GetOperatorId(&bind.CallOpts{}, operator)
}

func (r *AvsReader) GetOperatorSetParams(ctx context.Context, quorumNumber uint8) (txregistrycoordinator.IRegistryCoordinatorOperatorSetParam, error) {
	return r.AvsServiceBindings.RegistryCoordinator.GetOperatorSetParams(&bind.CallOpts{}, quorumNumber)
}

func (r *AvsReader) GetOperatorStatus(ctx context.Context, operator common.Address) (uint8, error) {
	return r.AvsServiceBindings.RegistryCoordinator.GetOperatorStatus(&bind.CallOpts{}, operator)
}

func (r *AvsReader) GetQuorumBitmapAtBlockNumberByIndex(ctx context.Context, operatorId [32]byte, blockNumber uint32, index *big.Int) (*big.Int, error) {
	return r.AvsServiceBindings.RegistryCoordinator.GetQuorumBitmapAtBlockNumberByIndex(&bind.CallOpts{}, operatorId, blockNumber, index)
}

func (r *AvsReader) GetQuorumBitmapHistoryLength(ctx context.Context, operatorId [32]byte) (*big.Int, error) {
	return r.AvsServiceBindings.RegistryCoordinator.GetQuorumBitmapHistoryLength(&bind.CallOpts{}, operatorId)
}

func (r *AvsReader) GetQuorumBitmapIndicesAtBlockNumber(ctx context.Context, blockNumber uint32, operatorIds [][32]byte) ([]uint32, error) {
	return r.AvsServiceBindings.RegistryCoordinator.GetQuorumBitmapIndicesAtBlockNumber(&bind.CallOpts{}, blockNumber, operatorIds)
}

func (r *AvsReader) GetQuorumBitmapUpdateByIndex(ctx context.Context, operatorId [32]byte, index *big.Int) (txregistrycoordinator.IRegistryCoordinatorQuorumBitmapUpdate, error) {
	return r.AvsServiceBindings.RegistryCoordinator.GetQuorumBitmapUpdateByIndex(&bind.CallOpts{}, operatorId, index)
}

func (r *AvsReader) IsChurnApproverSaltUsed(ctx context.Context, salt [32]byte) (bool, error) {
	return r.AvsServiceBindings.RegistryCoordinator.IsChurnApproverSaltUsed(&bind.CallOpts{}, salt)
}

func (r *AvsReader) GetLastEjectionTimestamp(ctx context.Context, operator common.Address) (*big.Int, error) {
	return r.AvsServiceBindings.RegistryCoordinator.LastEjectionTimestamp(&bind.CallOpts{}, operator)
}

func (r *AvsReader) GetPubkeyRegistrationMessageHash(ctx context.Context, operator common.Address) (txregistrycoordinator.BN254G1Point, error) {
	return r.AvsServiceBindings.RegistryCoordinator.PubkeyRegistrationMessageHash(&bind.CallOpts{}, operator)
}

func (r *AvsReader) GetQuorumCount(ctx context.Context) (uint8, error) {
	return r.AvsServiceBindings.RegistryCoordinator.QuorumCount(&bind.CallOpts{})
}

func (r *AvsReader) GetQuorumUpdateBlockNumber(ctx context.Context, quorumNumber uint8) (*big.Int, error) {
	return r.AvsServiceBindings.RegistryCoordinator.QuorumUpdateBlockNumber(&bind.CallOpts{}, quorumNumber)
}
