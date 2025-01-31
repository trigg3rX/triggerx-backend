package cmd

import (
	"fmt"
	"os"

	"github.com/ethereum/go-ethereum/common"
	"github.com/urfave/cli/v2"
	"gopkg.in/yaml.v3"
	"github.com/Layr-Labs/eigensdk-go/logging"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
    contractRegistryCoordinator "github.com/Layr-Labs/eigensdk-go/contracts/bindings/RegistryCoordinator"
    contractDelegationManager "github.com/Layr-Labs/eigensdk-go/contracts/bindings/DelegationManager"
)

// checkKeeperStatus verifies the registration status of a keeper both on TriggerX and EigenLayer.
// It checks if the keeper is properly registered as an operator in both systems.
// Returns an error if the keeper is not fully registered or if any checks fail.
func checkKeeperStatus(c *cli.Context) error {
    logger, err := logging.NewZapLogger(logging.Development)
    if err != nil {
        return fmt.Errorf("failed to create logger: %w", err)
    }

    logger.Info("keeper status check initiated")

    yamlFile, err := os.ReadFile(c.String("config"))
    if err != nil {
        logger.Error("config file read failed", "error", err)
        return fmt.Errorf("failed to read config file: %w", err)
    }

    var config types.NodeConfig 
    if err := yaml.Unmarshal(yamlFile, &config); err != nil {
        logger.Error("config file parse failed", "error", err)
        return fmt.Errorf("failed to unmarshal config file: %w", err)
    }

    keeperAddr := common.HexToAddress(config.KeeperAddress)

    client, err := ethclient.Dial(config.EthRpcUrl)
    if err != nil {
        return cli.Exit(fmt.Sprintf("Failed to connect to Ethereum client: %v", err), 1)
    }
    defer client.Close()

    registryCoordinatorContract, err := contractRegistryCoordinator.NewContractRegistryCoordinator(common.HexToAddress(config.RegistryCoordinatorAddress), client)
    if err != nil {
        return cli.Exit(fmt.Sprintf("Failed to create RegistryCoordinator contract instance: %v", err), 1)
    }

    delegationManager, err := contractDelegationManager.NewContractDelegationManager(common.HexToAddress(config.DelegationManagerAddress), client)
    if err != nil {
        return cli.Exit(fmt.Sprintf("Failed to create DelegationManager contract instance: %v", err), 1)
    }

    logger.Info("keeper configuration loaded", "address", keeperAddr.Hex())

    operatorStatus, err := registryCoordinatorContract.GetOperatorStatus(&bind.CallOpts{}, keeperAddr)
    if err != nil {
        logger.Error("operator status check failed", "error", err)
        return fmt.Errorf("failed to get operator status: %w", err)
    }

    statusMap := map[uint8]string{
        0: "Never Registered",
        1: "Registered",
        2: "Deregistered",
    }

    isOperator, err := delegationManager.IsOperator(&bind.CallOpts{}, keeperAddr)
    if err != nil {
        logger.Error("operator check failed", "error", err)
        return fmt.Errorf("failed to check operator status: %w", err)
    }

    if operatorStatus != 1 || !isOperator {
        logger.Info("keeper registration incomplete", 
            "triggerx_status", statusMap[operatorStatus], 
            "eigen_layer_status", isOperator)
        return cli.Exit("keeper is not fully registered", 1)
    }

    logger.Info("keeper registration verified", "status", "fully registered")
    return nil
}

// StatusCommand returns a CLI command that checks the registration status of a keeper.
// The command verifies the keeper's registration on both TriggerX and EigenLayer systems.
func StatusCommand() *cli.Command {
    return &cli.Command{
        Name:   "status",
        Usage:  "Check keeper registration status",
        Action: checkKeeperStatus,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "config",
				Usage: "Path to the config file (triggerx_keeper.yaml)",
				Required: true,
			},
		},
    }
}