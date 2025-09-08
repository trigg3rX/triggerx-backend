package utils

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/trigg3rX/triggerx-backend/internal/keeper/config"
)

func TestGetChainRpcUrl(t *testing.T) {
	tests := []struct {
		name     string
		chainID  string
		expected string
	}{
		{
			name:     "Ethereum Sepolia",
			chainID:  "11155111",
			expected: fmt.Sprintf("https://eth-sepolia.g.alchemy.com/v2/%s", config.GetAlchemyAPIKey()),
		},
		{
			name:     "Optimism Sepolia",
			chainID:  "11155420",
			expected: fmt.Sprintf("https://opt-sepolia.g.alchemy.com/v2/%s", config.GetAlchemyAPIKey()),
		},
		{
			name:     "Base Sepolia",
			chainID:  "84532",
			expected: fmt.Sprintf("https://base-sepolia.g.alchemy.com/v2/%s", config.GetAlchemyAPIKey()),
		},
		{
			name:     "Arbitrum Sepolia",
			chainID:  "421614",
			expected: fmt.Sprintf("https://arb-sepolia.g.alchemy.com/v2/%s", config.GetAlchemyAPIKey()),
		},
		{
			name:     "Unknown chain ID",
			chainID:  "12345",
			expected: "",
		},
		{
			name:     "Empty chain ID",
			chainID:  "",
			expected: "",
		},
		{
			name:     "Malformed chain ID",
			chainID:  "abc123",
			expected: "",
		},
		{
			name:     "Ethereum Mainnet (not supported)",
			chainID:  "1",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := GetChainRpcUrl(tt.chainID)
			assert.Equal(t, tt.expected, actual)
		})
	}
}
