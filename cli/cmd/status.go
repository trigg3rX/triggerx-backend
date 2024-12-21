package cmd

import (
	"fmt"

	"github.com/urfave/cli/v2"
)

func StatusCommand() *cli.Command {
	return &cli.Command{
		Name:   "status",
		Usage:  "Check operator status",
		Action: checkOperatorStatus,
	}
}

func checkOperatorStatus(c *cli.Context) error {
	fmt.Println("Checking operator status...")
	return nil
}
