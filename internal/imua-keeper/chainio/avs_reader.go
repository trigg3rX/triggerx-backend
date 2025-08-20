package chainio

import (
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	gethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	avs "github.com/trigg3rX/imua-contracts/bindings/contracts"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
)

type AvsReader interface {
	GetTaskInfo(
		opts *bind.CallOpts,
		avsAddress string,
		taskID uint64,
	) (avs.TaskInfo, error)
	IsOperator(
		opts *bind.CallOpts,
		operator string,
	) (bool, error)

	GetAVSEpochIdentifier(
		opts *bind.CallOpts,
		avsAddress string,
	) (string, error)
	GetCurrentEpoch(
		opts *bind.CallOpts,
		epochIdentifier string,
	) (int64, error)
	GetChallengeInfo(
		opts *bind.CallOpts,
		taskAddress string,
		taskID uint64,
	) (gethcommon.Address, error)
	GetOperatorTaskResponseList(
		opts *bind.CallOpts,
		taskAddress string,
		taskID uint64,
	) ([]avs.OperatorResInfo, error)
}

type ChainReader struct {
	logger     logging.Logger
	avsManager avs.TriggerXAvs
	ethClient  *ethclient.Client
}

// forces EthReader to implement the chainio.Reader interface
var _ AvsReader = (*ChainReader)(nil)

func NewChainReader(
	avsManager avs.TriggerXAvs,
	logger logging.Logger,
	ethClient *ethclient.Client,
) *ChainReader {
	return &ChainReader{
		avsManager: avsManager,
		logger:     logger,
		ethClient:  ethClient,
	}
}

func BuildChainReader(
	avsAddr gethcommon.Address,
	ethClient *ethclient.Client,
	logger logging.Logger,
) (*ChainReader, error) {
	contractBindings, err := NewContractBindings(
		avsAddr,
		ethClient,
		logger,
	)
	if err != nil {
		return nil, err
	}
	return NewChainReader(
		*contractBindings.AVSManager,
		logger,
		ethClient,
	), nil
}

func (r *ChainReader) GetTaskInfo(opts *bind.CallOpts, avsAddress string, taskID uint64) (avs.TaskInfo, error) {
	info, err := r.avsManager.GetTaskInfo(
		opts,
		gethcommon.HexToAddress(avsAddress), taskID)
	if err != nil {
		r.logger.Error("Failed to GetTaskInfo ", "err", err)
		return avs.TaskInfo{}, err
	}
	return info, nil
}

func (r *ChainReader) IsOperator(opts *bind.CallOpts, operator string) (bool, error) {
	flag, err := r.avsManager.IsOperator(
		opts,
		gethcommon.HexToAddress(operator))
	if err != nil {
		r.logger.Error("Failed to exec IsOperator ", "err", err)
		return false, err
	}
	return flag, nil
}

func (r *ChainReader) GetAVSEpochIdentifier(opts *bind.CallOpts, avsAddress string) (string, error) {
	epochIdentifier, err := r.avsManager.GetAVSEpochIdentifier(
		opts,
		gethcommon.HexToAddress(avsAddress))
	if err != nil {
		r.logger.Error("Failed to GetAVSEpochIdentifier ", "err", err)
		return "", err
	}
	return epochIdentifier, nil
}

func (r *ChainReader) GetCurrentEpoch(opts *bind.CallOpts, epochIdentifier string) (int64, error) {
	currentEpoch, err := r.avsManager.GetCurrentEpoch(
		opts,
		epochIdentifier)
	if err != nil {
		r.logger.Error("Failed to exec IsOperator ", "err", err)
		return 0, err
	}
	return currentEpoch, nil
}

func (r *ChainReader) GetChallengeInfo(opts *bind.CallOpts, taskAddress string, taskID uint64) (gethcommon.Address, error) {
	address, err := r.avsManager.GetChallengeInfo(
		opts,
		gethcommon.HexToAddress(taskAddress),
		taskID)
	if err != nil {
		r.logger.Error("Failed to exec IsOperator ", "err", err)
		return gethcommon.Address{}, err
	}
	return address, nil
}

func (r *ChainReader) GetOperatorTaskResponseList(opts *bind.CallOpts, taskAddress string, taskID uint64) ([]avs.OperatorResInfo, error) {
	res, err := r.avsManager.GetOperatorTaskResponseList(
		opts,
		gethcommon.HexToAddress(taskAddress),
		taskID,
	)
	if err != nil {
		r.logger.Error("Failed to GetOperatorTaskResponseList ", "err", err)
		return nil, err
	}
	return res, nil
}
