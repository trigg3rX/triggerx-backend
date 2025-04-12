package main

import (
	"fmt"

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
	logger.Info("Starting registrar service ...")

	config.Init()
	registrar.InitABI()

	avsGovernanceAddress := common.HexToAddress(config.AvsGovernanceAddress)
	attestationCenterAddress := common.HexToAddress(config.AttestationCenterAddress)

	logger.Info(fmt.Sprintf("AVS Governance     [L1]: %s", config.AvsGovernanceAddress))
	logger.Info(fmt.Sprintf("Attestation Center [L2]: %s", config.AttestationCenterAddress))

	go registrar.StartEventPolling(avsGovernanceAddress, attestationCenterAddress)

	logger.Info("Registrar service is running.")

	select {}
}
