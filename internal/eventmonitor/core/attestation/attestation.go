package attestation

import (
	"context"
	"encoding/hex"
	"fmt"
	"math/big"
	"strconv"
	"strings"
	"sync"
	"time"

	// "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	"github.com/trigg3rX/triggerx-backend/internal/eventmonitor/config"
	"github.com/trigg3rX/triggerx-backend/internal/eventmonitor/rpc/clients/taskmonitor"
	nodeclient "github.com/trigg3rX/triggerx-backend/pkg/client/nodeclient"
	"github.com/trigg3rX/triggerx-backend/pkg/ipfs"
	"github.com/trigg3rX/triggerx-backend/pkg/observability"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

// PermanentPoller polls Base networks (mainnet and sepolia) for AttestationCenter events
type PermanentPoller struct {
	logger            observability.Logger
	tracer            observability.Tracer
	taskMonitorClient *taskmonitor.Client
	ipfsClient        ipfs.IPFSClient
	ctx               context.Context
	cancel            context.CancelFunc
	wg                sync.WaitGroup
	mu                sync.RWMutex
	isRunning         bool
	lastBlocks        map[string]uint64 // chainID -> lastBlock
}

// NewPermanentPoller creates a new permanent poller for Base networks
func NewPermanentPoller(ctx context.Context, logger observability.Logger, tracer observability.Tracer, ipfsClient ipfs.IPFSClient) (*PermanentPoller, error) {
	ctx, cancel := context.WithCancel(ctx)

	// Create TaskMonitor RPC client
	taskMonitorClient, err := taskmonitor.NewClient(config.GetTaskMonitorRPCUrl(), logger, tracer)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to create task monitor client: %w", err)
	}

	return &PermanentPoller{
		logger:            logger,
		tracer:            tracer,
		taskMonitorClient: taskMonitorClient,
		ipfsClient:        ipfsClient,
		ctx:               ctx,
		cancel:            cancel,
		lastBlocks:        make(map[string]uint64),
	}, nil
}

// Using the processTransaction function instead of the startChainPoller function
// Start starts the permanent poller for Base networks
// func (p *PermanentPoller) Start() error {
// 	p.mu.Lock()
// 	defer p.mu.Unlock()

// 	if p.isRunning {
// 		return fmt.Errorf("permanent poller is already running")
// 	}

// 	// Start polling for Base mainnet (8453)
// 	p.wg.Add(1)
// 	go func() {
// 		defer p.wg.Done()
// 		p.startChainPoller("8453", "Base Mainnet", config.GetAttestationCenterAddress())
// 	}()

// 	// Start polling for Base sepolia (84532)
// 	p.wg.Add(1)
// 	go func() {
// 		defer p.wg.Done()
// 		p.startChainPoller("84532", "Base Sepolia", config.GetTestAttestationCenterAddress())
// 	}()

// 	p.isRunning = true
// 	p.logger.Info(p.ctx, "Permanent Base network poller started",
// 		observability.String("chain_8453", "Base Mainnet"),
// 		observability.String("chain_84532", "Base Sepolia"))

// 	return nil
// }

// Stop stops the permanent poller
// func (p *PermanentPoller) Stop() {
// 	p.mu.Lock()
// 	defer p.mu.Unlock()

// 	if !p.isRunning {
// 		return
// 	}

// 	p.logger.Info(p.ctx, "Stopping permanent Base network poller")
// 	p.cancel()
// 	p.wg.Wait()
// 	p.isRunning = false
// 	p.logger.Info(p.ctx, "Permanent Base network poller stopped")
// }

// startChainPoller starts polling for a specific chain
// func (p *PermanentPoller) startChainPoller(chainID, chainName, contractAddr string) {
// 	// Create node client
// 	rpcURLs := config.GetChainRPCUrls()
// 	rpcURL, exists := rpcURLs[chainID]
// 	if !exists {
// 		p.logger.Error(p.ctx, "RPC URL not found for chain", observability.String("chain_id", chainID))
// 		return
// 	}

// 	nodeCfg := &nodeclient.Config{
// 		APIKey:         "",
// 		BaseURL:        rpcURL,
// 		RequestTimeout: 30 * time.Second,
// 		Logger:         p.logger,
// 	}

// 	client, err := nodeclient.NewNodeClient(nodeCfg)
// 	if err != nil {
// 		p.logger.Error(p.ctx, "Failed to create node client", observability.String("chain_id", chainID), observability.Error(err))
// 		return
// 	}
// 	defer client.Close()

