package registrar

import (
	// "bytes"
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	// "io"
	"math/big"
	// "net/http"
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

	// "github.com/gocql/gocql"

	"github.com/trigg3rX/triggerx-backend/internal/registrar/config"
	"github.com/trigg3rX/triggerx-backend/pkg/bindings/contractAttestationCenter"
	"github.com/trigg3rX/triggerx-backend/pkg/database"
	"github.com/trigg3rX/triggerx-backend/pkg/ipfs"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

var (
	// dbMain   *gocql.Session
	// db       *gocql.Session
	db       *database.Connection
	dbLogger = logging.GetLogger(logging.Development, logging.DatabaseProcess)
	logger   = logging.GetLogger(logging.Development, logging.RegistrarProcess)

	ethClient  *ethclient.Client
	baseClient *ethclient.Client

	lastProcessedBlockEth  uint64
	lastProcessedBlockBase uint64
	blockProcessingMutex   sync.Mutex
)

type OperatorRegisteredEvent struct {
	Operator common.Address
	BlsKey   [4]*big.Int
	Raw      ethtypes.Log
}

type OperatorUnregisteredEvent struct {
	Operator common.Address
	Raw      ethtypes.Log
}

type TaskSubmittedEvent struct {
	Operator         common.Address
	TaskNumber       uint32
	ProofOfTask      string
	Data             []byte
	TaskDefinitionId uint16
	AttestersIds     []string
	Raw              ethtypes.Log
}

type TaskRejectedEvent struct {
	Operator         common.Address
	TaskNumber       uint32
	ProofOfTask      string
	Data             []byte
	TaskDefinitionId uint16
	AttestersIds     []string
	Raw              ethtypes.Log
}

// SetDatabaseConnection sets both database connections for the registrar package
// func SetDatabaseConnection(mainSession *gocql.Session, registrarSession *gocql.Session) {
// 	if mainSession == nil || registrarSession == nil {
// 		dbLogger.Fatal("Database sessions cannot be nil")
// 		return
// 	}

// 	// Close existing connections before reassigning
// 	if dbMain != nil {
// 		dbMain.Close()
// 	}
// 	dbMain = mainSession

// 	if db != nil {
// 		db.Close()
// 	}
// 	db = registrarSession

// 	dbLogger.Info("Database connections set for registrar package")
// }

// StartEventPolling begins the event polling process
func StartEventPolling(
	avsGovernanceAddress common.Address,
	attestationCenterAddress common.Address,
) {
	// Create clients (non-WebSocket)
	var err error
	ethClient, err = ethclient.Dial(config.EthRpcUrl)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to connect to Ethereum node: %v", err))
		return
	}
	logger.Info("Ethereum node connected")
	defer ethClient.Close()

	baseClient, err = ethclient.Dial(config.BaseRpcUrl)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to connect to Base node: %v", err))
		return
	}
	logger.Info("Base node connected")
	defer baseClient.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Get latest block numbers from each network
	ethLatestBlock, err := ethClient.BlockNumber(ctx)
	if err != nil {
		logger.Error(fmt.Sprintf("failed to get ETH latest block: %v", err))
	}

	baseLatestBlock, err := baseClient.BlockNumber(ctx)
	if err != nil {
		logger.Error(fmt.Sprintf("failed to get BASE latest block: %v", err))
	}

	// Initialize with current block numbers (could also load from database if you want to restore from a previous state)
	lastProcessedBlockEth = ethLatestBlock
	lastProcessedBlockBase = baseLatestBlock

	logger.Info("Starting event polling service...")

	// Create ticker for 20-minute interval
	ticker := time.NewTicker(20 * time.Second)
	defer ticker.Stop()

	dbConfig := &database.Config{
		Hosts:       []string{"localhost:9042"},
		Timeout:     time.Second * 30,
		Retries:     3,
		ConnectWait: time.Second * 20,
	}
	db, err = database.NewConnection(dbConfig)
	if err != nil {
		logger.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

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

	// Get current block numbers
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Get latest blocks from each network
	ethLatestBlock, err := ethClient.BlockNumber(ctx)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to get ETH latest block: %v", err))
		return
	}
	//logger.Info("Ethereum Last Block: %d", ethLatestBlock)

	baseLatestBlock, err := baseClient.BlockNumber(ctx)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to get BASE latest block: %v", err))
		return
	}
	//logger.Info("Base Last Block: %d", baseLatestBlock)

	// Process events from AVS Governance contract
	if ethLatestBlock > lastProcessedBlockEth {
		fromBlock := lastProcessedBlockEth + 1
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
		lastProcessedBlockEth = ethLatestBlock
		blockProcessingMutex.Unlock()
	}

	// Process events from Attestation Center contract
	if baseLatestBlock > lastProcessedBlockBase {
		fromBlock := lastProcessedBlockBase
		overlap := uint64(5)
		if fromBlock > overlap {
			fromBlock -= overlap
		}

		// Ensure we don't go past current block
		toBlock := baseLatestBlock

		logger.Info(fmt.Sprintf("Checking for Attestation Center events from block %d to %d", fromBlock, toBlock))

		// Process task submission events
		err = processTaskSubmittedEvents(baseClient, attestationCenterAddress, fromBlock, baseLatestBlock)
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

		// Sleep for 5 seconds to allow for network propagation
		logger.Info("Sleeping for 5 seconds to allow for network propagation...")
		time.Sleep(5 * time.Second)
		logger.Info("Resuming operation after sleep")
		// Get operator ID from the contract
		attestationCenter, err := contractAttestationCenter.NewAttestationCenter(contractAddress, ethClient)
		if err != nil {
			logger.Error(fmt.Sprintf("Failed to create AttestationCenter instance: %v", err))
		} else {
			operatorId, err := attestationCenter.OperatorsIdsByAddress(nil, event.Operator)
			if err != nil {
				logger.Error(fmt.Sprintf("Failed to get operator ID: %v", err))
			} else {
				logger.Info(fmt.Sprintf("Operator ID: %s", operatorId.String()))
			}
		}

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

	var currentKeeperID int64
	if err := db.Session().Query(`
		SELECT keeper_id FROM triggerx.keeper_data WHERE keeper_address = ? ALLOW FILTERING`,
		operatorAddress).Scan(&currentKeeperID); err != nil {
		dbLogger.Errorf("[CreateKeeperData] Error getting max keeper ID: %v", err)
		return err
	}

	dbLogger.Infof("[CreateKeeperData] Updating keeper with ID: %d", currentKeeperID)
	if err := db.Session().Query(`
        UPDATE triggerx.keeper_data SET 
            status = ?
        WHERE keeper_id = ?`,
		true, currentKeeperID).Exec(); err != nil {
		dbLogger.Errorf("[CreateKeeperData] Error creating keeper with ID %d: %v", currentKeeperID, err)
		return err
	}
	
	logger.Info(fmt.Sprintf("Successfully updated keeper %s status to unregistered", operatorAddress))
	return nil
}

