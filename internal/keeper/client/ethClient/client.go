package ethclient

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

// EthClient is an interface for Ethereum client operations
type EthClient interface {
	ChainID(ctx context.Context) (*big.Int, error)
	BalanceAt(ctx context.Context, account common.Address, blockNumber *big.Int) (*big.Int, error)
	FilterLogs(ctx context.Context, query ethereum.FilterQuery) ([]types.Log, error)
	SubscribeFilterLogs(ctx context.Context, query ethereum.FilterQuery, ch chan<- types.Log) (ethereum.Subscription, error)
	CallContract(ctx context.Context, call ethereum.CallMsg, blockNumber *big.Int) ([]byte, error)
	bind.ContractBackend
}

// Client wraps the ethclient.Client to implement our EthClient interface
type Client struct {
	*ethclient.Client
}

// NewClient creates a new Ethereum client
func NewClient(rpcURL string) (EthClient, error) {
	client, err := ethclient.Dial(rpcURL)
	if err != nil {
		return nil, err
	}
	return &Client{Client: client}, nil
}

// NewWebSocketClient creates a new WebSocket Ethereum client for event subscriptions
func NewWebSocketClient(wsURL string) (EthClient, error) {
	client, err := ethclient.Dial(wsURL)
	if err != nil {
		return nil, err
	}
	return &Client{Client: client}, nil
}
