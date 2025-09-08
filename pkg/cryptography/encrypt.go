package cryptography

import (
	"fmt"

	"crypto/rand"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/crypto/ecies"
)

func EncryptMessage(publicKeyHex string, message string) (string, error) {
	// Ensure the hex string has 0x prefix
	if len(publicKeyHex) >= 2 && publicKeyHex[:2] != "0x" {
		publicKeyHex = "0x" + publicKeyHex
	}

	publicKeyBytes, err := hexutil.Decode(publicKeyHex)
	if err != nil {
		return "", fmt.Errorf("invalid public key hex: %w", err)
	}

	pubKey, err := crypto.UnmarshalPubkey(publicKeyBytes)
	if err != nil {
		return "", fmt.Errorf("failed to unmarshal public key: %w", err)
	}

	eciesPubKey := ecies.ImportECDSAPublic(pubKey)

	encryptedBytes, err := ecies.Encrypt(rand.Reader, eciesPubKey, []byte(message), nil, nil)
	if err != nil {
		return "", fmt.Errorf("encryption failed: %w", err)
	}

	return hexutil.Encode(encryptedBytes), nil
}

func DecryptMessage(privateKey string, encryptedHex string) (string, error) {
	// Ensure the encrypted hex has 0x prefix
	if len(encryptedHex) >= 2 && encryptedHex[:2] != "0x" {
		encryptedHex = "0x" + encryptedHex
	}

	encryptedBytes, err := hexutil.Decode(encryptedHex)
	if err != nil {
		return "", fmt.Errorf("invalid encrypted hex: %w", err)
	}

	// Remove 0x prefix from private key if present (crypto.HexToECDSA doesn't accept it)
	if len(privateKey) >= 2 && privateKey[:2] == "0x" {
		privateKey = privateKey[2:]
	}

	privateKeyECDSA, err := crypto.HexToECDSA(privateKey)
	if err != nil {
		return "", fmt.Errorf("invalid private key: %w", err)
	}

	eciesPrivKey := ecies.ImportECDSA(privateKeyECDSA)
	decrypted, err := eciesPrivKey.Decrypt(encryptedBytes, nil, nil)
	if err != nil {
		return "", fmt.Errorf("decryption failed: %w", err)
	}

	return string(decrypted), nil
}
