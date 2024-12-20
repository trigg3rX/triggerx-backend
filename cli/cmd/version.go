package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
)

func displayVersion() {
	fmt.Println("TriggerX AVS CLI")
	fmt.Printf("Version:      %s\n", "v1.0.0")
	fmt.Printf("Build Date:   %s\n", "2024-11-26")
	fmt.Printf("Go Version:   %s\n", "1.23.1")
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Display version information",
	Run: func(cmd *cobra.Command, args []string) {
		displayVersion()
	},
} 