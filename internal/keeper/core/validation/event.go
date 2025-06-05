package validation

import (
	"context"
	"fmt"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

func (v *TaskValidator) ValidateEventBasedTask(job *types.SendTaskTargetData, trigger *types.SendTriggerData, ipfsData *types.IPFSData) (bool, error) {
	// Ensure this is an event-based job
	if job.TaskDefinitionID != 3 && job.TaskDefinitionID != 4 {
		return false, fmt.Errorf("not an event-based job: task definition ID %d", job.TaskDefinitionID)
	}

	v.logger.Infof("Validating event-based job %d (taskDefID: %d)", job.JobID, job.TaskDefinitionID)

	// Check if TriggerContractAddress and TriggerEvent are provided
	if trigger.EventTriggerContractAddress == "" {
		return false, fmt.Errorf("missing TriggerContractAddress for event-based job %d", job.JobID)
	}

	if trigger.EventTriggerFunction == "" {
		return false, fmt.Errorf("missing TriggerEvent for event-based job %d", job.JobID)
	}

	v.logger.Infof("Validating contract address '%s' and event '%s'", trigger.EventTriggerContractAddress, trigger.EventTriggerFunction)

	// Verify the contract address is valid
	if !common.IsHexAddress(trigger.EventTriggerContractAddress) {
		return false, fmt.Errorf("invalid Ethereum contract address: %s", trigger.EventTriggerContractAddress)
	}

	// Check if the contract exists on chain
	contractCode, err := v.ethClient.CodeAt(context.Background(), common.HexToAddress(trigger.EventTriggerContractAddress), nil)
	if err != nil {
		return false, fmt.Errorf("failed to check contract existence: %v", err)
	}

	if len(contractCode) == 0 {
		return false, fmt.Errorf("no contract found at address: %s", trigger.EventTriggerContractAddress)
	}

	// Optional: Check if the contract ABI contains the specified event
	// (This is a simple check to see if we can get the ABI, but we don't validate the event itself
	// since that would require parsing the ABI)
	_, err = v.fetchContractABI(trigger.EventTriggerContractAddress)
	if err != nil {
		v.logger.Warnf("Could not fetch ABI for contract %s: %v", trigger.EventTriggerContractAddress, err)
		// We don't fail validation just because we can't fetch the ABI
	}

	// Check if job is within its timeframe
	if trigger.Timestamp.After(job.ExpirationTime) {
		v.logger.Infof("Job %d is outside its timeframe (created: %s, timeframe: %d seconds)",
			job.JobID, trigger.Timestamp.Format(time.RFC3339), job.TimeFrame)
		return false, nil
	}

	v.logger.Infof("Event-based job %d validated successfully", job.JobID)
	return true, nil
}
