package cmd

import (
	"fmt"

	"github.com/urfave/cli/v2"
)

func VersionCommand() *cli.Command {
	return &cli.Command{
		Name:   "version",
		Usage:  "Display version information",
		Action: displayVersion,
	}
}

func displayVersion(c *cli.Context) error {
	fmt.Println("TriggerX AVS CLI")
	fmt.Printf("Version:      %s\n", "v1.0.0")
	fmt.Printf("Build Date:   %s\n", "2024-11-26")
	fmt.Printf("Go Version:   %s\n", "1.23.1")
	return nil
}
