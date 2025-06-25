package sync

import (
	"context"
	"fmt"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/trigg3rX/triggerx-backend/internal/registrar/config"
	"github.com/trigg3rX/triggerx-backend/internal/registrar/events"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
)

const (
	// Backfill configuration
	DefaultBackfillBatchSize = 1000 // blocks per batch
	DefaultBackfillDelay     = 100 * time.Millisecond
	MaxBackfillBlocks        = 100000 // safety limit
)

// BackfillManager handles historical event processing
type BackfillManager struct {
	ethClient  *ethclient.Client
	baseClient *ethclient.Client
	optClient  *ethclient.Client
	logger     logging.Logger
	batchSize  uint64
	delay      time.Duration
}

// BackfillConfig holds configuration for backfill operations
type BackfillConfig struct {
	BatchSize  uint64        `default:"1000"`
	Delay      time.Duration `default:"100ms"`
	MaxBlocks  uint64        `default:"100000"`
	StartBlock uint64        // Block to start backfill from
	EndBlock   uint64        // Block to end backfill at (0 = current)
	ChainID    string        // Chain to backfill
	EventTypes []string      // Specific events to process
}

// BackfillProgress tracks the progress of a backfill operation
type BackfillProgress struct {
	ChainID         string    `json:"chain_id"`
	StartBlock      uint64    `json:"start_block"`
	EndBlock        uint64    `json:"end_block"`
	CurrentBlock    uint64    `json:"current_block"`
	ProcessedBlocks uint64    `json:"processed_blocks"`
	TotalBlocks     uint64    `json:"total_blocks"`
	EventsFound     uint64    `json:"events_found"`
	StartTime       time.Time `json:"start_time"`
	LastUpdate      time.Time `json:"last_update"`
	Completed       bool      `json:"completed"`
	Error           string    `json:"error,omitempty"`
}

// NewBackfillManager creates a new backfill manager
func NewBackfillManager(ethClient, baseClient, optClient *ethclient.Client, logger logging.Logger) *BackfillManager {
	return &BackfillManager{
		ethClient:  ethClient,
		baseClient: baseClient,
		optClient:  optClient,
		logger:     logger,
		batchSize:  DefaultBackfillBatchSize,
		delay:      DefaultBackfillDelay,
	}
}

// SetBatchSize sets the batch size for backfill operations
func (bm *BackfillManager) SetBatchSize(size uint64) {
	bm.batchSize = size
}

// SetDelay sets the delay between batches
func (bm *BackfillManager) SetDelay(delay time.Duration) {
	bm.delay = delay
}

// BackfillEvents performs a backfill operation for the specified configuration
func (bm *BackfillManager) BackfillEvents(ctx context.Context, config BackfillConfig, progressCallback func(*BackfillProgress)) error {
	client := bm.getClientForChain(config.ChainID)
	if client == nil {
		return fmt.Errorf("no client available for chain %s", config.ChainID)
	}

	// Validate configuration
	if err := bm.validateBackfillConfig(ctx, client, &config); err != nil {
		return fmt.Errorf("invalid backfill config: %w", err)
	}

	totalBlocks := config.EndBlock - config.StartBlock + 1
	progress := &BackfillProgress{
		ChainID:      config.ChainID,
		StartBlock:   config.StartBlock,
		EndBlock:     config.EndBlock,
		CurrentBlock: config.StartBlock,
		TotalBlocks:  totalBlocks,
		StartTime:    time.Now().UTC(),
		LastUpdate:   time.Now().UTC(),
	}

	bm.logger.Infof("Starting backfill for chain %s from block %d to %d (%d blocks)",
		config.ChainID, config.StartBlock, config.EndBlock, totalBlocks)

	// Process events in batches
	for currentBlock := config.StartBlock; currentBlock <= config.EndBlock; {
		// Check for cancellation
		select {
		case <-ctx.Done():
			progress.Error = "cancelled"
			if progressCallback != nil {
				progressCallback(progress)
			}
			return ctx.Err()
		default:
		}

		// Calculate batch end block
		batchEnd := currentBlock + config.BatchSize - 1
		if batchEnd > config.EndBlock {
			batchEnd = config.EndBlock
		}

		// Process batch
		eventsFound, err := bm.processBatch(ctx, client, config.ChainID, currentBlock, batchEnd, config.EventTypes)
		if err != nil {
			progress.Error = err.Error()
			if progressCallback != nil {
				progressCallback(progress)
			}
			return fmt.Errorf("failed to process batch %d-%d: %w", currentBlock, batchEnd, err)
		}

		// Update progress
		progress.CurrentBlock = batchEnd
		progress.ProcessedBlocks = batchEnd - config.StartBlock + 1
		progress.EventsFound += eventsFound
		progress.LastUpdate = time.Now().UTC()

		if progressCallback != nil {
			progressCallback(progress)
		}

		bm.logger.Debugf("Processed batch %d-%d on chain %s, found %d events",
			currentBlock, batchEnd, config.ChainID, eventsFound)

		// Move to next batch
		currentBlock = batchEnd + 1

		// Add delay between batches to avoid overwhelming the RPC
		if bm.delay > 0 && currentBlock <= config.EndBlock {
			time.Sleep(bm.delay)
		}
	}

	progress.Completed = true
	progress.LastUpdate = time.Now().UTC()

	if progressCallback != nil {
		progressCallback(progress)
	}

	bm.logger.Infof("Completed backfill for chain %s, processed %d blocks, found %d events",
		config.ChainID, progress.ProcessedBlocks, progress.EventsFound)

	return nil
}

