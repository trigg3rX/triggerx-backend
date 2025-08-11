package utils

import (
	"fmt"
	"github.com/trigg3rX/triggerx-backend/internal/keeper/config"
)

func GetChainRpcUrl(chainID string) string {
	switch chainID {
	// Testnets
	case "11155111":
		return fmt.Sprintf("https://eth-sepolia.g.alchemy.com/v2/%s", config.GetAlchemyAPIKey())
	case "11155420":
		return fmt.Sprintf("https://opt-sepolia.g.alchemy.com/v2/%s", config.GetAlchemyAPIKey())
	case "84532":
		return fmt.Sprintf("https://base-sepolia.g.alchemy.com/v2/%s", config.GetAlchemyAPIKey())
	case "421614":
		return fmt.Sprintf("https://arb-sepolia.g.alchemy.com/v2/%s", config.GetAlchemyAPIKey())

	// Mainnets
	case "1":
		return fmt.Sprintf("https://eth-mainnet.g.alchemy.com/v2/%s", config.GetAlchemyAPIKey())
	case "10":
		return fmt.Sprintf("https://opt-mainnet.g.alchemy.com/v2/%s", config.GetAlchemyAPIKey())
	case "8453":
		return fmt.Sprintf("https://base-mainnet.g.alchemy.com/v2/%s", config.GetAlchemyAPIKey())
	case "42161":
		return fmt.Sprintf("https://arb-mainnet.g.alchemy.com/v2/%s", config.GetAlchemyAPIKey())
	default:
		return ""
	}
}
