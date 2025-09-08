package chainio

import (
	"cosmossdk.io/math"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	gethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/imua-xyz/imua-avs-sdk/logging"
	avs "github.com/trigg3rX/imua-contracts/bindings/contracts/TriggerXAvs"
	"github.com/trigg3rX/triggerx-backend/cli/core/chainio/eth"
)

type AvsReader interface {
	GetOptInOperators(
		opts *bind.CallOpts,
		avsAddress string,
	) ([]gethcommon.Address, error)

	GetRegisteredPubkey(
		opts *bind.CallOpts,
		operator string,
		avsAddress string,
	) ([]byte, error)
	GtAVSUSDValue(
		opts *bind.CallOpts,
		avsAddress string,
	) (math.LegacyDec, error)

	GetOperatorOptedUSDValue(
		opts *bind.CallOpts,
		avsAddress string,
		operatorAddr string,
	) (math.LegacyDec, error)
	GetAVSEpochIdentifier(
		opts *bind.CallOpts,
		avsAddress string,
	) (string, error)
	GetTaskInfo(
		opts *bind.CallOpts,
		avsAddress string,
		taskID uint64,
	) (avs.TaskInfo, error)
	IsOperator(
		opts *bind.CallOpts,
		operator string,
	) (bool, error)

	GetCurrentEpoch(
		opts *bind.CallOpts,
		epochIdentifier string,
	) (int64, error)
	GetChallengeInfo(
		opts *bind.CallOpts,
		taskAddress string,
		taskID uint64,
	) (gethcommon.Address, error)
	GetOperatorTaskResponse(
		opts *bind.CallOpts,
		taskAddress string,
		operatorAddress string,
		taskID uint64,
	) (avs.TaskResultInfo, error)
	GetOperatorTaskResponseList(
		opts *bind.CallOpts,
		taskAddress string,
		taskID uint64,
	) ([]avs.OperatorResInfo, error)
}

type ChainReader struct {
	logger     logging.Logger
	avsManager avs.TriggerXAvs
	ethClient  eth.EthClient
}

// forces EthReader to implement the chainio.Reader interface
var _ AvsReader = (*ChainReader)(nil)

func NewChainReader(
	avsManager avs.TriggerXAvs,
	logger logging.Logger,
	ethClient eth.EthClient,
) *ChainReader {
	return &ChainReader{
		avsManager: avsManager,
		logger:     logger,
		ethClient:  ethClient,
	}
}

func BuildChainReader(
	avsAddr gethcommon.Address,
	ethClient eth.EthClient,
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

func (r *ChainReader) GetOptInOperators(
	opts *bind.CallOpts,
	avsAddress string,
) ([]gethcommon.Address, error) {
	operators, err := r.avsManager.GetOptInOperators(
		opts,
		gethcommon.HexToAddress(avsAddress))
	if err != nil {
		r.logger.Error("Failed to GetOptInOperators ", "err", err)
		return nil, err
	}
	return operators, nil
}

func (r *ChainReader) GetRegisteredPubkey(opts *bind.CallOpts, operator string, avsAddress string) ([]byte, error) {
	pukKey, err := r.avsManager.GetRegisteredPubkey(
		opts,
		gethcommon.HexToAddress(operator), gethcommon.HexToAddress(avsAddress))
	if err != nil {
		r.logger.Error("Failed to GetRegisteredPubkey ", "err", err)
		return nil, err
	}
	return pukKey, nil
}

func (r *ChainReader) GtAVSUSDValue(opts *bind.CallOpts, avsAddress string) (math.LegacyDec, error) {
	amount, err := r.avsManager.GetAVSUSDValue(
		opts,
		gethcommon.HexToAddress(avsAddress))
	if err != nil {
		r.logger.Error("Failed to GtAVSUSDValue ", "err", err)
		return math.LegacyDec{}, err
	}
	return math.LegacyNewDecFromBigInt(amount), nil
}

func (r *ChainReader) GetOperatorOptedUSDValue(opts *bind.CallOpts, avsAddress string, operatorAddr string) (math.LegacyDec, error) {
	amount, err := r.avsManager.GetOperatorOptedUSDValue(
		opts,
		gethcommon.HexToAddress(avsAddress), gethcommon.HexToAddress(operatorAddr))
	if err != nil {
		r.logger.Error("Failed to GetOperatorOptedUSDValue ", "err", err)
		return math.LegacyDec{}, err
	}
	return math.LegacyNewDecFromBigInt(amount), nil
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

func (r *ChainReader) GetOperatorTaskResponse(opts *bind.CallOpts, taskAddress string, operatorAddress string, taskID uint64) (avs.TaskResultInfo, error) {
	res, err := r.avsManager.GetOperatorTaskResponse(
		opts,
		gethcommon.HexToAddress(taskAddress),
		gethcommon.HexToAddress(operatorAddress),
		taskID)
	if err != nil {
		r.logger.Error("Failed to exec IsOperator ", "err", err)
		return avs.TaskResultInfo{}, err
	}
	return res, nil
}

func (r *ChainReader) GetOperatorTaskResponseList(opts *bind.CallOpts, taskAddress string, taskID uint64) ([]avs.OperatorResInfo, error) {
	res, err := r.avsManager.GetOperatorTaskResponseList(
		opts,
		gethcommon.HexToAddress(taskAddress),
		taskID)
	if err != nil {
		r.logger.Error("Failed to exec IsOperator ", "err", err)
		return []avs.OperatorResInfo{}, err
	}
	return res, nil
}