// 	// Get ABI for AttestationCenter
// 	attABI, err := contractAttestationCenter.ContractAttestationCenterMetaData.GetAbi()
// 	if err != nil {
// 		p.logger.Error(p.ctx, "Failed to load AttestationCenter ABI", observability.Error(err))
// 		return
// 	}

// 	// Get event signatures
// 	taskSubmittedEvent, exists := attABI.Events["TaskSubmitted"]
// 	if !exists {
// 		p.logger.Error(p.ctx, "TaskSubmitted event not found in ABI")
// 		return
// 	}

// 	taskRejectedEvent, exists := attABI.Events["TaskRejected"]
// 	if !exists {
// 		p.logger.Error(p.ctx, "TaskRejected event not found in ABI")
// 		return
// 	}

// 	contractAddress := common.HexToAddress(contractAddr)

// 	// Initialize last block
// 	blockNumberHex, err := client.EthBlockNumber(p.ctx)
// 	if err != nil {
// 		p.logger.Error(p.ctx, "Failed to get current block number", observability.String("chain_id", chainID), observability.Error(err))
// 		return
// 	}
// 	lastBlock, err := hexToUint64(blockNumberHex)
// 	if err != nil {
// 		p.logger.Error(p.ctx, "Failed to parse block number", observability.String("chain_id", chainID), observability.Error(err))
// 		return
// 	}

// 	// Look back a few blocks on startup
// 	lookback := config.GetLookbackBlocks()
// 	if lastBlock > lookback {
// 		lastBlock = lastBlock - lookback
// 	} else {
// 		lastBlock = 0
// 	}

// 	p.mu.Lock()
// 	p.lastBlocks[chainID] = lastBlock
// 	p.mu.Unlock()

// 	p.logger.Info(p.ctx, "Starting poller",
// 		observability.String("chain_id", chainID),
// 		observability.String("chain_name", chainName),
// 		observability.String("contract_address", contractAddr),
// 		observability.Uint64("start_block", lastBlock))

// 	// Poll at configured interval
// 	pollInterval := config.GetPollInterval()
// 	ticker := time.NewTicker(pollInterval)
// 	defer ticker.Stop()

// 	for {
// 		select {
// 		case <-p.ctx.Done():
// 			return
// 		case <-ticker.C:
// 			// Get current block
// 			blockNumberHex, err := client.EthBlockNumber(p.ctx)
// 			if err != nil {
// 				p.logger.Error(p.ctx, "Failed to get current block number", observability.String("chain_id", chainID), observability.Error(err))
// 				continue
// 			}
// 			currentBlock, err := hexToUint64(blockNumberHex)
// 			if err != nil {
// 				p.logger.Error(p.ctx, "Failed to parse block number", observability.String("chain_id", chainID), observability.Error(err))
// 				continue
// 			}

// 			p.mu.RLock()
// 			fromBlock := p.lastBlocks[chainID]
// 			p.mu.RUnlock()

// 			if currentBlock <= fromBlock {
// 				continue
// 			}

// 			// Poll for TaskSubmitted events
// 			if err := p.pollEvent(client, chainID, chainName, contractAddress, taskSubmittedEvent, "TaskSubmitted", fromBlock+1, currentBlock); err != nil {
// 				p.logger.Error(p.ctx, "Failed to poll TaskSubmitted events", observability.String("chain_id", chainID), observability.Error(err))
// 			}

// 			// Poll for TaskRejected events
// 			if err := p.pollEvent(client, chainID, chainName, contractAddress, taskRejectedEvent, "TaskRejected", fromBlock+1, currentBlock); err != nil {
// 				p.logger.Error(p.ctx, "Failed to poll TaskRejected events", observability.String("chain_id", chainID), observability.Error(err))
// 			}

// 			// Update last block
// 			p.mu.Lock()
// 			p.lastBlocks[chainID] = currentBlock
// 			p.mu.Unlock()
// 		}
// 	}
// }

// pollEvent polls for a specific event
// func (p *PermanentPoller) pollEvent(client *nodeclient.NodeClient, chainID, chainName string, contractAddr common.Address, event abi.Event, eventName string, fromBlock, toBlock uint64) error {
// 	const maxRange uint64 = 10

// 	for cur := fromBlock; cur <= toBlock; {
// 		chunkEnd := cur + maxRange - 1
// 		if chunkEnd > toBlock {
// 			chunkEnd = toBlock
// 		}

