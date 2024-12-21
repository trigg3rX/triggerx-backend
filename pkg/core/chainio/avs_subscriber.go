package chainio

import (
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	gethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/event"

	sdklogging "github.com/Layr-Labs/eigensdk-go/logging"

	txservicemanager "github.com/trigg3rX/go-backend/pkg/avsinterface/bindings/TriggerXServiceManager"
	txtaskmanager "github.com/trigg3rX/go-backend/pkg/avsinterface/bindings/TriggerXTaskManager"
	sdkcommon "github.com/trigg3rX/go-backend/pkg/common"
	"github.com/trigg3rX/go-backend/pkg/core/config"
)

type AvsSubscriberer interface {
	SubscribeToNewTasks(newTaskCreatedChan chan *txtaskmanager.ContractTriggerXTaskManagerTaskCreated) event.Subscription
	SubscribeToTaskResponses(taskResponseLogs chan *txtaskmanager.ContractTriggerXTaskManagerTaskResponded) event.Subscription
	SubscribeToKeeperRegistered(keeperRegisteredChan chan *txservicemanager.ContractTriggerXServiceManagerKeeperAdded) event.Subscription
	SubscribeToKeeperDeregistered(keeperDeregisteredChan chan *txservicemanager.ContractTriggerXServiceManagerKeeperRemoved) event.Subscription

	ParseTaskResponded(rawLog types.Log) (*txtaskmanager.ContractTriggerXTaskManagerTaskResponded, error)
	ParseTaskCreated(rawLog types.Log) (*txtaskmanager.ContractTriggerXTaskManagerTaskCreated, error)
	ParseKeeperRegistered(rawLog types.Log) (*txservicemanager.ContractTriggerXServiceManagerKeeperAdded, error)
	ParseKeeperDeregistered(rawLog types.Log) (*txservicemanager.ContractTriggerXServiceManagerKeeperRemoved, error)
}

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
		s.logger.Error("Failed to subscribe to TaskCreated events", "err", err)
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
	s.logger.Info("Subscribed to TaskResponded events")
	return sub
}

func (s *AvsSubscriber) SubscribeToKeeperRegistered(keeperRegisteredChan chan *txservicemanager.ContractTriggerXServiceManagerKeeperAdded) event.Subscription {
	sub, err := s.AvsContractBindings.ServiceManager.WatchKeeperAdded(
		&bind.WatchOpts{},
		keeperRegisteredChan,
		[]gethcommon.Address{},
	)
	if err != nil {
		s.logger.Error("Failed to subscribe to KeeperRegistered events", "err", err)
	}
	s.logger.Info("Subscribed to KeeperRegistered events")
	return sub
}

func (s *AvsSubscriber) SubscribeToKeeperDeregistered(keeperDeregisteredChan chan *txservicemanager.ContractTriggerXServiceManagerKeeperRemoved) event.Subscription {
	sub, err := s.AvsContractBindings.ServiceManager.WatchKeeperRemoved(
		&bind.WatchOpts{},
		keeperDeregisteredChan,
		[]gethcommon.Address{},
	)
	if err != nil {
		s.logger.Error("Failed to subscribe to KeeperDeregistered events", "err", err)
	}
	s.logger.Info("Subscribed to KeeperDeregistered events")
	return sub
}

func (s *AvsSubscriber) ParseTaskResponded(rawLog types.Log) (*txtaskmanager.ContractTriggerXTaskManagerTaskResponded, error) {
	return s.AvsContractBindings.TaskManager.ContractTriggerXTaskManagerFilterer.ParseTaskResponded(rawLog)
}

func (s *AvsSubscriber) ParseTaskCreated(rawLog types.Log) (*txtaskmanager.ContractTriggerXTaskManagerTaskCreated, error) {
	return s.AvsContractBindings.TaskManager.ContractTriggerXTaskManagerFilterer.ParseTaskCreated(rawLog)
}

func (s *AvsSubscriber) ParseKeeperRegistered(rawLog types.Log) (*txservicemanager.ContractTriggerXServiceManagerKeeperAdded, error) {
	return s.AvsContractBindings.ServiceManager.ContractTriggerXServiceManagerFilterer.ParseKeeperAdded(rawLog)
}

func (s *AvsSubscriber) ParseKeeperDeregistered(rawLog types.Log) (*txservicemanager.ContractTriggerXServiceManagerKeeperRemoved, error) {
	return s.AvsContractBindings.ServiceManager.ContractTriggerXServiceManagerFilterer.ParseKeeperRemoved(rawLog)
}
