package security

import (
	"fmt"

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

	signature, err := crypto.Sign(messageHash.Bytes(), privateKeyECDSA)
	if err != nil {
		return "", fmt.Errorf("failed to sign message: %w", err)
	}

	signature[64] += 27

	return hexutil.Encode(signature), nil
}

func VerifySignature(message string, signatureHex string, expectedAddress string) (bool, error) {
	signature, err := hexutil.Decode(signatureHex)
	if err != nil {
		return false, fmt.Errorf("invalid signature: %w", err)
	}

	if len(signature) != 65 {
		return false, fmt.Errorf("invalid signature length")
	}

	if signature[64] >= 27 {
		signature[64] -= 27
	}

	messageHash := crypto.Keccak256Hash([]byte(fmt.Sprintf("\x19Ethereum Signed Message:\n%d%s", len(message), message)))

	pubKeyRaw, err := crypto.Ecrecover(messageHash.Bytes(), signature)
	if err != nil {
		return false, fmt.Errorf("failed to recover public key: %w", err)
	}

	pubKey, err := crypto.UnmarshalPubkey(pubKeyRaw)
	if err != nil {
		return false, fmt.Errorf("failed to unmarshal public key: %w", err)
	}

	recoveredAddr := crypto.PubkeyToAddress(*pubKey)

	checksumAddr := common.HexToAddress(expectedAddress)

	return checksumAddr == recoveredAddr, nil
}