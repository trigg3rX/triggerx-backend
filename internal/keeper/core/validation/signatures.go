package validation

import (
	"fmt"

	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

func (v *TaskValidator) ValidateSchedulerSignature(task *types.SendTaskDataToKeeper, traceID string) (bool, error) {
	logger := v.logger.With("traceID", traceID)

	// check if the scheduler signature is valid
	if task.SchedulerSignature == nil {
		logger.Error("Scheduler signature data is missing")
		return false, fmt.Errorf("scheduler signature data is missing")
	}

	if task.SchedulerSignature.SchedulerSignature == "" {
		logger.Error("Scheduler signature is empty")
		return false, fmt.Errorf("scheduler signature is empty")
	}

	if task.SchedulerSignature.SchedulerSigningAddress == "" {
		logger.Error("Scheduler signing address is empty")
		return false, fmt.Errorf("scheduler signing address is empty")
	}

	// Create a copy of the task data without the signature for verification
	taskDataForVerification := types.SendTaskDataToKeeper{
		TaskID:        task.TaskID,
		PerformerData: task.PerformerData,
		TargetData:    task.TargetData,
		TriggerData:   task.TriggerData,
		SchedulerSignature: &types.SchedulerSignatureData{
			TaskID:                  task.SchedulerSignature.TaskID,
			SchedulerSigningAddress: task.SchedulerSignature.SchedulerSigningAddress,
			// Note: SchedulerSignature field is intentionally left empty for verification
		},
	}

	// Convert the task data to JSON message format (same as signing process)
	isValid, err := v.crypto.VerifySignatureFromJSON(
		taskDataForVerification,
		task.SchedulerSignature.SchedulerSignature,
		task.SchedulerSignature.SchedulerSigningAddress,
	)
	if err != nil {
		logger.Error("Failed to verify scheduler signature", "error", err)
		return false, fmt.Errorf("failed to verify scheduler signature: %w", err)
	}

	if !isValid {
		logger.Error("Scheduler signature verification failed")
		return false, fmt.Errorf("scheduler signature verification failed")
	}

	logger.Info("Scheduler signature verification successful")
	return true, nil
}

func (v *TaskValidator) ValidatePerformerSignature(ipfsData types.IPFSData, traceID string) (bool, error) {
	logger := v.logger.With("traceID", traceID)

	if ipfsData.PerformerSignature == nil {
		logger.Error("Performer signature data is missing")
		return false, fmt.Errorf("performer signature data is missing")
	}

	if ipfsData.PerformerSignature.PerformerSignature == "" {
		logger.Error("Performer signature is empty")
		return false, fmt.Errorf("performer signature is empty")
	}

	if ipfsData.PerformerSignature.PerformerSigningAddress == "" {
		logger.Error("Performer signing address is empty")
		return false, fmt.Errorf("performer signing address is empty")
	}

	// TODO: Uncomment this when we have a way to get the Consensus address to perform the action with AA
	// check if the performer is the same as the the one assigned to the task
	// if ipfsData.PerformerSignature.PerformerSigningAddress != ipfsData.TaskData.PerformerData.KeeperAddress {
	// 	return false, fmt.Errorf("performer signing address does not match the assigned performer")
	// }

	// Create a copy of the ipfs data without the signature for verification
	ipfsDataForVerification := types.IPFSData{
		TaskData:   ipfsData.TaskData,
		ActionData: ipfsData.ActionData,
		ProofData:  ipfsData.ProofData,
		PerformerSignature: &types.PerformerSignatureData{
			TaskID:                  ipfsData.TaskData.TaskID,
			PerformerSigningAddress: ipfsData.PerformerSignature.PerformerSigningAddress,
			// Note: PerformerSignature field is intentionally left empty for verification
		},
	}

	// Convert the task data to JSON message format (same as signing process)
	isValid, err := v.crypto.VerifySignatureFromJSON(
		ipfsDataForVerification,
		ipfsData.PerformerSignature.PerformerSignature,
		ipfsData.PerformerSignature.PerformerSigningAddress,
	)
	if err != nil {
		logger.Error("Failed to verify performer signature", "error", err)
		return false, fmt.Errorf("failed to verify performer signature: %w", err)
	}

	if !isValid {
		logger.Error("Performer signature verification failed")
		return false, fmt.Errorf("performer signature verification failed")
	}

	logger.Info("Performer signature verification successful")
	return true, nil
}
