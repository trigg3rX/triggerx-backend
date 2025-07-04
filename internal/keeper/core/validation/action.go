package validation

import (
	"context"
	"fmt"
	// "time"

	"github.com/trigg3rX/triggerx-backend/pkg/types"

	"github.com/ethereum/go-ethereum/common"
	// ethTypes "github.com/ethereum/go-ethereum/core/types"
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

	// TODO: Investigate this err: transaction type not supported
	// check if the task was time, if yes, check if it was executed within the time interval + tolerance
	// if targetData.TaskDefinitionID == 1 || targetData.TaskDefinitionID == 2 {
	// 	const timeTolerance = 1100 * time.Millisecond
	// 	var block *ethTypes.Block
	// 	block, err = client.BlockByHash(context.Background(), receipt.BlockHash)
	// 	if err != nil {
	// 		return false, fmt.Errorf("failed to get block: %v", err)
	// 	}
	// 	txTimestamp := time.Unix(int64(block.Time()), 0)

	// 	if txTimestamp.After(triggerData.NextTriggerTimestamp.Add(timeTolerance)) {
	// 		return false, fmt.Errorf("transaction was made after the next execution timestamp")
	// 	}
	// 	if txTimestamp.Before(triggerData.NextTriggerTimestamp.Add(-timeTolerance)) {
	// 		return false, fmt.Errorf("transaction was made before the next execution timestamp")
	// 	}
	// 	return true, nil
	// }
	return true, nil
}
