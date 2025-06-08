package validation

import (
	// "context"
	"fmt"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

func (v *TaskValidator) ValidateEventBasedTask(ipfsData types.IPFSData) (bool, error) {
	targetData := ipfsData.TargetData
	triggerData := ipfsData.TriggerData

	// Ensure this is an event-based job
	if targetData.TaskDefinitionID != 3 && targetData.TaskDefinitionID != 4 {
		return false, fmt.Errorf("not an event-based job: task definition ID %d", targetData.TaskDefinitionID)
	}

	v.logger.Infof("Validating event-based job %d (taskDefID: %d)", targetData.TaskID, targetData.TaskDefinitionID)

	// Check if TriggerContractAddress and TriggerEvent are provided
	if triggerData.EventTriggerContractAddress == "" {
		return false, fmt.Errorf("missing TriggerContractAddress for event-based job %d", targetData.TaskID)
	}

	if triggerData.EventTriggerFunction == "" {
		return false, fmt.Errorf("missing TriggerEvent for event-based job %d", targetData.TaskID)
	}

	v.logger.Infof("Validating contract address '%s' and event '%s'", triggerData.EventTriggerContractAddress, triggerData.EventTriggerFunction)

	// Verify the contract address is valid
	if !common.IsHexAddress(triggerData.EventTriggerContractAddress) {
		return false, fmt.Errorf("invalid Ethereum contract address: %s", triggerData.EventTriggerContractAddress)
	}

	// Check if the contract exists on chain
	// contractCode, err := v.ethClient.CodeAt(context.Background(), common.HexToAddress(triggerData.EventTriggerContractAddress), nil)
	// if err != nil {
	// 	return false, fmt.Errorf("failed to check contract existence: %v", err)
	// }

	// if len(contractCode) == 0 {
	// 	return false, fmt.Errorf("no contract found at address: %s", triggerData.EventTriggerContractAddress)
	// }

	// Optional: Check if the contract ABI contains the specified event
	// (This is a simple check to see if we can get the ABI, but we don't validate the event itself
	// since that would require parsing the ABI)
	// _, err = v.fetchContractABI(triggerData.EventTriggerContractAddress)
	// if err != nil {
	// 	v.logger.Warnf("Could not fetch ABI for contract %s: %v", triggerData.EventTriggerContractAddress, err)
	// 	// We don't fail validation just because we can't fetch the ABI
	// }

	// Check if job is within its timeframe
	if triggerData.TriggerTimestamp.After(triggerData.ExpirationTime) {
		v.logger.Infof("Job %d is outside its timeframe (created: %s, timeframe: %d seconds)",
			targetData.TaskID, triggerData.TriggerTimestamp.Format(time.RFC3339), triggerData.TimeInterval)
		return false, nil
	}

	v.logger.Infof("Event-based job %d validated successfully", targetData.TaskID)
	return true, nil
}
