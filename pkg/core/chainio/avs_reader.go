package chainio

import (
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"

	sdkavsregistry "github.com/Layr-Labs/eigensdk-go/chainio/clients/avsregistry"
	logging "github.com/Layr-Labs/eigensdk-go/logging"
	"github.com/Layr-Labs/eigensdk-go/types"

	erc20mock "github.com/trigg3rX/triggerx-backend/pkg/avsinterface/bindings/ERC20Mock"
	txregistrycoordinator "github.com/trigg3rX/triggerx-backend/pkg/avsinterface/bindings/RegistryCoordinator"
	txtaskmanager "github.com/trigg3rX/triggerx-backend/pkg/avsinterface/bindings/TriggerXTaskManager"
	opstateretriever "github.com/Layr-Labs/eigensdk-go/contracts/bindings/OperatorStateRetriever"
	sdkcommon "github.com/trigg3rX/triggerx-backend/pkg/common"
	"github.com/trigg3rX/triggerx-backend/pkg/core/config"
)

type AvsReaderer interface {
	// TriggerXTaskManager methods
	GetTaskHash( // taskHashes, 0x304feba2
		opts *bind.CallOpts,
		taskId [8]byte,
	) ([32]byte, error)
	GetTaskResponseHash( // taskResponseHashes, 0xd82c7b5c
		opts *bind.CallOpts,
		taskId [8]byte,
	) ([32]byte, error)
	GetJobToTaskCounter( // jobToTaskCounter, 0x9f2d70df
		opts *bind.CallOpts,
		jobId uint32,
	) (uint32, error)
	GetTaskResponseWindowBlock( // TASK_RESPONSE_WINDOW_BLOCK, 0x1ad43189
		opts *bind.CallOpts,
	) (uint32, error)
	GenerateTaskId( // generateTaskId, 0x8e91269d
		opts *bind.CallOpts,
		jobId uint32,
		taskNum uint32,
	) ([8]byte, error)
	CheckSignatures( // checkSignatures, 0x6efb4636
		opts *bind.CallOpts,
		msgHash [32]byte,
		quorumNumbers []byte,
		referenceBlockNumber uint32,
		nonSignerStakesAndSignature txtaskmanager.IBLSSignatureCheckerNonSignerStakesAndSignature,
	) (txtaskmanager.IBLSSignatureCheckerQuorumStakeTotals, error)
	TrySignatureAndApkVerification( // trySignatureAndApkVerification, 0x171f1d5b
		opts *bind.CallOpts,
		msgHash [32]byte,
		apk txtaskmanager.BN254G1Point,
		apkG2 txtaskmanager.BN254G2Point,
		sigma txtaskmanager.BN254G1Point,
	) (bool, bool, error)

	// TriggerXServiceManager methods
	GetOperatorRestakedStrategies( // getOperatorRestakedStrategies, 0x33cfb7b7
		opts *bind.CallOpts,
		operator common.Address,
	) ([]common.Address, error)
	GetRestakeableStrategies( // getRestakeableStrategies, 0xe481af9d
		opts *bind.CallOpts,
	) ([]common.Address, error)
	IsBlackListed( // isBlackListed, 0xe47d6060
		opts *bind.CallOpts,
		operator common.Address,
	) (bool, error)
	GetQuorumManager( // quorumManager, 0xb5f7eb6b
		opts *bind.CallOpts,
	) (common.Address, error)
	GetTaskManager( // taskManager, 0xa50a640e
		opts *bind.CallOpts,
	) (common.Address, error)
	GetTaskValidator( // taskValidator, 0xfd38ec8c
		opts *bind.CallOpts,
	) (common.Address, error)

	// Misc
	GetErc20Mock(
		opts *bind.CallOpts,
		tokenAddr common.Address,
	) (*erc20mock.ContractERC20Mock, error)

	// TriggerXStakeRegistry read methods
	GetStake( // getStake, 0x7a766460
		opts *bind.CallOpts,
		user common.Address,
	) (struct {
		Amount *big.Int
		Exists bool
	}, error)

	// RegistryCoordinator read methods
	GetOperator(
		opts *bind.CallOpts,
		operator common.Address,
	) (txregistrycoordinator.IRegistryCoordinatorOperatorInfo, error)
	GetOperatorStatus(
		opts *bind.CallOpts,
		operator common.Address,
	) (uint8, error)
	GetQuorumCount(
		opts *bind.CallOpts,
	) (uint8, error)
	GetOperatorsStakeInQuorumsAtCurrentBlock(
		opts *bind.CallOpts,
		quorumNumbers types.QuorumNums,
	) ([][]opstateretriever.OperatorStateRetrieverOperator, error)
	GetOperatorsStakeInQuorumsAtBlock(
		opts *bind.CallOpts,
		quorumNumbers types.QuorumNums,
		blockNumber uint32,
	) ([][]opstateretriever.OperatorStateRetrieverOperator, error)
	GetOperatorAddrsInQuorumsAtCurrentBlock(
		opts *bind.CallOpts,
		quorumNumbers types.QuorumNums,
	) ([][]common.Address, error)
	GetOperatorsStakeInQuorumsOfOperatorAtBlock(
		opts *bind.CallOpts,
		operatorId types.OperatorId,
		blockNumber uint32,
	) (types.QuorumNums, [][]opstateretriever.OperatorStateRetrieverOperator, error)
	GetOperatorsStakeInQuorumsOfOperatorAtCurrentBlock(
		opts *bind.CallOpts,
		operatorId types.OperatorId,
	) (types.QuorumNums, [][]opstateretriever.OperatorStateRetrieverOperator, error)
	GetOperatorStakeInQuorumsOfOperatorAtCurrentBlock(
		opts *bind.CallOpts,
		operatorId types.OperatorId,
	) (map[types.QuorumNum]types.StakeAmount, error)
	GetCheckSignaturesIndices(
		opts *bind.CallOpts,
		referenceBlockNumber uint32,
		quorumNumbers types.QuorumNums,
		nonSignerOperatorIds []types.OperatorId,
	) (opstateretriever.OperatorStateRetrieverCheckSignaturesIndices, error)
	GetOperatorId(
		opts *bind.CallOpts,
		operator common.Address,
	) ([32]byte, error)
	GetOperatorFromId(
		opts *bind.CallOpts,
		operatorId [32]byte,
	) (common.Address, error)
	QueryRegistrationDetail(
		opts *bind.CallOpts,
		operator common.Address,
	) ([]bool, error)

	IsOperatorRegistered(
		opts *bind.CallOpts,
		operator common.Address,
	) (bool, error)
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
	opts *bind.CallOpts, msgHash [32]byte, quorumNumbers []byte, referenceBlockNumber uint32, nonSignerStakesAndSignature txtaskmanager.IBLSSignatureCheckerNonSignerStakesAndSignature,
) (txtaskmanager.IBLSSignatureCheckerQuorumStakeTotals, error) {
	stakeTotalsPerQuorum, _, err := r.AvsServiceBindings.TaskManager.CheckSignatures(
		opts, msgHash, quorumNumbers, referenceBlockNumber, nonSignerStakesAndSignature,
	)
	if err != nil {
		return txtaskmanager.IBLSSignatureCheckerQuorumStakeTotals{}, err
	}
	return stakeTotalsPerQuorum, nil
}