// 		// Build filter query
// 		fq := ethereum.FilterQuery{
// 			Addresses: []common.Address{contractAddr},
// 			Topics:    [][]common.Hash{{event.ID}},
// 			FromBlock: new(big.Int).SetUint64(cur),
// 			ToBlock:   new(big.Int).SetUint64(chunkEnd),
// 		}

// 		// Convert to EthGetLogsParams
// 		params := convertFilterQueryToEthGetLogsParams(fq, cur, chunkEnd)

// 		// Query logs
// 		logs, err := client.EthGetLogs(p.ctx, params)
// 		if err != nil {
// 			if p.ctx.Err() != nil {
// 				return nil
// 			}
// 			p.logger.Error(p.ctx, "EthGetLogs failed",
// 				observability.String("chain_id", chainID),
// 				observability.String("event_name", eventName),
// 				observability.Uint64("from_block", cur),
// 				observability.Uint64("to_block", chunkEnd),
// 				observability.String("contract_address", contractAddr.Hex()),
// 				observability.String("event_id", event.ID.Hex()),
// 				observability.Error(err))
// 			cur = chunkEnd + 1
// 			continue
// 		}

// 		// Process logs
// 		for _, nodeLog := range logs {
// 			if err := p.processLog(chainID, chainName, event, eventName, nodeLog); err != nil {
// 				p.logger.Error(p.ctx, "Failed to process log",
// 					observability.String("chain_id", chainID),
// 					observability.String("event_name", eventName),
// 					observability.String("tx_hash", nodeLog.TransactionHash),
// 					observability.Error(err))
// 			}
// 		}

// 		cur = chunkEnd + 1
// 	}

// 	return nil
// }

