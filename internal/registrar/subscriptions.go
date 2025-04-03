package registrar

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"

	// "github.com/ethereum/go-ethereum/event"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/gocql/gocql"
	shell "github.com/ipfs/go-ipfs-api"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

var (
	dbMain   *gocql.Session
	db       *gocql.Session
	loggerdb = logging.GetLogger(logging.Development, logging.DatabaseProcess)
	//logger   = logging.GetLogger(logging.Development, logging.EventService)

	// Variables to track the last processed block for each contract
	lastProcessedBlockAVS  uint64
	lastProcessedBlockBase uint64
	blockProcessingMutex   sync.Mutex
	//EthRpcUrl                  string // Regular RPC URL (not WebSocket)
	//BaseRpcUrl                 string // Regular RPC URL (not WebSocket)
	//DatabaseIPAddress          string // Your database API URL
)

type OperatorRegisteredEvent struct {
	Operator common.Address
	BlsKey   [4]*big.Int
	Raw      ethtypes.Log // This contains metadata about the event
}

type OperatorUnregisteredEvent struct {
	Operator common.Address
	Raw      ethtypes.Log
}

type TaskSubmittedEvent struct {
	Operator         common.Address
	TaskNumber       *big.Int
	ProofOfTask      string
	Data             []byte
	TaskDefinitionId uint16
	Raw              ethtypes.Log
}

type TaskRejectedEvent struct {
	Operator         common.Address
	TaskNumber       *big.Int
	ProofOfTask      string
	Data             []byte
	TaskDefinitionId uint16
	Raw              ethtypes.Log
}

// Add this init function
func init() {
	// Remove any database operations from here
	// Only initialize other necessary components
	// ... other non-database initializations ...
}

// SetDatabaseConnection sets both database connections for the registrar package
func SetDatabaseConnection(mainSession *gocql.Session, registrarSession *gocql.Session) {
	if mainSession == nil || registrarSession == nil {
		loggerdb.Fatal("Database sessions cannot be nil")
		return
	}

	// Close existing connections before reassigning
	if dbMain != nil {
		dbMain.Close()
	}
	dbMain = mainSession

	if db != nil {
		db.Close()
	}
	db = registrarSession

	loggerdb.Info("Database connections set for registrar package")
}

// GetDatabaseConnection returns the current database session
func GetDatabaseConnection() *gocql.Session {
	return db
}

// Helper function to check database connection
func isDatabaseConnected() bool {
	return db != nil && !db.Closed()
}

// For any function that needs database access, add a check:
func someFunction() error {
	if !isDatabaseConnected() {
		return fmt.Errorf("database connection not initialized")
	}
	// Use db safely here
	return nil
}

func InitEventProcessing(ethClient *ethclient.Client, baseClient *ethclient.Client) error {
	var err error

	// Get current block numbers
	blockProcessingMutex.Lock()
	defer blockProcessingMutex.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Get latest block numbers from each network
	ethLatestBlock, err := ethClient.BlockNumber(ctx)
	if err != nil {
		return fmt.Errorf("failed to get ETH latest block: %v", err)
	}

	baseLatestBlock, err := baseClient.BlockNumber(ctx)
	if err != nil {
		return fmt.Errorf("failed to get BASE latest block: %v", err)
	}

	// Initialize with current block numbers (could also load from database if you want to restore from a previous state)
	lastProcessedBlockAVS = ethLatestBlock
	lastProcessedBlockBase = baseLatestBlock

	logger.Info(fmt.Sprintf("Initialized event processing from ETH block %d and BASE block %d", lastProcessedBlockAVS, lastProcessedBlockBase))
	return nil
}

// StartEventPolling begins the event polling process
func StartEventPolling(
	avsGovernanceAddress common.Address,
	attestationCenterAddress common.Address,
) {
	logger.Info("Starting event polling service...")

	// Create ticker for 20-minute interval
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	// Initial poll immediately on startup
	pollEvents(avsGovernanceAddress, attestationCenterAddress)

	// Then poll according to the ticker
	for range ticker.C {
		pollEvents(avsGovernanceAddress, attestationCenterAddress)
	}
}

