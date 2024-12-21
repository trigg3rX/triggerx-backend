package main

import (
	"log"
	"os"

	"github.com/trigg3rX/triggerx-backend/cli/cmd"
	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name:  "triggerx",
		Usage: "TriggerX AVS CLI - Tool for managing operators on AVS",
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
