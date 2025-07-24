package chainio

import (

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	avs "github.com/trigg3rX/imua-contracts/bindings/contracts/TriggerXAvs"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
)

type ContractBindings struct {
	AvsAddr    common.Address
	AVSManager *avs.TriggerXAvs
}

func NewContractBindings(
	avsAddr common.Address,
	ethclient *ethclient.Client,
	logger logging.Logger,
) (*ContractBindings, error) {
	contractAvsManager, err := avs.NewTriggerXAvs(avsAddr, ethclient)
	if err != nil {
		logger.Error("Failed to fetch Avs contract", "err", err)
		return nil, err
	}

	return &ContractBindings{
		AvsAddr:    avsAddr,
		AVSManager: contractAvsManager,
	}, nil
}
