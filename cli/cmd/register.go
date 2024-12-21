package cmd

import (
	"fmt"
	"math/big"
	"os"

	"github.com/urfave/cli/v2"

	"github.com/Layr-Labs/eigensdk-go/crypto/ecdsa"
	"github.com/ethereum/go-ethereum/common"

	"github.com/trigg3rX/triggerx-backend/pkg/keeper"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

func RegisterCommand() *cli.Command {
	return &cli.Command{
		Name:  "register",
		Usage: "Register a new operator",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "keystore-path",
				Usage:   "Path to the keystore file",
				EnvVars: []string{"KEYSTORE_PATH"},
			},
			&cli.StringFlag{
				Name:    "eth-rpc-url",
				Usage:   "Ethereum RPC URL",
				EnvVars: []string{"ETH_RPC_URL"},
			},
			&cli.StringFlag{
				Name:    "eth-ws-url",
				Usage:   "Ethereum WebSocket URL",
				EnvVars: []string{"ETH_WS_URL"},
			},
			&cli.StringFlag{
				Name:    "bls-keystore-path",
				Usage:   "Path to the BLS keystore file",
				EnvVars: []string{"BLS_KEYSTORE_PATH"},
			},
			&cli.StringFlag{
				Name:    "keeper-address",
				Usage:   "Keeper address",
				EnvVars: []string{"KEEPER_ADDRESS"},
			},
			&cli.StringFlag{
				Name:    "avs-registry-coordinator-address",
				Usage:   "AVS Registry Coordinator contract address",
				EnvVars: []string{"AVS_REGISTRY_COORDINATOR_ADDRESS"},
			},
			&cli.StringFlag{
				Name:    "operator-state-retriever-address",
				Usage:   "Operator State Retriever contract address",
				EnvVars: []string{"OPERATOR_STATE_RETRIEVER_ADDRESS"},
			},
			&cli.StringFlag{
				Name:    "metrics-address",
				Usage:   "Metrics server address",
				EnvVars: []string{"METRICS_ADDRESS"},
			},
			&cli.StringFlag{
				Name:    "api-address",
				Usage:   "API server address",
				EnvVars: []string{"API_ADDRESS"},
			},
		},
		Action: registerOperator,
	}
}

func DeregisterCommand() *cli.Command {
	return &cli.Command{
		Name:   "deregister",
		Usage:  "Deregister an operator",
		Action: deregisterOperator,
	}
}

func registerOperator(c *cli.Context) error {
	fmt.Println(
		`Register a new operator to the TriggerX AVS. It is assumed that the operator has already been registered to the EigenLayer protocol.
If not, please register to EigenLayer first. Follow the instructions here: 
https://github.com/Layr-Labs/eigenlayer-cli/blob/master/README.md`)

	keystorePath := c.String("keystore-path")
	if keystorePath == "" {
		return cli.Exit("keystore-path flag is required", 1)
	}

	// Check if keystore file exists
	if _, err := os.Stat(keystorePath); os.IsNotExist(err) {
		return cli.Exit(fmt.Sprintf("keystore file not found at path: %s", keystorePath), 1)
	}

	passphrase := "pL6!oK5@iJ4#"
	mockTokenStrategyAddr := common.HexToAddress("0x80528D6e9A2BAbFc766965E0E26d5aB08D9CFaF9")

	privKey, err := ecdsa.ReadKey(keystorePath, passphrase)
	if err != nil {
		return cli.Exit(fmt.Sprintf("Error reading keystore file: %s", err), 1)
	}

	// Initialize keeper with NodeConfig
	config := types.NodeConfig{
		EcdsaPrivateKeyStorePath:      keystorePath,
		Passphrase:                    passphrase,
		EthRpcUrl:                     c.String("eth-rpc-url"), // Add these flags to your CLI
		EthWsUrl:                      c.String("eth-ws-url"),
		BlsPrivateKeyStorePath:        c.String("bls-keystore-path"),
		KeeperAddress:                 c.String("keeper-address"),
		ServiceManagerAddress:         c.String("service-manager-address"),
		OperatorStateRetrieverAddress: c.String("operator-state-retriever-address"),
		EnableMetrics:                 true,
		EigenMetricsIpPortAddress:     c.String("metrics-address"),
		NodeApiIpPortAddress:          c.String("api-address"),
		Production:                    false,
	}

	keeper, err := keeper.NewKeeperFromConfig(config)
	if err != nil {
		return cli.Exit(fmt.Sprintf("Error initializing keeper: %s", err), 1)
	}

	amount := big.NewInt(1)

	// Now use the keeper instance
	err = keeper.DepositIntoStrategy(mockTokenStrategyAddr, amount)
	if err != nil {
		return cli.Exit(fmt.Sprintf("Error depositing into strategy: %s", err), 1)
	}
	keeper.Logger.Infof("Deposited %s into strategy %s", amount, mockTokenStrategyAddr)

	err = keeper.RegisterOperatorWithAvs(privKey)
	if err != nil {
		return cli.Exit(fmt.Sprintf("Error registering operator with avs: %s", err), 1)
	}
	keeper.Logger.Infof("Registered operator with avs")

	return nil
}

func deregisterOperator(c *cli.Context) error {
	return nil
}
