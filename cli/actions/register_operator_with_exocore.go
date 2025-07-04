package actions

import (
	"log"

	"github.com/trigg3rX/triggerx-backend/cli/core/config"
	"github.com/trigg3rX/triggerx-backend/cli/operator"
	"github.com/trigg3rX/triggerx-backend/cli/types"
	"github.com/urfave/cli"
)

func RegisterOperatorWithChain(ctx *cli.Context) error {
	log.Println("Registering operator with chain...")

	// Initialize config from environment variables
	err := config.Init()
	if err != nil {
		return err
	}

	// Create a NodeConfig from our environment config
	nodeConfig := types.NodeConfig{
		Production:                       config.GetProduction(),
		AVSOwnerAddress:                  config.GetAvsOwnerAddress().Hex(),
		OperatorAddress:                  config.GetOperatorAddress().Hex(),
		AVSAddress:                       config.GetAvsAddress().Hex(),
		EthRpcUrl:                        config.GetEthHttpRpcUrl(),
		EthWsUrl:                         config.GetEthWsRpcUrl(),
		BlsPrivateKeyStorePath:           config.GetBlsPrivateKeyStorePath(),
		OperatorEcdsaPrivateKeyStorePath: config.GetEcdsaPrivateKeyStorePath(),
		RegisterOperatorOnStartup:        false, // We don't want to register on startup when using CLI
		NodeApiIpPortAddress:             config.GetNodeApiIpPortAddress(),
		EnableNodeApi:                    config.GetEnableNodeApi(),
	}

	log.Printf("Config loaded - Operator Address: %s", nodeConfig.OperatorAddress)

	o, err := operator.NewOperatorFromConfig(nodeConfig)
	if err != nil {
		return err
	}

	err = o.RegisterOperatorWithChain()
	if err != nil {
		return err
	}

	log.Println("Successfully registered operator with chain")
	return nil
}
