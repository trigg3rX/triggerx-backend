package chainio

import (
	"context"
	"errors"
	gethcommon "github.com/ethereum/go-ethereum/common"
	gethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/imua-xyz/imua-avs-sdk/client/txmgr"
	avs "github.com/trigg3rX/imua-contracts/bindings/contracts/TriggerXAvs"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
)

type AvsWriter interface {
	CreateNewTask(
		ctx context.Context,
		name string,
		taskDefinitionId uint8,
		taskResponsePeriod uint64,
		taskChallengePeriod uint64,
		thresholdPercentage uint8,
		taskStatisticalPeriod uint64,
	) (*gethtypes.Receipt, error)

	OperatorSubmitTask(
		ctx context.Context,
		taskID uint64,
		taskResponse []byte,
		blsSignature []byte,
		taskContractAddress string,
		phase uint8,
	) (*gethtypes.Receipt, error)
}

type ChainWriter struct {
	avsManager  avs.TriggerXAvs
	chainReader AvsReader
	ethClient   *ethclient.Client
	logger      logging.Logger
	txMgr       txmgr.TxManager
}

var _ AvsWriter = (*ChainWriter)(nil)

func NewChainWriter(
	avsManager avs.TriggerXAvs,
	chainReader AvsReader,
	ethClient *ethclient.Client,
	logger logging.Logger,
	txMgr txmgr.TxManager,
) *ChainWriter {
	return &ChainWriter{
		avsManager:  avsManager,
		chainReader: chainReader,
		logger:      logger,
		ethClient:   ethClient,
		txMgr:       txMgr,
	}
}

func BuildChainWriter(
	avsAddr gethcommon.Address,
	ethClient *ethclient.Client,
	logger logging.Logger,
	txMgr txmgr.TxManager,
) (*ChainWriter, error) {
	contractBindings, err := NewContractBindings(
		avsAddr,
		ethClient,
		logger,
	)
	if err != nil {
		return nil, err
	}
	chainReader := NewChainReader(
		*contractBindings.AVSManager,
		logger,
		ethClient,
	)
	return NewChainWriter(
		*contractBindings.AVSManager,
		chainReader,
		ethClient,
		logger,
		txMgr,
	), nil
}

func (w *ChainWriter) CreateNewTask(
	ctx context.Context,
	name string,
	taskDefinitionId uint8,
	taskResponsePeriod uint64,
	taskChallengePeriod uint64,
	thresholdPercentage uint8,
	taskStatisticalPeriod uint64,
) (*gethtypes.Receipt, error) {
	noSendTxOpts, err := w.txMgr.GetNoSendTxOpts()
	if err != nil {
		return nil, err
	}
	tx, err := w.avsManager.CreateTask(
		noSendTxOpts,
		name,
		taskDefinitionId,
		taskResponsePeriod,
		taskChallengePeriod,
		thresholdPercentage,
		taskStatisticalPeriod)
	if err != nil {
		return nil, err
	}
	receipt, err := w.txMgr.Send(ctx, tx)
	if err != nil {
		return nil, errors.New("failed to send tx with err: " + err.Error())
	}
	w.logger.Infof("tx hash: %s", tx.Hash().String())

	return receipt, nil
}

func (w *ChainWriter) OperatorSubmitTask(
	ctx context.Context,
	taskID uint64,
	taskResponse []byte,
	blsSignature []byte,
	taskContractAddress string,
	phase uint8,
) (*gethtypes.Receipt, error) {
	noSendTxOpts, err := w.txMgr.GetNoSendTxOpts()
	if err != nil {
		return nil, err
	}
	tx, err := w.avsManager.OperatorSubmitTask(
		noSendTxOpts,
		taskID,
		taskResponse,
		blsSignature,
		gethcommon.HexToAddress(taskContractAddress),
		phase)
	if err != nil {
		return nil, err
	}
	receipt, err := w.txMgr.Send(ctx, tx)
	if err != nil {
		return nil, errors.New("failed to send tx with err: " + err.Error())
	}
	w.logger.Infof("tx hash: %s", tx.Hash().String())

	return receipt, nil
}
