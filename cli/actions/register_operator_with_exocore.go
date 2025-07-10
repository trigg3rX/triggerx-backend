package actions

import (
	"encoding/json"
	"github.com/urfave/cli"
	"log"

	sdkutils "github.com/imua-xyz/imua-avs-sdk/utils"
	"github.com/trigg3rX/triggerx-backend/core/config"
	"github.com/trigg3rX/triggerx-backend/cli/operator"
	"github.com/trigg3rX/triggerx-backend/cli/types"
)

func RegisterOperatorWithChain(ctx *cli.Context) error {

	configPath := ctx.GlobalString(config.FileFlag.Name)
	nodeConfig := types.NodeConfig{}
	err := sdkutils.ReadYamlConfig(configPath, &nodeConfig)
	if err != nil {
		return err
	}
	// need to make sure we don't register the operator on startup
	// when using the cli commands to register the operator.
	nodeConfig.RegisterOperatorOnStartup = false
	configJson, err := json.MarshalIndent(nodeConfig, "", "  ")
	if err != nil {
		log.Fatalf(err.Error())
	}
	log.Println("Config:", string(configJson))

	o, err := operator.NewOperatorFromConfig(nodeConfig)
	if err != nil {
		return err
	}

	err = o.RegisterOperatorWithChain()
	if err != nil {
		return err
	}

	return nil
}