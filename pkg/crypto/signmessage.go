package crypto

import (
	// "crypto/ecdsa"
	"fmt"
	// "math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
)

// signMessage signs a message using the keeper's private key
func SignMessage(message string, privateKey string) (string, error) {
	// Convert private key from hex to ECDSA private key
	privateKeyECDSA, err := crypto.HexToECDSA(privateKey)
	if err != nil {
		return "", fmt.Errorf("invalid private key: %w", err)
	}

	// Create Ethereum specific signature:
	// keccak256("\x19Ethereum Signed Message:\n" + len(message) + message))
	messageHash := crypto.Keccak256Hash([]byte(fmt.Sprintf("\x19Ethereum Signed Message:\n%d%s", len(message), message)))

	// Sign the hashed message
	signature, err := crypto.Sign(messageHash.Bytes(), privateKeyECDSA)
	if err != nil {
		return "", fmt.Errorf("failed to sign message: %w", err)
	}

	// Add recovery id to the signature (v = sig[64] + 27)
	signature[64] += 27

	// Convert signature to hex string format
	return hexutil.Encode(signature), nil
}

// verifySignature verifies if the signature was signed by the expected address
func VerifySignature(message string, signatureHex string, expectedAddress string) (bool, error) {
	// Decode the signature from hex
	signature, err := hexutil.Decode(signatureHex)
	if err != nil {
		return false, fmt.Errorf("invalid signature: %w", err)
	}

	// Ensure signature is in the right format
	if len(signature) != 65 {
		return false, fmt.Errorf("invalid signature length")
	}

	// Ethereum's signature recovery ID adjustment
	if signature[64] >= 27 {
		signature[64] -= 27
	}

	// Create the message hash as done during signing
	messageHash := crypto.Keccak256Hash([]byte(fmt.Sprintf("\x19Ethereum Signed Message:\n%d%s", len(message), message)))

	// Recover public key from signature
	pubKeyRaw, err := crypto.Ecrecover(messageHash.Bytes(), signature)
	if err != nil {
		return false, fmt.Errorf("failed to recover public key: %w", err)
	}

	pubKey, err := crypto.UnmarshalPubkey(pubKeyRaw)
	if err != nil {
		return false, fmt.Errorf("failed to unmarshal public key: %w", err)
	}

	// Derive the Ethereum address from the public key
	recoveredAddr := crypto.PubkeyToAddress(*pubKey)

	// Convert expected address to checksum format for comparison
	checksumAddr := common.HexToAddress(expectedAddress)

	// Compare the recovered address with expected address
	return checksumAddr == recoveredAddr, nil
}