// Process TaskSubmitted events
func processTaskSubmittedEvents(
	baseClient *ethclient.Client,
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
			{TaskSubmittedEventSignature()},
			nil, // For operator (indexed)
			nil, // For taskDefinitionId (indexed)
		},
	}

	logger.Info(fmt.Sprintf("Filter query: FromBlock=%d, ToBlock=%d, Address=%s, Topics=%v",
		query.FromBlock, query.ToBlock, contractAddress.Hex(), query.Topics))
	logs, err := baseClient.FilterLogs(context.Background(), query)
	if err != nil {
		return fmt.Errorf("failed to filter TaskSubmitted logs: %v", err)
	}

	logger.Info(fmt.Sprintf("Found %d raw logs", len(logs)))
	for i, log := range logs {
		logger.Info(fmt.Sprintf("Log %d: Topics=%v, Data=%x", i, log.Topics, log.Data))
	}

	// Process each event
	for _, vLog := range logs {
		event, err := ParseTaskSubmitted(vLog)
		if err != nil {
			logger.Error(fmt.Sprintf("Failed to parse TaskSubmitted event: %v", err))
			continue
		}

		logger.Infof("Performer Address: %s", event.Operator)
		logger.Infof("Attester IDs: %v", event.AttestersIds)
		logger.Infof("Task Number: %d", event.TaskNumber)
		logger.Infof("Task Definition ID: %d", event.TaskDefinitionId)
		// logger.Infof("Data: %x", event.Data)

		decodedData := string(event.Data)
		logger.Infof("Decoded Data: %s", decodedData)

		ipfsContent, err := ipfs.FetchIPFSContent(decodedData)
		if err != nil {
			logger.Error(fmt.Sprintf("Failed to fetch IPFS content: %v", err))
			continue
		}

		var ipfsData types.IPFSData
		if err := json.Unmarshal([]byte(ipfsContent), &ipfsData); err != nil {
			logger.Errorf("Failed to parse IPFS content into IPFSData: %v", err)
			continue
		}

		// logger.Infof("IPFS Data: %+v", ipfsData)

		// Update points in database
		if err := updatePointsInDatabase(int(ipfsData.TriggerData.TaskID), event.Operator, event.AttestersIds); err != nil {
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

	logger.Info(fmt.Sprintf("Filter query: FromBlock=%d, ToBlock=%d, Address=%s, Topics=%v",
		query.FromBlock, query.ToBlock, contractAddress.Hex(), query.Topics))
	logs, err := baseClient.FilterLogs(context.Background(), query)
	if err != nil {
		return fmt.Errorf("failed to filter TaskRejected logs: %v", err)
	}

	// logger.Info(fmt.Sprintf("Found %d raw logs", len(logs)))
	// for i, log := range logs {
	// 	logger.Info(fmt.Sprintf("Log %d: Topics=%v, Data=%x", i, log.Topics, log.Data))
	// }

	// Process each event
	for _, vLog := range logs {
		event, err := ParseTaskRejected(vLog)
		if err != nil {
			logger.Error(fmt.Sprintf("Failed to parse TaskRejected event: %v", err))
			continue
		}

		// In processTaskRejectedEvents function, after parsing the event:
		// In processTaskRejectedEvents function:
		logger.Info("❌ Task Rejected Event Detected!")
		logger.Info(fmt.Sprintf("Operator Address: %s", event.Operator.Hex()))
		logger.Info(fmt.Sprintf("Task Number: %d", event.TaskNumber))
		logger.Info(fmt.Sprintf("Proof of Task: %s", event.ProofOfTask))
		logger.Info(fmt.Sprintf("Task Definition ID: %d", event.TaskDefinitionId))
		logger.Info(fmt.Sprintf("Attesters IDs: %v", event.AttestersIds)) // Now this will work

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
	return crypto.Keccak256Hash([]byte("TaskSubmitted(address,uint32,string,bytes,uint16,uint256[])"))
}

func TaskRejectedEventSignature() common.Hash {
	return crypto.Keccak256Hash([]byte("TaskRejected(address,uint32,string,bytes,uint16,uint256[])"))

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
		AttestersIds:     attestersIds,
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
		AttestersIds:     attestersIds, // Include the attesters IDs
		Raw:              log,
	}, nil
}

// Function to add a keeper to the database via the API when an operator is registered
func addKeeperToDatabase(operatorAddress string, blsKeysArray [4]*big.Int, txHash string) error {
	logger.Info(fmt.Sprintf("Adding operator %s to database as keeper", operatorAddress))

	// Convert BLS keys to string array for database
	blsKeys := make([]string, 4)
	for i, key := range blsKeysArray {
		blsKeys[i] = key.String()
	}

	var currentKeeperID int64
	if err := db.Session().Query(`
		SELECT keeper_id FROM triggerx.keeper_data WHERE keeper_address = ? ALLOW FILTERING`,
		operatorAddress).Scan(&currentKeeperID); err != nil {
		dbLogger.Errorf("[CreateKeeperData] Error getting max keeper ID: %v", err)
		return err
	}

	dbLogger.Infof("[CreateKeeperData] Updating keeper with ID: %d", currentKeeperID)
	if err := db.Session().Query(`
        UPDATE triggerx.keeper_data SET 
            registered_tx = ?, consensus_keys = ?, status = ?
        WHERE keeper_id = ?`,
		txHash, blsKeys, true, currentKeeperID).Exec(); err != nil {
		dbLogger.Errorf("[CreateKeeperData] Error creating keeper with ID %d: %v", currentKeeperID, err)
		return err
	}

	logger.Info(fmt.Sprintf("Successfully added keeper %s to database", operatorAddress))
	return nil
}

// New function to handle database updates
func updatePointsInDatabase(taskID int, performerAddress common.Address, attestersIds []string) error {
	if db == nil {
		logger.Error("Database connection is not initialized")
		return fmt.Errorf("database connection not initialized, please restart the service")
	}

	// Get task fee
	var taskFee int64
	err := db.Session().Query(`SELECT task_fee FROM triggerx.task_data WHERE task_id = ?`,
		taskID).Scan(&taskFee)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to get task fee for task ID %d: %v", taskID, err))
		return fmt.Errorf("failed to get task fee for task ID %d: %v", taskID, err)
	}

	logger.Info(fmt.Sprintf("Task ID %d has a fee of %d", taskID, taskFee))

	// Skip if performer address is empty
	if performerAddress != (common.Address{}) {
		if err := updatePerformerPoints(performerAddress.Hex(), taskFee); err != nil {
			return err
		}
	}

	// Process attesters
	for _, attesterId := range attestersIds {
		if attesterId != "" {
			if err := updateAttesterPoints(attesterId, taskFee); err != nil {
				logger.Error(fmt.Sprintf("Attester points update failed: %v", err))
				continue
			}
		}
	}

	return nil
}

