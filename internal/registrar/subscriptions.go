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
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/event"
	"github.com/gocql/gocql"
	shell "github.com/ipfs/go-ipfs-api"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)


var (
	dbMain *gocql.Session
	db     *gocql.Session
	loggerdb = logging.GetLogger(logging.Development, logging.DatabaseProcess)
)

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

// Setup subscription for OperatorRegistered events
func SetupRegisteredSubscription(
	ethClient *ethclient.Client,
	contractAddress common.Address,
	operatorRegisteredCh chan<- *OperatorRegistered,
) (event.Subscription, error) {
	sub, err := SubscribeOperatorRegistered(ethClient, contractAddress, operatorRegisteredCh)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to watch for OperatorRegistered events: %v", err))
		return nil, err
	}
	logger.Info("Started watching for OperatorRegistered events")
	return sub, nil
}

// Setup subscription for OperatorUnregistered events
func SetupUnregisteredSubscription(
	ethClient *ethclient.Client,
	contractAddress common.Address,
	operatorUnregisteredCh chan<- *OperatorUnregistered,
) (event.Subscription, error) {
	sub, err := SubscribeOperatorUnregistered(ethClient, contractAddress, operatorUnregisteredCh)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to watch for OperatorUnregistered events: %v", err))
		return nil, err
	}
	logger.Info("Started watching for OperatorUnregistered events")
	return sub, nil
}

// Setup subscription for TaskSubmitted events
func SetupTaskSubmittedSubscription(
	ethClient *ethclient.Client,
	contractAddress common.Address,
	taskSubmittedCh chan<- *TaskSubmitted,
) (event.Subscription, error) {
	sub, err := SubscribeTaskSubmitted(ethClient, contractAddress, taskSubmittedCh)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to watch for TaskSubmitted events: %v", err))
		return nil, err
	}
	logger.Info("Started watching for TaskSubmitted events")
	return sub, nil
}

// Setup subscription for TaskRejected events
func SetupTaskRejectedSubscription(
	ethClient *ethclient.Client,
	contractAddress common.Address,
	taskRejectedCh chan<- *TaskRejected,
) (event.Subscription, error) {
	sub, err := SubscribeTaskRejected(ethClient, contractAddress, taskRejectedCh)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to watch for TaskRejected events: %v", err))
		return nil, err
	}
	logger.Info("Started watching for TaskRejected events")
	return sub, nil
}

