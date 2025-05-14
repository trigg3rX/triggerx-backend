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

	"github.com/trigg3rX/triggerx-backend/internal/registrar/config"
	"github.com/trigg3rX/triggerx-backend/internal/registrar/database"

	"github.com/trigg3rX/triggerx-backend/pkg/ipfs"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

func ProcessOperatorRegisteredEvents(
	ethClient *ethclient.Client,
	contractAddress common.Address,
	fromBlock uint64,
	toBlock uint64,
) error {
	query := ethereum.FilterQuery{
		FromBlock: new(big.Int).SetUint64(fromBlock),
		ToBlock:   new(big.Int).SetUint64(toBlock),
		Addresses: []common.Address{contractAddress},
		Topics: [][]common.Hash{
			{OperatorRegisteredEventSignature()},
		},
	}

	logs, err := ethClient.FilterLogs(context.Background(), query)
	if err != nil {
		return fmt.Errorf("failed to filter OperatorRegistered logs: %v", err)
	}

	logger.Debugf("Found %d OperatorRegistered events", len(logs))

	for _, vLog := range logs {
		event, err := ParseOperatorRegistered(vLog)
		if err != nil {
			logger.Errorf("Failed to parse OperatorRegistered event: %v", err)
			continue
		}

		logger.Infof("Operator Registered Event Detected!")
		logger.Debugf("Operator Address: %s", event.Operator.Hex())
		logger.Debugf("Transaction Hash: %s", event.Raw.TxHash.Hex())
		logger.Debugf("Block Number: %d", event.Raw.BlockNumber)

		err = database.KeeperRegistered(event.Operator.Hex(), event.Raw.TxHash.Hex())
		if err != nil {
			logger.Errorf("Failed to add keeper to database: %v", err)
		}

		FetchOperatorDetailsAfterDelay(event.Operator, 3*time.Minute)
	}
	return nil
}

func ProcessOperatorUnregisteredEvents(
	ethClient *ethclient.Client,
	contractAddress common.Address,
	fromBlock uint64,
	toBlock uint64,
) error {
	query := ethereum.FilterQuery{
		FromBlock: new(big.Int).SetUint64(fromBlock),
		ToBlock:   new(big.Int).SetUint64(toBlock),
		Addresses: []common.Address{contractAddress},
		Topics: [][]common.Hash{
			{OperatorUnregisteredEventSignature()},
		},
	}

	logs, err := ethClient.FilterLogs(context.Background(), query)
	if err != nil {
		return fmt.Errorf("failed to filter OperatorUnregistered logs: %v", err)
	}

	logger.Debugf("Found %d OperatorUnregistered events", len(logs))

	for _, vLog := range logs {
		event, err := ParseOperatorUnregistered(vLog)
		if err != nil {
			logger.Errorf("Failed to parse OperatorUnregistered event: %v", err)
			continue
		}

		logger.Infof("Operator Unregistered Event Detected!")
		logger.Debugf("Operator Address: %s", event.Operator.Hex())
		logger.Debugf("Transaction Hash: %s", event.Raw.TxHash.Hex())
		logger.Debugf("Block Number: %d", event.Raw.BlockNumber)

		err = database.KeeperUnregistered(event.Operator.Hex())
		if err != nil {
			logger.Errorf("Failed to update keeper status in database: %v", err)
		}
	}
	return nil
}