// processLog processes a log and sends it to TaskMonitor
func (p *PermanentPoller) processLog(chainID, chainName string, event abi.Event, eventName string, nodeLog nodeclient.Log) error {
	// Convert nodeclient.Log to ethtypes.Log
	lg, err := convertNodeLogToTypesLog(nodeLog)
	if err != nil {
		return fmt.Errorf("failed to convert log: %w", err)
	}

	// Parse event data to get the IPFS CID from proofOfTask
	parsedData, err := p.parseEventData(event, lg)
	if err != nil {
		return fmt.Errorf("failed to parse event data: %w", err)
	}

	// Log parsed data for debugging
	p.logger.Debug(p.ctx, "Parsed event data",
		observability.String("tx_hash", lg.TxHash.Hex()),
		observability.Any("parsed_data_keys", getMapKeys(parsedData)))

	// Extract IPFS CID from data field (same field used when broadcasting to aggregator)
	ipfsCID, err := p.extractIPFSCID(parsedData)
	if err != nil {
		return fmt.Errorf("failed to extract IPFS CID: %w", err)
	}

	// Skip internal tasks (no IPFS data needed)
	taskDefinitionIdStr, ok := parsedData["taskDefinitionId"].(string)
	if ok {
		taskDefID, _ := strconv.ParseInt(taskDefinitionIdStr, 0, 64)
		if taskDefID == 10001 || taskDefID == 10002 {
			p.logger.Debug(p.ctx, "Skipping internal task", observability.Int64("task_definition_id", taskDefID))
			return nil
		}
	}

	// Fetch IPFS data
	ipfsData, err := p.ipfsClient.Fetch(p.ctx, ipfsCID)
	if err != nil {
		p.logger.Error(p.ctx, "Failed to fetch IPFS data",
			observability.String("ipfs_cid", ipfsCID),
			observability.String("tx_hash", lg.TxHash.Hex()),
			observability.Error(err))
		return fmt.Errorf("failed to fetch IPFS data: %w", err)
	}

	// Extract trace context from IPFS data and continue the trace
	ctx := p.ctx
	if ipfsData.TraceID != "" {
		ctx = observability.ContinueTrace(ctx, ipfsData.TraceID, ipfsData.SpanID)
	}

	// Create span for on-chain event processing (continues the task cycle trace)
	ctx, span := p.tracer.Start(ctx, "task.onchain.received",
		observability.WithSpanKind(trace.SpanKindConsumer),
		observability.WithAttributes(
			attribute.String("chain.id", chainID),
			attribute.String("chain.name", chainName),
			attribute.String("event.name", eventName),
			attribute.String("tx.hash", lg.TxHash.Hex()),
			attribute.Int64("block.number", int64(lg.BlockNumber)),
			attribute.String("ipfs.cid", ipfsCID),
		),
	)
	defer span.End()

	span.AddEvent("ipfs.data.fetched", observability.WithEventAttributes(
		attribute.String("ipfs.cid", ipfsCID),
	))

	// Determine if task was accepted based on event name
	isAccepted := eventName != "TaskRejected"

	// Extract taskNumber from parsed event data
	taskNumber := int64(0)
	if taskNumberVal, exists := parsedData["taskNumber"]; exists {
		switch v := taskNumberVal.(type) {
		case uint32:
			taskNumber = int64(v)
		case *big.Int:
			taskNumber = v.Int64()
		case big.Int:
			taskNumber = v.Int64()
		case uint64:
			taskNumber = int64(v)
		case uint:
			taskNumber = int64(v)
		case int64:
			taskNumber = v
		case int32:
			taskNumber = int64(v)
		case int:
			taskNumber = int64(v)
		default:
			p.logger.Warn(p.ctx, "Failed to parse taskNumber, using 0",
				observability.String("tx_hash", lg.TxHash.Hex()),
				observability.Any("task_number_type", fmt.Sprintf("%T", v)),
				observability.Any("task_number_value", v))
		}
	} else {
		p.logger.Warn(p.ctx, "taskNumber not found in parsed event data, using 0",
			observability.String("tx_hash", lg.TxHash.Hex()),
			observability.Any("parsed_data_keys", getMapKeys(parsedData)))
	}

	// Extract attesterIds from parsed event data
	attesterIds := []int64{}
	if attesterIdsVal, exists := parsedData["attestersIds"]; exists {
		switch v := attesterIdsVal.(type) {
		case []*big.Int:
			attesterIds = make([]int64, len(v))
			for i, id := range v {
				attesterIds[i] = id.Int64()
			}
		case []big.Int:
			attesterIds = make([]int64, len(v))
			for i, id := range v {
				attesterIds[i] = id.Int64()
			}
		case []uint64:
			attesterIds = make([]int64, len(v))
			for i, id := range v {
				attesterIds[i] = int64(id)
			}
		case []uint32:
			attesterIds = make([]int64, len(v))
			for i, id := range v {
				attesterIds[i] = int64(id)
			}
		case []int64:
			attesterIds = v
		case []int:
			attesterIds = make([]int64, len(v))
			for i, id := range v {
				attesterIds[i] = int64(id)
			}
		case []interface{}:
			attesterIds = make([]int64, 0, len(v))
			for _, idVal := range v {
				switch id := idVal.(type) {
				case *big.Int:
					attesterIds = append(attesterIds, id.Int64())
				case big.Int:
					attesterIds = append(attesterIds, id.Int64())
				case uint64:
					attesterIds = append(attesterIds, int64(id))
				case uint32:
					attesterIds = append(attesterIds, int64(id))
				case int64:
					attesterIds = append(attesterIds, id)
				case int:
					attesterIds = append(attesterIds, int64(id))
				default:
					p.logger.Warn(p.ctx, "Failed to parse attester ID, skipping",
						observability.String("tx_hash", lg.TxHash.Hex()),
						observability.Any("attester_id_type", fmt.Sprintf("%T", id)),
						observability.Any("attester_id_value", id))
				}
			}
		default:
			p.logger.Warn(p.ctx, "Failed to parse attesterIds, using empty slice",
				observability.String("tx_hash", lg.TxHash.Hex()),
				observability.Any("attester_ids_type", fmt.Sprintf("%T", v)),
				observability.Any("attester_ids_value", v))
		}
	} else {
		p.logger.Debug(p.ctx, "attestersIds not found in parsed event data, using empty slice",
			observability.String("tx_hash", lg.TxHash.Hex()),
			observability.Any("parsed_data_keys", getMapKeys(parsedData)))
	}

	// Send to TaskMonitor via RPC with event data and IPFS CID
	if err := p.taskMonitorClient.ReportTaskConsensusStatus(ctx, types.ReportTaskConsensusStatusRequest{
		TaskID:               ipfsData.ActionData.TaskID,
		Network:              string(ipfsData.TaskData.Network),
		TaskNumber:           taskNumber,
		TaskOpxActualCost:    ipfsData.ActionData.TotalFee,
		TaskSubmissionTxHash: lg.TxHash.Hex(),
		IsAccepted:           isAccepted,
		AttesterIds:          attesterIds,
		IPFSDataCID:          ipfsCID,
	}); err != nil {
		span.RecordError(err, observability.WithErrorAttributes(
			attribute.String("error.type", "rpc_call_failed"),
		))
		span.SetStatus(codes.Error, "failed to send consensus event to task monitor")
		return fmt.Errorf("failed to send to task monitor: %w", err)
	}

	span.AddEvent("consensus.event.sent")
	span.SetStatus(codes.Ok, "consensus event sent successfully")

	p.logger.Info(ctx, "Consensus event processed and sent to TaskMonitor",
		observability.String("tx_hash", lg.TxHash.Hex()),
		observability.String("event_name", eventName),
		observability.Bool("is_accepted", isAccepted),
		observability.String("ipfs_cid", ipfsCID))

	return nil
}

