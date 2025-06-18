package validation

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"github.com/trigg3rX/triggerx-backend/internal/keeper/config"
	"github.com/trigg3rX/triggerx-backend/pkg/proof"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

func (v *TaskValidator) ValidateProof(ipfsData types.IPFSData, traceID string) (bool, error) {
	proofData := ipfsData.ProofData
	if proofData == nil {
		return false, fmt.Errorf("proof data is missing")
	}
	if proofData.ProofOfTask == "" {
		return false, fmt.Errorf("proof of task is empty")
	}
	if proofData.CertificateHash == "" {
		return false, fmt.Errorf("certificate hash is empty")
	}

	// Validate TLS certificate by establishing connection to the same host/port
	// that was used during proof generation
	tlsConfig := proof.DefaultTLSProofConfig(config.GetTLSProofHost())
	tlsConfig.TargetPort = config.GetTLSProofPort()

	connState, err := proof.EstablishTLSConnection(tlsConfig)
	if err != nil {
		v.logger.Warn("Failed to establish TLS connection for validation", "trace_id", traceID, "error", err)
	}
	if connState == nil {
		return false, fmt.Errorf("failed to establish TLS connection for validation")
	}

	// Verify the certificate hash matches what was recorded in the proof
	if len(connState.PeerCertificates) == 0 {
		return false, fmt.Errorf("no peer certificates found during validation")
	}

	currentCertHash := sha256.Sum256(connState.PeerCertificates[0].Raw)
	currentCertHashStr := hex.EncodeToString(currentCertHash[:])

	if currentCertHashStr != proofData.CertificateHash {
		v.logger.Warn("Certificate hash mismatch during validation",
			"trace_id", traceID,
			"expected", proofData.CertificateHash,
			"actual", currentCertHashStr)
		// Certificate might have been renewed, this is not necessarily an error
		// but should be logged for investigation
	}

	// Validate the proof hash by regenerating it
	return v.validateProofHash(ipfsData, traceID)
}

func (v *TaskValidator) validateProofHash(ipfsData types.IPFSData, traceID string) (bool, error) {
	// Create a copy of IPFS data without the proof for hash validation
	ipfsDataForValidation := types.IPFSData{
		TaskData:           ipfsData.TaskData,
		ActionData:         ipfsData.ActionData,
		ProofData:          &types.ProofData{},
		PerformerSignature: &types.PerformerSignatureData{},
	}
	ipfsDataForValidation.ProofData.TaskID = ipfsData.TaskData.TaskID
	ipfsDataForValidation.PerformerSignature.TaskID = ipfsData.TaskData.TaskID
	ipfsDataForValidation.PerformerSignature.PerformerSigningAddress = ipfsData.PerformerSignature.PerformerSigningAddress

	// Regenerate the proof hash
	dataStr, err := proof.StringifyIPFSData(ipfsDataForValidation)
	if err != nil {
		return false, fmt.Errorf("failed to stringify IPFS data for validation: %w", err)
	}

	expectedProofHash := sha256.Sum256([]byte(dataStr))
	expectedProofHashStr := hex.EncodeToString(expectedProofHash[:])

	if expectedProofHashStr != ipfsData.ProofData.ProofOfTask {
		return false, fmt.Errorf("proof hash validation failed: expected %s, got %s",
			expectedProofHashStr, ipfsData.ProofData.ProofOfTask)
	}

	v.logger.Info("Proof validation passed", "trace_id", traceID, "task_id", ipfsData.TaskData.TaskID)
	return true, nil
}
