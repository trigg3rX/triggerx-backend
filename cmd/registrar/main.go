package main

import (
	"fmt"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/trigg3rX/triggerx-backend/internal/registrar"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
)

var logger logging.Logger

func main() {

	if err := logging.InitLogger(logging.Development, logging.RegistrarProcess); err != nil {
		panic(fmt.Sprintf("Failed to initialize logger: %v", err))
	}
	logger = logging.GetLogger(logging.Development, logging.RegistrarProcess)
	logger.Info("Starting registrar node (poll-based)...")

	// Initialize registrar configuration
	registrar.Init()

	// Connect to Ethereum network via HTTP RPC
	ethClient, err := ethclient.Dial(registrar.EthRpcUrl)
	if err != nil {
		logger.Fatal(fmt.Sprintf("Failed to connect to Ethereum RPC: %v", err))
	}
	logger.Info("Connected to Ethereum HTTP RPC")

	// Connect to Base network via HTTP RPC
	baseClient, err := ethclient.Dial(registrar.BaseRpcUrl)
	if err != nil {
		logger.Fatal(fmt.Sprintf("Failed to connect to Base RPC: %v", err))
	}
	logger.Info("Connected to Base HTTP RPC")

	// Get contract addresses
	avsGovernanceAddress := common.HexToAddress(registrar.AvsGovernanceAddress)
	attestationCenterAddress := common.HexToAddress(registrar.AttestationCenterAddress)

	logger.Info(fmt.Sprintf("Using AVS Governance contract at address: %s", registrar.AvsGovernanceAddress))
	logger.Info(fmt.Sprintf("Using Attestation Center contract at address: %s", registrar.AttestationCenterAddress))

	// Initialize event processing
	if err := registrar.InitEventProcessing(ethClient, baseClient); err != nil {
		logger.Fatal(fmt.Sprintf("Failed to initialize event processing: %v", err))
	}

	// Start the polling service in a goroutine
	go registrar.StartEventPolling(avsGovernanceAddress, attestationCenterAddress)

	// Keep the program running
	logger.Info("Registrar node is running. Press Ctrl+C to exit.")

	// Keep the main thread alive
	for {
		time.Sleep(time.Hour)
	}
}
