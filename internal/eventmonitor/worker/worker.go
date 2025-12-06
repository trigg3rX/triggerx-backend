package worker

import (
	"context"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/trigg3rX/triggerx-backend/internal/eventmonitor/config"
	"github.com/trigg3rX/triggerx-backend/internal/eventmonitor/types"
	"github.com/trigg3rX/triggerx-backend/internal/eventmonitor/webhook"
	nodeclient "github.com/trigg3rX/triggerx-backend/pkg/client/nodeclient"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
)

// Worker polls blockchain for events and distributes to subscribers
type Worker struct {
	entry         *types.RegistryEntry
	nodeClient    *nodeclient.NodeClient
	webhookClient *webhook.Client
	logger        logging.Logger
	ctx           context.Context
	cancel        context.CancelFunc
}

// NewWorker creates a new event worker
func NewWorker(
	entry *types.RegistryEntry,
	nodeClient *nodeclient.NodeClient,
	webhookClient *webhook.Client,
	logger logging.Logger,
) *Worker {
	ctx, cancel := context.WithCancel(entry.WorkerCtx)
	return &Worker{
		entry:         entry,
		nodeClient:    nodeClient,
		webhookClient: webhookClient,
		logger:        logger,
		ctx:           ctx,
		cancel:        cancel,
	}
}

// Start starts the worker polling loop
func (w *Worker) Start() {
	w.logger.Info("Starting event worker",
		"key", w.entry.Key,
		"chain_id", w.entry.ChainID)

	// Initialize last block if needed
	if w.entry.LastBlock == 0 {
		// Look back a few blocks on startup
		currentBlock, err := w.getCurrentBlock()
		if err != nil {
			w.logger.Error("Failed to get current block on startup", "error", err)
			// Set to 0, will retry on next poll
			w.entry.LastBlock = 0
		} else {
			lookback := config.GetLookbackBlocks()
			if currentBlock > lookback {
				w.entry.LastBlock = currentBlock - lookback
			} else {
				w.entry.LastBlock = 0
			}
			w.logger.Info("Initialized last block",
				"key", w.entry.Key,
				"last_block", w.entry.LastBlock,
				"current_block", currentBlock)
		}
	}

	ticker := time.NewTicker(config.GetPollInterval())
	defer ticker.Stop()

	for {
		select {
		case <-w.ctx.Done():
			w.logger.Info("Worker context cancelled, stopping", "key", w.entry.Key)
			return
		case <-ticker.C:
			if err := w.pollEvents(); err != nil {
				w.logger.Error("Error polling events", "key", w.entry.Key, "error", err)
			}
		}
	}
}

// Stop stops the worker
func (w *Worker) Stop() {
	w.cancel()
}

// pollEvents polls for new events
func (w *Worker) pollEvents() error {
	// Get current block number
	currentBlock, err := w.getCurrentBlock()
	if err != nil {
		return fmt.Errorf("failed to get current block number: %w", err)
	}

	// Check if there are new blocks to process
	if currentBlock <= w.entry.LastBlock {
		return nil // No new blocks
	}

	// Query logs in chunks
	maxBlockRange := config.GetMaxBlockRange()
	fromBlock := w.entry.LastBlock + 1

	for fromBlock <= currentBlock {
		toBlock := fromBlock + maxBlockRange - 1
		if toBlock > currentBlock {
			toBlock = currentBlock
		}

		// Query logs for this range
		logs, err := w.queryLogs(fromBlock, toBlock)
		if err != nil {
			w.logger.Error("Failed to query logs",
				"key", w.entry.Key,
				"from_block", fromBlock,
				"to_block", toBlock,
				"error", err)
			// Continue to next chunk
			fromBlock = toBlock + 1
			continue
		}

		// Process logs and notify subscribers
		for _, log := range logs {
			if err := w.processLog(log); err != nil {
				w.logger.Error("Failed to process log",
					"key", w.entry.Key,
					"tx_hash", log.TransactionHash,
					"log_index", log.LogIndex,
					"error", err)
			}
		}

		// Update last block
		w.entry.LastBlock = toBlock
		fromBlock = toBlock + 1
	}

	return nil
}

// getCurrentBlock gets the current block number
func (w *Worker) getCurrentBlock() (uint64, error) {
	blockHex, err := w.nodeClient.EthBlockNumber(w.ctx)
	if err != nil {
		return 0, err
	}

	return hexToUint64(blockHex)
}

// queryLogs queries logs for a block range
func (w *Worker) queryLogs(fromBlock, toBlock uint64) ([]nodeclient.Log, error) {
	// Build FilterQuery for conversion
	query := ethereum.FilterQuery{
		Addresses: []common.Address{w.entry.ContractAddr},
		Topics:    [][]common.Hash{{w.entry.EventSig}},
	}

	// Convert FilterQuery to EthGetLogsParams
	params := convertFilterQueryToEthGetLogsParams(query, fromBlock, toBlock)

	// Query logs
	logs, err := w.nodeClient.EthGetLogs(w.ctx, params)
	if err != nil {
		return nil, fmt.Errorf("eth_getLogs failed: %w", err)
	}

	return logs, nil
}

