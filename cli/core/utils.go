package core

import (
	"encoding/hex"
	"fmt"
	"regexp"
	"strings"

	"github.com/cosmos/btcutil/bech32"
	"github.com/ethereum/go-ethereum/common"
	"github.com/prysmaticlabs/prysm/v5/crypto/bls"
)

const BLSMessageToSign = "BLS12-381 Signed Message\nChainIDWithoutRevision: %s\nAccAddressBech32: %s"

func ValidateBLSPrivateKey(privateKeyHex string) error {
	// Remove 0x prefix if present
	privateKeyHex = strings.TrimPrefix(privateKeyHex, "0x")

	// Check length (BLS private key should be 32 bytes = 64 hex characters)
	if len(privateKeyHex) != 64 {
		return fmt.Errorf("BLS private key should be 64 hex characters, got %d", len(privateKeyHex))
	}
	privateKeyBytes := common.FromHex(privateKeyHex)
	_, err := bls.SecretKeyFromBytes(privateKeyBytes)
	if err != nil {
		return fmt.Errorf("invalid BLS private key format: %w", err)
	}

	return nil
}

func GetBLSPublicKeyFromPrivateKey(privateKeyHex string) ([]byte, error) {
	privateKeyHex = strings.TrimPrefix(privateKeyHex, "0x")
	privateKeyBytes := common.FromHex(privateKeyHex)

	privateKey, err := bls.SecretKeyFromBytes(privateKeyBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse BLS private key: %w", err)
	}

	publicKey := privateKey.PublicKey()
	return publicKey.Marshal(), nil
}

func FormatChainIDWithoutRevision(chainID string) string {
	// Find the last dash and check if it's followed by a number (revision)
	lastDash := strings.LastIndex(chainID, "-")
	if lastDash == -1 {
		return chainID
	}

	// Check if what follows the dash is a number
	revision := chainID[lastDash+1:]
	if len(revision) > 0 {
		// If it looks like a revision number, remove it
		for _, r := range revision {
			if r < '0' || r > '9' {
				// Not a number, return original
				return chainID
			}
		}
		// All characters after dash are numbers, so it's a revision
		return chainID[:lastDash]
	}

	return chainID
}

func SwitchEthAddressToImAddress(ethAddress string) (string, error) {
	b, err := hex.DecodeString(ethAddress[2:])
	if err != nil {
		return "", fmt.Errorf("failed to decode eth address: %w", err)
	}

	// Generate im address
	bech32Prefix := "im"
	imAddress, err := bech32.EncodeFromBase256(bech32Prefix, b)
	if err != nil {
		return "", fmt.Errorf("failed to encode bech32 address: %w", err)
	}

	return imAddress, nil
}

// ChainIDWithoutRevision returns the chainID without the revision number.
// For example, "imuachaintestnet_233-1" returns "imuachaintestnet_233".
func ChainIDWithoutRevision(chainID string) string {
	if !IsRevisionFormat(chainID) {
		return chainID
	}
	splitStr := strings.Split(chainID, "-")
	return splitStr[0]
}

var IsRevisionFormat = regexp.MustCompile(`^.*[^\n-]-{1}[1-9][0-9]*$`).MatchString
