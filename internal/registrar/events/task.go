package events

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/trigg3rX/triggerx-backend/internal/registrar/client"
	"github.com/trigg3rX/triggerx-backend/internal/registrar/config"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

// TaskProcessor handles task-related events
type TaskProcessor struct {
	*EventProcessor
}

// NewTaskProcessor creates a new task event processor
func NewTaskProcessor(base *EventProcessor) *TaskProcessor {
	if base == nil {
		panic("base processor cannot be nil")
	}
	return &TaskProcessor{
		EventProcessor: base,
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
			nil,
			nil,
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

			ipfsContent, err := FetchIPFSContent(config.GetIPFSHost(), dataCID)
			if err != nil {
				t.logger.Errorf("Failed to fetch IPFS content: %v", err)
				continue
			}

			var ipfsData types.IPFSData
			if err := json.Unmarshal([]byte(ipfsContent), &ipfsData); err != nil {
				t.logger.Errorf("Failed to parse IPFS content into IPFSData: %v", err)
				continue
			}

			if err := client.UpdatePointsInDatabase(int(ipfsData.TaskData.TriggerData[0].TaskID), event.Operator, ConvertBigIntToStrings(event.AttestersIds), true); err != nil {
				t.logger.Errorf("Failed to update points in database: %v", err)
				continue
			}

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

			ipfsContent, err := FetchIPFSContent(config.GetIPFSHost(), dataCID)
			if err != nil {
				t.logger.Errorf("Failed to fetch IPFS content: %v", err)
				continue
			}

			var ipfsData types.IPFSData
			if err := json.Unmarshal([]byte(ipfsContent), &ipfsData); err != nil {
				t.logger.Errorf("Failed to parse IPFS content into IPFSData: %v", err)
				continue
			}

			if err := client.UpdatePointsInDatabase(int(ipfsData.TaskData.TriggerData[0].TaskID), event.Operator, ConvertBigIntToStrings(event.AttestersIds), false); err != nil {
				t.logger.Errorf("Failed to update points in database: %v", err)
				continue
			}

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
	processor := NewTaskProcessor(NewEventProcessor(logger))
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
	processor := NewTaskProcessor(NewEventProcessor(logger))
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

func FetchIPFSContent(gateway string, cid string) (string, error) {
	if strings.HasPrefix(cid, "https://") {
		resp, err := http.Get(cid)
		if err != nil {
			return "", fmt.Errorf("failed to fetch IPFS content from URL: %v", err)
		}
		defer func() {
			if err := resp.Body.Close(); err != nil {
				// Log the error but don't return it since we're in a defer
				fmt.Printf("Error closing response body: %v\n", err)
			}
		}()

		if resp.StatusCode != http.StatusOK {
			return "", fmt.Errorf("failed to fetch IPFS content from URL: status code %d", resp.StatusCode)
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return "", fmt.Errorf("failed to read response body: %v", err)
		}

		return string(body), nil
	}

	ipfsGateway := "https://" + gateway + "/ipfs/" + cid
	resp, err := http.Get(ipfsGateway)
	if err != nil {
		return "", fmt.Errorf("failed to fetch IPFS content: %v", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			// Log the error but don't return it since we're in a defer
			fmt.Printf("Error closing response body: %v\n", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to fetch IPFS content: status code %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %v", err)
	}

	return string(body), nil
}

func ConvertBigIntToStrings(bigInts []*big.Int) []string {
	strings := make([]string, len(bigInts))
	for i, bigInt := range bigInts {
		strings[i] = bigInt.String()
	}
	return strings
}

// Pinata API helpers
func DeletePinataCID(cid string, logger logging.Logger) error {
	url := "https://api.pinata.cloud/pinning/unpin/" + cid
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create delete request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+config.GetPinataJWT())
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send delete request: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			// Log the error but don't return it since we're in a defer
			fmt.Printf("Error closing response body: %v\n", err)
		}
	}()
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted && resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to delete CID %s: status %d, body: %s", cid, resp.StatusCode, string(body))
	}
	logger.Infof("Deleted CID %s from Pinata", cid)
	return nil
}

// List pins from Pinata (returns a list of pin objects)
type pinataPin struct {
	IpfsHash   string `json:"ipfs_pin_hash"`
	DatePinned string `json:"date_pinned"`
}
type pinataListResponse struct {
	Rows []pinataPin `json:"rows"`
}

func listPinataPins(logger logging.Logger) ([]pinataPin, error) {
	url := "https://api.pinata.cloud/data/pinList?pageLimit=1000"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create list request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+config.GetPinataJWT())
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send list request: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			// Log the error but don't return it since we're in a defer
			fmt.Printf("Error closing response body: %v\n", err)
		}
	}()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to list pins: status %d, body: %s", resp.StatusCode, string(body))
	}
	var result pinataListResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode pinata list response: %w", err)
	}
	return result.Rows, nil
}

// Schedule a 24h delayed deletion for a CID
func scheduleCIDDeletion(cid string, logger logging.Logger) {
	go func() {
		logger.Infof("Scheduled deletion for CID %s in 24h", cid)
		time.Sleep(24 * time.Hour)
		err := DeletePinataCID(cid, logger)
		if err != nil {
			logger.Errorf("Failed to delete CID %s after 24h: %v", cid, err)
		}
	}()
}

// Weekly cleanup: delete pins older than 1 day every Sunday
func StartWeeklyPinataCleanup(logger logging.Logger) {
	go func() {
		for {
			now := time.Now().UTC()
			// Calculate next Sunday 00:05 UTC
			daysUntilSunday := (7 - int(now.Weekday())) % 7
			if daysUntilSunday == 0 && now.Hour() >= 0 && now.Minute() >= 5 {
				daysUntilSunday = 7
			}
			nextSunday := time.Date(now.Year(), now.Month(), now.Day(), 0, 5, 0, 0, time.UTC).AddDate(0, 0, daysUntilSunday)
			wait := nextSunday.Sub(now)
			logger.Infof("Weekly Pinata cleanup scheduled for: %v (in %v)", nextSunday, wait)
			time.Sleep(wait)

			logger.Info("Starting weekly Pinata cleanup: deleting pins older than 1 day...")
			pins, err := listPinataPins(logger)
			if err != nil {
				logger.Errorf("Failed to list Pinata pins: %v", err)
				continue
			}
			cutoff := time.Now().Add(-24 * time.Hour)
			for _, pin := range pins {
				pinTime, err := time.Parse(time.RFC3339, pin.DatePinned)
				if err != nil {
					logger.Warnf("Could not parse pin date for CID %s: %v", pin.IpfsHash, err)
					continue
				}
				if pinTime.Before(cutoff) {
					logger.Infof("Deleting old CID %s pinned at %s", pin.IpfsHash, pin.DatePinned)
					if err := DeletePinataCID(pin.IpfsHash, logger); err != nil {
						logger.Errorf("Failed to delete old CID %s: %v", pin.IpfsHash, err)
					}
				}
			}
		}
	}()
}
