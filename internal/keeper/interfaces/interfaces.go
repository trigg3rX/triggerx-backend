package interfaces

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	jobtypes "github.com/trigg3rX/triggerx-backend/pkg/types"
)

type ValidatorInterface interface {
	ValidateTimeBasedJob(job *jobtypes.HandleCreateJobData) (bool, error)
	ValidateEventBasedJob(job *jobtypes.HandleCreateJobData, ipfsData *jobtypes.IPFSData) (bool, error)
	ValidateAndPrepareJob(job *jobtypes.HandleCreateJobData, triggerData *jobtypes.TriggerData) (bool, error)
}

type Logger interface {
	Infof(format string, args ...interface{})
	Errorf(format string, args ...interface{})
	Warnf(format string, args ...interface{})
}

type EthClientInterface interface {
	PendingNonceAt(ctx context.Context, account common.Address) (uint64, error)
	SuggestGasPrice(ctx context.Context) (*big.Int, error)
	SendTransaction(ctx context.Context, tx *types.Transaction) error
	CodeAt(ctx context.Context, contract common.Address, blockNumber *big.Int) ([]byte, error)
	CallContract(ctx context.Context, call ethereum.CallMsg, blockNumber *big.Int) ([]byte, error)
	HeaderByNumber(ctx context.Context, number *big.Int) (*types.Header, error)
	TransactionReceipt(ctx context.Context, txHash common.Hash) (*types.Receipt, error)
}
