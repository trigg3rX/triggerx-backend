package validation

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
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

	// Check if the event name (topic hash) exists in any of the logs' topics
	eventFound := false
	eventHash := crypto.Keccak256Hash([]byte(triggerData.EventTriggerName))
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

	blockNumberHex := fmt.Sprintf("0x%x", receipt.BlockNumber)
	reqBody := map[string]interface{}{
		"jsonrpc": "2.0",
		"method":  "eth_getBlockByNumber",
		"params": []interface{}{
			blockNumberHex,
			false,
		},
		"id": 1,
	}
	reqBytes, err := json.Marshal(reqBody)
	if err != nil {
		return false, fmt.Errorf("failed to marshal eth_getBlockReceipts request: %v", err)
	}

	httpReq, err := http.NewRequest("POST", rpcURL, bytes.NewBuffer(reqBytes))
	if err != nil {
		return false, fmt.Errorf("failed to create eth_getBlockReceipts request: %v", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	httpClient := &http.Client{Timeout: 10 * time.Second}
	resp, err := httpClient.Do(httpReq)
	if err != nil {
		return false, fmt.Errorf("failed to call eth_getBlockReceipts: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return false, fmt.Errorf("eth_getBlockReceipts returned status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var rpcResp struct {
		Result json.RawMessage `json:"result"`
		Error  interface{}     `json:"error"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&rpcResp); err != nil {
		return false, fmt.Errorf("failed to decode eth_getBlockReceipts response: %v", err)
	}
	if rpcResp.Error != nil {
		return false, fmt.Errorf("eth_getBlockReceipts error: %v", rpcResp.Error)
	}

	var block map[string]interface{}
	if err := json.Unmarshal(rpcResp.Result, &block); err != nil {
		return false, fmt.Errorf("failed to unmarshal block: %v", err)
	}

	timestampHex, ok := block["timestamp"].(string)
	if !ok {
		return false, fmt.Errorf("block timestamp is not a string")
	}
	timestampInt, err := strconv.ParseInt(timestampHex, 0, 64)
	if err != nil {
		return false, fmt.Errorf("failed to parse block timestamp hex: %v", err)
	}
	txTimestamp := time.Unix(timestampInt, 0).UTC()
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