// Manage all subscriptions and handle reconnection logic
func ManageSubscriptions(
	ethWsClient *ethclient.Client,
	avsGovernanceAddress common.Address,
	baseWsClient *ethclient.Client,
	attestationCenterAddress common.Address,
	operatorRegisteredCh chan *OperatorRegistered,
	operatorUnregisteredCh chan *OperatorUnregistered,
	taskSubmittedCh chan *TaskSubmitted,
	taskRejectedCh chan *TaskRejected,
	regSub event.Subscription,
	unregSub event.Subscription,
	taskSubSub event.Subscription,
	taskRejSub event.Subscription,
) {
	// Create channels for subscription errors
	regSubErrCh := make(chan error)
	unregSubErrCh := make(chan error)
	taskSubErrCh := make(chan error)
	taskRejErrCh := make(chan error)

	// Forward subscription errors to the error channels
	if regSub != nil {
		go forwardSubscriptionErrors(regSub, regSubErrCh)
	}
	if unregSub != nil {
		go forwardSubscriptionErrors(unregSub, unregSubErrCh)
	}
	if taskSubSub != nil {
		go forwardSubscriptionErrors(taskSubSub, taskSubErrCh)
	}
	if taskRejSub != nil {
		go forwardSubscriptionErrors(taskRejSub, taskRejErrCh)
	}

	// Keep track of which subscriptions are being reconnected
	regSubReconnecting := false
	unregSubReconnecting := false
	taskSubSubReconnecting := false
	taskRejSubReconnecting := false

	// Background task to restore subscriptions if they're nil
	go func() {
		ticker := time.NewTicker(3 * time.Second)
		defer ticker.Stop()

		for {
			<-ticker.C
			// Check and reconnect each subscription individually
			if regSub == nil && !regSubReconnecting {
				logger.Info("Attempting to restore OperatorRegistered subscription...")
				go reconnectRegisteredSubscription(
					ethWsClient,
					avsGovernanceAddress,
					operatorRegisteredCh,
					&regSub,
					&regSubReconnecting,
				)
			}

			if unregSub == nil && !unregSubReconnecting {
				logger.Info("Attempting to restore OperatorUnregistered subscription...")
				go reconnectUnregisteredSubscription(
					ethWsClient,
					avsGovernanceAddress,
					operatorUnregisteredCh,
					&unregSub,
					&unregSubReconnecting,
				)
			}

			if taskSubSub == nil && !taskSubSubReconnecting {
				logger.Info("Attempting to restore TaskSubmitted subscription...")
				go reconnectTaskSubmittedSubscription(
					baseWsClient,
					attestationCenterAddress,
					taskSubmittedCh,
					&taskSubSub,
					&taskSubSubReconnecting,
				)
			}

			if taskRejSub == nil && !taskRejSubReconnecting {
				logger.Info("Attempting to restore TaskRejected subscription...")
				go reconnectTaskRejectedSubscription(
					baseWsClient,
					attestationCenterAddress,
					taskRejectedCh,
					&taskRejSub,
					&taskRejSubReconnecting,
				)
			}
		}
	}()

	for {
		select {
		// Handle OperatorRegistered subscription errors
		case err := <-regSubErrCh:
			logger.Error(fmt.Sprintf("Error in OperatorRegistered subscription: %v", err))
			regSub = nil
			if !regSubReconnecting {
				go reconnectRegisteredSubscription(
					ethWsClient,
					avsGovernanceAddress,
					operatorRegisteredCh,
					&regSub,
					&regSubReconnecting,
				)
			}

		// Handle OperatorUnregistered subscription errors
		case err := <-unregSubErrCh:
			logger.Error(fmt.Sprintf("Error in OperatorUnregistered subscription: %v", err))
			unregSub = nil
			if !unregSubReconnecting {
				go reconnectUnregisteredSubscription(
					ethWsClient,
					avsGovernanceAddress,
					operatorUnregisteredCh,
					&unregSub,
					&unregSubReconnecting,
				)
			}

		// Handle TaskSubmitted subscription errors
		case err := <-taskSubErrCh:
			logger.Error(fmt.Sprintf("Error in TaskSubmitted subscription: %v", err))
			taskSubSub = nil
			if !taskSubSubReconnecting {
				go reconnectTaskSubmittedSubscription(
					baseWsClient,
					attestationCenterAddress,
					taskSubmittedCh,
					&taskSubSub,
					&taskSubSubReconnecting,
				)
			}

		// Handle TaskRejected subscription errors
		case err := <-taskRejErrCh:
			logger.Error(fmt.Sprintf("Error in TaskRejected subscription: %v", err))
			taskRejSub = nil
			if !taskRejSubReconnecting {
				go reconnectTaskRejectedSubscription(
					baseWsClient,
					attestationCenterAddress,
					taskRejectedCh,
					&taskRejSub,
					&taskRejSubReconnecting,
				)
			}

		// Handle OperatorRegistered events
		case event := <-operatorRegisteredCh:
			// Log operator registration details
			logger.Info("🟢 Operator Registered Event Detected!")
			logger.Info(fmt.Sprintf("Operator Address: %s", event.Operator.Hex()))

			// Log BLS Key details (array of 4 big integers)
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
			err := addKeeperToDatabase(event.Operator.Hex(), event.BlsKey, event.Raw.TxHash.Hex())
			if err != nil {
				logger.Error(fmt.Sprintf("Failed to add keeper to database: %v", err))
			}

		// Handle OperatorUnregistered events
		case event := <-operatorUnregisteredCh:
			// Log operator unregistration details
			logger.Info("🔴 Operator Unregistered Event Detected!")
			logger.Info(fmt.Sprintf("Operator Address: %s", event.Operator.Hex()))

			// Log transaction details
			logger.Info(fmt.Sprintf("Transaction Hash: %s", event.Raw.TxHash.Hex()))
			logger.Info(fmt.Sprintf("Block Number: %d", event.Raw.BlockNumber))

			// Update keeper status in database
			// err := updateKeeperStatusAsUnregistered(event.Operator.Hex())
			// if err != nil {
			// 	logger.Error(fmt.Sprintf("Failed to update keeper status in database: %v", err))
			// }

		// Handle TaskSubmitted events
		case event := <-taskSubmittedCh:
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

			///====================================================================

			// Log transaction details
			logger.Info(fmt.Sprintf("Transaction Hash: %s", event.Raw.TxHash.Hex()))
			logger.Info(fmt.Sprintf("Block Number: %d", event.Raw.BlockNumber))
			tx, isPending, err := ethWsClient.TransactionByHash(context.Background(), event.Raw.TxHash)
			if err != nil {
				logger.Error(fmt.Sprintf("Failed to fetch transaction: %v", err))
				return
			}
			if isPending {
				logger.Info("Transaction is pending.")
			}

			// Decode the input data to extract performer address and operator IDs
			performerAddress, proofOfTask, data, taskDefinitionId, isApproved, tpSignature, taSignature, attestersIds, err := decodeTaskInputData(tx.Data())
			if err != nil {
				logger.Error(fmt.Sprintf("Failed to decode input data: %v", err))
				return
			}

			// Fetch the task information from IPFS using the CID
			cid := string(data) // Assuming data is the CID
			IPFSData, err := fetchDataFromCID(cid)
			if err != nil {
				logger.Error(fmt.Sprintf("Failed to fetch task data from CID: %v", err))
				return
			}

			// Unmarshal the IPFS data into the IPFSData struct
			var ipfsData types.IPFSData
			err = json.Unmarshal([]byte(IPFSData), &ipfsData)
			if err != nil {
				logger.Error(fmt.Sprintf("Failed to unmarshal IPFS data: %v", err))
				return
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

			// Replace the database update section with this single call
			if err := updatePointsInDatabase(taskID, performerAddress, attestersIds); err != nil {
				logger.Error(fmt.Sprintf("Failed to update points in database: %v", err))
				return
			}

		// Handle TaskRejected events
		case event := <-taskRejectedCh:
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
	}
}

// Forward subscription errors to a dedicated channel
func forwardSubscriptionErrors(sub event.Subscription, errCh chan<- error) {
	errCh <- <-sub.Err()
}

// Reconnect OperatorRegistered subscription
func reconnectRegisteredSubscription(
	ethWsClient *ethclient.Client,
	avsGovernanceAddress common.Address,
	operatorRegisteredCh chan *OperatorRegistered,
	regSub *event.Subscription,
	isReconnecting *bool,
) {
	// Set reconnecting flag
	*isReconnecting = true
	defer func() { *isReconnecting = false }()

	// Clean up any existing subscription
	if *regSub != nil {
		(*regSub).Unsubscribe()
		*regSub = nil
	}

	// Wait before reconnecting to avoid hammering the server
	time.Sleep(2 * time.Second)

	// Reconnect to Ethereum WebSocket if needed
	if ethWsClient == nil {
		var err error
		ethWsClient, err = ethclient.Dial(EthWsRpcUrl)
		if err != nil {
			logger.Error(fmt.Sprintf("Failed to reconnect to Ethereum WebSocket for OperatorRegistered: %v", err))
			return
		}
	}

	// Restart watching for OperatorRegistered events
	newRegSub, err := SetupRegisteredSubscription(ethWsClient, avsGovernanceAddress, operatorRegisteredCh)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to restart watching for OperatorRegistered events: %v", err))
	} else {
		*regSub = newRegSub
		logger.Info("Successfully reconnected OperatorRegistered subscription")
	}
}

// Reconnect OperatorUnregistered subscription
func reconnectUnregisteredSubscription(
	ethWsClient *ethclient.Client,
	avsGovernanceAddress common.Address,
	operatorUnregisteredCh chan *OperatorUnregistered,
	unregSub *event.Subscription,
	isReconnecting *bool,
) {
	// Set reconnecting flag
	*isReconnecting = true
	defer func() { *isReconnecting = false }()

	// Clean up any existing subscription
	if *unregSub != nil {
		(*unregSub).Unsubscribe()
		*unregSub = nil
	}

	// Wait before reconnecting to avoid hammering the server
	time.Sleep(2 * time.Second)

	// Reconnect to Ethereum WebSocket if needed
	if ethWsClient == nil {
		var err error
		ethWsClient, err = ethclient.Dial(EthWsRpcUrl)
		if err != nil {
			logger.Error(fmt.Sprintf("Failed to reconnect to Ethereum WebSocket for OperatorUnregistered: %v", err))
			return
		}
	}

	// Restart watching for OperatorUnregistered events
	newUnregSub, err := SetupUnregisteredSubscription(ethWsClient, avsGovernanceAddress, operatorUnregisteredCh)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to restart watching for OperatorUnregistered events: %v", err))
	} else {
		*unregSub = newUnregSub
		logger.Info("Successfully reconnected OperatorUnregistered subscription")
	}
}

