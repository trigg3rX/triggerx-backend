package registrar

import (
	"context"
	"fmt"
	"math/big"
	"os"
	"strings"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/event"
)

var (
	avsGovernanceABI     abi.ABI
	attestationCenterABI abi.ABI
)

// Custom event structures to replace the binding-generated ones
type OperatorRegistered struct {
	Operator common.Address
	BlsKey   [4]*big.Int
	Raw      types.Log
}

type OperatorUnregistered struct {
	Operator common.Address
	Raw      types.Log
}

type TaskSubmitted struct {
	Operator         common.Address
	TaskNumber       uint32
	ProofOfTask      string
	Data             []byte
	TaskDefinitionId uint16
	Raw              types.Log
}

type TaskRejected struct {
	Operator         common.Address
	TaskNumber       uint32
	ProofOfTask      string
	Data             []byte
	TaskDefinitionId uint16
	Raw              types.Log
}

// Initialize ABI parsers
func InitABI() error {
	// Load AvsGovernance ABI
	avsGovernanceABIJSON, err := os.ReadFile("pkg/bindings/abi/AvsGovernance.json")
	if err != nil {
		return fmt.Errorf("failed to read AvsGovernance ABI: %v", err)
	}
	avsGovernanceABI, err = abi.JSON(strings.NewReader(string(avsGovernanceABIJSON)))
	if err != nil {
		return fmt.Errorf("failed to parse AvsGovernance ABI: %v", err)
	}

	// Log discovered events for debugging
	logger.Info(fmt.Sprintf("AvsGovernance ABI loaded with %d events", len(avsGovernanceABI.Events)))

	// Note about event structures based on our observations
	logger.Info("Note: OperatorUnregistered event has empty data with operator address as second topic")

	// Load AttestationCenter ABI
	attestationCenterABIJSON, err := os.ReadFile("pkg/bindings/abi/AttestationCenter.json")
	if err != nil {
		return fmt.Errorf("failed to read AttestationCenter ABI: %v", err)
	}
	attestationCenterABI, err = abi.JSON(strings.NewReader(string(attestationCenterABIJSON)))
	if err != nil {
		return fmt.Errorf("failed to parse AttestationCenter ABI: %v", err)
	}

	// Log discovered events for AttestationCenter
	logger.Info(fmt.Sprintf("AttestationCenter ABI loaded with %d events", len(attestationCenterABI.Events)))

	logger.Info("Successfully loaded and parsed contract ABIs")
	return nil
}

// Subscribe to OperatorRegistered events
func SubscribeOperatorRegistered(
	client *ethclient.Client,
	contractAddr common.Address,
	sink chan<- *OperatorRegistered,
) (event.Subscription, error) {
	// Create the event signature hash
	eventSignature := []byte("OperatorRegistered(address,uint256[4])")
	topic := crypto.Keccak256Hash(eventSignature)

	// Create the filter query
	query := ethereum.FilterQuery{
		Addresses: []common.Address{contractAddr},
		Topics:    [][]common.Hash{{topic}},
	}

	// Subscribe to logs
	logs := make(chan types.Log)
	sub, err := client.SubscribeFilterLogs(context.Background(), query, logs)
	if err != nil {
		return nil, fmt.Errorf("failed to subscribe to OperatorRegistered events: %v", err)
	}

	// Process logs in background
	go func() {
		for {
			select {
			case err := <-sub.Err():
				logger.Error(fmt.Sprintf("Error in event subscription: %v", err))
				return
			case vLog := <-logs:
				// Parse the event
				event := new(OperatorRegistered)
				err := unpackOperatorRegistered(event, vLog)
				if err != nil {
					logger.Error(fmt.Sprintf("Failed to unpack OperatorRegistered event: %v", err))
					continue
				}
				event.Raw = vLog
				sink <- event
			}
		}
	}()

	return sub, nil
}

// Subscribe to OperatorUnregistered events
func SubscribeOperatorUnregistered(
	client *ethclient.Client,
	contractAddr common.Address,
	sink chan<- *OperatorUnregistered,
) (event.Subscription, error) {
	// Create the event signature hash based on the ABI definition
	// Our logs show that the operator is indexed (appears in topics) even if not marked in ABI
	eventSignature := []byte("OperatorUnregistered(address)")
	topic := crypto.Keccak256Hash(eventSignature)

	// Create the filter query
	query := ethereum.FilterQuery{
		Addresses: []common.Address{contractAddr},
		Topics:    [][]common.Hash{{topic}},
	}

	// Subscribe to logs
	logs := make(chan types.Log)
	sub, err := client.SubscribeFilterLogs(context.Background(), query, logs)
	if err != nil {
		return nil, fmt.Errorf("failed to subscribe to OperatorUnregistered events: %v", err)
	}

	// Process logs in background
	go func() {
		for {
			select {
			case err := <-sub.Err():
				logger.Error(fmt.Sprintf("Error in event subscription: %v", err))
				return
			case vLog := <-logs:
				// Parse the event
				event := new(OperatorUnregistered)
				err := unpackOperatorUnregistered(event, vLog)
				if err != nil {
					logger.Error(fmt.Sprintf("Failed to unpack OperatorUnregistered event: %v", err))
					continue
				}
				event.Raw = vLog
				sink <- event
			}
		}
	}()

	return sub, nil
}

