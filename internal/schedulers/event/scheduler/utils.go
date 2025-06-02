package scheduler

import (
	"context"
	"fmt"

	"github.com/trigg3rX/triggerx-backend/internal/schedulers/event/config"
	schedulerTypes "github.com/trigg3rX/triggerx-backend/internal/schedulers/event/scheduler/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

// initChainClients initializes RPC clients for supported chains
func (s *EventBasedScheduler) initChainClients() error {
	s.clientsMutex.Lock()
	defer s.clientsMutex.Unlock()

	for chainID, rpcURL := range config.GetChainRPCUrls() {
		client, err := ethclient.Dial(rpcURL)
		if err != nil {
			s.logger.Error("Failed to connect to chain", "chain_id", chainID, "rpc_url", rpcURL, "error", err)
			continue
		}

		// Test connection
		_, err = client.ChainID(context.Background())
		if err != nil {
			s.logger.Error("Failed to get chain ID", "chain_id", chainID, "error", err)
			client.Close()
			continue
		}

		s.chainClients[chainID] = client
		s.logger.Info("Connected to chain", "chain_id", chainID, "rpc_url", rpcURL)
	}

	if len(s.chainClients) == 0 {
		return fmt.Errorf("no chain clients initialized successfully")
	}

	return nil
}

// getCachedOrFetchBlockNumber gets block number from cache or fetches from chain
func (s *EventBasedScheduler) getCachedOrFetchBlockNumber(client *ethclient.Client, chainID string) (uint64, error) {
	cacheKey := fmt.Sprintf("block_number_%s", chainID)

	if s.cache != nil {
		if cached, err := s.cache.Get(cacheKey); err == nil {
			var blockNum uint64
			if _, err := fmt.Sscanf(cached, "%d", &blockNum); err == nil {
				return blockNum, nil
			}
		}
	}

	// Fetch from chain
	currentBlock, err := client.BlockNumber(context.Background())
	if err != nil {
		return 0, err
	}

	// Cache the result
	if s.cache != nil {
		err := s.cache.Set(cacheKey, fmt.Sprintf("%d", currentBlock), schedulerTypes.BlockCacheTTL)
		if err != nil {
			s.logger.Errorf("Error caching block number: %v", err)
		}
	}

	return currentBlock, nil
}