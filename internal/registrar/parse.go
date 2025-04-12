package registrar

import (
	"encoding/binary"
	"fmt"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
)

func ParseOperatorRegistered(log ethtypes.Log) (*OperatorRegisteredEvent, error) {
	// Define the event ABI exactly as emitted by the contract
	eventABI := `[{
        "anonymous": false,
        "inputs": [
            {"indexed": true, "name": "operator", "type": "address"},
            {"indexed": false, "name": "blsKey", "type": "uint256[4]"}
        ],
        "name": "OperatorRegistered",
        "type": "event"
    }]`

	parsedABI, err := abi.JSON(strings.NewReader(eventABI))
	if err != nil {
		return nil, fmt.Errorf("failed to parse ABI: %v", err)
	}

	// Verify the event signature
	expectedTopic := crypto.Keccak256Hash([]byte("OperatorRegistered(address,uint256[4])"))
	if log.Topics[0] != expectedTopic {
		return nil, fmt.Errorf("unexpected event signature")
	}

	// The operator address is in the first indexed parameter (topic 1)
	if len(log.Topics) < 2 {
		return nil, fmt.Errorf("missing operator address in topics")
	}
	operator := common.BytesToAddress(log.Topics[1].Bytes())

	// We'll unpack just the BLS key data
	var blsKey struct {
		BlsKey [4]*big.Int
	}

	err = parsedABI.UnpackIntoInterface(&blsKey, "OperatorRegistered", log.Data)
	if err != nil {
		return nil, fmt.Errorf("failed to unpack log data: %v", err)
	}

	return &OperatorRegisteredEvent{
		Operator: operator,
		BlsKey:   blsKey.BlsKey,
		Raw:      log,
	}, nil
}

// ParseOperatorUnregistered parses a log into the OperatorUnregistered event
// ParseOperatorUnregistered parses a log into the OperatorUnregistered event
func ParseOperatorUnregistered(log ethtypes.Log) (*OperatorUnregisteredEvent, error) {
	// Verify the event signature
	expectedTopic := OperatorUnregisteredEventSignature()
	if log.Topics[0] != expectedTopic {
		return nil, fmt.Errorf("unexpected event signature")
	}

	// The operator address is in the first indexed parameter (topic 1)
	if len(log.Topics) < 2 {
		return nil, fmt.Errorf("missing operator address in topics")
	}
	operator := common.BytesToAddress(log.Topics[1].Bytes())

	return &OperatorUnregisteredEvent{
		Operator: operator,
		Raw:      log,
	}, nil
}

// ParseTaskSubmitted parses a log into the TaskSubmitted event
func ParseTaskSubmitted(log ethtypes.Log) (*TaskSubmittedEvent, error) {
	// Define the complete event ABI including attestersIds
	eventABI := `[{
        "anonymous": false,
        "inputs": [
            {"indexed": true, "name": "operator", "type": "address"},
            {"indexed": false, "name": "taskNumber", "type": "uint32"},
            {"indexed": false, "name": "proofOfTask", "type": "string"},
            {"indexed": false, "name": "data", "type": "bytes"},
            {"indexed": true, "name": "taskDefinitionId", "type": "uint16"},
            {"indexed": false, "name": "attestersIds", "type": "uint256[]"}
        ],
        "name": "TaskSubmitted",
        "type": "event"
    }]`

	parsedABI, err := abi.JSON(strings.NewReader(eventABI))
	if err != nil {
		return nil, fmt.Errorf("failed to parse ABI: %v", err)
	}

	// Verify the event signature
	expectedTopic := TaskSubmittedEventSignature()
	if log.Topics[0] != expectedTopic {
		return nil, fmt.Errorf("unexpected event signature: got %s, expected %s",
			log.Topics[0].Hex(), expectedTopic.Hex())
	}

	// Extract indexed parameters
	if len(log.Topics) < 3 {
		return nil, fmt.Errorf("not enough topics for TaskSubmitted event")
	}

	operator := common.BytesToAddress(log.Topics[1].Bytes())
	taskDefinitionId := binary.BigEndian.Uint16(log.Topics[2].Bytes()[30:32]) // Last 2 bytes

	// Unpack non-indexed parameters
	var unpacked struct {
		TaskNumber   uint32
		ProofOfTask  string
		Data         []byte
		AttestersIds []*big.Int
	}

	if err := parsedABI.UnpackIntoInterface(&unpacked, "TaskSubmitted", log.Data); err != nil {
		return nil, fmt.Errorf("failed to unpack event data: %v", err)
	}

	// Convert attestersIds to strings
	attestersIds := make([]string, len(unpacked.AttestersIds))
	for i, id := range unpacked.AttestersIds {
		attestersIds[i] = id.String()
	}

	return &TaskSubmittedEvent{
		Operator:         operator,
		TaskNumber:       unpacked.TaskNumber,
		ProofOfTask:      unpacked.ProofOfTask,
		Data:             unpacked.Data,
		TaskDefinitionId: taskDefinitionId,
		AttestersIds:     unpacked.AttestersIds,
		Raw:              log,
	}, nil
}

// ParseTaskRejected parses a log into the TaskRejected event
func ParseTaskRejected(log ethtypes.Log) (*TaskRejectedEvent, error) {
	// Define the complete event ABI including attestersIds
	eventABI := `[{
        "anonymous": false,
        "inputs": [
            {"indexed": true, "name": "operator", "type": "address"},
            {"indexed": false, "name": "taskNumber", "type": "uint32"},
            {"indexed": false, "name": "proofOfTask", "type": "string"},
            {"indexed": false, "name": "data", "type": "bytes"},
            {"indexed": true, "name": "taskDefinitionId", "type": "uint16"},
            {"indexed": false, "name": "attestersIds", "type": "uint256[]"}
        ],
        "name": "TaskRejected",
        "type": "event"
    }]`

	parsedABI, err := abi.JSON(strings.NewReader(eventABI))
	if err != nil {
		return nil, fmt.Errorf("failed to parse ABI: %v", err)
	}

	// Verify the event signature
	expectedTopic := TaskRejectedEventSignature()
	if log.Topics[0] != expectedTopic {
		return nil, fmt.Errorf("unexpected event signature: got %s, expected %s",
			log.Topics[0].Hex(), expectedTopic.Hex())
	}

	// Extract indexed parameters
	if len(log.Topics) < 3 {
		return nil, fmt.Errorf("not enough topics for TaskRejected event")
	}

	operator := common.BytesToAddress(log.Topics[1].Bytes())
	taskDefinitionId := binary.BigEndian.Uint16(log.Topics[2].Bytes()[30:32]) // Last 2 bytes

	// Unpack non-indexed parameters
	var unpacked struct {
		TaskNumber   uint32
		ProofOfTask  string
		Data         []byte
		AttestersIds []*big.Int
	}

	if err := parsedABI.UnpackIntoInterface(&unpacked, "TaskRejected", log.Data); err != nil {
		return nil, fmt.Errorf("failed to unpack event data: %v", err)
	}

	// Convert attestersIds to strings
	attestersIds := make([]string, len(unpacked.AttestersIds))
	for i, id := range unpacked.AttestersIds {
		attestersIds[i] = id.String()
	}

	return &TaskRejectedEvent{
		Operator:         operator,
		TaskNumber:       unpacked.TaskNumber,
		ProofOfTask:      unpacked.ProofOfTask,
		Data:             unpacked.Data,
		TaskDefinitionId: taskDefinitionId,
		AttestersIds:     unpacked.AttestersIds, // Include the attesters IDs
		Raw:              log,
	}, nil
}
