package worker

import (
	"context"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	nodeclient "github.com/trigg3rX/triggerx-backend/pkg/client/nodeclient"
	"github.com/trigg3rX/triggerx-backend/pkg/observability"

	"github.com/trigg3rX/triggerx-backend/internal/schedulers/condition/metrics"
)

// checkForEvents checks for new events since the last processed block
func (w *EventWorker) checkForEvents(ctx context.Context, contractAddr common.Address, eventSig common.Hash) error {
	// Get current block number
	blockHex, err := w.ChainClient.EthBlockNumber(ctx)
	if err != nil {
		metrics.TrackCriticalError("rpc_block_number_failed")
		return fmt.Errorf("failed to get current block number: %w", err)
	}
	currentBlock, err := hexToUint64(blockHex)
	if err != nil {
		metrics.TrackCriticalError("rpc_block_number_parse_failed")
		return fmt.Errorf("failed to parse block number: %w", err)
	}

	// Check if there are new blocks to process
	if currentBlock <= w.LastBlock {
		return nil // No new blocks to process
	}

	// Query logs for events in chunks to respect API limits
	// Alchemy free tier allows max 10 blocks per eth_getLogs request
	const maxBlockRange = 10
	fromBlock := w.LastBlock + 1
	allLogs := []types.Log{}

	for fromBlock <= currentBlock {
		toBlock := fromBlock + maxBlockRange - 1
		if toBlock > currentBlock {
			toBlock = currentBlock
		}

		// Build FilterQuery for conversion
		query := ethereum.FilterQuery{
			Addresses: []common.Address{contractAddr},
			Topics:    [][]common.Hash{{eventSig}},
		}

		w.Logger.Debug(ctx, "Querying block range",
			observability.String("job_id", w.EventWorkerData.JobID.String()),
			observability.Uint64("from_block", fromBlock),
			observability.Uint64("to_block", toBlock),
			observability.Uint64("range_size", toBlock-fromBlock+1),
		)

		// Convert FilterQuery to EthGetLogsParams
		params := convertFilterQueryToEthGetLogsParams(query, fromBlock, toBlock)

		nodeLogs, err := w.ChainClient.EthGetLogs(ctx, params)
		if err != nil {
			metrics.TrackCriticalError("rpc_filter_logs_failed")
			return fmt.Errorf("failed to filter logs for blocks %d-%d: %w", fromBlock, toBlock, err)
		}

		// Convert nodeclient.Log to types.Log
		for _, nodeLog := range nodeLogs {
			log, err := convertNodeLogToTypesLog(nodeLog)
			if err != nil {
				w.Logger.Error(ctx, "Failed to convert log",
					observability.String("job_id", w.EventWorkerData.JobID.String()),
					observability.Error(err),
				)
				continue
			}
			allLogs = append(allLogs, log)
		}

		fromBlock = toBlock + 1
	}

	logs := allLogs

	// Process each event
	for _, log := range logs {
		w.Logger.Info(ctx, "Found raw event",
			observability.String("job_id", w.EventWorkerData.JobID.String()),
			observability.String("tx_hash", log.TxHash.Hex()),
			observability.Uint64("block", log.BlockNumber),
			observability.Int("topics", len(log.Topics)),
			observability.String("contract", log.Address.Hex()),
		)

		// Apply event filtering if configured
		if w.shouldFilterEvent() {
			w.Logger.Info(ctx, "Applying event filter",
				observability.String("job_id", w.EventWorkerData.JobID.String()),
				observability.String("filter_param", w.EventWorkerData.EventFilterParaName),
				observability.String("filter_value", w.EventWorkerData.EventFilterValue),
			)
			if !w.matchesEventFilter(ctx, log) {
				w.Logger.Info(ctx, "Event filtered out",
					observability.String("job_id", w.EventWorkerData.JobID.String()),
					observability.String("tx_hash", log.TxHash.Hex()),
					observability.Uint64("block", log.BlockNumber),
					observability.String("filter_param", w.EventWorkerData.EventFilterParaName),
					observability.String("filter_value", w.EventWorkerData.EventFilterValue),
				)
				continue
			} else {
				w.Logger.Info(ctx, "Event passed filter",
					observability.String("job_id", w.EventWorkerData.JobID.String()),
					observability.String("tx_hash", log.TxHash.Hex()),
				)
			}
		}

		if err := w.processEvent(ctx, log); err != nil {
			w.Logger.Error(ctx, "Failed to process event",
				observability.String("job_id", w.EventWorkerData.JobID.String()),
				observability.String("tx_hash", log.TxHash.Hex()),
				observability.Uint64("block", log.BlockNumber),
				observability.Error(err),
			)
			metrics.TrackCriticalError("event_processing_failed")
		}
	}

	// Update last processed block
	w.LastBlock = currentBlock

	fromBlock = w.LastBlock + 1
	if currentBlock <= w.LastBlock {
		fromBlock = w.LastBlock
	}

	w.Logger.Info(ctx, "Processed blocks",
		observability.String("job_id", w.EventWorkerData.JobID.String()),
		observability.Uint64("from_block", fromBlock),
		observability.Uint64("to_block", currentBlock),
		observability.Int64("blocks_scanned", int64(currentBlock)-int64(w.LastBlock)),
		observability.Int("events_found", len(logs)),
		observability.String("contract_address", contractAddr.Hex()),
		observability.String("event_signature", eventSig.Hex()),
	)

	return nil
}