// extractIPFSCID extracts the IPFS CID from the data field (same field used when broadcasting to aggregator).
// The CID is stored in the Data field as bytes when broadcasting (executor.go:385: Data: []byte(cid))
func (p *PermanentPoller) extractIPFSCID(parsedData map[string]interface{}) (string, error) {
	v, exists := parsedData["data"]
	if !exists {
		return "", fmt.Errorf("data field not found in parsed event data")
	}
	cid, err := p.decodeValueToCID("data", v)
	if err != nil {
		return "", fmt.Errorf("decode data: %w", err)
	}
	if cid == "" {
		return "", fmt.Errorf("data decoded to empty CID")
	}
	p.logger.Debug(p.ctx, "Using CID from data field",
		observability.String("cid", cid))
	return cid, nil
}

// decodeValueToCID decodes a value to IPFS CID using keeper-style logic (copied from keeper validation):
// hex string (0x-prefix) -> hex.DecodeString(s[2:]) -> string -> CID;
// []byte -> string -> CID; plain string -> CID.
func (p *PermanentPoller) decodeValueToCID(fieldName string, v interface{}) (string, error) {
	switch val := v.(type) {
	case []byte:
		cid := string(val)
		p.logger.Debug(p.ctx, "Decoded field for CID",
			observability.String("field", fieldName),
			observability.String("type", "[]byte"),
			observability.Int("cid_len", len(cid)))
		return cid, nil
	case string:
		if len(val) > 2 && val[:2] == "0x" {
			// Decode hex string (remove "0x" prefix before decoding) - same as keeper validation
			dataBytes, err := hex.DecodeString(val[2:])
			if err != nil {
				return "", fmt.Errorf("hex decode failed: %w", err)
			}
			cid := string(dataBytes)
			p.logger.Debug(p.ctx, "Decoded field for CID",
				observability.String("field", fieldName),
				observability.String("type", "hex"),
				observability.Int("cid_len", len(cid)))
			return cid, nil
		}
		// Plain string - use as-is (same as keeper validation)
		p.logger.Debug(p.ctx, "Decoded field for CID",
			observability.String("field", fieldName),
			observability.String("type", "string"),
			observability.Int("cid_len", len(val)))
		return val, nil
	default:
		return "", fmt.Errorf("unexpected type %T", v)
	}
}

// parseEventData parses event data from log
func (p *PermanentPoller) parseEventData(event abi.Event, lg ethtypes.Log) (map[string]interface{}, error) {
	// Decode non-indexed fields
	nonIndexedArgs := make(abi.Arguments, 0)
	for _, input := range event.Inputs {
		if !input.Indexed {
			nonIndexedArgs = append(nonIndexedArgs, input)
		}
	}

	nonIndexed := make(map[string]interface{})
	if len(nonIndexedArgs) > 0 {
		allUnpacked := make(map[string]interface{})
		if err := event.Inputs.UnpackIntoMap(allUnpacked, lg.Data); err != nil {
			return nil, fmt.Errorf("failed to unpack event data: %w", err)
		}

		for _, input := range event.Inputs {
			if !input.Indexed {
				if value, exists := allUnpacked[input.Name]; exists {
					nonIndexed[input.Name] = value
				}
			}
		}
	}

	// Parse indexed parameters from topics
	parsedData := make(map[string]interface{})
	for k, v := range nonIndexed {
		parsedData[k] = v
	}

	// Parse indexed parameters from topics (skip topics[0] which is the signature)
	topicIndex := 1
	for _, input := range event.Inputs {
		if input.Indexed {
			if topicIndex < len(lg.Topics) {
				parsedData[input.Name] = p.parseTopicData(input, lg.Topics[topicIndex])
				topicIndex++
			}
		}
	}

	return parsedData, nil
}

