package validation

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
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

	switch triggerData.TaskDefinitionID {
	case 1, 2:
		isValid, err := e.IsValidTimeBasedTrigger(triggerData)
		if !isValid {
			return isValid, err
		}
	case 3, 4:
		isValid, err := e.IsValidEventBasedTrigger(triggerData)
		if !isValid {
			return isValid, err
		}
	case 5, 6:
		isValid, err := e.IsValidConditionBasedTrigger(triggerData)
		if !isValid {
			return isValid, err
		}
	default:
		return false, fmt.Errorf("invalid task definition id: %d", triggerData.TaskDefinitionID)
	}

	return true, nil
}

func (v *TaskValidator) IsValidTimeBasedTrigger(triggerData *types.TaskTriggerData) (bool, error) {
	// check if expiration time is before trigger timestamp
	if triggerData.ExpirationTime.Before(triggerData.NextTriggerTimestamp) {
		return false, errors.New("expiration time is before trigger timestamp")
	}

	// rest validation is handled when we validate the action

	return true, nil
}

func (v *TaskValidator) IsValidEventBasedTrigger(triggerData *types.TaskTriggerData) (bool, error) {
	// check if expiration time is before trigger timestamp
	if triggerData.ExpirationTime.Before(triggerData.NextTriggerTimestamp) {
		return false, errors.New("expiration time is before trigger timestamp")
	}

	rpcURL := utils.GetChainRpcUrl(triggerData.EventChainId)
	client, err := ethclient.Dial(rpcURL)
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
	if receipt.Logs[0].Address != common.HexToAddress(triggerData.EventTriggerContractAddress) {
		return false, fmt.Errorf("transaction was not made to correct target contract")
	}

	txTimestamp, err := v.getBlockTimestamp(receipt, rpcURL)
	if err != nil {
		return false, fmt.Errorf("failed to get block timestamp: %v", err)
	}

	expirationTime := triggerData.ExpirationTime.UTC()

	if txTimestamp.After(expirationTime.Add(timeTolerance)) {
		return false, fmt.Errorf("transaction was made after the expiration time (tx: %v, exp+tolerance: %v)",
			txTimestamp.Format(time.RFC3339),
			expirationTime.Add(timeTolerance).Format(time.RFC3339))
	}

	return true, nil
}

func (v *TaskValidator) IsValidConditionBasedTrigger(triggerData *types.TaskTriggerData) (bool, error) {
	// check if expiration time is before trigger timestamp
	if triggerData.ExpirationTime.Before(triggerData.NextTriggerTimestamp) {
		return false, errors.New("expiration time is before trigger timestamp")
	}
	// v.logger.Infof("trigger data: %+v", triggerData)
	v.logger.Infof("value: %v | upper limit: %v | lower limit: %v", triggerData.ConditionSatisfiedValue, triggerData.ConditionUpperLimit, triggerData.ConditionLowerLimit)

	// check if the condition was satisfied by the value
	if triggerData.ConditionType == ConditionEquals {
		return triggerData.ConditionSatisfiedValue == triggerData.ConditionUpperLimit, nil
	}
	if triggerData.ConditionType == ConditionNotEquals {
		return triggerData.ConditionSatisfiedValue != triggerData.ConditionUpperLimit, nil
	}
	if triggerData.ConditionType == ConditionGreaterThan {
		return triggerData.ConditionSatisfiedValue > triggerData.ConditionUpperLimit, nil
	}
	if triggerData.ConditionType == ConditionLessThan {
		return triggerData.ConditionSatisfiedValue < triggerData.ConditionUpperLimit, nil
	}
	if triggerData.ConditionType == ConditionGreaterEqual {
		return triggerData.ConditionSatisfiedValue >= triggerData.ConditionUpperLimit, nil
	}
	if triggerData.ConditionType == ConditionLessEqual {
		return triggerData.ConditionSatisfiedValue <= triggerData.ConditionUpperLimit, nil
	}
	if triggerData.ConditionType == ConditionBetween {
		return triggerData.ConditionSatisfiedValue >= triggerData.ConditionLowerLimit && triggerData.ConditionSatisfiedValue <= triggerData.ConditionUpperLimit, nil
	}

	// TODO: add: to fetch the data at trigger timestamp, and check if the values fetched is true
	// oracles would be easy, apis would not be possible if there is no support for it

	return false, nil
}
