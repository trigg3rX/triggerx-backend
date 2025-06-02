package validation

import (
	"context"
	"fmt"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

func (v *TaskValidator) ValidateEventBasedTask(job *types.HandleCreateJobData, ipfsData *types.IPFSData) (bool, error) {
	// Ensure this is an event-based job
	if job.TaskDefinitionID != 3 && job.TaskDefinitionID != 4 {
		return false, fmt.Errorf("not an event-based job: task definition ID %d", job.TaskDefinitionID)
	}

	v.logger.Infof("Validating event-based job %d (taskDefID: %d)", job.JobID, job.TaskDefinitionID)

	// For non-recurring jobs, check if job has already been executed and shouldn't run again
	if !job.Recurring && !job.LastExecutedAt.IsZero() {
		v.logger.Infof("Job %d is non-recurring and has already been executed on %s",
			job.JobID, job.LastExecutedAt.Format(time.RFC3339))
		return false, nil
	}

	// Check if TriggerContractAddress and TriggerEvent are provided
	if job.TriggerContractAddress == "" {
		return false, fmt.Errorf("missing TriggerContractAddress for event-based job %d", job.JobID)
	}

	if job.TriggerEvent == "" {
		return false, fmt.Errorf("missing TriggerEvent for event-based job %d", job.JobID)
	}

	v.logger.Infof("Validating contract address '%s' and event '%s'", job.TriggerContractAddress, job.TriggerEvent)

	// Verify the contract address is valid
	if !common.IsHexAddress(job.TriggerContractAddress) {
		return false, fmt.Errorf("invalid Ethereum contract address: %s", job.TriggerContractAddress)
	}

	// Check if the contract exists on chain
	contractCode, err := v.ethClient.CodeAt(context.Background(), common.HexToAddress(job.TriggerContractAddress), nil)
	if err != nil {
		return false, fmt.Errorf("failed to check contract existence: %v", err)
	}

	if len(contractCode) == 0 {
		return false, fmt.Errorf("no contract found at address: %s", job.TriggerContractAddress)
	}

	// Optional: Check if the contract ABI contains the specified event
	// (This is a simple check to see if we can get the ABI, but we don't validate the event itself
	// since that would require parsing the ABI)
	_, err = v.fetchContractABI(job.TriggerContractAddress)
	if err != nil {
		v.logger.Warnf("Could not fetch ABI for contract %s: %v", job.TriggerContractAddress, err)
		// We don't fail validation just because we can't fetch the ABI
	}

	// Check if job is within its timeframe
	now := time.Now().UTC()
	if job.TimeFrame > 0 {
		endTime := job.CreatedAt.Add(time.Duration(job.TimeFrame) * time.Second)
		if now.After(endTime) {
			v.logger.Infof("Job %d is outside its timeframe (created: %s, timeframe: %d seconds)",
				job.JobID, job.CreatedAt.Format(time.RFC3339), job.TimeFrame)
			return false, nil
		}
	}

	v.logger.Infof("Event-based job %d validated successfully", job.JobID)
	return true, nil
}
