package actions

import (
	"encoding/json"
	"log"
	"os"

	sdkecdsa "github.com/imua-xyz/imua-avs-sdk/crypto/ecdsa"
	sdkutils "github.com/imua-xyz/imua-avs-sdk/utils"
	"github.com/trigg3rX/triggerx-backend/core/config"
	"github.com/trigg3rX/triggerx-backend/operator"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
	"github.com/urfave/cli"
)

func RegisterOperatorWithAvs(ctx *cli.Context) error {

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

	ecdsaKeyPassword, ok := os.LookupEnv("OPERATOR_ECDSA_KEY_PASSWORD")
	if !ok {
		log.Printf("OPERATOR_ECDSA_KEY_PASSWORD env var not set. using empty string")
	}
	operatorEcdsaPrivKey, err := sdkecdsa.ReadKey(
		nodeConfig.OperatorEcdsaPrivateKeyStorePath,
		ecdsaKeyPassword,
	)
	if err != nil {
		return err
	}
	log.Printf(operatorEcdsaPrivKey.D.String())

	err = o.RegisterOperatorWithAvs()
	if err != nil {
		return err
	}

	return nil
}