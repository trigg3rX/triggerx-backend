package encrypt

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/crypto/ecies"
	"crypto/rand"
)

func EncryptMessage(publicKeyHex string, message string) (string, error) {
	publicKeyBytes, err := hexutil.Decode(fmt.Sprintf("0x%s", publicKeyHex))
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

func DecryptMessageForKeeper(privateKey string, encryptedHex string) (string, error) {
	encryptedBytes, err := hexutil.Decode(encryptedHex)
	if err != nil {
		return "", fmt.Errorf("invalid encrypted hex: %w", err)
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