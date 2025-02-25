package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"net/http"
	"os"
	"strings"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	gethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/joho/godotenv"

	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

const OperatorRegisteredSig = "OperatorRegistered(address,uint256[4])"
const OperatorUnregisteredSig = "OperatorUnregistered(address)"
const TaskSubmittedSig = "TaskSubmitted(address,uint32,string,bytes,uint16)"
const TaskRejectedSig = "TaskRejected(address,uint32,string,bytes,uint16)"

type KeeperRegisteredEvent struct {
	Operator common.Address
	BlsKey   [4]*big.Int
}

type KeeperUnregisteredEvent struct {
	Operator common.Address
}

type TaskSubmittedEvent struct {
	Operator         common.Address
	TaskID           uint32
	TaskName         string
	TaskData         []byte
	TaskDefinitionID int
}

type TaskRejectedEvent struct {
	Operator         common.Address
	TaskID           uint32
	TaskName         string
	TaskData         []byte
	TaskDefinitionID int
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	avsGovernanceAddress := os.Getenv("AVS_GOVERNANCE_ADDRESS")
	attestationCenterAddress := os.Getenv("ATTESTATION_CENTER_ADDRESS")

	ethClient, err := ethclient.Dial(os.Getenv("L1_RPC"))
	if err != nil {
		log.Fatalf("Failed to connect to the Ethereum client: %v", err)
	}

	baseClient, err := ethclient.Dial(os.Getenv("L2_RPC"))
	if err != nil {
		log.Fatalf("Failed to connect to the Base client: %v", err)
	}

	registeredQuery := ethereum.FilterQuery{
		Addresses: []common.Address{common.HexToAddress(avsGovernanceAddress)},
		Topics: [][]common.Hash{{
			common.HexToHash("0x" + generateEventSignatureHash(OperatorRegisteredSig)),
		}},
	}

	unregisteredQuery := ethereum.FilterQuery{
		Addresses: []common.Address{common.HexToAddress(avsGovernanceAddress)},
		Topics: [][]common.Hash{{
			common.HexToHash("0x" + generateEventSignatureHash(OperatorUnregisteredSig)),
		}},
	}

	taskSubmittedQuery := ethereum.FilterQuery{
		Addresses: []common.Address{common.HexToAddress(attestationCenterAddress)},
		Topics: [][]common.Hash{{
			common.HexToHash("0x" + generateEventSignatureHash(TaskSubmittedSig)),
		}},
	}

	taskRejectedQuery := ethereum.FilterQuery{
		Addresses: []common.Address{common.HexToAddress(attestationCenterAddress)},
		Topics: [][]common.Hash{{
			common.HexToHash("0x" + generateEventSignatureHash(TaskRejectedSig)),
		}},
	}

	regLogs := make(chan gethtypes.Log)
	unregLogs := make(chan gethtypes.Log)
	taskSubmittedLogs := make(chan gethtypes.Log)
	taskRejectedLogs := make(chan gethtypes.Log)

	registeredSub, err := ethClient.SubscribeFilterLogs(context.Background(), registeredQuery, regLogs)
	if err != nil {
		log.Fatalf("Failed to subscribe to logs: %v", err)
	}
	defer registeredSub.Unsubscribe()

	unregisteredSub, err := ethClient.SubscribeFilterLogs(context.Background(), unregisteredQuery, unregLogs)
	if err != nil {
		log.Fatalf("Failed to subscribe to logs: %v", err)
	}
	defer unregisteredSub.Unsubscribe()

	taskSubmittedSub, err := baseClient.SubscribeFilterLogs(context.Background(), taskSubmittedQuery, taskSubmittedLogs)
	if err != nil {
		log.Fatalf("Failed to subscribe to logs: %v", err)
	}
	defer taskSubmittedSub.Unsubscribe()

	taskRejectedSub, err := baseClient.SubscribeFilterLogs(context.Background(), taskRejectedQuery, taskRejectedLogs)
	if err != nil {
		log.Fatalf("Failed to subscribe to logs: %v", err)
	}
	defer taskRejectedSub.Unsubscribe()

	fmt.Println("Listening for Events...")

	go func() {
		for {
			select {
			case err := <-registeredSub.Err():
				log.Fatal(err)
			case vLog := <-regLogs:
				event, err := parseOperatorRegisteredEvent(vLog)
				if err != nil {
					log.Printf("Error parsing event: %v", err)
					continue
				}
				log.Printf("Processed operator registration: %s", event.Operator.Hex())
			case err := <-unregisteredSub.Err():
				log.Fatal(err)
			case vLog := <-unregLogs:
				event, err := parseOperatorUnregisteredEvent(vLog)
				if err != nil {
					log.Printf("Error parsing event: %v", err)
					continue
				}
				log.Printf("Processed operator unregistration: %s", event.Operator.Hex())
			case err := <-taskSubmittedSub.Err():
				log.Fatal(err)
			case vLog := <-taskSubmittedLogs:
				event, err := parseTaskSubmittedEvent(vLog)
				if err != nil {
					log.Printf("Error parsing event: %v", err)
					continue
				}
				log.Printf("Processed task submission: ID=%d", event.TaskID)
			case err := <-taskRejectedSub.Err():
				log.Fatal(err)
			case vLog := <-taskRejectedLogs:
				event, err := parseTaskRejectedEvent(vLog)
				if err != nil {
					log.Printf("Error parsing event: %v", err)
					continue
				}
				log.Printf("Processed task rejection: ID=%d", event.TaskID)
			}
		}
	}()

	select {}
}

