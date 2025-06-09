package validation

import (
	"fmt"

	"github.com/trigg3rX/triggerx-backend/pkg/cryptography"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

func (v *TaskValidator) ValidateSchedulerSignature(task *types.SendTaskDataToKeeper, traceID string) (bool, error) {
	// check if the scheduler signature is valid
	if task.SchedulerSignature == nil {
		return false, fmt.Errorf("scheduler signature data is missing")
	}

	if task.SchedulerSignature.SchedulerSignature == "" {
		return false, fmt.Errorf("scheduler signature is empty")
	}

	if task.SchedulerSignature.SchedulerSigningAddress == "" {
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
	isValid, err := cryptography.VerifySignatureFromJSON(
		taskDataForVerification,
		task.SchedulerSignature.SchedulerSignature,
		task.SchedulerSignature.SchedulerSigningAddress,
	)
	if err != nil {
		return false, fmt.Errorf("failed to verify scheduler signature: %w", err)
	}

	return isValid, nil
}

func (v *TaskValidator) ValidatePerformerSignature(ipfsData types.IPFSData, traceID string) (bool, error) {
	if ipfsData.PerformerSignature == nil {
		return false, fmt.Errorf("performer signature data is missing")
	}

	if ipfsData.PerformerSignature.PerformerSignature == "" {
		return false, fmt.Errorf("performer signature is empty")
	}

	if ipfsData.PerformerSignature.PerformerSigningAddress == "" {
		return false, fmt.Errorf("performer signing address is empty")
	}

	// check if the performer is the same as the the one assigned to the task
	if ipfsData.PerformerSignature.PerformerSigningAddress != ipfsData.TaskData.PerformerData.KeeperAddress {
		return false, fmt.Errorf("performer signing address does not match the assigned performer")
	}

	// Create a copy of the ipfs data without the signature for verification
	ipfsDataForVerification := types.IPFSData{
		TaskData: ipfsData.TaskData,
		ActionData: ipfsData.ActionData,
		ProofData: ipfsData.ProofData,
		PerformerSignature: &types.PerformerSignatureData{
			PerformerSigningAddress: ipfsData.PerformerSignature.PerformerSigningAddress,
			// Note: PerformerSignature field is intentionally left empty for verification
		},
	}

	// Convert the task data to JSON message format (same as signing process)
	isValid, err := cryptography.VerifySignatureFromJSON(
		ipfsDataForVerification,
		ipfsData.PerformerSignature.PerformerSignature,
		ipfsData.PerformerSignature.PerformerSigningAddress,
	)
	if err != nil {
		return false, fmt.Errorf("failed to verify performer signature: %w", err)
	}

	return isValid, nil
}