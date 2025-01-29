// metrics/metrics.go
package metrics

import (
	"context"
	"crypto/ecdsa"
	"fmt"

	"github.com/Layr-Labs/eigensdk-go/chainio/clients"
	"github.com/Layr-Labs/eigensdk-go/chainio/clients/eth"
	"github.com/Layr-Labs/eigensdk-go/logging"
	eigenmetrics "github.com/Layr-Labs/eigensdk-go/metrics"
	"github.com/Layr-Labs/eigensdk-go/metrics/collectors/economic"
	rpccalls "github.com/Layr-Labs/eigensdk-go/metrics/collectors/rpc_calls"
	"github.com/Layr-Labs/eigensdk-go/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/prometheus/client_golang/prometheus"
)

type MetricsConfig struct {
	AvsName                   string
	EthRpcUrl                 string
	EthWsUrl                  string
	RegistryCoordinatorAddress   string
	OperatorStateRetrieverAddress string
}

type MetricsService struct {
	logger         logging.Logger
	eigenMetrics   *eigenmetrics.EigenMetrics
	clientSet      *clients.Clients
	ethClient      *eth.InstrumentedClient
	operatorAddr   common.Address
	metricsAddress string
}

func NewMetricsService(
	logger logging.Logger,
	ecdsaPrivateKey *ecdsa.PrivateKey,
	operatorAddr common.Address,
	config *MetricsConfig,
) (*MetricsService, error) {
	chainioConfig := clients.BuildAllConfig{
		EthHttpUrl:                 config.EthRpcUrl,
		EthWsUrl:                  config.EthWsUrl,
		RegistryCoordinatorAddr:    config.RegistryCoordinatorAddress,
		OperatorStateRetrieverAddr: config.OperatorStateRetrieverAddress,
		AvsName:                    config.AvsName,
		PromMetricsIpPortAddress:   ":9092", // Fixed metrics port
	}

	clientSet, err := clients.BuildAll(chainioConfig, ecdsaPrivateKey, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to build chainio clients: %w", err)
	}

	reg := prometheus.NewRegistry()
	eigenMetrics := eigenmetrics.NewEigenMetrics(config.AvsName, ":9092", reg, logger)

	// Setup quorum names
	quorumNames := map[types.QuorumNum]string{
		0: "quorum1",
		1: "quorum2",
		2: "quorum3",
		3: "quorum4",
		4: "quorum5",
	}

	// Initialize economic metrics collector
	economicMetricsCollector := economic.NewCollector(
		clientSet.ElChainReader,
		clientSet.AvsRegistryChainReader,
		config.AvsName,
		logger,
		operatorAddr,
		quorumNames,
	)
	reg.MustRegister(economicMetricsCollector)

	// Initialize RPC calls collector
	rpcCallsCollector := rpccalls.NewCollector(config.AvsName, reg)
	instrumentedEthClient, err := eth.NewInstrumentedClient(config.EthRpcUrl, rpcCallsCollector)
	if err != nil {
		return nil, fmt.Errorf("failed to create instrumented ETH client: %w", err)
	}

	return &MetricsService{
		logger:         logger,
		eigenMetrics:   eigenMetrics,
		clientSet:      clientSet,
		ethClient:      instrumentedEthClient,
		operatorAddr:   operatorAddr,
		metricsAddress: ":9092",
	}, nil
}

func (m *MetricsService) Start(ctx context.Context) error {
	reg := prometheus.NewRegistry()
	m.eigenMetrics.Start(ctx, reg)
	return nil
}

