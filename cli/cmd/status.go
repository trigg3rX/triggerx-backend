package cmd

import (
	"fmt"
	"os"

	"github.com/urfave/cli/v2"
	"gopkg.in/yaml.v3"

	"github.com/Layr-Labs/eigensdk-go/logging"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

func StatusCommand() *cli.Command {
	return &cli.Command{
		Name:   "status",
		Usage:  "Check keeper status",
		Action: checkKeeperStatus,
	}
}

func checkKeeperStatus(c *cli.Context) error {
	logger, err := logging.NewZapLogger(logging.Development)
	if err != nil {
		return fmt.Errorf("failed to create logger: %w", err)
	}

	logger.Info("Checking keeper status...")

	yamlFile, err := os.ReadFile("config-files/triggerx_keeper.yaml")
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	var config types.NodeConfig
	if err := yaml.Unmarshal(yamlFile, &config); err != nil {
		return fmt.Errorf("failed to unmarshal config file: %w", err)
	}

	logger.Info("Keeper Address: ", config.KeeperAddress)

	

	return nil
}

