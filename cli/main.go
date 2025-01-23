package main

// TODO: 
// - Use (eigensdk-go/chainio/avsregistry).UpdateSocket() to update socket
// - Add more cmds and use:
// - (eigensdk-go/chainio/avsregistry).GetOperatorsStakeInQuorumsAtCurrentBlock()
// - ().GetOperatorsStakeInQuorumsAtBlock()

import (
	"log"
	"os"

	"github.com/trigg3rX/triggerx-backend/cli/cmd"
	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name:  "triggerx",
		Usage: "TriggerX Keeper CLI - Tool for managing keepers on TriggerX",
		Commands: []*cli.Command{
			cmd.RegisterCommand(),
			cmd.DeregisterCommand(),
			cmd.StatusCommand(),
			cmd.VersionCommand(),
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
