package validation

import (
	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/trigg3rX/triggerx-backend/pkg/logging"
)

type TaskValidator struct {
	logger    logging.Logger
	ethClient *ethclient.Client
}

func NewTaskValidator(logger logging.Logger, ethClient *ethclient.Client) *TaskValidator {
	return &TaskValidator{
		logger:    logger,
		ethClient: ethClient,
	}
}