func generateEventSignatureHash(sig string) string {
	return strings.TrimPrefix(common.HexToHash(sig).Hex(), "0x")
}

func parseOperatorRegisteredEvent(vLog gethtypes.Log) (*KeeperRegisteredEvent, error) {
	if len(vLog.Topics) < 2 {
		return nil, fmt.Errorf("invalid number of topics for OperatorRegistered event")
	}

	event := &KeeperRegisteredEvent{}
	event.Operator = common.HexToAddress(vLog.Topics[1].Hex())

	err := json.Unmarshal(vLog.Data, &event.BlsKey)
	if err != nil {
		return nil, fmt.Errorf("error unmarshaling BLS key: %v", err)
	}
	log.Printf("Operator registered: %s", event.Operator.Hex())

	err = createKeeper(types.CreateKeeperData{
		KeeperAddress:  event.Operator.Hex(),
		RegisteredTx:   vLog.TxHash.Hex(),
		RewardsAddress: event.Operator.Hex(),
		ConsensusKeys:  []string{},
	})
	if err != nil {
		log.Printf("Error creating keeper: %v", err)
	}

	return event, nil
}

func parseOperatorUnregisteredEvent(vLog gethtypes.Log) (*KeeperUnregisteredEvent, error) {
	if len(vLog.Topics) < 2 {
		return nil, fmt.Errorf("invalid number of topics for OperatorUnregistered event")
	}

	event := &KeeperUnregisteredEvent{}

	event.Operator = common.HexToAddress(vLog.Topics[1].Hex())

	log.Printf("Operator unregistered: %s", event.Operator.Hex())

	// TODO: Update keeper in DB

	return event, nil
}

func parseTaskSubmittedEvent(vLog gethtypes.Log) (*TaskSubmittedEvent, error) {
	if len(vLog.Topics) < 2 {
		return nil, fmt.Errorf("invalid number of topics for TaskSubmitted event")
	}

	event := &TaskSubmittedEvent{}

	event.Operator = common.HexToAddress(vLog.Topics[1].Hex())

	var data struct {
		TaskID           uint32
		TaskName         string
		TaskData         []byte
		TaskDefinitionID int
	}

	err := json.Unmarshal(vLog.Data, &data)
	if err != nil {
		return nil, fmt.Errorf("error unmarshaling task data: %v", err)
	}

	event.TaskID = data.TaskID
	event.TaskName = data.TaskName
	event.TaskData = data.TaskData
	event.TaskDefinitionID = data.TaskDefinitionID

	log.Printf("Task submitted: ID=%d, Name=%s, Operator=%s",
		event.TaskID, event.TaskName, event.Operator.Hex())

	// TODO: Update task in DB

	return event, nil
}

func parseTaskRejectedEvent(vLog gethtypes.Log) (*TaskRejectedEvent, error) {
	if len(vLog.Topics) < 2 {
		return nil, fmt.Errorf("invalid number of topics for TaskRejected event")
	}

	event := &TaskRejectedEvent{}

	event.Operator = common.HexToAddress(vLog.Topics[1].Hex())

	var data struct {
		TaskID           uint32
		TaskName         string
		TaskData         []byte
		TaskDefinitionID int
	}

	err := json.Unmarshal(vLog.Data, &data)
	if err != nil {
		return nil, fmt.Errorf("error unmarshaling task data: %v", err)
	}

	event.TaskID = data.TaskID
	event.TaskName = data.TaskName
	event.TaskData = data.TaskData
	event.TaskDefinitionID = data.TaskDefinitionID

	log.Printf("Task rejected: ID=%d, Name=%s, Operator=%s",
		event.TaskID, event.TaskName, event.Operator.Hex())

	// TODO: Update task in DB

	return event, nil
}

func createKeeper(data types.CreateKeeperData) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("error marshaling data: %v", err)
	}

	resp, err := http.Post("http://data.triggerx.network/api/keepers",
		"application/json",
		bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("error sending request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}