// processEvent processes a single event and notifies the scheduler
func (w *EventWorker) processEvent(ctx context.Context, log types.Log) error {
	w.Logger.Info(ctx, "Event detected",
		observability.String("job_id", w.EventWorkerData.JobID.String()),
		observability.String("tx_hash", log.TxHash.Hex()),
		observability.Uint64("block", log.BlockNumber),
		observability.Uint("index", log.Index),
		observability.String("chain_id", w.EventWorkerData.TriggerChainID),
		observability.String("event", w.EventWorkerData.TriggerEvent),
	)

	// Notify scheduler about the event
	if w.TriggerCallback != nil {
		notification := &TriggerNotification{
			JobID:         w.EventWorkerData.JobID.ToBigInt(),
			TriggerTxHash: log.TxHash.Hex(),
			TriggeredAt:   time.Now(),
		}

		if err := w.TriggerCallback(ctx, notification); err != nil {
			w.Logger.Error(ctx, "Failed to notify scheduler about event",
				observability.String("job_id", w.EventWorkerData.JobID.String()),
				observability.String("tx_hash", log.TxHash.Hex()),
				observability.Error(err),
			)
			metrics.TrackCriticalError("event_notification_failed")
			return err
		} else {
			w.Logger.Info(ctx, "Successfully notified scheduler about event",
				observability.String("job_id", w.EventWorkerData.JobID.String()),
				observability.String("tx_hash", log.TxHash.Hex()),
			)
		}
	} else {
		w.Logger.Warn(ctx, "No trigger callback configured for event worker",
			observability.String("job_id", w.EventWorkerData.JobID.String()),
		)
	}

	// For non-recurring jobs, stop the worker after triggering
	if !w.EventWorkerData.Recurring {
		w.Logger.Info(ctx, "Non-recurring job triggered, stopping worker",
			observability.String("job_id", w.EventWorkerData.JobID.String()),
		)
		go w.Stop(ctx) // Stop in a goroutine to avoid deadlock
	}

	return nil
}

// shouldFilterEvent checks if event filtering is enabled
func (w *EventWorker) shouldFilterEvent() bool {
	return strings.TrimSpace(w.EventWorkerData.EventFilterParaName) != "" &&
		strings.TrimSpace(w.EventWorkerData.EventFilterValue) != ""
}

