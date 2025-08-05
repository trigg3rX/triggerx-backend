package validation

import (
	"context"
	"fmt"
	"time"

	"github.com/trigg3rX/triggerx-backend/internal/imua-keeper/utils"
	"github.com/trigg3rX/triggerx-backend/pkg/types"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

const timeTolerance = 2200 * time.Millisecond
const expirationTimeTolerance = 5 * time.Second

func (v *TaskValidator) ValidateAction(targetData *types.TaskTargetData, triggerData *types.TaskTriggerData, actionData *types.PerformerActionData, client *ethclient.Client, traceID string) (bool, error) {
	// v.logger.Infof("txHash: %s", actionData.ActionTxHash)
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
	// if receipt.Status != 1 {
	// 	return false, fmt.Errorf("transaction is not successful")
	// }

	// TODO: get the action tx check right
	// check if the tx was made to correct target contract
	// fetch the AA contract address and the transaction from there to complete the flow

	txTimestamp, err := v.getBlockTimestamp(receipt, utils.GetChainRpcUrl(targetData.TargetChainID))
	if err != nil {
		return false, fmt.Errorf("failed to get block timestamp: %v", err)
	}

	// check if the tx was made before expiration time + tolerance
	if txTimestamp.After(triggerData.ExpirationTime.Add(expirationTimeTolerance)) {
		return false, fmt.Errorf("transaction was made after the expiration time by %v", txTimestamp.Sub(triggerData.ExpirationTime.Add(expirationTimeTolerance)))
	}

	// check if the task was time, if yes, check if it was executed within the time interval + tolerance
	if targetData.TaskDefinitionID == 1 || targetData.TaskDefinitionID == 2 {
		if txTimestamp.After(triggerData.NextTriggerTimestamp.Add(timeTolerance)) {
			return false, fmt.Errorf("transaction was made after the next execution timestamp by %v", txTimestamp.Sub(triggerData.NextTriggerTimestamp.Add(timeTolerance)))
		}
		// if txTimestamp.Before(triggerData.NextTriggerTimestamp.Add(-timeTolerance)) {
		// 	return false, fmt.Errorf("transaction was made before the next execution timestamp by %v", triggerData.NextTriggerTimestamp.Add(-timeTolerance).Sub(txTimestamp))
		// }
	}
	return true, nil
}