func pollEvents(
	avsGovernanceAddress common.Address,
	attestationCenterAddress common.Address,
) {
	logger.Info("Polling for new events...")

	// Create clients (non-WebSocket)
	ethClient, err := ethclient.Dial(EthRpcUrl)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to connect to Ethereum node: %v", err))
		return
	}
	defer ethClient.Close()

	baseClient, err := ethclient.Dial(BaseRpcUrl)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to connect to Base node: %v", err))
		return
	}
	defer baseClient.Close()

	// Get current block numbers
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Get latest blocks from each network
	ethLatestBlock, err := ethClient.BlockNumber(ctx)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to get ETH latest block: %v", err))
		return
	}

	baseLatestBlock, err := baseClient.BlockNumber(ctx)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to get BASE latest block: %v", err))
		return
	}

	// Process events from AVS Governance contract
	if ethLatestBlock > lastProcessedBlockAVS {
		fromBlock := lastProcessedBlockAVS + 1
		logger.Info(fmt.Sprintf("Checking for AVS Governance events from block %d to %d", fromBlock, ethLatestBlock))

		// Process operator registration events
		err = processOperatorRegisteredEvents(ethClient, avsGovernanceAddress, fromBlock, ethLatestBlock)
		if err != nil {
			logger.Error(fmt.Sprintf("Failed to process OperatorRegistered events: %v", err))
		}

		// Process operator unregistration events
		err = processOperatorUnregisteredEvents(ethClient, avsGovernanceAddress, fromBlock, ethLatestBlock)
		if err != nil {
			logger.Error(fmt.Sprintf("Failed to process OperatorUnregistered events: %v", err))
		}

		// Update last processed block for AVS
		blockProcessingMutex.Lock()
		lastProcessedBlockAVS = ethLatestBlock
		blockProcessingMutex.Unlock()
	}

	// Process events from Attestation Center contract
	if baseLatestBlock > lastProcessedBlockBase {
		fromBlock := lastProcessedBlockBase + 1
		logger.Info(fmt.Sprintf("Checking for Attestation Center events from block %d to %d", fromBlock, baseLatestBlock))

		// Process task submission events
		err = processTaskSubmittedEvents(baseClient, ethClient, attestationCenterAddress, fromBlock, baseLatestBlock)
		if err != nil {
			logger.Error(fmt.Sprintf("Failed to process TaskSubmitted events: %v", err))
		}

		// Process task rejection events
		err = processTaskRejectedEvents(baseClient, attestationCenterAddress, fromBlock, baseLatestBlock)
		if err != nil {
			logger.Error(fmt.Sprintf("Failed to process TaskRejected events: %v", err))
		}

		// Update last processed block for Base
		blockProcessingMutex.Lock()
		lastProcessedBlockBase = baseLatestBlock
		blockProcessingMutex.Unlock()
	}

	logger.Info("Event polling completed")
}

func processOperatorRegisteredEvents(
	ethClient *ethclient.Client,
	contractAddress common.Address,
	fromBlock uint64,
	toBlock uint64,
) error {
	// Create filter query for OperatorRegistered events
	query := ethereum.FilterQuery{
		FromBlock: new(big.Int).SetUint64(fromBlock),
		ToBlock:   new(big.Int).SetUint64(toBlock),
		Addresses: []common.Address{contractAddress},
		Topics: [][]common.Hash{
			{OperatorRegisteredEventSignature()}, // Event signature for OperatorRegistered
		},
	}

	logs, err := ethClient.FilterLogs(context.Background(), query)
	if err != nil {
		return fmt.Errorf("failed to filter OperatorRegistered logs: %v", err)
	}

	logger.Info(fmt.Sprintf("Found %d OperatorRegistered events", len(logs)))

	// Process each event
	for _, vLog := range logs {
		event, err := ParseOperatorRegistered(vLog)
		if err != nil {
			logger.Error(fmt.Sprintf("Failed to parse OperatorRegistered event: %v", err))
			continue
		}

		// Log operator registration details
		logger.Info("🟢 Operator Registered Event Detected!")
		logger.Info(fmt.Sprintf("Operator Address: %s", event.Operator.Hex()))

		// Log BLS Key details
		blsKeyStr := "BLS Key: ["
		for i, num := range event.BlsKey {
			if i > 0 {
				blsKeyStr += ", "
			}
			blsKeyStr += num.String()
		}
		blsKeyStr += "]"
		logger.Info(blsKeyStr)

		// Log transaction details
		logger.Info(fmt.Sprintf("Transaction Hash: %s", event.Raw.TxHash.Hex()))
		logger.Info(fmt.Sprintf("Block Number: %d", event.Raw.BlockNumber))

		// Add keeper to database
		err = addKeeperToDatabase(event.Operator.Hex(), event.BlsKey, event.Raw.TxHash.Hex())
		if err != nil {
			logger.Error(fmt.Sprintf("Failed to add keeper to database: %v", err))
		}
	}

	return nil
}

