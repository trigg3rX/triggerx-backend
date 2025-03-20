package main

import (
	"fmt"

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
	logger.Info("Starting registrar node...")

	registrar.Init()

	// Initialize ABI parsers
	if err := registrar.InitABI(); err != nil {
		logger.Fatal(fmt.Sprintf("Failed to initialize ABI parsers: %v", err))
	}

	// Check if WebSocket URL is available
	if registrar.EthWsRpcUrl == "" {
		logger.Fatal("WebSocket URL (L1_WS_RPC) is required for event subscriptions but not provided in .env")
	}

	// Connect to Ethereum network via WebSocket - for event subscriptions
	ethWsClient, err := ethclient.Dial(registrar.EthWsRpcUrl)
	if err != nil {
		logger.Fatal(fmt.Sprintf("Failed to connect to Ethereum WebSocket: %v", err))
	}
	logger.Info("Connected to Ethereum WebSocket RPC")

	// Connect to Base network via WebSocket - for event subscriptions
	baseWsClient, err := ethclient.Dial(registrar.BaseWsRpcUrl)
	if err != nil {
		logger.Fatal(fmt.Sprintf("Failed to connect to Base WebSocket: %v", err))
	}
	logger.Info("Connected to Base WebSocket RPC")

	// Get contract addresses
	avsGovernanceAddress := common.HexToAddress(registrar.AvsGovernanceAddress)
	attestationCenterAddress := common.HexToAddress(registrar.AttestationCenterAddress)

	logger.Info(fmt.Sprintf("Using AVS Governance contract at address: %s", registrar.AvsGovernanceAddress))
	logger.Info(fmt.Sprintf("Using Attestation Center contract at address: %s", registrar.AttestationCenterAddress))

	// Create channels for events
	operatorRegisteredCh := make(chan *registrar.OperatorRegistered)
	operatorUnregisteredCh := make(chan *registrar.OperatorUnregistered)
	taskSubmittedCh := make(chan *registrar.TaskSubmitted)
	taskRejectedCh := make(chan *registrar.TaskRejected)

	// Set up event subscriptions
	regSub, err := registrar.SetupRegisteredSubscription(
		ethWsClient,
		avsGovernanceAddress,
		operatorRegisteredCh,
	)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to set up initial OperatorRegistered subscription: %v", err))
	}

	unregSub, err := registrar.SetupUnregisteredSubscription(
		ethWsClient,
		avsGovernanceAddress,
		operatorUnregisteredCh,
	)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to set up initial OperatorUnregistered subscription: %v", err))
	}

	taskSubSub, err := registrar.SetupTaskSubmittedSubscription(
		baseWsClient,
		attestationCenterAddress,
		taskSubmittedCh,
	)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to set up initial TaskSubmitted subscription: %v", err))
	}

	taskRejSub, err := registrar.SetupTaskRejectedSubscription(
		baseWsClient,
		attestationCenterAddress,
		taskRejectedCh,
	)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to set up initial TaskRejected subscription: %v", err))
	}

	// Manage all subscriptions
	go registrar.ManageSubscriptions(
		ethWsClient,
		avsGovernanceAddress,
		baseWsClient,
		attestationCenterAddress,
		operatorRegisteredCh,
		operatorUnregisteredCh,
		taskSubmittedCh,
		taskRejectedCh,
		regSub,
		unregSub,
		taskSubSub,
		taskRejSub,
	)

	select {} // Block forever
}
