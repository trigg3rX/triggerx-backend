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
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

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
			performerAddress, operatorIDs, err := decodeTaskInputData(tx.Data())
			if err != nil {
				logger.Error(fmt.Sprintf("Failed to decode input data: %v", err))
				return
			}

			logger.Info(fmt.Sprintf("Performer Address: %s", performerAddress.Hex()))
			logger.Info(fmt.Sprintf("Operator IDs: %v", operatorIDs))

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
func decodeTaskInputData(data []byte) (common.Address, []string, error) {
	// Define the ABI of the contract containing the submitTask function
	const submitTaskABI = `[{"inputs":[{"internalType":"address","name":"performer","type":"address"},{"internalType":"string[]","name":"operatorIds","type":"string[]"}],"name":"submitTask","outputs":[],"stateMutability":"nonpayable","type":"function"}]`

	// Parse the ABI
	parsedABI, err := abi.JSON(strings.NewReader(submitTaskABI))
	if err != nil {
		return common.Address{}, nil, fmt.Errorf("failed to parse ABI: %v", err)
	}

	// Variables to hold the decoded values
	var performerAddress common.Address
	var operatorIDs []string

	// Decode the input data
	err = parsedABI.UnpackIntoInterface(&struct {
		Performer   *common.Address
		OperatorIds *[]string
	}{&performerAddress, &operatorIDs}, "submitTask", data)
	if err != nil {
		return common.Address{}, nil, fmt.Errorf("failed to unpack input data: %v", err)
	}

	return performerAddress, operatorIDs, nil
}
