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
	"github.com/trigg3rX/triggerx-backend/internal/registrar/config"
	"github.com/trigg3rX/triggerx-backend/pkg/ipfs"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
)

// TaskProcessor handles task-related events
type TaskProcessor struct {
	*EventProcessor
	logger logging.Logger
}

// NewTaskProcessor creates a new task event processor
func NewTaskProcessor(base *EventProcessor, logger logging.Logger) *TaskProcessor {
	if base == nil {
		panic("base processor cannot be nil")
	}
	return &TaskProcessor{
		EventProcessor: base,
		logger:         logger,
	}
}

func (t *TaskProcessor) processEventsInBatches(
	client *ethclient.Client,
	contractAddress common.Address,
	fromBlock uint64,
	toBlock uint64,
	batchSize uint64,
	processFunc func(client *ethclient.Client, contractAddress common.Address, start uint64, end uint64) error,
) error {
	for start := fromBlock; start < toBlock; start += batchSize {
		end := start + batchSize - 1
		if end > toBlock {
			end = toBlock
		}

		if err := processFunc(client, contractAddress, start, end); err != nil {
			return fmt.Errorf("failed to process batch from %d to %d: %v", start, end, err)
		}
	}
	return nil
}

func (t *TaskProcessor) ProcessTaskSubmittedEvents(
	baseClient *ethclient.Client,
	contractAddress common.Address,
	fromBlock uint64,
	toBlock uint64,
) error {
	return t.processEventsInBatches(baseClient, contractAddress, fromBlock, toBlock, 500, t.processTaskSubmittedBatch)
}

func (t *TaskProcessor) processTaskSubmittedBatch(
	baseClient *ethclient.Client,
	contractAddress common.Address,
	fromBlock uint64,
	toBlock uint64,
) error {
	query := ethereum.FilterQuery{
		FromBlock: new(big.Int).SetUint64(fromBlock),
		ToBlock:   new(big.Int).SetUint64(toBlock),
		Addresses: []common.Address{contractAddress},
		Topics: [][]common.Hash{
			{TaskSubmittedEventSignature()},
		},
	}

	logs, err := baseClient.FilterLogs(context.Background(), query)
	if err != nil {
		return fmt.Errorf("failed to filter TaskSubmitted logs: %v", err)
	}

	t.logger.Debugf("Found %d TaskSubmitted events in batch [%d-%d]", len(logs), fromBlock, toBlock)

	for _, vLog := range logs {
		event, err := ParseTaskSubmitted(vLog)
		if err != nil {
			t.logger.Errorf("Failed to parse TaskSubmitted event: %v", err)
			continue
		}

		t.logger.Infof("Task Submitted Event Detected!")
		t.logger.Debugf("Performer Address: %s", event.Operator)
		t.logger.Debugf("Attester IDs: %v", event.AttestersIds)
		t.logger.Debugf("Task Number: %d", event.TaskNumber)
		t.logger.Debugf("Task Definition ID: %d", event.TaskDefinitionId)

		if event.TaskDefinitionId == 10001 || event.TaskDefinitionId == 10002 {
		} else {
			dataCID := string(event.Data)
			t.logger.Debugf("Decoded Data: %s", dataCID)

			ipfsData, err := ipfs.FetchIPFSContent(config.GetPinataHost(), dataCID)
			if err != nil {
				t.logger.Errorf("Failed to fetch IPFS content: %v", err)
				continue
			}

			// Get execution timestamp from IPFS data
			var executionTimestamp time.Time
			if ipfsData.ActionData != nil && !ipfsData.ActionData.ExecutionTimestamp.IsZero() {
				executionTimestamp = ipfsData.ActionData.ExecutionTimestamp
				t.logger.Debugf("Using execution timestamp from IPFS data: %v", executionTimestamp)
			} else {
				// Fallback to current time if not available in IPFS data
				executionTimestamp = time.Now()
				t.logger.Warnf("No execution timestamp found in IPFS data, using current time")
			}

			// Get execution transaction hash from IPFS data
			var executionTxHash string
			if ipfsData.ActionData != nil && ipfsData.ActionData.ActionTxHash != "" {
				executionTxHash = ipfsData.ActionData.ActionTxHash
				t.logger.Debugf("Using execution tx hash from IPFS data: %s", executionTxHash)
			} else {
				// Fallback to event transaction hash if not available in IPFS data
				executionTxHash = event.Raw.TxHash.Hex()
				t.logger.Warnf("No execution tx hash found in IPFS data, using event tx hash")
			}

			// Get proof of task from IPFS data
			var proofOfTask string
			if ipfsData.ProofData != nil && ipfsData.ProofData.ProofOfTask != "" {
				proofOfTask = ipfsData.ProofData.ProofOfTask
				t.logger.Debugf("Using proof of task from IPFS data: %s", proofOfTask[:min(50, len(proofOfTask))]+"...")
			} else {
				t.logger.Warnf("No proof of task found in IPFS data")
			}

			// Update task number and success status to true (submitted) in database
			taskID := int(ipfsData.TaskData.TriggerData[0].TaskID)
			if err := client.UpdateTaskNumberAndStatus(taskID, int64(event.TaskNumber), true, event.Raw.TxHash.Hex(),
				event.Operator.Hex(), ConvertBigIntToStrings(event.AttestersIds), executionTxHash, executionTimestamp, proofOfTask); err != nil {
				t.logger.Errorf("Failed to update task execution details in database: %v", err)
				continue
			}

			// Update points for performers and attesters
			if err := client.UpdatePointsInDatabase(taskID, event.Operator, ConvertBigIntToStrings(event.AttestersIds), true); err != nil {
				t.logger.Errorf("Failed to update points in database: %v", err)
				continue
			}

			// Update job status to "running"
			if err := client.UpdateJobStatus(ipfsData.TaskData.TriggerData[0].TaskID, "running"); err != nil {
				t.logger.Errorf("Failed to update job status: %v", err)
				continue
			}

			// Schedule 24h delayed deletion
			scheduleCIDDeletion(dataCID, t.logger)
		}
	}
	return nil
}