// BackfillMissingBlocks identifies and backfills missing blocks since last processed block
func (bm *BackfillManager) BackfillMissingBlocks(ctx context.Context, stateManager *StateManager) error {
	chains := []struct {
		chainID      string
		client       *ethclient.Client
		getLastBlock func(context.Context) (uint64, error)
	}{
		{"17000", bm.ethClient, stateManager.GetLastPolledEthBlock},
		{"84532", bm.baseClient, stateManager.GetLastPolledBaseBlock},
		{"11155420", bm.optClient, stateManager.GetLastPolledOptBlock},
	}

	for _, chain := range chains {
		if err := bm.backfillChainMissingBlocks(ctx, chain.chainID, chain.client, chain.getLastBlock); err != nil {
			bm.logger.Errorf("Failed to backfill missing blocks for chain %s: %v", chain.chainID, err)
			// Continue with other chains instead of failing completely
		}
	}

	return nil
}

// backfillChainMissingBlocks backfills missing blocks for a specific chain
func (bm *BackfillManager) backfillChainMissingBlocks(ctx context.Context, chainID string, client *ethclient.Client, getLastBlock func(context.Context) (uint64, error)) error {
	// Get current blockchain block
	currentBlock, err := client.BlockNumber(ctx)
	if err != nil {
		return fmt.Errorf("failed to get current block for chain %s: %w", chainID, err)
	}

	// Get last processed block
	lastProcessed, err := getLastBlock(ctx)
	if err != nil {
		return fmt.Errorf("failed to get last processed block for chain %s: %w", chainID, err)
	}

	// Check if backfill is needed
	if currentBlock <= lastProcessed {
		bm.logger.Debugf("No missing blocks for chain %s (current: %d, last processed: %d)",
			chainID, currentBlock, lastProcessed)
		return nil
	}

	missingBlocks := currentBlock - lastProcessed
	bm.logger.Infof("Found %d missing blocks for chain %s (from %d to %d)",
		missingBlocks, chainID, lastProcessed+1, currentBlock)

	// Perform backfill
	config := BackfillConfig{
		BatchSize:  bm.batchSize,
		Delay:      bm.delay,
		StartBlock: lastProcessed + 1,
		EndBlock:   currentBlock,
		ChainID:    chainID,
		EventTypes: bm.getEventTypesForChain(chainID),
	}

	progressCallback := func(progress *BackfillProgress) {
		bm.logger.Infof("Backfill progress for chain %s: %d/%d blocks (%.1f%%)",
			chainID, progress.ProcessedBlocks, progress.TotalBlocks,
			float64(progress.ProcessedBlocks)/float64(progress.TotalBlocks)*100)
	}

	return bm.BackfillEvents(ctx, config, progressCallback)
}

