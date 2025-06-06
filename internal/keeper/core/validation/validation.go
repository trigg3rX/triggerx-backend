package validation

import (
	"github.com/trigg3rX/triggerx-backend/pkg/client/aggregator"
	"github.com/trigg3rX/triggerx-backend/pkg/docker"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
)

type TaskValidator struct {
	alchemyAPIKey   string
	etherscanAPIKey string
	codeExecutor    *docker.CodeExecutor
	aggregatorClient *aggregator.AggregatorClient
	logger          logging.Logger
}

func NewTaskValidator(alchemyAPIKey string, etherscanAPIKey string, codeExecutor *docker.CodeExecutor, aggregatorClient *aggregator.AggregatorClient, logger logging.Logger) *TaskValidator {
	return &TaskValidator{
		alchemyAPIKey:   alchemyAPIKey,
		etherscanAPIKey: etherscanAPIKey,
		codeExecutor:    codeExecutor,
		aggregatorClient: aggregatorClient,
		logger:          logger,
	}
}
