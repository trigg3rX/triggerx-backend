package main

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common"

	"github.com/trigg3rX/triggerx-backend/internal/registrar"
	"github.com/trigg3rX/triggerx-backend/internal/registrar/config"
	"github.com/trigg3rX/triggerx-backend/internal/registrar/database"
	dbpkg "github.com/trigg3rX/triggerx-backend/pkg/database"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
)

var logger logging.Logger

func main() {
	config.Init()

	if config.DevMode {
		if err := logging.InitLogger(logging.Development, logging.RegistrarProcess); err != nil {
			panic(fmt.Sprintf("Failed to initialize logger: %v", err))
		}
		logger = logging.GetLogger(logging.Development, logging.RegistrarProcess)
	} else {
		if err := logging.InitLogger(logging.Production, logging.RegistrarProcess); err != nil {
			panic(fmt.Sprintf("Failed to initialize logger: %v", err))
		}
		logger = logging.GetLogger(logging.Production, logging.RegistrarProcess)
	}
	logger.Info("Starting registrar service ...")

	registrar.InitABI()

	dbConfig := dbpkg.NewConfig(config.DatabaseHost, config.DatabaseHostPort)

	dbConn, err := dbpkg.NewConnection(dbConfig)
	if err != nil {
		logger.Fatalf("Failed to connect to database: %v", err)
	}
	defer dbConn.Close()

	database.SetDatabaseConnection(dbConn)
	logger.Info("Database connection initialized")

	avsGovernanceAddress := common.HexToAddress(config.AvsGovernanceAddress)
	attestationCenterAddress := common.HexToAddress(config.AttestationCenterAddress)

	logger.Info(fmt.Sprintf("AVS Governance     [L1]: %s", config.AvsGovernanceAddress))
	logger.Info(fmt.Sprintf("Attestation Center [L2]: %s", config.AttestationCenterAddress))

	go func() {
		registrar.StartEventPolling(avsGovernanceAddress, attestationCenterAddress)
	}()

	go func() {
		registrar.StartDailyRewardsPoints()
	}()

	logger.Info("Registrar service is running.")

	select {}
}