// getMapKeys returns the keys of a map as a slice (for logging)
func getMapKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

// ProcessTransaction processes a transaction by hash and extracts TaskSubmitted/TaskRejected events.
// When isRejected is non-nil (from aggregator log), it overrides chain-derived event type so TaskRejected is sent correctly.
func (p *PermanentPoller) ProcessTransaction(ctx context.Context, txHash, chainID string, isRejected bool) (string, error) {
	// Get chain name
	chainName := getChainName(chainID)

	// Get contract address for the chain
	contractAddr := getContractAddress(chainID)
	if contractAddr == "" {
		return "", fmt.Errorf("contract address not configured for chain: %s", chainID)
	}

	// Create node client for the chain
	rpcURLs := config.GetChainRPCUrls()
	rpcURL, exists := rpcURLs[chainID]
	if !exists {
		return "", fmt.Errorf("RPC URL not found for chain: %s", chainID)
	}

	nodeCfg := &nodeclient.Config{
		APIKey:         "",
		BaseURL:        rpcURL,
		RequestTimeout: 30 * time.Second,
		Logger:         p.logger,
	}

	client, err := nodeclient.NewNodeClient(nodeCfg)
	if err != nil {
		return "", fmt.Errorf("failed to create node client: %w", err)
	}
	defer client.Close()

	// Use hardcoded ABI events
	taskSubmittedEvent := TaskSubmittedEvent
	taskRejectedEvent := TaskRejectedEvent
	
	p.logger.Debug(ctx, "TaskSubmitted event found in ABI",
		observability.String("ABI", taskSubmittedEvent.Sig))
	p.logger.Debug(ctx, "TaskRejected event found in ABI",
		observability.String("ABI", taskRejectedEvent.Sig))

	// Get transaction receipt
	receipt, err := client.EthGetTransactionReceipt(ctx, txHash)
	if err != nil {
		return "", fmt.Errorf("failed to get transaction receipt: %w", err)
	}
	if receipt == nil {
		return "", fmt.Errorf("transaction receipt not found for hash: %s", txHash)
	}

	contractAddress := common.HexToAddress(contractAddr)
	contractAddrLower := strings.ToLower(contractAddress.Hex())

	// Get expected event IDs (normalize to lowercase for comparison)
	taskSubmittedEventID := strings.ToLower(taskSubmittedEvent.ID.Hex())
	taskRejectedEventID := strings.ToLower(taskRejectedEvent.ID.Hex())

	// Find TaskSubmitted or TaskRejected events in the receipt logs
	var eventName string
	var eventLog nodeclient.Log
	var event abi.Event

	for i, log := range receipt.Logs {
		// Check if log is from the contract address (case-insensitive)
		logAddrLower := strings.ToLower(log.Address)
		if logAddrLower != contractAddrLower {
			p.logger.Debug(ctx, "Log address mismatch",
				observability.Int("log_index", i),
				observability.String("log_address", log.Address),
				observability.String("expected_address", contractAddress.Hex()))
			continue
		}

		// Check if log has topics
		if len(log.Topics) == 0 {
			p.logger.Debug(ctx, "Log has no topics",
				observability.Int("log_index", i))
			continue
		}

		// Normalize topic to lowercase for comparison
		topic0Lower := strings.ToLower(log.Topics[0])

		p.logger.Debug(ctx, "Checking log topic",
			observability.Int("log_index", i),
			observability.String("topic0", log.Topics[0]),
			observability.String("topic0_lower", topic0Lower),
			observability.Any("all_topics", log.Topics))

		// Try to parse the log with TaskSubmitted event first
		// Convert nodeclient.Log to ethtypes.Log for parsing
		ethtypesLog, err := convertNodeLogToTypesLog(log)
		if err != nil {
			p.logger.Debug(ctx, "Failed to convert log",
				observability.Int("log_index", i),
				observability.Error(err))
			continue
		}

		// Try TaskSubmitted event - check if topic matches OR try to parse
		if topic0Lower == taskSubmittedEventID {
			// Topic matches, verify by trying to parse
			_, err := p.parseEventData(taskSubmittedEvent, ethtypesLog)
			if err == nil {
				eventName = "TaskSubmitted"
				eventLog = log
				event = taskSubmittedEvent
				p.logger.Info(ctx, "Found TaskSubmitted event",
					observability.String("tx_hash", txHash),
					observability.Int("log_index", i))
				break
			}
			p.logger.Debug(ctx, "Topic matched but parsing failed for TaskSubmitted",
				observability.Int("log_index", i),
				observability.Error(err))
		}

		// Try TaskRejected event
		if topic0Lower == taskRejectedEventID {
			// Topic matches, verify by trying to parse
			_, err := p.parseEventData(taskRejectedEvent, ethtypesLog)
			if err == nil {
				eventName = "TaskRejected"
				eventLog = log
				event = taskRejectedEvent
				p.logger.Info(ctx, "Found TaskRejected event",
					observability.String("tx_hash", txHash),
					observability.Int("log_index", i))
				break
			}
			p.logger.Debug(ctx, "Topic matched but parsing failed for TaskRejected",
				observability.Int("log_index", i),
				observability.Error(err))
		}

		// If topic doesn't match, try parsing anyway (in case ABI signature is wrong)
		// This is a fallback to handle ABI mismatches
		if eventName == "" {
			// Try TaskSubmitted
			_, err := p.parseEventData(taskSubmittedEvent, ethtypesLog)
			if err == nil {
				// Check if topic0 matches (even if case differs)
				actualTopic := strings.ToLower(ethtypesLog.Topics[0].Hex())
				if actualTopic == topic0Lower {
					eventName = "TaskSubmitted"
					eventLog = log
					event = taskSubmittedEvent
					p.logger.Info(ctx, "Found TaskSubmitted event (via parsing fallback)",
						observability.String("tx_hash", txHash),
						observability.Int("log_index", i),
						observability.String("actual_topic", actualTopic),
						observability.String("expected_topic", taskSubmittedEventID))
					break
				}
			}

			// Try TaskRejected
			_, err = p.parseEventData(taskRejectedEvent, ethtypesLog)
			if err == nil {
				actualTopic := strings.ToLower(ethtypesLog.Topics[0].Hex())
				if actualTopic == topic0Lower {
					eventName = "TaskRejected"
					eventLog = log
					event = taskRejectedEvent
					p.logger.Info(ctx, "Found TaskRejected event (via parsing fallback)",
						observability.String("tx_hash", txHash),
						observability.Int("log_index", i),
						observability.String("actual_topic", actualTopic),
						observability.String("expected_topic", taskRejectedEventID))
					break
				}
			}
		}
	}

	if eventName == "" {
		// Log all topics for debugging
		allTopics := make([]string, 0)
		for _, log := range receipt.Logs {
			if strings.ToLower(log.Address) == contractAddrLower && len(log.Topics) > 0 {
				allTopics = append(allTopics, log.Topics[0])
			}
		}
		topicsStr := strings.Join(allTopics, ", ")
		p.logger.Error(ctx, "No TaskSubmitted or TaskRejected event found in transaction",
			observability.String("tx_hash", txHash),
			observability.String("chain_id", chainID),
			observability.String("contract_address", contractAddr),
			observability.String("found_topics", topicsStr),
			observability.String("expected_task_submitted", taskSubmittedEventID),
			observability.String("expected_task_rejected", taskRejectedEventID))
		return "", fmt.Errorf("no TaskSubmitted or TaskRejected event found in transaction: %s", txHash)
	}

	// Use aggregator-log hint when provided so TaskRejected is sent correctly (chain topic may not distinguish)
	if isRejected {
		eventName = "TaskRejected"
		event = taskRejectedEvent
	} else {
		eventName = "TaskSubmitted"
		event = taskSubmittedEvent
	}
	p.logger.Debug(ctx, "Using event type from aggregator log",
		observability.String("event_name", eventName),
		observability.Bool("is_rejected", isRejected))

	// Process the log
	if err := p.processLog(chainID, chainName, event, eventName, eventLog); err != nil {
		return "", fmt.Errorf("failed to process log: %w", err)
	}

	return eventName, nil
}