func processEventsInBatches(
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

func ProcessTaskSubmittedEvents(
	baseClient *ethclient.Client,
	contractAddress common.Address,
	fromBlock uint64,
	toBlock uint64,
) error {
	return processEventsInBatches(baseClient, contractAddress, fromBlock, toBlock, 500, processTaskSubmittedBatch)
}

func processTaskSubmittedBatch(
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

	logger.Debugf("Found %d TaskSubmitted events in batch [%d-%d]", len(logs), fromBlock, toBlock)

	for _, vLog := range logs {
		event, err := ParseTaskSubmitted(vLog)
		if err != nil {
			logger.Errorf("Failed to parse TaskSubmitted event: %v", err)
			continue
		}

		logger.Infof("Task Submitted Event Detected!")
		logger.Debugf("Performer Address: %s", event.Operator)
		logger.Debugf("Attester IDs: %v", event.AttestersIds)
		logger.Debugf("Task Number: %d", event.TaskNumber)
		logger.Debugf("Task Definition ID: %d", event.TaskDefinitionId)

		if event.TaskDefinitionId == 10001 || event.TaskDefinitionId == 10002 {
		} else {
			dataCID := string(event.Data)
			logger.Debugf("Decoded Data: %s", dataCID)

			ipfsContent, err := ipfs.FetchIPFSContent(config.IpfsHost, dataCID)
			if err != nil {
				logger.Errorf("Failed to fetch IPFS content: %v", err)
				continue
			}

			var ipfsData types.IPFSData
			if err := json.Unmarshal([]byte(ipfsContent), &ipfsData); err != nil {
				logger.Errorf("Failed to parse IPFS content into IPFSData: %v", err)
				continue
			}

			if err := database.UpdatePointsInDatabase(int(ipfsData.TriggerData.TaskID), event.Operator, convertBigIntToStrings(event.AttestersIds), true); err != nil {
				logger.Errorf("Failed to update points in database: %v", err)
				continue
			}
		}
	}
	return nil
}

func ProcessTaskRejectedEvents(
	baseClient *ethclient.Client,
	contractAddress common.Address,
	fromBlock uint64,
	toBlock uint64,
) error {
	return processEventsInBatches(baseClient, contractAddress, fromBlock, toBlock, 500, processTaskRejectedBatch)
}

func processTaskRejectedBatch(
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

	logger.Debugf("Found %d TaskRejected events in batch [%d-%d]", len(logs), fromBlock, toBlock)

	for _, vLog := range logs {
		event, err := ParseTaskRejected(vLog)
		if err != nil {
			logger.Errorf("Failed to parse TaskRejected event: %v", err)
			continue
		}

		logger.Infof("Task Rejected Event Detected!")
		logger.Debugf("Performer Address: %s", event.Operator)
		logger.Debugf("Attester IDs: %v", event.AttestersIds)
		logger.Debugf("Task Number: %d", event.TaskNumber)
		logger.Debugf("Task Definition ID: %d", event.TaskDefinitionId)

		if event.TaskDefinitionId == 10001 || event.TaskDefinitionId == 10002 {
		} else {
			dataCID := string(event.Data)
			logger.Debugf("Decoded Data: %s", dataCID)

			ipfsContent, err := ipfs.FetchIPFSContent(config.IpfsHost, dataCID)
			if err != nil {
				logger.Errorf("Failed to fetch IPFS content: %v", err)
				continue
			}

			var ipfsData types.IPFSData
			if err := json.Unmarshal([]byte(ipfsContent), &ipfsData); err != nil {
				logger.Errorf("Failed to parse IPFS content into IPFSData: %v", err)
				continue
			}

			if err := database.UpdatePointsInDatabase(int(ipfsData.TriggerData.TaskID), event.Operator, convertBigIntToStrings(event.AttestersIds), false); err != nil {
				logger.Errorf("Failed to update points in database: %v", err)
				continue
			}
		}
	}
	return nil
}

func OperatorRegisteredEventSignature() common.Hash {
	return crypto.Keccak256Hash([]byte("OperatorRegistered(address,uint256[4])"))
}

func OperatorUnregisteredEventSignature() common.Hash {
	return crypto.Keccak256Hash([]byte("OperatorUnregistered(address)"))
}

func TaskSubmittedEventSignature() common.Hash {
	return crypto.Keccak256Hash([]byte("TaskSubmitted(address,uint32,string,bytes,uint16,uint256[])"))
}

func TaskRejectedEventSignature() common.Hash {
	return crypto.Keccak256Hash([]byte("TaskRejected(address,uint32,string,bytes,uint16,uint256[])"))
}

func convertBigIntToStrings(bigInts []*big.Int) []string {
	result := make([]string, len(bigInts))
	for i, v := range bigInts {
		if v != nil {
			result[i] = v.String()
		}
	}
	return result
}
