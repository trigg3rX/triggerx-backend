package validation

import (
	"fmt"

	"github.com/trigg3rX/triggerx-backend/internal/imua-keeper/config"
	"github.com/trigg3rX/triggerx-backend/pkg/cryptography"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

func (v *TaskValidator) ValidateManagerSignature(task *types.SendTaskDataToKeeper, traceID string) (bool, error) {
	logger := v.logger.With("traceID", traceID)

	// check if the manager signature is valid
	if task.ManagerSignature == "" {
		logger.Error("Manager signature data is missing")
		return false, fmt.Errorf("manager signature data is missing")
	}

	// Create a copy of the task data without the signature for verification
	taskDataForVerification := types.SendTaskDataToKeeper{
		TaskID:        task.TaskID,
		PerformerData: task.PerformerData,
		TargetData:    task.TargetData,
		TriggerData:   task.TriggerData,
		SchedulerID:   task.SchedulerID,
	}

	// Convert the task data to JSON message format (same as signing process)
	isValid, err := cryptography.VerifySignatureFromJSON(
		taskDataForVerification,
		task.ManagerSignature,
		config.GetManagerSigningAddress(),
	)
	if err != nil {
		logger.Error("Failed to verify manager signature", "error", err)
		return false, fmt.Errorf("failed to verify manager signature: %w", err)
	}

	if !isValid {
		logger.Error("Manager signature verification failed")
		return false, fmt.Errorf("manager signature verification failed")
	}

	logger.Info("Manager signature verification successful")
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
			TaskID:                  ipfsData.PerformerSignature.TaskID,
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
