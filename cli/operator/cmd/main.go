package main

import (
	"context"
	"encoding/json"
	"log"
	"os"

	"github.com/urfave/cli"

	"github.com/trigg3rX/triggerx-backend/cli/core/config"
	"github.com/trigg3rX/triggerx-backend/cli/operator"
	"github.com/trigg3rX/triggerx-backend/cli/types"

	sdkutils "github.com/imua-xyz/imua-avs-sdk/utils"
)

func main() {
	app := cli.NewApp()
	app.Flags = []cli.Flag{config.FileFlag}
	app.Name = "triggerx-operator"
	app.Usage = "triggerx Operator"
	app.Description = "Service that operator listens to AVS contract events, signs tasks, and submits results."

	app.Action = operatorMain
	err := app.Run(os.Args)
	if err != nil {
		log.Fatalln("Application failed. Message:", err)
	}
}

func operatorMain(ctx *cli.Context) error {
	log.Println("Initializing Operator")
	configPath := ctx.GlobalString(config.FileFlag.Name)
	nodeConfig := types.NodeConfig{}
	err := sdkutils.ReadYamlConfig(configPath, &nodeConfig)
	if err != nil {
		return err
	}
	configJson, err := json.MarshalIndent(nodeConfig, "", "  ")
	if err != nil {
		log.Fatalf(err.Error())
	}
	log.Println("Config:", string(configJson))

	log.Println("initializing operator")
	operator, err := operator.NewOperatorFromConfig(nodeConfig)
	if err != nil {
		return err
	}
	log.Println("initialized operator")

	log.Println("starting operator")
	err = operator.Start(context.Background())
	if err != nil {
		return err
	}
	log.Println("started operator")

	return nil
}
