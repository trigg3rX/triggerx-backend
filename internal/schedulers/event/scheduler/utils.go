package scheduler

import (
	"context"
	"fmt"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/trigg3rX/triggerx-backend/internal/schedulers/event/config"
	"github.com/trigg3rX/triggerx-backend/internal/schedulers/event/metrics"
)

// initChainClients initializes RPC clients for supported chains
func (s *EventBasedScheduler) initChainClients() error {
	s.clientsMutex.Lock()
	defer s.clientsMutex.Unlock()

	for chainID, rpcURL := range config.GetChainRPCUrls() {
		client, err := ethclient.Dial(rpcURL)
		if err != nil {
			s.logger.Error("Failed to connect to chain", "chain_id", chainID, "rpc_url", rpcURL, "error", err)
			metrics.TrackChainConnection(chainID, "failed")
			metrics.TrackConnectionFailure(chainID)
			continue
		}

		// Test connection
		_, err = client.ChainID(context.Background())
		if err != nil {
			s.logger.Error("Failed to get chain ID", "chain_id", chainID, "error", err)
			client.Close()
			metrics.TrackChainConnection(chainID, "failed")
			metrics.TrackConnectionFailure(chainID)
			continue
		}

		s.chainClients[chainID] = client
		metrics.TrackChainConnection(chainID, "success")
		s.logger.Info("Connected to chain", "chain_id", chainID, "rpc_url", rpcURL)
	}

	if len(s.chainClients) == 0 {
		return fmt.Errorf("no chain clients initialized successfully")
	}

	return nil
}
