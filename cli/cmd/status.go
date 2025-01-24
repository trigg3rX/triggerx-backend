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

func checkKeeperStatus(c *cli.Context) error {
    logger, err := logging.NewZapLogger(logging.Development)
    if err != nil {
        return fmt.Errorf("failed to create logger: %w", err)
    }

    logger.Info("Starting Keeper Status Check")

    // Config Loading
    yamlFile, err := os.ReadFile("config-files/triggerx_keeper.yaml")
    if err != nil {
        logger.Error("Failed to read config file", "error", err)
        return fmt.Errorf("failed to read config file: %w", err)
    }

    var config types.NodeConfig 
    if err := yaml.Unmarshal(yamlFile, &config); err != nil {
        logger.Error("Failed to parse config file", "error", err)
        return fmt.Errorf("failed to unmarshal config file: %w", err)
    }

    keeperAddr := common.HexToAddress(config.KeeperAddress)
    logger.Info("Keeper Configuration Loaded", 
        "keeperAddress", keeperAddr.Hex())

    client, err := ethclient.Dial(config.EthRpcUrl)
    if err != nil {
        return cli.Exit(fmt.Sprintf("Failed to connect to Ethereum client: %v", err), 1)
    }
    defer client.Close()

    registryCoordinatorContract, err := contractRegistryCoordinator.NewContractRegistryCoordinator(common.HexToAddress(config.RegistryCoordinatorAddress), client)
    if err != nil {
        return cli.Exit(fmt.Sprintf("Failed to create RegistryCoordinator contract instance: %v", err), 1)
    }

    operatorStatus, err := registryCoordinatorContract.GetOperatorStatus(&bind.CallOpts{}, keeperAddr)
    if err != nil {
        logger.Error("Failed to get operator status", "error", err)
        return fmt.Errorf("failed to get operator status: %w", err)
    }

    statusMap := map[uint8]string{
        0: "Never Registered",
        1: "Registered",
        2: "Deregistered",
    }

    delegationManager, err := contractDelegationManager.NewContractDelegationManager(common.HexToAddress(config.DelegationManagerAddress), client)
    if err != nil {
        return cli.Exit(fmt.Sprintf("Failed to create DelegationManager contract instance: %v", err), 1)
    }

    isOperator, err := delegationManager.IsOperator(&bind.CallOpts{}, keeperAddr)
    if err != nil {
        logger.Error("Failed to check operator status", "error", err)
        return fmt.Errorf("failed to check operator status: %w", err)
    }

    // Exit with non-zero status if not fully registered
    if operatorStatus != 1 || !isOperator {
        logger.Info("Keeper Registration Status", 
            "TriggerX", statusMap[operatorStatus], 
            "Eigen Layer", isOperator)
        return cli.Exit("Keeper is not fully registered", 1)
    }

    logger.Info("✅ Keeper is fully registered")
    return nil
}

func StatusCommand() *cli.Command {
    return &cli.Command{
        Name:   "status",
        Usage:  "Check keeper registration status",
        Action: checkKeeperStatus,
    }
}