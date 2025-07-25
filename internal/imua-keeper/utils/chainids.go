package utils

import (
	"fmt"

	"github.com/trigg3rX/triggerx-backend/internal/imua-keeper/config"
)

func GetChainRpcUrl(chainID string) string {
	switch chainID {
	case "11155111":
		return fmt.Sprintf("https://eth-sepolia.g.alchemy.com/v2/%s", config.GetAlchemyAPIKey())
	case "11155420":
		return fmt.Sprintf("https://opt-sepolia.g.alchemy.com/v2/%s", config.GetAlchemyAPIKey())
	case "84532":
		return fmt.Sprintf("https://base-sepolia.g.alchemy.com/v2/%s", config.GetAlchemyAPIKey())
	default:
		return ""
	}
}