// matchesEventFilter checks if the event matches the configured filter
func (w *EventWorker) matchesEventFilter(ctx context.Context, log types.Log) bool {
	filterParam := strings.TrimSpace(w.EventWorkerData.EventFilterParaName)
	filterValue := strings.TrimSpace(w.EventWorkerData.EventFilterValue)

	if filterParam == "" || filterValue == "" {
		return true // No filtering configured
	}

	// For basic filtering, we'll check indexed topics and data fields
	// This is a simplified implementation that works with common event patterns

	// Convert filter value to compare with event data
	filterValueLower := strings.ToLower(filterValue)

	// Check indexed topics (topics[1], topics[2], topics[3] are indexed parameters)
	// topics[0] is the event signature
	for i, topic := range log.Topics {
		if i == 0 {
			continue // Skip event signature
		}

		// Convert topic to string representation for comparison
		topicStr := strings.ToLower(topic.Hex())
		topicBigInt := topic.Big()

		// Check if this might be the parameter we're looking for
		// We'll do a fuzzy match since parameter names aren't directly available in logs
		if strings.Contains(topicStr, filterValueLower) {
			w.Logger.Debug(ctx, "Event filter matched in topic",
				observability.String("job_id", w.EventWorkerData.JobID.String()),
				observability.Int("topic_index", i),
				observability.String("topic_value", topicStr),
				observability.String("filter_value", filterValue),
			)
			return true
		}

		// Also check if the filter value matches the big int representation
		if topicBigInt.String() == filterValue {
			w.Logger.Debug(ctx, "Event filter matched in topic (big int)",
				observability.String("job_id", w.EventWorkerData.JobID.String()),
				observability.Int("topic_index", i),
				observability.String("topic_value", topicBigInt.String()),
				observability.String("filter_value", filterValue),
			)
			return true
		}
	}

	// Check event data for non-indexed parameters
	dataStr := strings.ToLower(common.Bytes2Hex(log.Data))
	if strings.Contains(dataStr, filterValueLower) {
		w.Logger.Debug(ctx, "Event filter matched in data",
			observability.String("job_id", w.EventWorkerData.JobID.String()),
			observability.String("data_snippet", dataStr[:min(len(dataStr), 100)]), // Log first 100 chars for debugging
			observability.String("filter_value", filterValue),
		)
		return true
	}

	// Check if the filter value is a hex address
	if common.IsHexAddress(filterValue) {
		filterAddr := common.HexToAddress(filterValue)
		// Convert address to hash for topic comparison (addresses are left-padded to 32 bytes in topics)
		addressHash := common.HexToHash(filterAddr.Hex())
		w.Logger.Info(ctx, "Checking address filter",
			observability.String("job_id", w.EventWorkerData.JobID.String()),
			observability.String("filter_address", filterAddr.Hex()),
			observability.String("address_hash", addressHash.Hex()),
			observability.Int("num_topics", len(log.Topics)),
		)
		for i, topic := range log.Topics {
			if i == 0 {
				continue
			}
			w.Logger.Info(ctx, "Comparing topic",
				observability.String("job_id", w.EventWorkerData.JobID.String()),
				observability.Int("topic_index", i),
				observability.String("topic_value", topic.Hex()),
				observability.String("expected_hash", addressHash.Hex()),
				observability.Bool("matches", topic == addressHash),
			)
			if topic == addressHash {
				w.Logger.Info(ctx, "Event filter matched address in topic",
					observability.String("job_id", w.EventWorkerData.JobID.String()),
					observability.Int("topic_index", i),
					observability.String("address", filterAddr.Hex()),
				)
				return true
			}
		}
	}

	w.Logger.Debug(ctx, "Event filter did not match",
		observability.String("job_id", w.EventWorkerData.JobID.String()),
		observability.String("filter_param", filterParam),
		observability.String("filter_value", filterValue),
		observability.String("tx_hash", log.TxHash.Hex()),
	)

	return false
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// uint64ToHex converts uint64 to hex string with 0x prefix
func uint64ToHex(val uint64) string {
	return fmt.Sprintf("0x%x", val)
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

// convertNodeLogToTypesLog converts nodeclient.Log to types.Log
func convertNodeLogToTypesLog(nodeLog nodeclient.Log) (types.Log, error) {
	// Parse block number
	blockNumber, err := hexToUint64(nodeLog.BlockNumber)
	if err != nil {
		return types.Log{}, fmt.Errorf("failed to parse block number: %w", err)
	}

	// Parse log index
	logIndex, err := hexToUint64(nodeLog.LogIndex)
	if err != nil {
		return types.Log{}, fmt.Errorf("failed to parse log index: %w", err)
	}

	// Parse transaction index
	txIndex, err := hexToUint64(nodeLog.TransactionIndex)
	if err != nil {
		return types.Log{}, fmt.Errorf("failed to parse transaction index: %w", err)
	}

	// Parse address
	address := common.HexToAddress(nodeLog.Address)

	// Parse topics
	topics := make([]common.Hash, len(nodeLog.Topics))
	for i, topicStr := range nodeLog.Topics {
		topics[i] = common.HexToHash(topicStr)
	}

	// Parse data
	var data []byte
	if len(nodeLog.Data) >= 2 && nodeLog.Data[:2] == "0x" {
		data, err = hex.DecodeString(nodeLog.Data[2:]) // Remove 0x prefix
	} else {
		data, err = hex.DecodeString(nodeLog.Data)
	}
	if err != nil {
		return types.Log{}, fmt.Errorf("failed to decode data: %w", err)
	}

	// Parse block hash
	blockHash := common.HexToHash(nodeLog.BlockHash)

	// Parse transaction hash
	txHash := common.HexToHash(nodeLog.TransactionHash)

	return types.Log{
		Address:     address,
		Topics:      topics,
		Data:        data,
		BlockNumber: blockNumber,
		TxHash:      txHash,
		TxIndex:     uint(txIndex),
		BlockHash:   blockHash,
		Index:       uint(logIndex),
		Removed:     nodeLog.Removed,
	}, nil
}