func (t *TaskProcessor) ProcessTaskRejectedEvents(
	baseClient *ethclient.Client,
	contractAddress common.Address,
	fromBlock uint64,
	toBlock uint64,
) error {
	return t.processEventsInBatches(baseClient, contractAddress, fromBlock, toBlock, 500, t.processTaskRejectedBatch)
}

func (t *TaskProcessor) processTaskRejectedBatch(
	baseClient *ethclient.Client,
	contractAddress common.Address,
	fromBlock uint64,
	toBlock uint64,
) error {
	query := ethereum.FilterQuery{
		FromBlock: new(big.Int).SetUint64(fromBlock),
		ToBlock:   new(big.Int).SetUint64(toBlock),
		Addresses: []common.Address{contractAddress},
		Topics: [][]common.Hash{
			{TaskRejectedEventSignature()},
		},
	}

	logs, err := baseClient.FilterLogs(context.Background(), query)
	if err != nil {
		return fmt.Errorf("failed to filter TaskRejected logs: %v", err)
	}

	t.logger.Debugf("Found %d TaskRejected events in batch [%d-%d]", len(logs), fromBlock, toBlock)

	for _, vLog := range logs {
		event, err := ParseTaskRejected(vLog)
		if err != nil {
			t.logger.Errorf("Failed to parse TaskRejected event: %v", err)
			continue
		}

		t.logger.Infof("Task Rejected Event Detected!")
		t.logger.Debugf("Performer Address: %s", event.Operator)
		t.logger.Debugf("Attester IDs: %v", event.AttestersIds)
		t.logger.Debugf("Task Number: %d", event.TaskNumber)
		t.logger.Debugf("Task Definition ID: %d", event.TaskDefinitionId)

		if event.TaskDefinitionId == 10001 || event.TaskDefinitionId == 10002 {
		} else {
			dataCID := string(event.Data)
			t.logger.Debugf("Decoded Data: %s", dataCID)

			ipfsData, err := ipfs.FetchIPFSContent(config.GetPinataHost(), dataCID)
			if err != nil {
				t.logger.Errorf("Failed to fetch IPFS content: %v", err)
				continue
			}

			// Get execution timestamp from IPFS data
			var executionTimestamp time.Time
			if ipfsData.ActionData != nil && !ipfsData.ActionData.ExecutionTimestamp.IsZero() {
				executionTimestamp = ipfsData.ActionData.ExecutionTimestamp
				t.logger.Debugf("Using execution timestamp from IPFS data: %v", executionTimestamp)
			} else {
				// Fallback to current time if not available in IPFS data
				executionTimestamp = time.Now()
				t.logger.Warnf("No execution timestamp found in IPFS data, using current time")
			}

			// Get execution transaction hash from IPFS data
			var executionTxHash string
			if ipfsData.ActionData != nil && ipfsData.ActionData.ActionTxHash != "" {
				executionTxHash = ipfsData.ActionData.ActionTxHash
				t.logger.Debugf("Using execution tx hash from IPFS data: %s", executionTxHash)
			} else {
				// Fallback to event transaction hash if not available in IPFS data
				executionTxHash = event.Raw.TxHash.Hex()
				t.logger.Warnf("No execution tx hash found in IPFS data, using event tx hash")
			}

			// Get proof of task from IPFS data
			var proofOfTask string
			if ipfsData.ProofData != nil && ipfsData.ProofData.ProofOfTask != "" {
				proofOfTask = ipfsData.ProofData.ProofOfTask
				t.logger.Debugf("Using proof of task from IPFS data: %s", proofOfTask[:min(50, len(proofOfTask))]+"...")
			} else {
				t.logger.Warnf("No proof of task found in IPFS data")
			}

			// Update task number and success status to false (rejected) in database
			taskID := int(ipfsData.TaskData.TriggerData[0].TaskID)
			if err := client.UpdateTaskNumberAndStatus(taskID, int64(event.TaskNumber), false, event.Raw.TxHash.Hex(),
				event.Operator.Hex(), ConvertBigIntToStrings(event.AttestersIds), executionTxHash, executionTimestamp, proofOfTask); err != nil {
				t.logger.Errorf("Failed to update task execution details in database: %v", err)
				continue
			}

			// Update points for performers and attesters (with rejection penalty)
			if err := client.UpdatePointsInDatabase(taskID, event.Operator, ConvertBigIntToStrings(event.AttestersIds), false); err != nil {
				t.logger.Errorf("Failed to update points in database: %v", err)
				continue
			}

			// Update job status to "failed" for rejected tasks
			if err := client.UpdateJobStatus(ipfsData.TaskData.TriggerData[0].TaskID, "failed"); err != nil {
				t.logger.Errorf("Failed to update job status: %v", err)
				continue
			}

			// Schedule 24h delayed deletion
			scheduleCIDDeletion(dataCID, t.logger)
		}
	}
	return nil
}

