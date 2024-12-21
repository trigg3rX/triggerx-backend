package chainio

import (
	"context"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"

	sdkavsregistry "github.com/Layr-Labs/eigensdk-go/chainio/clients/avsregistry"
	logging "github.com/Layr-Labs/eigensdk-go/logging"

	erc20mock "github.com/Layr-Labs/incredible-squaring-avs/contracts/bindings/ERC20Mock"
	txtaskmanager "github.com/trigg3rX/go-backend/pkg/avsinterface/bindings/TriggerXTaskManager"
	sdkcommon "github.com/trigg3rX/go-backend/pkg/common"
	"github.com/trigg3rX/go-backend/pkg/core/config"
)

type AvsReaderer interface {
	// TriggerXTaskManager methods
	GetTaskHash(ctx context.Context, taskId [8]byte) ([32]byte, error)
	GetTaskResponseHash(ctx context.Context, taskId [8]byte) ([32]byte, error)
	GetJobToTaskCounter(ctx context.Context, jobId uint32) (uint32, error)
	GenerateTaskId(jobId uint32, taskNum uint32) ([8]byte, error)

	// TriggerXServiceManager methods
	IsOperatorBlacklisted(ctx context.Context, operator common.Address) (bool, error)
	GetTaskManager(ctx context.Context) (common.Address, error)
	GetTaskValidator(ctx context.Context) (common.Address, error)
	GetQuorumManager(ctx context.Context) (common.Address, error)

	// Existing methods
	CheckSignatures(
		ctx context.Context, msgHash [32]byte, quorumNumbers []byte, referenceBlockNumber uint32, nonSignerStakesAndSignature txtaskmanager.IBLSSignatureCheckerNonSignerStakesAndSignature,
	) (txtaskmanager.IBLSSignatureCheckerQuorumStakeTotals, error)
	GetErc20Mock(ctx context.Context, tokenAddr common.Address) (*erc20mock.ContractERC20Mock, error)
	GetOperatorId(opts *bind.CallOpts, operatorAddress common.Address) ([32]byte, error)
	IsOperatorRegistered(opts *bind.CallOpts, operatorAddress common.Address) (bool, error)
}

type AvsReader struct {
	sdkavsregistry.ChainReader
	AvsServiceBindings *AvsManagersBindings
	logger             logging.Logger
}

//var _ AvsReaderer = (*AvsReader)(nil)

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

func (r *AvsReader) GenerateTaskId(jobId uint32, taskNum uint32) ([8]byte, error) {
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
