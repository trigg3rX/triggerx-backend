package chainio

import (
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	gethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/event"

	sdklogging "github.com/Layr-Labs/eigensdk-go/logging"

	sdkcommon "github.com/trigg3rX/triggerx-keeper/pkg/common"
	txtaskmanager "github.com/trigg3rX/go-backend/pkg/avsinterface/bindings/TriggerXTaskManager"
	"github.com/trigg3rX/triggerx-keeper/pkg/core/config"
)

type AvsSubscriberer interface {
	SubscribeToNewTasks(newTaskCreatedChan chan *txtaskmanager.ContractTriggerXTaskManagerTaskCreated) event.Subscription
	SubscribeToTaskResponses(taskResponseLogs chan *txtaskmanager.ContractTriggerXTaskManagerTaskResponded) event.Subscription
	ParseTaskResponded(rawLog types.Log) (*txtaskmanager.ContractTriggerXTaskManagerTaskResponded, error)
}

// Subscribers use a ws connection instead of http connection like Readers
// kind of stupid that the geth client doesn't have a unified interface for both...
// it takes a single url, so the bindings, even though they have watcher functions, those can't be used
// with the http connection... seems very very stupid. Am I missing something?
type AvsSubscriber struct {
	AvsContractBindings *AvsManagersBindings
	logger              sdklogging.Logger
}

func BuildAvsSubscriberFromConfig(config *config.Config) (*AvsSubscriber, error) {
	return BuildAvsSubscriber(
		config.TriggerXServiceManagerAddr,
		config.OperatorStateRetrieverAddr,
		&config.EthWsClient,
		config.Logger,
	)
}

func BuildAvsSubscriber(registryCoordinatorAddr, blsOperatorStateRetrieverAddr gethcommon.Address, ethclient sdkcommon.EthClientInterface, logger sdklogging.Logger) (*AvsSubscriber, error) {
	avsContractBindings, err := NewAvsManagersBindings(registryCoordinatorAddr, blsOperatorStateRetrieverAddr, ethclient, logger)
	if err != nil {
		logger.Errorf("Failed to create contract bindings", "err", err)
		return nil, err
	}
	return NewAvsSubscriber(avsContractBindings, logger), nil
}

func NewAvsSubscriber(avsContractBindings *AvsManagersBindings, logger sdklogging.Logger) *AvsSubscriber {
	return &AvsSubscriber{
		AvsContractBindings: avsContractBindings,
		logger:              logger,
	}
}

func (s *AvsSubscriber) SubscribeToNewTasks(newTaskCreatedChan chan *txtaskmanager.ContractTriggerXTaskManagerTaskCreated) event.Subscription {
	sub, err := s.AvsContractBindings.TaskManager.WatchTaskCreated(
		&bind.WatchOpts{}, newTaskCreatedChan,
	)
	if err != nil {
		s.logger.Error("Failed to subscribe to new TaskManager tasks", "err", err)
	}
	s.logger.Infof("Subscribed to new TaskManager tasks")
	return sub
}

func (s *AvsSubscriber) SubscribeToTaskResponses(taskResponseChan chan *txtaskmanager.ContractTriggerXTaskManagerTaskResponded) event.Subscription {
	sub, err := s.AvsContractBindings.TaskManager.WatchTaskResponded(
		&bind.WatchOpts{}, taskResponseChan,
	)
	if err != nil {
		s.logger.Error("Failed to subscribe to TaskResponded events", "err", err)
	}
	s.logger.Infof("Subscribed to TaskResponded events")
	return sub
}

func (s *AvsSubscriber) ParseTaskResponded(rawLog types.Log) (*txtaskmanager.ContractTriggerXTaskManagerTaskResponded, error) {
	return s.AvsContractBindings.TaskManager.ContractTriggerXTaskManagerFilterer.ParseTaskResponded(rawLog)
}
