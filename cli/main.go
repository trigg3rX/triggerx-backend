package main

import (
	"log"
	"os"

	"github.com/trigg3rX/triggerx-backend/cli/actions"
	"github.com/trigg3rX/triggerx-backend/cli/core/config"
	"github.com/urfave/cli"
)

func main() {
	err := config.Init()
	if err != nil {
		log.Fatal(err)
	}

	app := cli.NewApp()
	app.Name = "triggerx"
	app.Usage = "TriggerX Operator CLI"

	app.Commands = []cli.Command{
		{
			Name:   "generate-keys",
			Usage:  "Generate BLS and ECDSA keystore files",
			Action: actions.GenerateKeys,
		},
		// Imuachain Integration
		{
			Name:   "setup-imua-keys",
			Usage:  "Setup and create required Imuachain validator keys",
			Action: actions.SetupImuaKeys,
		},
		{
			Name:   "fund-imua-account",
			Usage:  "Fund Imuachain validator account with IMUA tokens from faucet",
			Action: actions.FundImuaAccount,
		},
		{
			Name:   "check-imua-balance",
			Usage:  "Check IMUA token balance of validator account",
			Action: actions.CheckImuaBalance,
		},
		{
			Name:   "register-imua-operator",
			Usage:  "Register operator on Imuachain",
			Action: actions.RegisterImuaOperator,
		},
		// Token Management Commands
		{
			Name:   "get-imeth-tokens",
			Usage:  "Get imETH tokens from faucet (required before depositing)",
			Action: actions.GetImethTokens,
		},
		{
			Name:   "deposit-tokens",
			Usage:  "Deposit tokens from Ethereum Sepolia to Imuachain",
			Action: actions.DepositTokens,
		},
		{
			Name:   "delegate-tokens",
			Usage:  "Delegate deposited tokens to your Imua validator",
			Action: actions.DelegateTokens,
		},
		// AVS Management
		{
			Name:   "opt-in-to-avs",
			Usage:  "Opt-in to AVS on Imuachain as a validator",
			Action: actions.OptInToAVS,
		},
		{
			Name:   "associate-operator",
			Usage:  "Associate operator with EVM staker (post-bootstrap phase)",
			Action: actions.AssociateOperator,
		},
		// BLS Key Management
		{
			Name:   "register-bls-key",
			Usage:  "Register BLS public key for AVS operator",
			Action: actions.RegisterBLSPublicKey,
		},
	}

	err = app.Run(os.Args)
	if err != nil {
		panic(err)
	}
}
