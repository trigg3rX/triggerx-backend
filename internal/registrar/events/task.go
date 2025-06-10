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

			ipfsContent, err := FetchIPFSContent(config.GetIPFSHost(), dataCID, t.logger)
			if err != nil {
				t.logger.Errorf("Failed to fetch IPFS content: %v", err)
				continue
			}

			var ipfsData types.IPFSData
			if err := json.Unmarshal([]byte(ipfsContent), &ipfsData); err != nil {
				t.logger.Errorf("Failed to parse IPFS content into IPFSData: %v", err)
				continue
			}

			if err := client.UpdatePointsInDatabase(int(ipfsData.TaskData.TriggerData.TaskID), event.Operator, ConvertBigIntToStrings(event.AttestersIds), true); err != nil {
				t.logger.Errorf("Failed to update points in database: %v", err)
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

			ipfsContent, err := FetchIPFSContent(config.GetIPFSHost(), dataCID, t.logger)
			if err != nil {
				t.logger.Errorf("Failed to fetch IPFS content: %v", err)
				continue
			}

			var ipfsData types.IPFSData
			if err := json.Unmarshal([]byte(ipfsContent), &ipfsData); err != nil {
				t.logger.Errorf("Failed to parse IPFS content into IPFSData: %v", err)
				continue
			}

			if err := client.UpdatePointsInDatabase(int(ipfsData.TaskData.TriggerData.TaskID), event.Operator, ConvertBigIntToStrings(event.AttestersIds), false); err != nil {
				t.logger.Errorf("Failed to update points in database: %v", err)
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

func FetchIPFSContent(gateway string, cid string, logger logging.Logger) (string, error) {
	if strings.HasPrefix(cid, "https://") {
		resp, err := http.Get(cid)
		if err != nil {
			return "", fmt.Errorf("failed to fetch IPFS content from URL: %v", err)
		}
		defer func() {
			if err := resp.Body.Close(); err != nil {
				logger.Error("Error closing response body", "error", err)
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
			logger.Error("Error closing response body", "error", err)
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

// Pinata v3 API structures
type PinataFile struct {
	ID        string            `json:"id"`
	Name      string            `json:"name"`
	CID       string            `json:"cid"`
	Size      int64             `json:"size"`
	MimeType  string            `json:"mime_type"`
	GroupID   string            `json:"group_id,omitempty"`
	Keyvalues map[string]string `json:"keyvalues,omitempty"`
	CreatedAt time.Time         `json:"created_at"`
}

type PinataListResponse struct {
	Data struct {
		Files         []PinataFile `json:"files"`
		NextPageToken string       `json:"next_page_token,omitempty"`
	} `json:"data"`
}

type PinataDeleteResponse struct {
	Data interface{} `json:"data"`
}

// Helper function to make authenticated requests to Pinata
func makePinataRequest(method, url string, body io.Reader, logger logging.Logger) (*http.Response, error) {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+config.GetPinataJWT())
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	return resp, nil
}

// Find file ID by CID using Pinata v3 API
func findPinataFileIDByCID(cid string, logger logging.Logger) (string, error) {
	network := config.GetPinataHost() // Use configurable network from config
	url := fmt.Sprintf("https://api.pinata.cloud/v3/files/%s?cid=%s&limit=1", network, cid)

	resp, err := makePinataRequest("GET", url, nil, logger)
	if err != nil {
		return "", fmt.Errorf("failed to search for CID %s: %w", cid, err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			logger.Debugf("failed to close response body: %v", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("failed to search for CID %s: status %d, body: %s",
			cid, resp.StatusCode, string(body))
	}

	var listResp PinataListResponse
	if err := json.NewDecoder(resp.Body).Decode(&listResp); err != nil {
		return "", fmt.Errorf("failed to decode search response for CID %s: %w", cid, err)
	}

	if len(listResp.Data.Files) == 0 {
		return "", fmt.Errorf("no file found with CID %s", cid)
	}

	return listResp.Data.Files[0].ID, nil
}

// Delete file by ID using Pinata v3 API
func deletePinataFileByID(fileID string, logger logging.Logger) error {
	network := config.GetPinataHost() // Use configurable network from config
	url := fmt.Sprintf("https://api.pinata.cloud/v3/files/%s/%s", network, fileID)

	resp, err := makePinataRequest("DELETE", url, nil, logger)
	if err != nil {
		return fmt.Errorf("failed to delete file %s: %w", fileID, err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			logger.Debugf("failed to close response body: %v", err)
		}
	}()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
		return fmt.Errorf("failed to delete file %s: status %d, body: %s",
			fileID, resp.StatusCode, string(body))
	}

	var deleteResp PinataDeleteResponse
	if err := json.Unmarshal(body, &deleteResp); err != nil {
		// If we can't parse the response but got a success status, that's still OK
		logger.Debugf("Could not parse delete response for file %s, but status was %d", fileID, resp.StatusCode)
	}

	logger.Infof("Successfully deleted file %s from Pinata", fileID)
	return nil
}

// Delete file by CID (finds ID first, then deletes)
func DeletePinataCID(cid string, logger logging.Logger) error {
	fileID, err := findPinataFileIDByCID(cid, logger)
	if err != nil {
		return fmt.Errorf("failed to find file ID for CID %s: %w", cid, err)
	}

	return deletePinataFileByID(fileID, logger)
}

// List all files from Pinata v3 API
func listPinataFiles(logger logging.Logger) ([]PinataFile, error) {
	network := config.GetPinataHost() // Use configurable network from config
	url := fmt.Sprintf("https://api.pinata.cloud/v3/files/%s?limit=1000", network)

	var allFiles []PinataFile
	nextPageToken := ""

	for {
		requestURL := url
		if nextPageToken != "" {
			requestURL = fmt.Sprintf("%s&pageToken=%s", url, nextPageToken)
		}

		resp, err := makePinataRequest("GET", requestURL, nil, logger)
		if err != nil {
			return nil, fmt.Errorf("failed to list files: %w", err)
		}

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			if err := resp.Body.Close(); err != nil {
				logger.Debugf("failed to close response body: %v", err)
			}
			return nil, fmt.Errorf("failed to list files: status %d, body: %s",
				resp.StatusCode, string(body))
		}

		var listResp PinataListResponse
		if err := json.NewDecoder(resp.Body).Decode(&listResp); err != nil {
			if err := resp.Body.Close(); err != nil {
				logger.Debugf("failed to close response body: %v", err)
			}
			return nil, fmt.Errorf("failed to decode list response: %w", err)
		}
		if err := resp.Body.Close(); err != nil {
			logger.Debugf("failed to close response body: %v", err)
		}

		allFiles = append(allFiles, listResp.Data.Files...)

		// Check if there are more pages
		if listResp.Data.NextPageToken == "" {
			break
		}
		nextPageToken = listResp.Data.NextPageToken
	}

	return allFiles, nil
}

// Schedule a 24h delayed deletion for a CID
func scheduleCIDDeletion(cid string, logger logging.Logger) {
	go func() {
		logger.Infof("Scheduled deletion for CID %s in 24h", cid)
		time.Sleep(24 * time.Hour)

		if err := DeletePinataCID(cid, logger); err != nil {
			logger.Errorf("Failed to delete CID %s after 24h: %v", cid, err)
		} else {
			logger.Infof("Successfully deleted CID %s after 24h delay", cid)
		}
	}()
}

// Weekly cleanup: delete files older than 1 day every Sunday
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

			logger.Info("Starting weekly Pinata cleanup: deleting files older than 1 day...")
			files, err := listPinataFiles(logger)
			if err != nil {
				logger.Errorf("Failed to list Pinata files: %v", err)
				continue
			}

			cutoff := time.Now().Add(-24 * time.Hour)
			deletedCount := 0

			for _, file := range files {
				if file.CreatedAt.Before(cutoff) {
					logger.Infof("Deleting old file %s (CID: %s) created at %s",
						file.ID, file.CID, file.CreatedAt)

					if err := deletePinataFileByID(file.ID, logger); err != nil {
						logger.Errorf("Failed to delete old file %s: %v", file.ID, err)
					} else {
						deletedCount++
					}

					// Add small delay between deletions to avoid rate limiting
					time.Sleep(100 * time.Millisecond)
				}
			}

			logger.Infof("Weekly cleanup completed. Deleted %d old files.", deletedCount)
		}
	}()
}
