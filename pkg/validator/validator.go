package validator

import (
	"regexp"
	"strings"
)

func IsEmpty(value string) bool {
	return value == ""
}

func IsValidEmail(email string) bool {
	matched, _ := regexp.MatchString("^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}$", email)
	return matched
}

func IsValidAddress(address string) bool {
	matched, _ := regexp.MatchString("^0x[0-9a-fA-F]{40}$", address)
	return matched
}

func IsValidPrivateKey(privateKey string) bool {
	matched, _ := regexp.MatchString("^[0-9a-fA-F]{64}$", privateKey)
	return matched
}

func IsValidIPAddress(ipAddress string) bool {
	ipPattern := `^((25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)$`
	matched, _ := regexp.MatchString(ipPattern, ipAddress)
	return matched
}

func IsValidPort(port string) bool {
	matched, _ := regexp.MatchString("^(102[4-9]|10[3-9][0-9]|1[1-9][0-9]{2}|[2-9][0-9]{3}|[1-5][0-9]{4}|6[0-4][0-9]{3}|65[0-4][0-9]{2}|655[0-2][0-9]|6553[0-5])$", port)
	return matched
}

func IsValidRPCAddress(rpcAddress string) bool {
	matched, _ := regexp.MatchString("^https?://", rpcAddress)
	return matched
}

func IsValidPeerID(peerID string) bool {
	if !strings.HasPrefix(peerID, "12D3") {
		return false
	}
	return len(peerID) == len("12D3KooWBNFG1QjuF3UKAKvqhdXcxh9iBmj88cM5eU2EK5Pa91KB")
}

func IsValidRPCUrl(url string) bool {
	if url == "" {
		return false
	}

	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		return false
	}
	urlWithoutProtocol := strings.TrimPrefix(strings.TrimPrefix(url, "http://"), "https://")
	parts := strings.Split(urlWithoutProtocol, ":")
	if len(parts) != 2 {
		return false
	}

	if !IsValidIPAddress(parts[0]) {
		return false
	}
	if !IsValidPort(parts[1]) {
		return false
	}

	return true
}