// Reconnect TaskSubmitted subscription
func reconnectTaskSubmittedSubscription(
	baseWsClient *ethclient.Client,
	attestationCenterAddress common.Address,
	taskSubmittedCh chan *TaskSubmitted,
	taskSubSub *event.Subscription,
	isReconnecting *bool,
) {
	// Set reconnecting flag
	*isReconnecting = true
	defer func() { *isReconnecting = false }()

	// Clean up any existing subscription
	if *taskSubSub != nil {
		(*taskSubSub).Unsubscribe()
		*taskSubSub = nil
	}

	// Wait before reconnecting to avoid hammering the server
	time.Sleep(2 * time.Second)

	// Reconnect to Base WebSocket if needed
	if baseWsClient == nil {
		var err error
		baseWsClient, err = ethclient.Dial(BaseWsRpcUrl)
		if err != nil {
			logger.Error(fmt.Sprintf("Failed to reconnect to Base WebSocket for TaskSubmitted: %v", err))
			return
		}
	}

	// Restart watching for TaskSubmitted events
	newTaskSubSub, err := SetupTaskSubmittedSubscription(baseWsClient, attestationCenterAddress, taskSubmittedCh)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to restart watching for TaskSubmitted events: %v", err))
	} else {
		*taskSubSub = newTaskSubSub
		logger.Info("Successfully reconnected TaskSubmitted subscription")
	}
}