// Process OperatorUnregistered events
func processOperatorUnregisteredEvents(
	ethClient *ethclient.Client,
	contractAddress common.Address,
	fromBlock uint64,
	toBlock uint64,
) error {
	// Create filter query for OperatorUnregistered events
	query := ethereum.FilterQuery{
		FromBlock: new(big.Int).SetUint64(fromBlock),
		ToBlock:   new(big.Int).SetUint64(toBlock),
		Addresses: []common.Address{contractAddress},
		Topics: [][]common.Hash{
			{OperatorUnregisteredEventSignature()}, // Event signature for OperatorUnregistered
		},
	}

	logs, err := ethClient.FilterLogs(context.Background(), query)
	if err != nil {
		return fmt.Errorf("failed to filter OperatorUnregistered logs: %v", err)
	}

	logger.Info(fmt.Sprintf("Found %d OperatorUnregistered events", len(logs)))

	// Process each event
	for _, vLog := range logs {
		event, err := ParseOperatorUnregistered(vLog)
		if err != nil {
			logger.Error(fmt.Sprintf("Failed to parse OperatorUnregistered event: %v", err))
			continue
		}

		// Log operator unregistration details
		logger.Info("🔴 Operator Unregistered Event Detected!")
		logger.Info(fmt.Sprintf("Operator Address: %s", event.Operator.Hex()))

		// Log transaction details
		logger.Info(fmt.Sprintf("Transaction Hash: %s", event.Raw.TxHash.Hex()))
		logger.Info(fmt.Sprintf("Block Number: %d", event.Raw.BlockNumber))

		// Update keeper status in database
		err = updateKeeperStatusAsUnregistered(event.Operator.Hex())
		if err != nil {
			logger.Error(fmt.Sprintf("Failed to update keeper status in database: %v", err))
		}
	}

	return nil
}

// Function to update keeper status as unregistered in the database via API
func updateKeeperStatusAsUnregistered(operatorAddress string) error {
	logger.Info(fmt.Sprintf("Updating operator %s status to unregistered in database", operatorAddress))

	// Create the request payload
	updateData := map[string]interface{}{
		"status": false,
	}

	// Convert to JSON
	jsonData, err := json.Marshal(updateData)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to marshal update data: %v", err))
		return err
	}

	// Create the HTTP request
	dbServerURL := fmt.Sprintf("%s/api/keepers/%s/status", DatabaseIPAddress, operatorAddress)
	req, err := http.NewRequest("PUT", dbServerURL, bytes.NewBuffer(jsonData))
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to create HTTP request: %v", err))
		return err
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")

	// Send the request
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to send HTTP request: %v", err))
		return err
	}
	defer resp.Body.Close()

	// Read the response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to read response body: %v", err))
		return err
	}

	// Check the response status
	if resp.StatusCode != http.StatusOK {
		logger.Error(fmt.Sprintf("Failed to update keeper status. Status: %d, Response: %s", resp.StatusCode, string(body)))
		return fmt.Errorf("failed to update keeper status: %s", string(body))
	}

	logger.Info(fmt.Sprintf("Successfully updated keeper %s status to unregistered", operatorAddress))
	return nil
}

