package types

import "regexp"

// Ethereum Address
func IsValidEthAddress(address string) bool {
	matched, _ := regexp.MatchString("^0x[0-9a-fA-F]{40}$", address)
	return matched
}

// Job ID
func IsValidJobID(jobID string) bool {
	matched, _ := regexp.MatchString("^[1-9][0-9]*$", jobID)
	return matched
}

// Chain ID
func IsValidChainID(chainID string) bool {
	if chainID == "1" || chainID == "8453" || chainID == "10" || chainID == "42161" {
		return true
	}
	if chainID == "11155111" || chainID == "84532" || chainID == "11155420" || chainID == "421614" {
		return true
	}
	return false
}