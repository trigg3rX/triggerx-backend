package validation

import (
	"context"
	"fmt"
	
	"github.com/trigg3rX/triggerx-backend/internal/keeper/utils"
	"github.com/trigg3rX/triggerx-backend/pkg/types"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

func (v *TaskValidator) ValidateAction(targetData *types.TaskTargetData, triggerData *types.TaskTriggerData, actionData *types.PerformerActionData, client *ethclient.Client, traceID string) (bool, error) {
	v.logger.Infof("txHash: %s", actionData.ActionTxHash)
	// time.Sleep(10 * time.Second)
	// Fetch the tx details from the action data
	txHash := common.HexToHash(actionData.ActionTxHash)
	receipt, err := client.TransactionReceipt(context.Background(), txHash)
	if err != nil || receipt == nil {
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

	// TODO: get the action tx check right
	// check if the tx was made to correct target contract
	// fetch the AA contract address and the transaction from there to complete the flow

	txTimestamp, err := v.getBlockTimestamp(receipt, utils.GetChainRpcUrl(targetData.TargetChainID))
	if err != nil {
		return false, fmt.Errorf("failed to get block timestamp: %v", err)
	}

	// check if the tx was made within expiration time + the time tolerance
	if triggerData.ExpirationTime.Before(txTimestamp.Add(timeTolerance)) {
		return false, fmt.Errorf("transaction was made after the expiration time")
	}

	// check if the task was time, if yes, check if it was executed within the time interval + tolerance
	if targetData.TaskDefinitionID == 1 || targetData.TaskDefinitionID == 2 {
		if txTimestamp.After(triggerData.NextTriggerTimestamp.Add(timeTolerance)) {
			return false, fmt.Errorf("transaction was made after the next execution timestamp")
		}
		if txTimestamp.Before(triggerData.NextTriggerTimestamp.Add(-timeTolerance)) {
			return false, fmt.Errorf("transaction was made before the next execution timestamp")
		}
	}
	return true, nil
}