// Reconnect TaskRejected subscription
func reconnectTaskRejectedSubscription(
	baseWsClient *ethclient.Client,
	attestationCenterAddress common.Address,
	taskRejectedCh chan *TaskRejected,
	taskRejSub *event.Subscription,
	isReconnecting *bool,
) {
	// Set reconnecting flag
	*isReconnecting = true
	defer func() { *isReconnecting = false }()

	// Clean up any existing subscription
	if *taskRejSub != nil {
		(*taskRejSub).Unsubscribe()
		*taskRejSub = nil
	}

	// Wait before reconnecting to avoid hammering the server
	time.Sleep(2 * time.Second)

	// Reconnect to Base WebSocket if needed
	if baseWsClient == nil {
		var err error
		baseWsClient, err = ethclient.Dial(BaseWsRpcUrl)
		if err != nil {
			logger.Error(fmt.Sprintf("Failed to reconnect to Base WebSocket for TaskRejected: %v", err))
			return
		}
	}

	// Restart watching for TaskRejected events
	newTaskRejSub, err := SetupTaskRejectedSubscription(baseWsClient, attestationCenterAddress, taskRejectedCh)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to restart watching for TaskRejected events: %v", err))
	} else {
		*taskRejSub = newTaskRejSub
		logger.Info("Successfully reconnected TaskRejected subscription")
	}
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
	if resp.StatusCode != http.StatusCreated {
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