func TaskSubmittedEventSignature() common.Hash {
	return crypto.Keccak256Hash([]byte("TaskSubmitted(address,uint32,string,bytes,uint16,uint256[])"))
}

func TaskRejectedEventSignature() common.Hash {
	return crypto.Keccak256Hash([]byte("TaskRejected(address,uint32,string,bytes,uint16,uint256[])"))
}

// ProcessTaskSubmittedEvents processes TaskSubmitted events from the blockchain
func ProcessTaskSubmittedEvents(
	client *ethclient.Client,
	contractAddress common.Address,
	fromBlock uint64,
	toBlock uint64,
	logger logging.Logger,
) error {
	logger = logger.With("processor", "task_submitted")
	processor := NewTaskProcessor(NewEventProcessor(logger), logger)
	return processor.ProcessTaskSubmittedEvents(client, contractAddress, fromBlock, toBlock)
}

// ProcessTaskRejectedEvents processes TaskRejected events from the blockchain
func ProcessTaskRejectedEvents(
	client *ethclient.Client,
	contractAddress common.Address,
	fromBlock uint64,
	toBlock uint64,
	logger logging.Logger,
) error {
	logger = logger.With("processor", "task_rejected")
	processor := NewTaskProcessor(NewEventProcessor(logger), logger)
	return processor.ProcessTaskRejectedEvents(client, contractAddress, fromBlock, toBlock)
}

