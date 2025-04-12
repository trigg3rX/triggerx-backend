package registrar

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/trigg3rX/triggerx-backend/pkg/ipfs"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

func ProcessOperatorRegisteredEvents(
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
		// Get operator ID from the contract - commenting out for now as binding is missing
		/*
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
		*/

		// Add keeper to database
		err = AddKeeperToDatabase(event.Operator.Hex(), event.BlsKey, event.Raw.TxHash.Hex())
		if err != nil {
			logger.Error(fmt.Sprintf("Failed to add keeper to database: %v", err))
		}
	}

	return nil
}

// Process OperatorUnregistered events
func ProcessOperatorUnregisteredEvents(
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
		err = UpdateKeeperStatusAsUnregistered(event.Operator.Hex())
		if err != nil {
			logger.Error(fmt.Sprintf("Failed to update keeper status in database: %v", err))
		}
	}

	return nil
}

// Function to update keeper status as unregistered in the database via API
func UpdateKeeperStatusAsUnregistered(operatorAddress string) error {
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
func ProcessTaskSubmittedEvents(
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
		if err := UpdatePointsInDatabase(int(ipfsData.TriggerData.TaskID), event.Operator, convertBigIntToStrings(event.AttestersIds)); err != nil {
			logger.Error(fmt.Sprintf("Failed to update points in database: %v", err))
			continue
		}
	}

	return nil
}

// Process TaskRejected events
func ProcessTaskRejectedEvents(
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
		decodedData := string(event.Data)
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

		// Get task fee
		var taskFee int64
		err = db.Session().Query(`SELECT task_fee FROM triggerx.task_data WHERE task_id = ?`,
			ipfsData.TriggerData.TaskID).Scan(&taskFee)
		if err != nil {
			logger.Error(fmt.Sprintf("Failed to get task fee for task ID %d: %v", ipfsData.TriggerData.TaskID, err))
			continue
		}

		logger.Info(fmt.Sprintf("Task ID %d has a fee of %d", ipfsData.TriggerData.TaskID, taskFee))

		// Process attesters - they still get points even if task was rejected
		for _, attesterId := range event.AttestersIds {
			if attesterId != nil && attesterId.String() != "0" {
				if err := UpdateAttesterPoints(attesterId.String(), taskFee); err != nil {
					logger.Error(fmt.Sprintf("Attester points update failed: %v", err))
					continue
				}
			}
		}
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

// Helper function to convert []*big.Int to []string
func convertBigIntToStrings(bigInts []*big.Int) []string {
	result := make([]string, len(bigInts))
	for i, v := range bigInts {
		if v != nil {
			result[i] = v.String()
		}
	}
	return result
}