// Process TaskSubmitted events
func processTaskSubmittedEvents(
	baseClient *ethclient.Client,
	ethClient *ethclient.Client,
	contractAddress common.Address,
	fromBlock uint64,
	toBlock uint64,
) error {
	// Create filter query for TaskSubmitted events
	query := ethereum.FilterQuery{
		FromBlock: new(big.Int).SetUint64(fromBlock),
		ToBlock:   new(big.Int).SetUint64(toBlock),
		Addresses: []common.Address{contractAddress},
		Topics: [][]common.Hash{
			{TaskSubmittedEventSignature()}, // Event signature for TaskSubmitted
		},
	}

	logs, err := baseClient.FilterLogs(context.Background(), query)
	if err != nil {
		return fmt.Errorf("failed to filter TaskSubmitted logs: %v", err)
	}

	logger.Info(fmt.Sprintf("Found %d TaskSubmitted events", len(logs)))

	// Process each event
	for _, vLog := range logs {
		event, err := ParseTaskSubmitted(vLog)
		if err != nil {
			logger.Error(fmt.Sprintf("Failed to parse TaskSubmitted event: %v", err))
			continue
		}

		// Log task submission details
		logger.Info("📥 Task Submitted Event Detected!")
		logger.Info(fmt.Sprintf("Operator Address: %s", event.Operator.Hex()))
		logger.Info(fmt.Sprintf("Task Number: %d", event.TaskNumber))
		logger.Info(fmt.Sprintf("Proof of Task: %s", event.ProofOfTask))
		logger.Info(fmt.Sprintf("Task Definition ID: %d", event.TaskDefinitionId))

		// Log data as hex if non-empty
		if len(event.Data) > 0 {
			logger.Info(fmt.Sprintf("Data: 0x%x", event.Data))
		} else {
			logger.Info("Data: <empty>")
		}

		// Log transaction details
		logger.Info(fmt.Sprintf("Transaction Hash: %s", event.Raw.TxHash.Hex()))
		logger.Info(fmt.Sprintf("Block Number: %d", event.Raw.BlockNumber))

		// Fetch the transaction to decode input data
		tx, isPending, err := baseClient.TransactionByHash(context.Background(), event.Raw.TxHash)
		if err != nil {
			logger.Error(fmt.Sprintf("Failed to fetch transaction: %v", err))
			continue
		}
		if isPending {
			logger.Info("Transaction is pending.")
			continue
		}

		// Decode the input data to extract performer address and operator IDs
		performerAddress, proofOfTask, data, taskDefinitionId, isApproved, tpSignature, taSignature, attestersIds, err := decodeTaskInputData(tx.Data())
		if err != nil {
			logger.Error(fmt.Sprintf("Failed to decode input data: %v", err))
			continue
		}

		// Fetch the task information from IPFS using the CID
		cid := string(data) // Assuming data is the CID
		IPFSData, err := fetchDataFromCID(cid)
		if err != nil {
			logger.Error(fmt.Sprintf("Failed to fetch task data from CID: %v", err))
			continue
		}

		// Unmarshal the IPFS data into the IPFSData struct
		var ipfsData types.IPFSData
		err = json.Unmarshal([]byte(IPFSData), &ipfsData)
		if err != nil {
			logger.Error(fmt.Sprintf("Failed to unmarshal IPFS data: %v", err))
			continue
		}

		// Now you can access the TaskID from the TriggerData
		taskID := ipfsData.TriggerData.TaskID
		logger.Info(fmt.Sprintf("Task ID: %d", taskID))
		logger.Info(fmt.Sprintf("Performer Address: %s", performerAddress.Hex()))
		logger.Info(fmt.Sprintf("Proof of Task: %s", proofOfTask))
		logger.Info(fmt.Sprintf("Data: %x", data))
		logger.Info(fmt.Sprintf("Task Definition ID: %d", taskDefinitionId))
		logger.Info(fmt.Sprintf("Is Approved: %v", isApproved))
		logger.Info(fmt.Sprintf("TP Signature: %x", tpSignature))
		logger.Info(fmt.Sprintf("TA Signature: %x", taSignature))
		logger.Info(fmt.Sprintf("Attesters IDs: %v", attestersIds))

		// Update points in database
		if err := updatePointsInDatabase(taskID, performerAddress, attestersIds); err != nil {
			logger.Error(fmt.Sprintf("Failed to update points in database: %v", err))
			continue
		}
	}

	return nil
}

