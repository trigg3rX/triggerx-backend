package utils

import (
	"fmt"
	"github.com/trigg3rX/triggerx-backend/internal/keeper/config"
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

func GetExecutionContractAddress(chainID string) string {
	switch chainID {
	case "11155111":
		return "0x68605feB94a8FeBe5e1fBEF0A9D3fE6e80cEC126"
	case "11155420":
		return "0x68605feB94a8FeBe5e1fBEF0A9D3fE6e80cEC126"
	case "84532":
		return "0x68605feB94a8FeBe5e1fBEF0A9D3fE6e80cEC126"
	default:
		return ""
	}
}