// getChainName returns the chain name for a given chain ID
func getChainName(chainID string) string {
	switch chainID {
	case "8453":
		return "Base Mainnet"
	case "84532":
		return "Base Sepolia"
	default:
		return fmt.Sprintf("Chain %s", chainID)
	}
}

// getContractAddress returns the contract address for a given chain ID
func getContractAddress(chainID string) string {
	switch chainID {
	case "8453":
		return config.GetAttestationCenterAddress()
	case "84532":
		return config.GetTestAttestationCenterAddress()
	default:
		return ""
	}
}

// parseTopicData parses topic data based on the input type
func (p *PermanentPoller) parseTopicData(input abi.Argument, topic common.Hash) interface{} {
	switch input.Type.String() {
	case "address":
		return common.HexToAddress(topic.Hex()).Hex()
	case "uint256", "uint128", "uint64", "uint32", "uint16", "uint8":
		return new(big.Int).SetBytes(topic.Bytes()).String()
	case "int256", "int128", "int64", "int32", "int16", "int8":
		value := new(big.Int).SetBytes(topic.Bytes())
		if value.Bit(255) == 1 {
			max := new(big.Int).Lsh(big.NewInt(1), 256)
			value.Sub(value, max)
		}
		return value.String()
	case "bytes32":
		return topic.Hex()
	case "bool":
		return topic.Big().Cmp(big.NewInt(0)) != 0
	default:
		return topic.Hex()
	}
}