// Process TaskRejected events
func processTaskRejectedEvents(
	baseClient *ethclient.Client,
	contractAddress common.Address,
	fromBlock uint64,
	toBlock uint64,
) error {
	// Create filter query for TaskRejected events
	query := ethereum.FilterQuery{
		FromBlock: new(big.Int).SetUint64(fromBlock),
		ToBlock:   new(big.Int).SetUint64(toBlock),
		Addresses: []common.Address{contractAddress},
		Topics: [][]common.Hash{
			{TaskRejectedEventSignature()}, // Event signature for TaskRejected
		},
	}

	logs, err := baseClient.FilterLogs(context.Background(), query)
	if err != nil {
		return fmt.Errorf("failed to filter TaskRejected logs: %v", err)
	}

	logger.Info(fmt.Sprintf("Found %d TaskRejected events", len(logs)))

	// Process each event
	for _, vLog := range logs {
		event, err := ParseTaskRejected(vLog)
		if err != nil {
			logger.Error(fmt.Sprintf("Failed to parse TaskRejected event: %v", err))
			continue
		}

		// Log task rejection details
		logger.Info("❌ Task Rejected Event Detected!")
		logger.Info(fmt.Sprintf("Operator Address: %s", event.Operator.Hex()))
		logger.Info(fmt.Sprintf("Task Number: %d", event.TaskNumber))
		logger.Info(fmt.Sprintf("Proof of Task: %s", event.ProofOfTask))
		logger.Info(fmt.Sprintf("Task Definition ID: %d", event.TaskDefinitionId))

		// Log data as hex if non-empty
		if len(event.Data) > 0 {
			logger.Info(fmt.Sprintf("Data: 0x%x", event.Data))
		} else {
			logger.Info("Data: <empty>")
		}

		// Log transaction details
		logger.Info(fmt.Sprintf("Transaction Hash: %s", event.Raw.TxHash.Hex()))
		logger.Info(fmt.Sprintf("Block Number: %d", event.Raw.BlockNumber))
	}

	return nil
}

// Helper functions for getting event signatures

// OperatorRegisteredEventSignature returns the signature hash for the OperatorRegistered event
func OperatorRegisteredEventSignature() common.Hash {
	return crypto.Keccak256Hash([]byte("OperatorRegistered(address,uint256[4])"))
}

// OperatorUnregisteredEventSignature returns the signature hash for the OperatorUnregistered event
func OperatorUnregisteredEventSignature() common.Hash {
	return crypto.Keccak256Hash([]byte("OperatorUnregistered(address)"))
}

// TaskSubmittedEventSignature returns the signature hash for the TaskSubmitted event
func TaskSubmittedEventSignature() common.Hash {
	return crypto.Keccak256Hash([]byte("TaskSubmitted(address,uint256,string,bytes,uint16)"))
}