// Helper function to update performer points
func updatePerformerPoints(performerAddress string, taskFee int64) error {
	var performerPoints int64
	var performerId int64

	// First, get the keeper_id using the keeper_address (requires a scan with ALLOW FILTERING)
	if err := db.Session().Query(`
		SELECT keeper_id, keeper_points FROM triggerx.keeper_data 
		WHERE keeper_address = ? ALLOW FILTERING`,
		performerAddress).Scan(&performerId, &performerPoints); err != nil {
		logger.Error(fmt.Sprintf("Failed to get performer ID and points: %v", err))
		return err
	}

	//multiplyer 2 till 7th April
	newPerformerPoints := performerPoints + 2*taskFee

	// Now update using the primary key (keeper_id)
	if err := db.Session().Query(`
		UPDATE triggerx.keeper_data 
		SET keeper_points = ? 
		WHERE keeper_id = ?`,
		newPerformerPoints, performerId).Exec(); err != nil {
		logger.Error(fmt.Sprintf("Failed to update performer points: %v", err))
		return err
	}

	logger.Info(fmt.Sprintf("Added %d points to performer %s (ID: %d)", taskFee, performerAddress, performerId))
	return nil
}

// Helper function to update attester points
func updateAttesterPoints(attesterId string, taskFee int64) error {
	var attesterPoints int64
	// Convert attesterId from string to int64
	var attesterIdInt int64
	_, err := fmt.Sscanf(attesterId, "%d", &attesterIdInt)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to convert attester ID %s to integer: %v", attesterId, err))
		return err
	}

	if err := db.Session().Query(`
		SELECT keeper_points FROM triggerx.keeper_data 
		WHERE keeper_id = ?`,
		attesterIdInt).Scan(&attesterPoints); err != nil {
		logger.Error(fmt.Sprintf("Failed to get attester points for ID %s: %v", attesterId, err))
		return err
	}

	//multiplyer 2 till 7th April
	newAttesterPoints := attesterPoints + 2*taskFee

	if err := db.Session().Query(`
		UPDATE triggerx.keeper_data 
		SET keeper_points = ? 
		WHERE keeper_id = ?`,
		newAttesterPoints, attesterIdInt).Exec(); err != nil {
		logger.Error(fmt.Sprintf("Failed to update attester points for ID %s: %v", attesterId, err))
		return err
	}

	logger.Info(fmt.Sprintf("Added %d points to attester ID %s", taskFee, attesterId))
	return nil
}
