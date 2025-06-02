package events

import (
	"context"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/trigg3rX/triggerx-backend/internal/registrar/client"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
)

// OperatorProcessor handles operator-related events
type OperatorProcessor struct {
	*EventProcessor
}

// NewOperatorProcessor creates a new operator event processor
func NewOperatorProcessor(base *EventProcessor) *OperatorProcessor {
	if base == nil {
		panic("base processor cannot be nil")
	}
	return &OperatorProcessor{
		EventProcessor: base,
	}
}

// ProcessOperatorRegisteredEvents processes OperatorRegistered events from the blockchain
func (p *OperatorProcessor) ProcessOperatorRegisteredEvents(
	ctx context.Context,
	ethClient *ethclient.Client,
	contractAddress common.Address,
	fromBlock uint64,
	toBlock uint64,
) error {
	query := ethereum.FilterQuery{
		FromBlock: new(big.Int).SetUint64(fromBlock),
		ToBlock:   new(big.Int).SetUint64(toBlock),
		Addresses: []common.Address{contractAddress},
		Topics: [][]common.Hash{
			{OperatorRegisteredEventSignature()},
		},
	}

	logs, err := ethClient.FilterLogs(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to filter OperatorRegistered logs: %w", err)
	}

	p.logger.Debug("Processing OperatorRegistered events",
		"count", len(logs),
		"fromBlock", fromBlock,
		"toBlock", toBlock,
	)

	for _, vLog := range logs {
		event, err := ParseOperatorRegistered(vLog)
		if err != nil {
			p.logger.Error("Failed to parse OperatorRegistered event", "error", err)
			continue
		}

		p.logger.Info("Operator registered",
			"operator", event.Operator.Hex(),
			"txHash", event.Raw.TxHash.Hex(),
			"blockNumber", event.Raw.BlockNumber,
		)

		if err := client.KeeperRegistered(event.Operator.Hex(), event.Raw.TxHash.Hex()); err != nil {
			p.logger.Error("Failed to add keeper to database",
				"error", err,
				"operator", event.Operator.Hex(),
			)
			continue
		}

		// Schedule fetching operator details
		if err := FetchOperatorDetailsAfterDelay(event.Operator, 4*time.Minute, p.logger); err != nil {
			p.logger.Error("Failed to fetch operator details after delay",
				"error", err,
				"operator", event.Operator.Hex(),
			)
		}
	}
	return nil
}

// ProcessOperatorUnregisteredEvents processes OperatorUnregistered events from the blockchain
func (p *OperatorProcessor) ProcessOperatorUnregisteredEvents(
	ctx context.Context,
	ethClient *ethclient.Client,
	contractAddress common.Address,
	fromBlock uint64,
	toBlock uint64,
) error {
	query := ethereum.FilterQuery{
		FromBlock: new(big.Int).SetUint64(fromBlock),
		ToBlock:   new(big.Int).SetUint64(toBlock),
		Addresses: []common.Address{contractAddress},
		Topics: [][]common.Hash{
			{OperatorUnregisteredEventSignature()},
		},
	}

	logs, err := ethClient.FilterLogs(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to filter OperatorUnregistered logs: %w", err)
	}

	p.logger.Debug("Processing OperatorUnregistered events",
		"count", len(logs),
		"fromBlock", fromBlock,
		"toBlock", toBlock,
	)

	for _, vLog := range logs {
		event, err := ParseOperatorUnregistered(vLog)
		if err != nil {
			p.logger.Error("Failed to parse OperatorUnregistered event", "error", err)
			continue
		}

		p.logger.Info("Operator unregistered",
			"operator", event.Operator.Hex(),
			"txHash", event.Raw.TxHash.Hex(),
			"blockNumber", event.Raw.BlockNumber,
		)

		if err := client.KeeperUnregistered(event.Operator.Hex()); err != nil {
			p.logger.Error("Failed to update keeper status in database",
				"error", err,
				"operator", event.Operator.Hex(),
			)
		}
	}
	return nil
}

// Event signatures
func OperatorRegisteredEventSignature() common.Hash {
	return crypto.Keccak256Hash([]byte("OperatorRegistered(address,uint256[4])"))
}

func OperatorUnregisteredEventSignature() common.Hash {
	return crypto.Keccak256Hash([]byte("OperatorUnregistered(address)"))
}

// ProcessOperatorRegisteredEvents processes OperatorRegistered events from the blockchain
func ProcessOperatorRegisteredEvents(
	client *ethclient.Client,
	contractAddress common.Address,
	fromBlock uint64,
	toBlock uint64,
	logger logging.Logger,
) error {
	logger = logger.With("processor", "operator_registered")
	processor := NewOperatorProcessor(NewEventProcessor(logger))
	return processor.ProcessOperatorRegisteredEvents(context.Background(), client, contractAddress, fromBlock, toBlock)
}

// ProcessOperatorUnregisteredEvents processes OperatorUnregistered events from the blockchain
func ProcessOperatorUnregisteredEvents(
	client *ethclient.Client,
	contractAddress common.Address,
	fromBlock uint64,
	toBlock uint64,
	logger logging.Logger,
) error {
	logger = logger.With("processor", "operator_unregistered")
	processor := NewOperatorProcessor(NewEventProcessor(logger))
	return processor.ProcessOperatorUnregisteredEvents(context.Background(), client, contractAddress, fromBlock, toBlock)
}

func ParseOperatorRegistered(log ethtypes.Log) (*OperatorRegisteredEvent, error) {
	expectedTopic := crypto.Keccak256Hash([]byte("OperatorRegistered(address,uint256[4])"))
	if log.Topics[0] != expectedTopic {
		return nil, fmt.Errorf("unexpected event signature")
	}

	if len(log.Topics) < 2 {
		return nil, fmt.Errorf("missing operator address in topics")
	}
	operator := common.BytesToAddress(log.Topics[1].Bytes())

	var blsKey struct {
		BlsKey [4]*big.Int
	}

	err := AvsGovernanceABI.UnpackIntoInterface(&blsKey, "OperatorRegistered", log.Data)
	if err != nil {
		return nil, fmt.Errorf("failed to unpack log data: %v", err)
	}

	return &OperatorRegisteredEvent{
		Operator: operator,
		BlsKey:   blsKey.BlsKey,
		Raw:      log,
	}, nil
}

func ParseOperatorUnregistered(log ethtypes.Log) (*OperatorUnregisteredEvent, error) {
	expectedTopic := OperatorUnregisteredEventSignature()
	if log.Topics[0] != expectedTopic {
		return nil, fmt.Errorf("unexpected event signature")
	}

	if len(log.Topics) < 2 {
		return nil, fmt.Errorf("missing operator address in topics")
	}
	operator := common.BytesToAddress(log.Topics[1].Bytes())

	return &OperatorUnregisteredEvent{
		Operator: operator,
		Raw:      log,
	}, nil
}