// TaskRejectedEventSignature returns the signature hash for the TaskRejected event
func TaskRejectedEventSignature() common.Hash {
	return crypto.Keccak256Hash([]byte("TaskRejected(address,uint256,string,bytes,uint16)"))
}

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
	// Define the event ABI
	eventSignature := "TaskSubmitted(address,uint256,string,bytes,uint16)"
	eventABI := fmt.Sprintf(`[{"name":"TaskSubmitted","type":"event","inputs":[
        {"name":"operator","type":"address","indexed":false},
        {"name":"taskNumber","type":"uint256","indexed":false},
        {"name":"proofOfTask","type":"string","indexed":false},
        {"name":"data","type":"bytes","indexed":false},
        {"name":"taskDefinitionId","type":"uint16","indexed":false}
    ]}]`)

	parsedABI, err := abi.JSON(strings.NewReader(eventABI))
	if err != nil {
		return nil, fmt.Errorf("failed to parse ABI: %v", err)
	}

	// Verify the event signature
	expectedTopic := crypto.Keccak256Hash([]byte(eventSignature))
	if log.Topics[0] != expectedTopic {
		return nil, fmt.Errorf("unexpected event signature")
	}

	// Prepare event result
	var event TaskSubmittedEvent
	event.Raw = log

	// Unpack the non-indexed fields
	unpacked := make(map[string]interface{})
	err = parsedABI.UnpackIntoMap(unpacked, "TaskSubmitted", log.Data)
	if err != nil {
		return nil, fmt.Errorf("failed to unpack into map: %v", err)
	}

	// Extract the fields from the unpacked data
	if operator, ok := unpacked["operator"].(common.Address); ok {
		event.Operator = operator
	} else {
		return nil, fmt.Errorf("failed to extract operator address")
	}

	if taskNumber, ok := unpacked["taskNumber"].(*big.Int); ok {
		event.TaskNumber = taskNumber
	} else {
		return nil, fmt.Errorf("failed to extract task number")
	}

	if proofOfTask, ok := unpacked["proofOfTask"].(string); ok {
		event.ProofOfTask = proofOfTask
	} else {
		return nil, fmt.Errorf("failed to extract proof of task")
	}

	if data, ok := unpacked["data"].([]byte); ok {
		event.Data = data
	} else {
		return nil, fmt.Errorf("failed to extract data")
	}

	if taskDefinitionId, ok := unpacked["taskDefinitionId"].(uint16); ok {
		event.TaskDefinitionId = taskDefinitionId
	} else {
		return nil, fmt.Errorf("failed to extract task definition ID")
	}

	return &event, nil
}

// ParseTaskRejected parses a log into the TaskRejected event
func ParseTaskRejected(log ethtypes.Log) (*TaskRejectedEvent, error) {
	// Define the event ABI
	eventSignature := "TaskRejected(address,uint256,string,bytes,uint16)"
	eventABI := fmt.Sprintf(`[{"name":"TaskRejected","type":"event","inputs":[
        {"name":"operator","type":"address","indexed":false},
        {"name":"taskNumber","type":"uint256","indexed":false},
        {"name":"proofOfTask","type":"string","indexed":false},
        {"name":"data","type":"bytes","indexed":false},
        {"name":"taskDefinitionId","type":"uint16","indexed":false}
    ]}]`)

	parsedABI, err := abi.JSON(strings.NewReader(eventABI))
	if err != nil {
		return nil, fmt.Errorf("failed to parse ABI: %v", err)
	}

	// Verify the event signature
	expectedTopic := crypto.Keccak256Hash([]byte(eventSignature))
	if log.Topics[0] != expectedTopic {
		return nil, fmt.Errorf("unexpected event signature")
	}

	// Prepare event result
	var event TaskRejectedEvent
	event.Raw = log

	// Unpack the non-indexed fields
	unpacked := make(map[string]interface{})
	err = parsedABI.UnpackIntoMap(unpacked, "TaskRejected", log.Data)
	if err != nil {
		return nil, fmt.Errorf("failed to unpack into map: %v", err)
	}

	// Extract the fields from the unpacked data
	if operator, ok := unpacked["operator"].(common.Address); ok {
		event.Operator = operator
	} else {
		return nil, fmt.Errorf("failed to extract operator address")
	}

	if taskNumber, ok := unpacked["taskNumber"].(*big.Int); ok {
		event.TaskNumber = taskNumber
	} else {
		return nil, fmt.Errorf("failed to extract task number")
	}

	if proofOfTask, ok := unpacked["proofOfTask"].(string); ok {
		event.ProofOfTask = proofOfTask
	} else {
		return nil, fmt.Errorf("failed to extract proof of task")
	}

	if data, ok := unpacked["data"].([]byte); ok {
		event.Data = data
	} else {
		return nil, fmt.Errorf("failed to extract data")
	}

	if taskDefinitionId, ok := unpacked["taskDefinitionId"].(uint16); ok {
		event.TaskDefinitionId = taskDefinitionId
	} else {
		return nil, fmt.Errorf("failed to extract task definition ID")
	}

	return &event, nil
}

