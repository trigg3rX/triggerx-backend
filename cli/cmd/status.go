package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
)

func checkOperatorStatus() {
	fmt.Println("Checking operator status...")
}

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Check operator status",
	Run: func(cmd *cobra.Command, args []string) {
		checkOperatorStatus()
	},
}