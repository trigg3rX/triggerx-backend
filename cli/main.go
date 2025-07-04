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
	app.Usage = "CLI for IMUA TRIGGERX AVS operator registration"
	app.Version = "1.0.0"

	app.Commands = []cli.Command{
		{
			Name:    "register-operator-with-chain",
			Aliases: []string{"chain"},
			Usage:   "registers operator with chain",
			Action:  actions.RegisterOperatorWithChain,
		},
		{
			Name:    "register-operator-with-avs",
			Aliases: []string{"avs"},
			Usage:   "registers operator with AVS (opt-in to AVS)",
			Action:  actions.RegisterOperatorWithAvs,
		},
	}

	err = app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
