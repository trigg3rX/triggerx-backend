package chainio

import (
	"github.com/Layr-Labs/eigensdk-go/chainio/clients/eth"
	"github.com/Layr-Labs/eigensdk-go/logging"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	gethcommon "github.com/ethereum/go-ethereum/common"

	regcoord "github.com/Layr-Labs/eigensdk-go/contracts/bindings/RegistryCoordinator"
	sdkcommon "github.com/Layr-Labs/triggerx-keeper/pkg/common"
	erc20mock "github.com/Layr-Labs/incredible-squaring-avs/contracts/bindings/ERC20Mock"
	txservicemanager "github.com/trigg3rX/go-backend/pkg/avsinterface/bindings/TriggerXServiceManager"
	txtaskmanager "github.com/trigg3rX/go-backend/pkg/avsinterface/bindings/TriggerXTaskManager"
)

type AvsManagersBindings struct {
	TaskManager    *txtaskmanager.ContractTriggerXTaskManager
	ServiceManager *txservicemanager.ContractTriggerXServiceManager
	ethClient      eth.HttpBackend
	logger         logging.Logger
}

func NewAvsManagersBindings(registryCoordinatorAddr, operatorStateRetrieverAddr gethcommon.Address, ethclient sdkcommon.EthClientInterface, logger logging.Logger) (*AvsManagersBindings, error) {
	contractRegistryCoordinator, err := regcoord.NewContractRegistryCoordinator(registryCoordinatorAddr, ethclient)
	if err != nil {
		return nil, err
	}
	serviceManagerAddr, err := contractRegistryCoordinator.ServiceManager(&bind.CallOpts{})
	if err != nil {
		return nil, err
	}
	contractServiceManager, err := txservicemanager.NewContractTriggerXServiceManager(serviceManagerAddr, ethclient)
	if err != nil {
		logger.Error("Failed to fetch IServiceManager contract", "err", err)
		return nil, err
	}

	taskManagerAddr, err := contractServiceManager.TaskManagerContract(&bind.CallOpts{})
	if err != nil {
		logger.Error("Failed to fetch TaskManager address", "err", err)
		return nil, err
	}
	contractTaskManager, err := txtaskmanager.NewContractTriggerXTaskManager(taskManagerAddr, ethclient)
	if err != nil {
		logger.Error("Failed to fetch TriggerXTaskManager contract", "err", err)
		return nil, err
	}

	return &AvsManagersBindings{
		ServiceManager: contractServiceManager,
		TaskManager:    contractTaskManager,
		ethClient:      ethclient,
		logger:         logger,
	}, nil
}

func (b *AvsManagersBindings) GetErc20Mock(tokenAddr common.Address) (*erc20mock.ContractERC20Mock, error) {
	contractErc20Mock, err := erc20mock.NewContractERC20Mock(tokenAddr, b.ethClient)
	if err != nil {
		b.logger.Error("Failed to fetch ERC20Mock contract", "err", err)
		return nil, err
	}
	return contractErc20Mock, nil
}