package scheduler

import (
	"fmt"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/trigg3rX/triggerx-backend-imua/internal/schedulers/condition/config"
	"github.com/trigg3rX/triggerx-backend-imua/internal/schedulers/condition/scheduler/worker"
	"github.com/trigg3rX/triggerx-backend-imua/pkg/retry"
)

// Helper functions
func isValidConditionType(conditionType string) bool {
	validTypes := []string{
		worker.ConditionGreaterThan, worker.ConditionLessThan, worker.ConditionBetween,
		worker.ConditionEquals, worker.ConditionNotEquals, worker.ConditionGreaterEqual, worker.ConditionLessEqual,
	}
	for _, valid := range validTypes {
		if conditionType == valid {
			return true
		}
	}
	return false
}

func isValidSourceType(sourceType string) bool {
	validTypes := []string{worker.SourceTypeAPI, worker.SourceTypeOracle, worker.SourceTypeStatic}
	for _, valid := range validTypes {
		if sourceType == valid {
			return true
		}
	}
	return false
}

func (s *ConditionBasedScheduler) initRetryClient() error {
	retryConfig := retry.DefaultHTTPRetryConfig()
	var err error
	s.HTTPClient, err = retry.NewHTTPClient(retryConfig, s.logger)
	if err != nil {
		return fmt.Errorf("failed to initialize retry client: %w", err)
	}
	return nil
}

// initChainClients initializes blockchain clients for different chains
func (s *ConditionBasedScheduler) initChainClients() error {
	// Get chain RPC URLs from config
	chainRPCs := config.GetChainRPCUrls()

	for chainID, rpcURL := range chainRPCs {
		client, err := ethclient.Dial(rpcURL)
		if err != nil {
			s.logger.Warn("Failed to connect to chain",
				"chain_id", chainID,
				"rpc_url", rpcURL,
				"error", err,
			)
			continue
		}

		s.chainClients[chainID] = client
		s.logger.Info("Connected to chain",
			"chain_id", chainID,
			"rpc_url", rpcURL,
		)
	}

	if len(s.chainClients) == 0 {
		return fmt.Errorf("failed to connect to any blockchain networks")
	}

	return nil
}