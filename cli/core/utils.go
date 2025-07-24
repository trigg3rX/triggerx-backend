package core

import (
	"encoding/hex"
	"fmt"
	"regexp"
	"strings"

	"github.com/cosmos/btcutil/bech32"
)

const BLSMessageToSign = "BLS12-381 Signed Message\nChainIDWithoutRevision: %s\nAccAddressBech32: %s"

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
