package main

import (
	"fmt"
	"time"

	"github.com/ethereum/go-ethereum/common"

	"github.com/trigg3rX/triggerx-backend/internal/registrar"
	"github.com/trigg3rX/triggerx-backend/internal/registrar/config"
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
	config.Init()

	// Get contract addresses
	avsGovernanceAddress := common.HexToAddress(config.AvsGovernanceAddress)
	attestationCenterAddress := common.HexToAddress(config.AttestationCenterAddress)

	logger.Info(fmt.Sprintf("Using AVS Governance contract at address: %s", config.AvsGovernanceAddress))
	logger.Info(fmt.Sprintf("Using Attestation Center contract at address: %s", config.AttestationCenterAddress))

	// Start the polling service in a goroutine
	go registrar.StartEventPolling(avsGovernanceAddress, attestationCenterAddress)

	// Keep the program running
	logger.Info("Registrar node is running.")

	// Keep the main thread alive
	for {
		time.Sleep(time.Hour)
	}
}