// Function to add a keeper to the database via the API when an operator is registered
func addKeeperToDatabase(operatorAddress string, blsKeysArray [4]*big.Int, txHash string) error {
	logger.Info(fmt.Sprintf("Adding operator %s to database as keeper", operatorAddress))

	// Convert BLS keys to string array for database
	blsKeys := make([]string, 4)
	for i, key := range blsKeysArray {
		blsKeys[i] = key.String()
	}

	// Create the request payload
	keeperData := types.CreateKeeperData{
		KeeperAddress:  operatorAddress,
		RegisteredTx:   txHash,
		RewardsAddress: operatorAddress,
		ConsensusKeys:  blsKeys,
	}

	// Convert to JSON
	jsonData, err := json.Marshal(keeperData)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to marshal keeper data: %v", err))
		return err
	}

	// Create the HTTP request
	dbServerURL := fmt.Sprintf("%s/api/keepers", DatabaseIPAddress)
	req, err := http.NewRequest("POST", dbServerURL, bytes.NewBuffer(jsonData))
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to create HTTP request: %v", err))
		return err
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")

	// Send the request
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to send HTTP request: %v", err))
		return err
	}
	defer resp.Body.Close()

	// Read the response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to read response body: %v", err))
		return err
	}

	// Check the response status
	if resp.StatusCode != http.StatusOK {
		logger.Error(fmt.Sprintf("Failed to add keeper to database. Status: %d, Response: %s", resp.StatusCode, string(body)))
		return fmt.Errorf("failed to add keeper to database: %s", string(body))
	}

	logger.Info(fmt.Sprintf("Successfully added keeper %s to database", operatorAddress))
	return nil
}

// Function to decode the input data of the task submission
func decodeTaskInputData(data []byte) (common.Address, string, []byte, uint16, bool, [2]*big.Int, [2]*big.Int, []string, error) {
	// Define the ABI of the contract containing the submitTask function
	const submitTaskABI = `[{"inputs":[{"components":[{"internalType":"string","name":"proofOfTask","type":"string"},{"internalType":"bytes","name":"data","type":"bytes"},{"internalType":"address","name":"taskPerformer","type":"address"},{"internalType":"uint16","name":"taskDefinitionId","type":"uint16"}],"internalType":"tuple","name":"_taskInfo","type":"tuple"},{"components":[{"internalType":"bool","name":"isApproved","type":"bool"},{"internalType":"uint256[2]","name":"tpSignature","type":"uint256[2]"},{"internalType":"uint256[2]","name":"taSignature","type":"uint256[2]"},{"internalType":"string[]","name":"attestersIds","type":"string[]"}],"internalType":"tuple","name":"_blsTaskSubmissionDetails","type":"tuple"}],"name":"submitTask","outputs":[],"stateMutability":"nonpayable","type":"function"}]`

	// Parse the ABI
	parsedABI, err := abi.JSON(strings.NewReader(submitTaskABI))
	if err != nil {
		return common.Address{}, "", nil, 0, false, [2]*big.Int{}, [2]*big.Int{}, nil, fmt.Errorf("failed to parse ABI: %v", err)
	}

	// Variables to hold the decoded values
	var taskInfo struct {
		ProofOfTask      string
		Data             []byte
		TaskPerformer    common.Address
		TaskDefinitionId uint16
	}
	var blsTaskSubmissionDetails struct {
		IsApproved   bool
		TpSignature  [2]*big.Int
		TaSignature  [2]*big.Int
		AttestersIds []string
	}

	// Decode the input data
	err = parsedABI.UnpackIntoInterface(&struct {
		TaskInfo *struct {
			ProofOfTask      string
			Data             []byte
			TaskPerformer    common.Address
			TaskDefinitionId uint16
		}
		BlsTaskSubmissionDetails *struct {
			IsApproved   bool
			TpSignature  [2]*big.Int
			TaSignature  [2]*big.Int
			AttestersIds []string
		}
	}{&taskInfo, &blsTaskSubmissionDetails}, "submitTask", data)
	if err != nil {
		return common.Address{}, "", nil, 0, false, [2]*big.Int{}, [2]*big.Int{}, nil, fmt.Errorf("failed to unpack input data: %v", err)
	}

	return taskInfo.TaskPerformer, taskInfo.ProofOfTask, taskInfo.Data, taskInfo.TaskDefinitionId, blsTaskSubmissionDetails.IsApproved, blsTaskSubmissionDetails.TpSignature, blsTaskSubmissionDetails.TaSignature, blsTaskSubmissionDetails.AttestersIds, nil
}

