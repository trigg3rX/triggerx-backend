package services

import (
	"github.com/trigg3rX/triggerx-backend/internal/keeper/config"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
)

var logger = logging.GetLogger(logging.Development, logging.KeeperProcess)

func Init() {
	config.Init()
	logger.Info("Config Initialized")
}
