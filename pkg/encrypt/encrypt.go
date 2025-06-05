package encrypt

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/crypto/ecies"
	"crypto/rand"
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

func EncryptMessageForKeeper(publicKeyHex string, ipfsHost string, pinataJWT string) (string, error) {
	publicKeyBytes, err := hexutil.Decode(fmt.Sprintf("0x%s", publicKeyHex))
	if err != nil {
		return "", fmt.Errorf("invalid public key hex: %w", err)
	}

	pubKey, err := crypto.UnmarshalPubkey(publicKeyBytes)
	if err != nil {
		return "", fmt.Errorf("failed to unmarshal public key: %w", err)
	}

	eciesPubKey := ecies.ImportECDSAPublic(pubKey)

	message := fmt.Sprintf("%s:%s", ipfsHost, pinataJWT)

	encryptedBytes, err := ecies.Encrypt(rand.Reader, eciesPubKey, []byte(message), nil, nil)
	if err != nil {
		return "", fmt.Errorf("encryption failed: %w", err)
	}

	return hexutil.Encode(encryptedBytes), nil
}
