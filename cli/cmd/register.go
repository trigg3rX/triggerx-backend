package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"net/http"

	"github.com/Layr-Labs/eigensdk-go/crypto/ecdsa"
	"github.com/ethereum/go-ethereum/common"
	"github.com/trigg3rX/triggerx-backend/cli/utils"
	"github.com/trigg3rX/triggerx-backend/pkg/keeper"
	"github.com/trigg3rX/triggerx-backend/pkg/registration"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
	"github.com/urfave/cli/v2"
)

func RegisterCommand() *cli.Command {
	return &cli.Command{
		Name:  "register",
		Usage: "Register a new operator",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "token-address",
				Usage:    "Address of the token to deposit into the strategy",
				Required: true,
			},
			&cli.StringFlag{
				Name:     "amount",
				Usage:    "Amount of tokens to deposit into the strategy",
				Required: true,
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
	// Check if required flags are provided
	tokenAddress := c.String("token-address")
	amountStr := c.String("amount")

	fmt.Println(
		`Register a new operator to the TriggerX AVS. It is assumed that the operator has already been registered to the EigenLayer protocol.
If not, please register to EigenLayer first. Follow the instructions here: 
https://github.com/Layr-Labs/eigenlayer-cli/blob/master/README.md`)

	configPath := "config-files/triggerx_operator.yaml"

	// Create keeper instance from config file
	keeper, err := keeper.NewKeeperFromConfigFile(configPath)
	if err != nil {
		return cli.Exit(fmt.Sprintf("Error creating keeper: %s", err), 1)
	}

	// Create registration instance
	registration := registration.NewRegistration(
		keeper.Logger,
		keeper.EthClient,
		keeper.AvsReader,
		keeper.AvsWriter,
		keeper.EigenlayerReader,
		keeper.KeeperAddr,
		keeper.BlsKeypair,
	)

	if registration == nil {
		return cli.Exit("registration was not properly initialized", 1)
	}

	// Convert amount string to big.Int (assuming amount is in ETH)
	amount := new(big.Int)
	amountFloat, ok := new(big.Float).SetString(amountStr)
	if !ok {
		return cli.Exit("Invalid amount format", 1)
	}
	// Convert ETH to Wei (multiply by 10^18)
	amountFloat.Mul(amountFloat, new(big.Float).SetInt(new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil)))
	amountFloat.Int(amount) // Convert float to int

	// Prompt for ECDSA keystore passphrase
	ecdsaPassphrase, err := utils.PasswordPrompt("Enter passphrase for ECDSA keystore: ", false)
	if err != nil {
		return cli.Exit(fmt.Sprintf("Error reading ECDSA passphrase: %s", err), 1)
	}

	mockTokenStrategyAddr := common.HexToAddress(tokenAddress)

	privKey, err := ecdsa.ReadKey("/home/nite-sky/.eigenlayer/operator_keys/frodo_keeper1.ecdsa.key.json", ecdsaPassphrase)
	if err != nil {
		return cli.Exit(fmt.Sprintf("Error reading keystore file: %s", err), 1)
	}

	// Deposit into strategy
	err = registration.DepositIntoStrategy(mockTokenStrategyAddr, amount)
	if err != nil {
		return cli.Exit(fmt.Sprintf("Error depositing into strategy: %s", err), 1)
	}
	keeper.Logger.Infof("Deposited %s into strategy %s", amount, mockTokenStrategyAddr)

	// Register operator with AVS
	_, txHash, err := registration.RegisterOperatorWithAvs(privKey)
	if err != nil {
		return cli.Exit(fmt.Sprintf("Error registering operator with avs: %s", err), 1)
	}
	keeper.Logger.Infof("Registered operator with avs")

	// Create keeper data record in database
	keeperData := types.KeeperData{
		KeeperID:          0,
		WithdrawalAddress: keeper.KeeperAddr.String(),
		Stakes:            []float64{float64(amount.Int64())},
		Strategies:        []string{mockTokenStrategyAddr.String()},
		Verified:          true,
		CurrentQuorumNo:   0,

		RegisteredTx:      txHash,
		Status:            true,
		BlsSigningKeys:    []string{keeper.BlsKeypair.GetPubKeyG1().String()},
		ConnectionAddress: "",
	}

	// Make HTTP request to create keeper data
	jsonData, err := json.Marshal(keeperData)
	if err != nil {
		return cli.Exit(fmt.Sprintf("Error marshaling keeper data: %s", err), 1)
	}

	resp, err := http.Post("http://localhost:8080/api/v1/keepers", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return cli.Exit(fmt.Sprintf("Error creating keeper record: %s", err), 1)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return cli.Exit(fmt.Sprintf("Error response from API: %s", string(body)), 1)
	}
	keeper.Logger.Info("Created keeper record in database")

	return nil
}

func deregisterOperator(c *cli.Context) error {
	return nil
}