func (r *AvsReader) GetErc20Mock(opts *bind.CallOpts, tokenAddr common.Address) (*erc20mock.ContractERC20Mock, error) {
	erc20Mock, err := r.AvsServiceBindings.GetErc20Mock(tokenAddr)
	if err != nil {
		r.logger.Error("Failed to fetch ERC20Mock contract", "err", err)
		return nil, err
	}
	return erc20Mock, nil
}

func (r *AvsReader) GetTaskHash(opts *bind.CallOpts, taskId [8]byte) ([32]byte, error) {
	return r.AvsServiceBindings.TaskManager.TaskHashes(opts, taskId)
}

func (r *AvsReader) GetTaskResponseHash(opts *bind.CallOpts, taskId [8]byte) ([32]byte, error) {
	return r.AvsServiceBindings.TaskManager.TaskResponseHashes(opts, taskId)
}

func (r *AvsReader) GetJobToTaskCounter(opts *bind.CallOpts, jobId uint32) (uint32, error) {
	return r.AvsServiceBindings.TaskManager.JobToTaskCounter(opts, jobId)
}

func (r *AvsReader) GenerateTaskId(opts *bind.CallOpts, jobId uint32, taskNum uint32) ([8]byte, error) {
	return r.AvsServiceBindings.TaskManager.GenerateTaskId(opts, jobId, taskNum)
}

func (r *AvsReader) IsOperatorBlacklisted(opts *bind.CallOpts, operator common.Address) (bool, error) {
	return r.AvsServiceBindings.ServiceManager.IsBlackListed(opts, operator)
}

func (r *AvsReader) GetTaskManager(opts *bind.CallOpts) (common.Address, error) {
	return r.AvsServiceBindings.ServiceManager.TaskManager(opts)
}

func (r *AvsReader) GetTaskValidator(opts *bind.CallOpts) (common.Address, error) {
	return r.AvsServiceBindings.ServiceManager.TaskValidator(opts)
}

func (r *AvsReader) GetQuorumManager(opts *bind.CallOpts) (common.Address, error) {
	return r.AvsServiceBindings.ServiceManager.QuorumManager(opts)
}

func (r *AvsReader) GetTaskResponseWindowBlock(opts *bind.CallOpts) (uint32, error) {
	return r.AvsServiceBindings.TaskManager.TASKRESPONSEWINDOWBLOCK(opts)
}

