package main

import (
	"log"
	"os"

	"github.com/trigg3rX/triggerx-backend/cli/actions"
	"github.com/trigg3rX/triggerx-backend/cli/core/config"
	"github.com/urfave/cli"
)

func main() {
	err := config.Init()
	if err != nil {
		log.Fatal(err)
	}

	app := cli.NewApp()
	app.Name = "triggerx"
	app.Usage = "TriggerX Operator CLI"

	app.Commands = []cli.Command{
		{
			Name:   "generate-keys",
			Usage:  "Generate BLS and ECDSA keystore files",
			Action: actions.GenerateKeys,
		},
		{
			Name:   "register-operator-with-chain",
			Usage:  "Register operator with the chain",
			Action: actions.RegisterOperatorWithChain,
		},
		{
			Name:   "register-operator-with-avs",
			Usage:  "Register operator with AVS",
			Action: actions.RegisterOperatorWithAvs,
		},
		{
			Name:   "complete-registration",
			Usage:  "Generate keystores from existing keys and complete entire registration process",
			Action: actions.CompleteRegistration,
		},
		{
			Name:   "verify-address",
			Usage:  "Verify the operator address derived from the private key in environment",
			Action: actions.VerifyAddress,
		},
	}

	err = app.Run(os.Args)
	if err != nil {
		panic(err)
	}
}
