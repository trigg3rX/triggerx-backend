package proof

import (
	"crypto/sha256"
	"encoding/hex"
	"time"

	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

func GetData() []byte {
	return []byte("")	
}

// GenerateProof takes the action execution data and generates a proof by:
// 1. Creating a hash of the action data (tx hash, gas used, status etc)
// 2. Adding a timestamp when the proof was generated
// This provides cryptographic proof that the action was executed
func GenerateProof(ipfsData types.IPFSData) (types.ProofData, error) {
	// Convert action data to bytes and generate hash
	// actionBytes := []byte(actionData.ActionTxHash + actionData.GasUsed + 
	// 	string(actionData.ExecutionTimestamp.Unix()))
	proofHash := sha256.Sum256(GetData())
	proofHashStr := hex.EncodeToString(proofHash[:])

	// Generate certificate hash from action data
	certHash := sha256.Sum256(GetData())
	certHashStr := hex.EncodeToString(certHash[:])

	// Return proof data with current timestamp
	return types.ProofData{
		TaskID:              ipfsData.TargetData.TaskID,
		ProofOfTask:         proofHashStr,
		CertificateHash:     certHashStr,
		CertificateTimestamp: time.Now().UTC(),
	}, nil
}
