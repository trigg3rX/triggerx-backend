package chainio

import (
	gethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/imua-xyz/imua-avs-sdk/logging"

	avs "github.com/trigg3rX/imua-contracts/bindings/contracts/TriggerXAvs"
	"github.com/trigg3rX/triggerx-backend/cli/core/chainio/eth"
)

type ContractBindings struct {
	AvsAddr    gethcommon.Address
	AVSManager *avs.TriggerXAvs
}

func NewContractBindings(
	avsAddr gethcommon.Address,
	ethclient eth.EthClient,
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
