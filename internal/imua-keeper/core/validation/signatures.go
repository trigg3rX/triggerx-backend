package validation

import (
	"encoding/json"
	"fmt"

	"github.com/prysmaticlabs/prysm/v5/crypto/bls"
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

	// Convert the ipfs data to JSON message format (same as signing process)
	ipfsDataBytes, err := json.Marshal(ipfsDataForVerification)
	if err != nil {
		logger.Error("Failed to marshal ipfs data for verification", "error", err)
		return false, fmt.Errorf("failed to marshal ipfs data for verification: %w", err)
	}

	// Parse the signature from string to BLS signature
	signatureBytes := []byte(ipfsData.PerformerSignature.PerformerSignature)
	_, err = bls.SignatureFromBytes(signatureBytes)
	if err != nil {
		logger.Error("Failed to parse BLS signature", "error", err)
		return false, fmt.Errorf("failed to parse BLS signature: %w", err)
	}

	// Parse the public key from string to BLS public key
	publicKeyBytes := []byte(ipfsData.PerformerSignature.PerformerSigningAddress)
	publicKey, err := bls.PublicKeyFromBytes(publicKeyBytes)
	if err != nil {
		logger.Error("Failed to parse BLS public key", "error", err)
		return false, fmt.Errorf("failed to parse BLS public key: %w", err)
	}

	// Hash the message to 32 bytes for BLS verification
	var messageHash [32]byte
	copy(messageHash[:], ipfsDataBytes)

	// Verify the BLS signature
	isValid, err := bls.VerifySignature(signatureBytes, messageHash, publicKey)
	if err != nil {
		logger.Error("Failed to verify BLS signature", "error", err)
		return false, fmt.Errorf("failed to verify BLS signature: %w", err)
	}

	if !isValid {
		logger.Error("Performer signature verification failed")
		return false, fmt.Errorf("performer signature verification failed")
	}

	logger.Info("Performer signature verification successful")
	return true, nil
}