// processLog processes a log and notifies subscribers
func (w *Worker) processLog(log nodeclient.Log) error {
	// Convert nodeclient.Log to EventNotification
	blockNumber, err := hexToUint64(log.BlockNumber)
	if err != nil {
		return fmt.Errorf("failed to parse block number: %w", err)
	}

	logIndex, err := hexToUint(log.LogIndex)
	if err != nil {
		return fmt.Errorf("failed to parse log index: %w", err)
	}

	// Get subscribers
	w.entry.Mu.RLock()
	subscribers := make([]*types.Subscriber, 0, len(w.entry.Subscribers))
	for _, sub := range w.entry.Subscribers {
		subscribers = append(subscribers, sub)
	}
	w.entry.Mu.RUnlock()

	// Notify each subscriber
	for _, subscriber := range subscribers {
		// Check if event matches filter
		if !w.matchesFilter(log, subscriber) {
			continue
		}

		notification := &types.EventNotification{
			RequestID:    subscriber.RequestID,
			ChainID:      w.entry.ChainID,
			ContractAddr: w.entry.ContractAddr.Hex(),
			EventSig:     w.entry.EventSig.Hex(),
			BlockNumber:  blockNumber,
			TxHash:       log.TransactionHash,
			LogIndex:     logIndex,
			Topics:       log.Topics,
			Data:         log.Data,
			Timestamp:    time.Now(),
		}

		// Send webhook (non-blocking)
		go func(sub *types.Subscriber, notif *types.EventNotification) {
			if err := w.webhookClient.Send(sub.WebhookURL, notif); err != nil {
				w.logger.Error("Failed to send webhook",
					"request_id", sub.RequestID,
					"webhook_url", sub.WebhookURL,
					"error", err)
			}
		}(subscriber, notification)
	}

	return nil
}

// matchesFilter checks if a log matches the subscriber's filter
func (w *Worker) matchesFilter(log nodeclient.Log, subscriber *types.Subscriber) bool {
	if subscriber.FilterParam == "" || subscriber.FilterValue == "" {
		return true // No filter, match all
	}

	// For MVP, we only support topic filtering (indexed parameters)
	// FilterParam should be the topic index (0, 1, 2, etc.)
	// FilterValue should be the hex value to match

	// Topic 0 is always the event signature, so filter params start at topic 1
	// But we'll allow topic 0 for flexibility
	topicIndex := 0
	switch subscriber.FilterParam {
	case "topic0", "0":
		topicIndex = 0
	case "topic1", "1":
		topicIndex = 1
	case "topic2", "2":
		topicIndex = 2
	case "topic3", "3":
		topicIndex = 3
	default:
		// Try to parse as number
		var err error
		if strings.HasPrefix(subscriber.FilterParam, "topic") {
			_, err = fmt.Sscanf(subscriber.FilterParam, "topic%d", &topicIndex)
		} else {
			_, err = fmt.Sscanf(subscriber.FilterParam, "%d", &topicIndex)
		}
		if err != nil {
			w.logger.Warn("Invalid filter param, matching all",
				"filter_param", subscriber.FilterParam)
			return true
		}
	}

	// Check if topic index exists
	if topicIndex >= len(log.Topics) {
		return false
	}

	// Normalize filter value (remove 0x prefix if present, add if missing)
	filterValue := strings.ToLower(subscriber.FilterValue)
	if !strings.HasPrefix(filterValue, "0x") {
		filterValue = "0x" + filterValue
	}

	// Compare topic
	topicValue := strings.ToLower(log.Topics[topicIndex])
	return topicValue == filterValue
}

// convertFilterQueryToEthGetLogsParams converts ethereum.FilterQuery to nodeclient.EthGetLogsParams
func convertFilterQueryToEthGetLogsParams(fq ethereum.FilterQuery, fromBlock, toBlock uint64) nodeclient.EthGetLogsParams {
	fromHex := uint64ToHex(fromBlock)
	toHex := uint64ToHex(toBlock)
	fromBlockNum := nodeclient.BlockNumber(fromHex)
	toBlockNum := nodeclient.BlockNumber(toHex)

	params := nodeclient.EthGetLogsParams{
		FromBlock: &fromBlockNum,
		ToBlock:   &toBlockNum,
	}

	// Convert addresses
	if len(fq.Addresses) > 0 {
		if len(fq.Addresses) == 1 {
			params.Address = fq.Addresses[0].Hex()
		} else {
			addrs := make([]string, len(fq.Addresses))
			for i, addr := range fq.Addresses {
				addrs[i] = addr.Hex()
			}
			params.Address = addrs
		}
	}

	// Convert topics
	if len(fq.Topics) > 0 {
		topics := make([]interface{}, len(fq.Topics))
		for i, topicGroup := range fq.Topics {
			if len(topicGroup) == 0 {
				continue
			}
			if len(topicGroup) == 1 {
				topics[i] = topicGroup[0].Hex()
			} else {
				topicStrs := make([]string, len(topicGroup))
				for j, topic := range topicGroup {
					topicStrs[j] = topic.Hex()
				}
				topics[i] = topicStrs
			}
		}
		params.Topics = topics
	}

	return params
}

// hexToUint64 converts hex string to uint64
func hexToUint64(hexStr string) (uint64, error) {
	hexStr = strings.TrimPrefix(hexStr, "0x")
	value := new(big.Int)
	value, ok := value.SetString(hexStr, 16)
	if !ok {
		return 0, fmt.Errorf("invalid hex string: %s", hexStr)
	}
	return value.Uint64(), nil
}

// hexToUint converts hex string to uint
func hexToUint(hexStr string) (uint, error) {
	val, err := hexToUint64(hexStr)
	if err != nil {
		return 0, err
	}
	return uint(val), nil
}

// uint64ToHex converts uint64 to hex string with 0x prefix
func uint64ToHex(val uint64) string {
	return fmt.Sprintf("0x%x", val)
}