// ParseTaskSubmitted parses a TaskSubmitted event from the log
func ParseTaskSubmitted(log ethtypes.Log) (*TaskSubmittedEvent, error) {
	expectedTopic := TaskSubmittedEventSignature()
	if log.Topics[0] != expectedTopic {
		return nil, fmt.Errorf("unexpected event signature: got %s, expected %s",
			log.Topics[0].Hex(), expectedTopic.Hex())
	}

	if len(log.Topics) < 3 {
		return nil, fmt.Errorf("not enough topics for TaskSubmitted event")
	}

	operator := common.BytesToAddress(log.Topics[1].Bytes())
	bigIntValue := new(big.Int).SetBytes(log.Topics[2].Bytes())
	uint64Value := bigIntValue.Uint64()
	if uint64Value > uint64(^uint16(0)) {
		return nil, fmt.Errorf("taskDefinitionId value %d exceeds uint16 maximum", uint64Value)
	}
	taskDefinitionId := uint16(uint64Value)

	var unpacked struct {
		TaskNumber   uint32
		ProofOfTask  string
		Data         []byte
		AttestersIds []*big.Int
	}

	if err := AttestationCenterABI.UnpackIntoInterface(&unpacked, "TaskSubmitted", log.Data); err != nil {
		return nil, fmt.Errorf("failed to unpack event data: %v", err)
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

// ParseTaskRejected parses a TaskRejected event from the log
func ParseTaskRejected(log ethtypes.Log) (*TaskRejectedEvent, error) {
	expectedTopic := TaskRejectedEventSignature()
	if log.Topics[0] != expectedTopic {
		return nil, fmt.Errorf("unexpected event signature: got %s, expected %s",
			log.Topics[0].Hex(), expectedTopic.Hex())
	}

	if len(log.Topics) < 3 {
		return nil, fmt.Errorf("not enough topics for TaskRejected event")
	}

	operator := common.BytesToAddress(log.Topics[1].Bytes())
	bigIntValue := new(big.Int).SetBytes(log.Topics[2].Bytes())
	uint64Value := bigIntValue.Uint64()
	if uint64Value > uint64(^uint16(0)) {
		return nil, fmt.Errorf("taskDefinitionId value %d exceeds uint16 maximum", uint64Value)
	}
	taskDefinitionId := uint16(uint64Value)

	var unpacked struct {
		TaskNumber   uint32
		ProofOfTask  string
		Data         []byte
		AttestersIds []*big.Int
	}

	if err := AttestationCenterABI.UnpackIntoInterface(&unpacked, "TaskRejected", log.Data); err != nil {
		return nil, fmt.Errorf("failed to unpack event data: %v", err)
	}

	return &TaskRejectedEvent{
		Operator:         operator,
		TaskNumber:       unpacked.TaskNumber,
		ProofOfTask:      unpacked.ProofOfTask,
		Data:             unpacked.Data,
		TaskDefinitionId: taskDefinitionId,
		AttestersIds:     unpacked.AttestersIds,
		Raw:              log,
	}, nil
}

func ConvertBigIntToStrings(bigInts []*big.Int) []string {
	strings := make([]string, len(bigInts))
	for i, bigInt := range bigInts {
		strings[i] = bigInt.String()
	}
	return strings
}