// Subscribe to TaskSubmitted events
func SubscribeTaskSubmitted(
	client *ethclient.Client,
	contractAddr common.Address,
	sink chan<- *TaskSubmitted,
) (event.Subscription, error) {
	// Create the event signature hash
	eventSignature := []byte("TaskSubmitted(address,uint32,string,bytes,uint16)")
	topic := crypto.Keccak256Hash(eventSignature)

	// Create the filter query
	query := ethereum.FilterQuery{
		Addresses: []common.Address{contractAddr},
		Topics:    [][]common.Hash{{topic}},
	}

	// Subscribe to logs
	logs := make(chan types.Log)
	sub, err := client.SubscribeFilterLogs(context.Background(), query, logs)
	if err != nil {
		return nil, fmt.Errorf("failed to subscribe to TaskSubmitted events: %v", err)
	}

	// Process logs in background
	go func() {
		for {
			select {
			case err := <-sub.Err():
				logger.Error(fmt.Sprintf("Error in event subscription: %v", err))
				return
			case vLog := <-logs:
				// Parse the event
				event := new(TaskSubmitted)
				err := unpackTaskSubmitted(event, vLog)
				if err != nil {
					logger.Error(fmt.Sprintf("Failed to unpack TaskSubmitted event: %v", err))
					continue
				}
				event.Raw = vLog
				sink <- event
			}
		}
	}()

	return sub, nil
}

// Subscribe to TaskRejected events
func SubscribeTaskRejected(
	client *ethclient.Client,
	contractAddr common.Address,
	sink chan<- *TaskRejected,
) (event.Subscription, error) {
	// Create the event signature hash
	eventSignature := []byte("TaskRejected(address,uint32,string,bytes,uint16)")
	topic := crypto.Keccak256Hash(eventSignature)

	// Create the filter query
	query := ethereum.FilterQuery{
		Addresses: []common.Address{contractAddr},
		Topics:    [][]common.Hash{{topic}},
	}

	// Subscribe to logs
	logs := make(chan types.Log)
	sub, err := client.SubscribeFilterLogs(context.Background(), query, logs)
	if err != nil {
		return nil, fmt.Errorf("failed to subscribe to TaskRejected events: %v", err)
	}

	// Process logs in background
	go func() {
		for {
			select {
			case err := <-sub.Err():
				logger.Error(fmt.Sprintf("Error in event subscription: %v", err))
				return
			case vLog := <-logs:
				// Parse the event
				event := new(TaskRejected)
				err := unpackTaskRejected(event, vLog)
				if err != nil {
					logger.Error(fmt.Sprintf("Failed to unpack TaskRejected event: %v", err))
					continue
				}
				event.Raw = vLog
				sink <- event
			}
		}
	}()

	return sub, nil
}

// Helper functions to unpack events

func unpackOperatorRegistered(event *OperatorRegistered, log types.Log) error {
	err := avsGovernanceABI.UnpackIntoInterface(event, "OperatorRegistered", log.Data)
	if err != nil {
		return err
	}

	// Handle indexed fields (if any)
	if len(log.Topics) > 1 {
		event.Operator = common.HexToAddress(log.Topics[1].Hex())
	}

	return nil
}

// Unpacks an OperatorUnregistered event from a log
func unpackOperatorUnregistered(event *OperatorUnregistered, log types.Log) error {
	// Based on our logs, we now know the OperatorUnregistered event has:
	// - Empty data
	// - Operator address in the second topic

	if len(log.Topics) > 1 {
		event.Operator = common.HexToAddress(log.Topics[1].Hex())
		logger.Debug(fmt.Sprintf("Extracted operator from topic: %s", event.Operator.Hex()))
	} else {
		return fmt.Errorf("unable to extract operator address: no topic found for operator address")
	}

	// Set the raw log for transaction details
	event.Raw = log
	return nil
}

func unpackTaskSubmitted(event *TaskSubmitted, log types.Log) error {
	err := attestationCenterABI.UnpackIntoInterface(event, "TaskSubmitted", log.Data)
	if err != nil {
		logger.Warn(fmt.Sprintf("Error unpacking TaskSubmitted data: %v - will try fallback", err))

		// If we can't unpack the data, we'll still provide the raw log for analysis
		// Check if operator address is in topics
		if len(log.Topics) > 1 {
			event.Operator = common.HexToAddress(log.Topics[1].Hex())
			logger.Debug(fmt.Sprintf("Extracted operator from topic: %s", event.Operator.Hex()))
		}
	}

	// Set the raw log regardless of unpacking success
	event.Raw = log
	return nil
}

func unpackTaskRejected(event *TaskRejected, log types.Log) error {
	err := attestationCenterABI.UnpackIntoInterface(event, "TaskRejected", log.Data)
	if err != nil {
		logger.Warn(fmt.Sprintf("Error unpacking TaskRejected data: %v - will try fallback", err))

		// If we can't unpack the data, we'll still provide the raw log for analysis
		// Check if operator address is in topics
		if len(log.Topics) > 1 {
			event.Operator = common.HexToAddress(log.Topics[1].Hex())
			logger.Debug(fmt.Sprintf("Extracted operator from topic: %s", event.Operator.Hex()))
		}
	}

	// Set the raw log regardless of unpacking success
	event.Raw = log
	return nil
}
