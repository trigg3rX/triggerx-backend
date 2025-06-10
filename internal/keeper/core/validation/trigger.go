package validation

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/ethereum/go-ethereum/common"
	ethTypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/trigg3rX/triggerx-backend/internal/keeper/utils"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

// Supported condition types
const (
	ConditionGreaterThan  = "greater_than"
	ConditionLessThan     = "less_than"
	ConditionBetween      = "between"
	ConditionEquals       = "equals"
	ConditionNotEquals    = "not_equals"
	ConditionGreaterEqual = "greater_equal"
	ConditionLessEqual    = "less_equal"
)

func (e *TaskValidator) ValidateTrigger(triggerData *types.TaskTriggerData, traceID string) (bool, error) {
	e.logger.Info("Validating trigger data", "task_id", triggerData.TaskID, "trace_id", traceID)

	if triggerData.TaskDefinitionID == 1 || triggerData.TaskDefinitionID == 2 {
		return e.IsValidTimeBasedTrigger(triggerData)
	}

	if triggerData.TaskDefinitionID == 3 || triggerData.TaskDefinitionID == 4 {
		return e.IsValidEventBasedTrigger(triggerData)
	}

	if triggerData.TaskDefinitionID == 5 || triggerData.TaskDefinitionID == 6 {
		return e.IsValidConditionBasedTrigger(triggerData)
	}

	return false, fmt.Errorf("invalid task definition ID: %d", triggerData.TaskDefinitionID)
}

func (v *TaskValidator) IsValidTimeBasedTrigger(triggerData *types.TaskTriggerData) (bool, error) {
	// check if expiration time is before trigger timestamp
	if triggerData.ExpirationTime.Before(triggerData.TriggerTimestamp) {
		return false, errors.New("expiration time is before trigger timestamp")
	}

	// rest validation is handled when we validate the action

	return true, nil
}

func (v *TaskValidator) IsValidEventBasedTrigger(triggerData *types.TaskTriggerData) (bool, error) {
	rpcURL := utils.GetChainRpcUrl(triggerData.EventChainId)
	client, err := v.ethClientMaker(rpcURL)
	if err != nil {
		return false, fmt.Errorf("failed to connect to chain: %v", err)
	}
	defer client.Close()

	// Check if the contract exists on chain
	contractCode, err := client.CodeAt(context.Background(), common.HexToAddress(triggerData.EventTriggerContractAddress), nil)
	if err != nil {
		return false, fmt.Errorf("failed to check contract existence: %v", err)
	}

	if len(contractCode) == 0 {
		return false, fmt.Errorf("no contract found at address: %s", triggerData.EventTriggerContractAddress)
	}

	// check if the tx is successful
	txHash := common.HexToHash(triggerData.EventTxHash)
	receipt, err := client.TransactionReceipt(context.Background(), txHash)
	if err != nil {
		_, isPending, err := client.TransactionByHash(context.Background(), txHash)
		if err != nil {
			return false, fmt.Errorf("failed to get transaction: %v", err)
		}
		if isPending {
			return false, fmt.Errorf("transaction is pending")
		}
		return false, fmt.Errorf("transaction is not found")
	}

	// check if the tx is successful
	if receipt.Status != 1 {
		return false, fmt.Errorf("transaction is not successful")
	}

	// check if the tx was made to correct target contract
	if receipt.ContractAddress != common.HexToAddress(triggerData.EventTriggerContractAddress) {
		return false, fmt.Errorf("transaction was not made to correct target contract")
	}

	// Check if the event name (topic hash) exists in any of the logs' topics
	eventFound := false
	eventHash := common.HexToHash(triggerData.EventTriggerName)
	for _, log := range receipt.Logs {
		for _, topic := range log.Topics {
			if topic == eventHash {
				eventFound = true
				break
			}
		}
		if eventFound {
			break
		}
	}
	if !eventFound {
		return false, fmt.Errorf("event name is not correct")
	}

	// check if the tx was made within expiration time + the time tolerance
	const timeTolerance = 1100 * time.Millisecond
	var block *ethTypes.Block
	if receipt.BlockNumber == nil {
		block, err = client.BlockByNumber(context.Background(), receipt.BlockNumber)
	} else {
		block, err = client.BlockByHash(context.Background(), receipt.BlockHash)
	}
	if err != nil {
		return false, fmt.Errorf("failed to get block: %v", err)
	}
	txTimestamp := time.Unix(int64(block.Time()), 0)

	if txTimestamp.After(triggerData.ExpirationTime.Add(timeTolerance)) {
		return false, fmt.Errorf("transaction was made after the expiration time")
	}

	return true, nil
}

func (v *TaskValidator) IsValidConditionBasedTrigger(triggerData *types.TaskTriggerData) (bool, error) {
	// check if the condition was satisfied by the value
	if triggerData.ConditionSourceType == ConditionEquals {
		return triggerData.ConditionSatisfiedValue == triggerData.ConditionUpperLimit, nil
	}
	if triggerData.ConditionSourceType == ConditionNotEquals {
		return triggerData.ConditionSatisfiedValue != triggerData.ConditionUpperLimit, nil
	}
	if triggerData.ConditionSourceType == ConditionGreaterThan {
		return triggerData.ConditionSatisfiedValue > triggerData.ConditionUpperLimit, nil
	}
	if triggerData.ConditionSourceType == ConditionLessThan {
		return triggerData.ConditionSatisfiedValue < triggerData.ConditionUpperLimit, nil
	}
	if triggerData.ConditionSourceType == ConditionGreaterEqual {
		return triggerData.ConditionSatisfiedValue >= triggerData.ConditionUpperLimit, nil
	}
	if triggerData.ConditionSourceType == ConditionLessEqual {
		return triggerData.ConditionSatisfiedValue <= triggerData.ConditionUpperLimit, nil
	}
	if triggerData.ConditionSourceType == ConditionBetween {
		return triggerData.ConditionSatisfiedValue >= triggerData.ConditionLowerLimit && triggerData.ConditionSatisfiedValue <= triggerData.ConditionUpperLimit, nil
	}

	// TODO: add: to fetch the data at trigger timestamp, and check if the values fetched is true
	// oracles would be easy, apis would not be possible if there is no support for it

	return false, nil
}
