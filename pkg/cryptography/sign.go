package cryptography

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
)

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