// Helper functions

func hexToUint64(hexStr string) (uint64, error) {
	if len(hexStr) >= 2 && hexStr[:2] == "0x" {
		hexStr = hexStr[2:]
	}
	return strconv.ParseUint(hexStr, 16, 64)
}

// func convertFilterQueryToEthGetLogsParams(fq ethereum.FilterQuery, fromBlock, toBlock uint64) nodeclient.EthGetLogsParams {
// 	fromHex := uint64ToHex(fromBlock)
// 	toHex := uint64ToHex(toBlock)
// 	fromBlockNum := nodeclient.BlockNumber(fromHex)
// 	toBlockNum := nodeclient.BlockNumber(toHex)

// 	params := nodeclient.EthGetLogsParams{
// 		FromBlock: &fromBlockNum,
// 		ToBlock:   &toBlockNum,
// 	}

// 	if len(fq.Addresses) > 0 {
// 		if len(fq.Addresses) == 1 {
// 			params.Address = fq.Addresses[0].Hex()
// 		} else {
// 			addrs := make([]string, len(fq.Addresses))
// 			for i, addr := range fq.Addresses {
// 				addrs[i] = addr.Hex()
// 			}
// 			params.Address = addrs
// 		}
// 	}

// 	if len(fq.Topics) > 0 {
// 		topics := make([]interface{}, len(fq.Topics))
// 		for i, topicGroup := range fq.Topics {
// 			if len(topicGroup) == 0 {
// 				continue
// 			}
// 			if len(topicGroup) == 1 {
// 				topics[i] = topicGroup[0].Hex()
// 			} else {
// 				topicStrs := make([]string, len(topicGroup))
// 				for j, topic := range topicGroup {
// 					topicStrs[j] = topic.Hex()
// 				}
// 				topics[i] = topicStrs
// 			}
// 		}
// 		params.Topics = topics
// 	}

// 	return params
// }

// func uint64ToHex(val uint64) string {
// 	return fmt.Sprintf("0x%x", val)
// }

func convertNodeLogToTypesLog(nodeLog nodeclient.Log) (ethtypes.Log, error) {
	blockNumber, err := hexToUint64(nodeLog.BlockNumber)
	if err != nil {
		return ethtypes.Log{}, fmt.Errorf("failed to parse block number: %w", err)
	}

	logIndex, err := hexToUint64(nodeLog.LogIndex)
	if err != nil {
		return ethtypes.Log{}, fmt.Errorf("failed to parse log index: %w", err)
	}

	txIndex, err := hexToUint64(nodeLog.TransactionIndex)
	if err != nil {
		return ethtypes.Log{}, fmt.Errorf("failed to parse transaction index: %w", err)
	}

	address := common.HexToAddress(nodeLog.Address)

	topics := make([]common.Hash, len(nodeLog.Topics))
	for i, topicStr := range nodeLog.Topics {
		topics[i] = common.HexToHash(topicStr)
	}

	var data []byte
	if len(nodeLog.Data) >= 2 && nodeLog.Data[:2] == "0x" {
		data, err = hex.DecodeString(nodeLog.Data[2:])
	} else {
		data, err = hex.DecodeString(nodeLog.Data)
	}
	if err != nil {
		return ethtypes.Log{}, fmt.Errorf("failed to decode data: %w", err)
	}

	blockHash := common.HexToHash(nodeLog.BlockHash)
	txHash := common.HexToHash(nodeLog.TransactionHash)

	return ethtypes.Log{
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
