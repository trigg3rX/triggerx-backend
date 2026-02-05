package cryptography

import (
	"encoding/json"
	"fmt"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
)

// SignContractExecution creates a signature that the TaskExecutionHub contract can verify.
// This signature authorizes a specific keeper to execute a specific job.
// The hash is constructed to match the contract's verification:
//
//	keccak256(abi.encode(jobId, target, deadline, keeperAddress, chainId))
//
// Parameters:
//   - privateKey: hex-encoded private key (without 0x prefix)
//   - jobId: the job ID for this execution
//   - target: the target contract address for the execution
//   - deadline: unix timestamp when this signature expires
//   - keeperAddress: the keeper authorized to submit this execution
//   - chainId: the chain ID where execution happens (cross-chain replay protection)
//
// Returns a 65-byte signature with v-value adjusted for Ethereum (+27)
func SignContractExecution(
	privateKey string,
	jobId *big.Int,
	target common.Address,
	deadline *big.Int,
	keeperAddress common.Address,
	chainId *big.Int,
) ([]byte, error) {
	privateKeyECDSA, err := crypto.HexToECDSA(privateKey)
	if err != nil {
		return nil, fmt.Errorf("invalid private key: %w", err)
	}

	// Define ABI types for encoding
	uint256Ty, _ := abi.NewType("uint256", "", nil)
	addressTy, _ := abi.NewType("address", "", nil)

	// ABI encode the parameters in the exact order the contract expects:
	// keccak256(abi.encode(jobId, target, deadline, msg.sender, block.chainid))
	arguments := abi.Arguments{
		{Type: uint256Ty}, // jobId
		{Type: addressTy}, // target
		{Type: uint256Ty}, // deadline
		{Type: addressTy}, // keeperAddress (msg.sender in contract)
		{Type: uint256Ty}, // chainId
	}

	packed, err := arguments.Pack(jobId, target, deadline, keeperAddress, chainId)
	if err != nil {
		return nil, fmt.Errorf("failed to ABI encode parameters: %w", err)
	}

	// Hash the packed data
	hash := crypto.Keccak256(packed)

	// Apply Ethereum signed message prefix (matching toEthSignedMessageHash in contract)
	prefixedHash := crypto.Keccak256([]byte(fmt.Sprintf("\x19Ethereum Signed Message:\n%d%s", len(hash), string(hash))))

	// Sign the prefixed hash
	signature, err := crypto.Sign(prefixedHash, privateKeyECDSA)
	if err != nil {
		return nil, fmt.Errorf("failed to sign: %w", err)
	}

	// Adjust v value for Ethereum (add 27)
	signature[64] += 27

	return signature, nil
}

func SignMessage(message string, privateKey string) (string, error) {
	privateKeyECDSA, err := crypto.HexToECDSA(privateKey)
	if err != nil {
		return "", fmt.Errorf("invalid private key: %w", err)
	}

	messageHash := crypto.Keccak256Hash([]byte(fmt.Sprintf("\x19Ethereum Signed Message:\n%d%s", len(message), message)))
	messageBytes := messageHash.Bytes()

	signature, err := crypto.Sign(messageBytes, privateKeyECDSA)
	if err != nil {
		return "", fmt.Errorf("failed to sign message: %w", err)
	}

	signature[64] += 27

	return hexutil.Encode(signature), nil
}

func SignJSONMessage(jsonData interface{}, privateKey string) (string, error) {
	jsonDataMap := make(map[string]interface{})

	jsonBytes, err := json.Marshal(jsonData)
	if err != nil {
		return "", fmt.Errorf("failed to marshal input data: %w", err)
	}

	if err := json.Unmarshal(jsonBytes, &jsonDataMap); err != nil {
		return "", fmt.Errorf("failed to unmarshal to map: %w", err)
	}

	convertToLower(jsonDataMap)

	jsonDataBytes, err := json.Marshal(jsonDataMap)
	if err != nil {
		return "", fmt.Errorf("failed to marshal json data: %w", err)
	}

	message := string(jsonDataBytes)

	return SignMessage(message, privateKey)
}

func VerifySignature(message string, signature string, signerAddress string) (bool, error) {
	messageHash := crypto.Keccak256Hash([]byte(fmt.Sprintf("\x19Ethereum Signed Message:\n%d%s", len(message), message)))
	messageBytes := messageHash.Bytes()

	signatureBytes, err := hexutil.Decode(signature)
	if err != nil {
		return false, fmt.Errorf("invalid signature: %w", err)
	}

	if len(signatureBytes) != 65 {
		return false, fmt.Errorf("invalid signature length")
	}

	if signatureBytes[64] >= 27 {
		signatureBytes[64] -= 27
	}

	pubKeyRaw, err := crypto.Ecrecover(messageBytes, signatureBytes)
	if err != nil {
		return false, fmt.Errorf("failed to recover public key: %w", err)
	}

	pubKey, err := crypto.UnmarshalPubkey(pubKeyRaw)
	if err != nil {
		return false, fmt.Errorf("failed to unmarshal public key: %w", err)
	}

	recoveredAddr := crypto.PubkeyToAddress(*pubKey)

	checksumAddr := common.HexToAddress(signerAddress)

	return checksumAddr == recoveredAddr, nil
}

func VerifySignatureFromJSON(jsonData interface{}, signature string, signerAddress string) (bool, error) {
	jsonDataMap := make(map[string]interface{})

	jsonBytes, err := json.Marshal(jsonData)
	if err != nil {
		return false, fmt.Errorf("failed to marshal input data: %w", err)
	}

	if err := json.Unmarshal(jsonBytes, &jsonDataMap); err != nil {
		return false, fmt.Errorf("failed to unmarshal to map: %w", err)
	}

	convertToLower(jsonDataMap)

	jsonDataBytes, err := json.Marshal(jsonDataMap)
	if err != nil {
		return false, fmt.Errorf("failed to marshal json data: %w", err)
	}

	message := string(jsonDataBytes)

	return VerifySignature(message, signature, signerAddress)
}

func convertToLower(data map[string]interface{}) {
	for k, v := range data {
		if s, ok := v.(string); ok {
			data[k] = strings.ToLower(s)
		} else if m, ok := v.(map[string]interface{}); ok {
			convertToLower(m)
		}
	}
}
