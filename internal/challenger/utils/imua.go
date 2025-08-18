package utils

import (
	"encoding/hex"
	"fmt"
	"github.com/btcsuite/btcutil/bech32"
)

func SwitchEthAddressToImAddress(ethAddress string) (string, error) {
	b, err := hex.DecodeString(ethAddress[2:])
	if err != nil {
		return "", fmt.Errorf("failed to decode eth address: %w", err)
	}

	// Generate im address
	bech32Prefix := "im"
	imAddress, err := bech32.Encode(bech32Prefix, b)
	if err != nil {
		return "", fmt.Errorf("failed to encode bech32 address: %w", err)
	}

	return imAddress, nil
}
