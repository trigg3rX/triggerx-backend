package chainio

import (
	"context"
	"errors"

	gethcommon "github.com/ethereum/go-ethereum/common"
	gethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/imua-xyz/imua-avs-sdk/client/txmgr"
	"github.com/imua-xyz/imua-avs-sdk/logging"
	avs "github.com/trigg3rX/imua-contracts/bindings/contracts/TriggerXAvs"
	"github.com/trigg3rX/triggerx-backend/cli/core/chainio/eth"
)

type AvsWriter interface {
	RegisterAVSToChain(
		ctx context.Context,
		params avs.AVSParams,
	) (*gethtypes.Receipt, error)

	RegisterBLSPublicKey(
		ctx context.Context,
		avsAddr string,
		pubKey []byte,
		pubKeyRegistrationSignature []byte,
	) (*gethtypes.Receipt, error)

	CreateNewTask(
		ctx context.Context,
		name string,
		numberToBeSquared uint8,
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

	// Challenge(
	// 	ctx context.Context,
	// 	req avs.AvsServiceContractChallengeReq,
	// ) (*gethtypes.Receipt, error)

	RegisterOperatorToAVS(
		ctx context.Context,
	) (*gethtypes.Receipt, error)
}

type ChainWriter struct {
	avsManager  avs.TriggerXAvs
	chainReader AvsReader
	ethClient   eth.EthClient
	logger      logging.Logger
	txMgr       txmgr.TxManager
}

var _ AvsWriter = (*ChainWriter)(nil)

func NewChainWriter(
	avsManager avs.TriggerXAvs,
	chainReader AvsReader,
	ethClient eth.EthClient,
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
	ethClient eth.EthClient,
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

func (w *ChainWriter) RegisterAVSToChain(
	ctx context.Context,
	params avs.AVSParams,
) (*gethtypes.Receipt, error) {

	noSendTxOpts, err := w.txMgr.GetNoSendTxOpts()
	if err != nil {
		return nil, err
	}
	tx, err := w.avsManager.RegisterAVS(
		noSendTxOpts,
		params)
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
func (w *ChainWriter) RegisterBLSPublicKey(
	ctx context.Context,
	avsAddr string,
	pubKey []byte,
	pubKeyRegistrationSignature []byte,
) (*gethtypes.Receipt, error) {
	noSendTxOpts, err := w.txMgr.GetNoSendTxOpts()
	if err != nil {
		return nil, err
	}
	tx, err := w.avsManager.RegisterBLSPublicKey(
		noSendTxOpts,
		gethcommon.HexToAddress(avsAddr),
		pubKey,
		pubKeyRegistrationSignature)
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
func (w *ChainWriter) CreateNewTask(
	ctx context.Context,
	name string,
	numberToBeSquared uint8,
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
		numberToBeSquared,
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

// func (w *ChainWriter) Challenge(ctx context.Context, req avs.AvsServiceContractChallengeReq) (*gethtypes.Receipt, error) {
// 	noSendTxOpts, err := w.txMgr.GetNoSendTxOpts()
// 	if err != nil {
// 		return nil, err
// 	}
// 	tx, err := w.avsManager.RaiseAndResolveChallenge(
// 		noSendTxOpts,
// 		req)
// 	if err != nil {
// 		return nil, err
// 	}
// 	receipt, err := w.txMgr.Send(ctx, tx)
// 	if err != nil {
// 		return nil, errors.New("failed to send tx with err: " + err.Error())
// 	}
// 	w.logger.Infof("tx hash: %s", tx.Hash().String())

// 	return receipt, nil
// }

func (w *ChainWriter) RegisterOperatorToAVS(
	ctx context.Context,
) (*gethtypes.Receipt, error) {
	noSendTxOpts, err := w.txMgr.GetNoSendTxOpts()
	if err != nil {
		return nil, err
	}
	tx, err := w.avsManager.RegisterOperatorToAVS(
		noSendTxOpts)
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

// func DeployAVS(
// 	ethClient eth.EthClient,
// 	logger logging.Logger,
// 	key ecdsa.PrivateKey,
// 	chainID *big.Int,
// ) (gethcommon.Address, string, error) {
// 	auth, err := bind.NewKeyedTransactorWithChainID(&key, chainID)
// 	if err != nil {
// 		logger.Fatalf("Failed to make transactor: %v", err)
// 	}

// 	address, tx, _, err := avs.DeployContracttriggerX(auth, ethClient)
// 	if err != nil {
// 		logger.Infof("deploy err: %s", err.Error())
// 		return gethcommon.Address{}, "", errors.New("failed to deploy contract with err: " + err.Error())
// 	}
// 	logger.Infof("tx hash: %s", tx.Hash().String())
// 	logger.Infof("contract address: %s", address.String())

// 	return address, tx.Hash().String(), nil
// }
