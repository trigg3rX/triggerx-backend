package scheduler

import (
	"fmt"
	"strings"
	"time"

	"github.com/trigg3rX/triggerx-backend/internal/schedulers/condition/config"
	"github.com/trigg3rX/triggerx-backend/internal/schedulers/condition/scheduler/worker"
	nodeclient "github.com/trigg3rX/triggerx-backend/pkg/client/nodeclient"
	httppkg "github.com/trigg3rX/triggerx-backend/pkg/http"
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
	validTypes := []string{worker.SourceTypeAPI, worker.SourceTypeOracle, worker.SourceTypeStatic, worker.SourceTypeWebSocket}
	for _, valid := range validTypes {
		if sourceType == valid {
			return true
		}
	}
	return false
}

func (s *ConditionBasedScheduler) initRetryClient() error {
	retryConfig := httppkg.DefaultHTTPRetryConfig()
	var err error
	s.HTTPClient, err = httppkg.NewHTTPClient(retryConfig, s.logger)
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
		// Extract API key from RPC URL if it's an Alchemy/Blast URL
		apiKey := extractAPIKeyFromURL(rpcURL)

		// Create node client config
		nodeCfg := nodeclient.DefaultConfig(apiKey, "", s.logger)
		nodeCfg.BaseURL = rpcURL
		nodeCfg.RequestTimeout = 30 * time.Second

		client, err := nodeclient.NewNodeClient(nodeCfg)
		if err != nil {
			s.logger.Warn("Failed to create node client for chain",
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

// extractAPIKeyFromURL extracts API key from RPC URL
func extractAPIKeyFromURL(url string) string {
	// For Alchemy/Blast URLs, the API key is typically at the end
	// Format: https://base-mainnet.g.alchemy.com/v2/API_KEY
	// or: https://base-mainnet.blastapi.io/API_KEY
	parts := strings.Split(url, "/")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return ""
}
