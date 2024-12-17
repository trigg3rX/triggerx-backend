package chainio

import (
	"context"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	gethcommon "github.com/ethereum/go-ethereum/common"

	sdkavsregistry "github.com/Layr-Labs/eigensdk-go/chainio/clients/avsregistry"
	logging "github.com/Layr-Labs/eigensdk-go/logging"

	sdkcommon "github.com/trigg3rX/triggerx-keeper/pkg/common"
	erc20mock "github.com/Layr-Labs/incredible-squaring-avs/contracts/bindings/ERC20Mock"
	txtaskmanager "github.com/trigg3rX/go-backend/pkg/avsinterface/bindings/TriggerXTaskManager"
	"github.com/Layr-Labs/triggerx-keeper/pkg/core/config"
)

type AvsReaderer interface {
	//sdkavsregistry.ChainReader

	CheckSignatures(
		ctx context.Context, msgHash [32]byte, quorumNumbers []byte, referenceBlockNumber uint32, nonSignerStakesAndSignature txtaskmanager.IBLSSignatureCheckerNonSignerStakesAndSignature,
	) (txtaskmanager.IBLSSignatureCheckerQuorumStakeTotals, error)
	GetErc20Mock(ctx context.Context, tokenAddr gethcommon.Address) (*erc20mock.ContractERC20Mock, error)
	GetOperatorId(
		opts *bind.CallOpts,
		operatorAddress common.Address,
	) ([32]byte, error)
	IsOperatorRegistered(
		opts *bind.CallOpts,
		operatorAddress common.Address,
	) (bool, error)
}

type AvsReader struct {
	sdkavsregistry.ChainReader
	AvsServiceBindings *AvsManagersBindings
	logger             logging.Logger
}

//var _ AvsReaderer = (*AvsReader)(nil)

func BuildAvsReaderFromConfig(c *config.Config) (*AvsReader, error) {
	return BuildAvsReader(c.IncredibleSquaringRegistryCoordinatorAddr, c.OperatorStateRetrieverAddr, &c.EthHttpClient, c.Logger)
}
func BuildAvsReader(registryCoordinatorAddr, operatorStateRetrieverAddr gethcommon.Address, ethHttpClient sdkcommon.EthClientInterface, logger logging.Logger) (*AvsReader, error) {
	avsManagersBindings, err := NewAvsManagersBindings(registryCoordinatorAddr, operatorStateRetrieverAddr, ethHttpClient, logger)
	if err != nil {
		return nil, err
	}
	avsRegistryReader, err := sdkavsregistry.BuildAvsRegistryChainReader(registryCoordinatorAddr, operatorStateRetrieverAddr, ethHttpClient, logger)
	if err != nil {
		return nil, err
	}
	return NewAvsReader(*avsRegistryReader, avsManagersBindings, logger)
}
func NewAvsReader(avsRegistryReader sdkavsregistry.ChainReader, avsServiceBindings *AvsManagersBindings, logger logging.Logger) (*AvsReader, error) {
	return &AvsReader{
		ChainReader:        avsRegistryReader,
		AvsServiceBindings: avsServiceBindings,
		logger:             logger,
	}, nil
}

func (r *AvsReader) CheckSignatures(
	ctx context.Context, msgHash [32]byte, quorumNumbers []byte, referenceBlockNumber uint32, nonSignerStakesAndSignature txtaskmanager.IBLSSignatureCheckerNonSignerStakesAndSignature,
) (txtaskmanager.IBLSSignatureCheckerQuorumStakeTotals, error) {
	stakeTotalsPerQuorum, _, err := r.AvsServiceBindings.TaskManager.CheckSignatures(
		&bind.CallOpts{}, msgHash, quorumNumbers, referenceBlockNumber, nonSignerStakesAndSignature,
	)
	if err != nil {
		return txtaskmanager.IBLSSignatureCheckerQuorumStakeTotals{}, err
	}
	return stakeTotalsPerQuorum, nil
}

func (r *AvsReader) GetErc20Mock(ctx context.Context, tokenAddr gethcommon.Address) (*erc20mock.ContractERC20Mock, error) {
	erc20Mock, err := r.AvsServiceBindings.GetErc20Mock(tokenAddr)
	if err != nil {
		r.logger.Error("Failed to fetch ERC20Mock contract", "err", err)
		return nil, err
	}
	return erc20Mock, nil
}