package validation

import (
	"fmt"

	"github.com/trigg3rX/triggerx-backend/pkg/client/aggregator"
	"github.com/trigg3rX/triggerx-backend/pkg/docker"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
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

func (v *TaskValidator) ValidateTask(ipfsData types.IPFSData) (bool, error) {
	switch ipfsData.TargetData.TaskDefinitionID {
	case 1, 2:
		return v.ValidateTimeBasedTask(ipfsData)
	case 3, 4:
		return v.ValidateEventBasedTask(ipfsData)
	case 5, 6:
		return v.ValidateConditionBasedTask(ipfsData)
	default:
		return false, fmt.Errorf("unsupported task definition id: %d", ipfsData.TargetData.TaskDefinitionID)
	}
}