func (r *AvsReader) TrySignatureAndApkVerification(
	opts *bind.CallOpts,
	msgHash [32]byte,
	apk txtaskmanager.BN254G1Point,
	apkG2 txtaskmanager.BN254G2Point,
	sigma txtaskmanager.BN254G1Point,
) (bool, bool, error) {
	result, err := r.AvsServiceBindings.TaskManager.TrySignatureAndApkVerification(
		opts,
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

func (r *AvsReader) GetOperatorRestakedStrategies(opts *bind.CallOpts, operator common.Address) ([]common.Address, error) {
	return r.AvsServiceBindings.ServiceManager.GetOperatorRestakedStrategies(opts, operator)
}

func (r *AvsReader) GetRestakeableStrategies(opts *bind.CallOpts) ([]common.Address, error) {
	return r.AvsServiceBindings.ServiceManager.GetRestakeableStrategies(opts)
}

func (r *AvsReader) IsBlackListed(opts *bind.CallOpts, operator common.Address) (bool, error) {
	return r.AvsServiceBindings.ServiceManager.IsBlackListed(opts, operator)
}

func (r *AvsReader) GetStake(opts *bind.CallOpts, user common.Address) (struct {
	Amount *big.Int
	Exists bool
}, error) {
	return r.AvsServiceBindings.StakeRegistry.GetStake(opts, user)
}

// RegistryCoordinator
func (r *AvsReader) GetOperator(opts *bind.CallOpts, operator common.Address) (txregistrycoordinator.IRegistryCoordinatorOperatorInfo, error) {
	return r.AvsServiceBindings.RegistryCoordinator.GetOperator(opts, operator)
}

func (r *AvsReader) GetOperatorStatus(opts *bind.CallOpts, operator common.Address) (uint8, error) {
	return r.AvsServiceBindings.RegistryCoordinator.GetOperatorStatus(opts, operator)
}

func (r *AvsReader) GetQuorumCount(opts *bind.CallOpts) (uint8, error) {
	return r.ChainReader.GetQuorumCount(opts)
}

func (r *AvsReader) GetOperatorsStakeInQuorumsAtCurrentBlock(opts *bind.CallOpts, quorumNumbers types.QuorumNums) ([][]opstateretriever.OperatorStateRetrieverOperator, error) {
	return r.ChainReader.GetOperatorsStakeInQuorumsAtCurrentBlock(opts, quorumNumbers)
}

func (r *AvsReader) GetOperatorsStakeInQuorumsAtBlock(opts *bind.CallOpts, quorumNumbers types.QuorumNums, blockNumber uint32) ([][]opstateretriever.OperatorStateRetrieverOperator, error) {
	return r.ChainReader.GetOperatorsStakeInQuorumsAtBlock(opts, quorumNumbers, blockNumber)
}

func (r *AvsReader) GetOperatorAddrsInQuorumsAtCurrentBlock(opts *bind.CallOpts, quorumNumbers types.QuorumNums) ([][]common.Address, error) {
	return r.ChainReader.GetOperatorAddrsInQuorumsAtCurrentBlock(opts, quorumNumbers)
}

func (r *AvsReader) GetOperatorsStakeInQuorumsOfOperatorAtBlock(opts *bind.CallOpts, operatorId types.OperatorId, blockNumber uint32) (types.QuorumNums, [][]opstateretriever.OperatorStateRetrieverOperator, error) {
	return r.ChainReader.GetOperatorsStakeInQuorumsOfOperatorAtBlock(opts, operatorId, blockNumber)
}

func (r *AvsReader) GetOperatorsStakeInQuorumsOfOperatorAtCurrentBlock(opts *bind.CallOpts, operatorId types.OperatorId) (types.QuorumNums, [][]opstateretriever.OperatorStateRetrieverOperator, error) {
	return r.ChainReader.GetOperatorsStakeInQuorumsOfOperatorAtCurrentBlock(opts, operatorId)
}

func (r *AvsReader) GetOperatorStakeInQuorumsOfOperatorAtCurrentBlock(opts *bind.CallOpts, operatorId types.OperatorId) (map[types.QuorumNum]types.StakeAmount, error) {
	return r.ChainReader.GetOperatorStakeInQuorumsOfOperatorAtCurrentBlock(opts, operatorId)
}

func (r *AvsReader) GetCheckSignaturesIndices(opts *bind.CallOpts, referenceBlockNumber uint32, quorumNumbers types.QuorumNums, nonSignerOperatorIds []types.OperatorId) (opstateretriever.OperatorStateRetrieverCheckSignaturesIndices, error) {
	return r.ChainReader.GetCheckSignaturesIndices(opts, referenceBlockNumber, quorumNumbers, nonSignerOperatorIds)
}

func (r *AvsReader) GetOperatorId(opts *bind.CallOpts, operator common.Address) ([32]byte, error) {
	return r.ChainReader.GetOperatorId(opts, operator)
}

func (r *AvsReader) GetOperatorFromId(opts *bind.CallOpts, operatorId [32]byte) (common.Address, error) {
	return r.ChainReader.GetOperatorFromId(opts, operatorId)
}

func (r *AvsReader) QueryRegistrationDetails(opts *bind.CallOpts, operator common.Address) ([]bool, error) {
	return r.ChainReader.QueryRegistrationDetail(opts, operator)
}

func (r *AvsReader) IsOperatorRegistered(opts *bind.CallOpts, operator common.Address) (bool, error) {
	return r.ChainReader.IsOperatorRegistered(opts, operator)
}