// processBatch processes a batch of blocks for events
func (bm *BackfillManager) processBatch(ctx context.Context, client *ethclient.Client, chainID string, fromBlock, toBlock uint64, eventTypes []string) (uint64, error) {
	var eventsFound uint64

	switch chainID {
	case "17000": // Ethereum Holesky
		avsAddr := common.HexToAddress(config.GetAvsGovernanceAddress())

		for _, eventType := range eventTypes {
			switch eventType {
			case "OperatorRegistered":
				if err := events.ProcessOperatorRegisteredEvents(client, avsAddr, fromBlock, toBlock, bm.logger); err != nil {
					return eventsFound, fmt.Errorf("failed to process OperatorRegistered events: %w", err)
				}
				eventsFound++
			case "OperatorUnregistered":
				if err := events.ProcessOperatorUnregisteredEvents(client, avsAddr, fromBlock, toBlock, bm.logger); err != nil {
					return eventsFound, fmt.Errorf("failed to process OperatorUnregistered events: %w", err)
				}
				eventsFound++
			}
		}

	case "84532": // Base Sepolia
		attAddr := common.HexToAddress(config.GetAttestationCenterAddress())

		for _, eventType := range eventTypes {
			switch eventType {
			case "TaskSubmitted":
				if err := events.ProcessTaskSubmittedEvents(client, attAddr, fromBlock, toBlock, bm.logger); err != nil {
					return eventsFound, fmt.Errorf("failed to process TaskSubmitted events: %w", err)
				}
				eventsFound++
			case "TaskRejected":
				if err := events.ProcessTaskRejectedEvents(client, attAddr, fromBlock, toBlock, bm.logger); err != nil {
					return eventsFound, fmt.Errorf("failed to process TaskRejected events: %w", err)
				}
				eventsFound++
			}
		}

	case "11155420": // Optimism Sepolia
		// Add Optimism-specific event processing here
		bm.logger.Debugf("Processed OPT blocks %d-%d (events processing not implemented)", fromBlock, toBlock)
	}

	return eventsFound, nil
}

// validateBackfillConfig validates the backfill configuration
func (bm *BackfillManager) validateBackfillConfig(ctx context.Context, client *ethclient.Client, config *BackfillConfig) error {
	// Set end block to current if not specified
	if config.EndBlock == 0 {
		currentBlock, err := client.BlockNumber(ctx)
		if err != nil {
			return fmt.Errorf("failed to get current block: %w", err)
		}
		config.EndBlock = currentBlock
	}

	// Validate block range
	if config.StartBlock > config.EndBlock {
		return fmt.Errorf("start block (%d) cannot be greater than end block (%d)",
			config.StartBlock, config.EndBlock)
	}

	// Check safety limits
	totalBlocks := config.EndBlock - config.StartBlock + 1
	if totalBlocks > MaxBackfillBlocks {
		return fmt.Errorf("backfill range too large (%d blocks, max %d)",
			totalBlocks, MaxBackfillBlocks)
	}

	// Set default event types if not specified
	if len(config.EventTypes) == 0 {
		config.EventTypes = bm.getEventTypesForChain(config.ChainID)
	}

	// Validate batch size
	if config.BatchSize == 0 {
		config.BatchSize = DefaultBackfillBatchSize
	}

	return nil
}

// getClientForChain returns the appropriate client for a chain
func (bm *BackfillManager) getClientForChain(chainID string) *ethclient.Client {
	switch chainID {
	case "17000":
		return bm.ethClient
	case "84532":
		return bm.baseClient
	case "11155420":
		return bm.optClient
	default:
		return nil
	}
}

// getEventTypesForChain returns the default event types for a chain
func (bm *BackfillManager) getEventTypesForChain(chainID string) []string {
	switch chainID {
	case "17000": // Ethereum Holesky
		return []string{"OperatorRegistered", "OperatorUnregistered"}
	case "84532": // Base Sepolia
		return []string{"TaskSubmitted", "TaskRejected"}
	case "11155420": // Optimism Sepolia
		return []string{} // Add Optimism events when implemented
	default:
		return []string{}
	}
}

// GetBackfillHealth returns health information about backfill operations
func (bm *BackfillManager) GetBackfillHealth() map[string]interface{} {
	return map[string]interface{}{
		"batch_size":       bm.batchSize,
		"delay":            bm.delay.String(),
		"max_blocks":       MaxBackfillBlocks,
		"supported_chains": []string{"17000", "84532", "11155420"},
		"available_events": map[string][]string{
			"17000":    bm.getEventTypesForChain("17000"),
			"84532":    bm.getEventTypesForChain("84532"),
			"11155420": bm.getEventTypesForChain("11155420"),
		},
	}
}