// Function to fetch data from IPFS using CID
func fetchDataFromCID(cid string) ([]byte, error) {
	// Create a new IPFS shell instance
	sh := shell.NewShell("https://api.pinata.cloud/pinning/pinJSONToIPFS") // Adjust the address if your IPFS node is running elsewhere

	// Fetch the data from IPFS
	data, err := sh.Cat(cid)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch data from IPFS: %v", err)
	}
	defer data.Close()

	// Read the data into a byte slice
	body, err := io.ReadAll(data)
	if err != nil {
		return nil, fmt.Errorf("failed to read data from IPFS: %v", err)
	}

	return body, nil
}

// New function to handle database updates
func updatePointsInDatabase(taskID int64, performerAddress common.Address, attestersIds []string) error {
	// Get task fee from task_data table
	var taskFee int64
	if err := db.Query(`
		SELECT task_fee FROM triggerx.task_data WHERE task_id = ?`,
		taskID).Scan(&taskFee); err != nil {
		logger.Error(fmt.Sprintf("Failed to get task fee for task ID %d: %v", taskID, err))
		return err
	}

	// Update performer points
	if err := updatePerformerPoints(performerAddress.Hex(), taskFee); err != nil {
		return err
	}

	// Update attester points
	for _, attesterId := range attestersIds {
		if err := updateAttesterPoints(attesterId, taskFee); err != nil {
			logger.Error(fmt.Sprintf("Failed to update points for attester %s: %v", attesterId, err))
			continue
		}
	}

	return nil
}

// Helper function to update performer points
func updatePerformerPoints(performerAddress string, taskFee int64) error {
	var performerPoints int64
	if err := db.Query(`
		SELECT keeper_points FROM triggerx.keeper_data 
		WHERE keeper_address = ? ALLOW FILTERING`,
		performerAddress).Scan(&performerPoints); err != nil {
		logger.Error(fmt.Sprintf("Failed to get performer points: %v", err))
		return err
	}

	newPerformerPoints := performerPoints + taskFee

	if err := db.Query(`
		UPDATE triggerx.keeper_data 
		SET keeper_points = ? 
		WHERE keeper_address = ?`,
		newPerformerPoints, performerAddress).Exec(); err != nil {
		logger.Error(fmt.Sprintf("Failed to update performer points: %v", err))
		return err
	}

	logger.Info(fmt.Sprintf("Added %d points to performer %s", taskFee, performerAddress))
	return nil
}

// Helper function to update attester points
func updateAttesterPoints(attesterId string, taskFee int64) error {
	var attesterPoints int64
	if err := db.Query(`
		SELECT keeper_points FROM triggerx.keeper_data 
		WHERE keeper_id = ?`,
		attesterId).Scan(&attesterPoints); err != nil {
		logger.Error(fmt.Sprintf("Failed to get attester points for ID %s: %v", attesterId, err))
		return err
	}

	newAttesterPoints := attesterPoints + taskFee

	if err := db.Query(`
		UPDATE triggerx.keeper_data 
		SET keeper_points = ? 
		WHERE keeper_id = ?`,
		newAttesterPoints, attesterId).Exec(); err != nil {
		logger.Error(fmt.Sprintf("Failed to update attester points for ID %s: %v", attesterId, err))
		return err
	}

	logger.Info(fmt.Sprintf("Added %d points to attester ID %s", taskFee, attesterId))
	return nil
}
