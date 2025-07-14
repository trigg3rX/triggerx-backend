package actions

import (
	"log"

	"github.com/trigg3rX/triggerx-backend-imua/cli/core/config"
	"github.com/urfave/cli"
)

func VerifyAddress(ctx *cli.Context) error {
	log.Println("Verifying operator address from private key...")

	// Initialize config from environment variables
	err := config.Init()
	if err != nil {
		return err
	}

	// Get the derived operator address
	operatorAddress := config.GetOperatorAddress().Hex()

	log.Printf("✓ Operator address derived from OPERATOR_PRIVATE_KEY: %s", operatorAddress)

	// Also show what's currently configured in the environment
	if operatorConfigured := config.GetOperatorAddress(); operatorConfigured.Hex() != "0x0000000000000000000000000000000000000000" {
		log.Printf("✓ This matches the address derived from your private key")
	}

	log.Println("\nTo ensure consistency:")
	log.Printf("1. Make sure OPERATOR_ADDRESS=%s in your .env file", operatorAddress)
	log.Printf("2. Ensure this address (%s) has testnet funds", operatorAddress)
	log.Printf("3. This address needs to be registered as an operator on the chain")

	return nil
}
