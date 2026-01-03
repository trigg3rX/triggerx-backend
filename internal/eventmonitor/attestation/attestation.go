package attestation

import (
	"context"
	"encoding/hex"
	"fmt"
	"math/big"
	"strconv"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	"github.com/trigg3rX/triggerx-backend/internal/eventmonitor/config"
	"github.com/trigg3rX/triggerx-backend/internal/eventmonitor/taskmonitor"
	tmTypes "github.com/trigg3rX/triggerx-backend/internal/taskmonitor/types"
	nodeclient "github.com/trigg3rX/triggerx-backend/pkg/client/nodeclient"
	"github.com/trigg3rX/triggerx-backend/pkg/observability"

	// Contract bindings
	contractAttestationCenter "github.com/trigg3rX/triggerx-contracts/bindings/contracts/AttestationCenter"
)

// PermanentPoller polls Base networks (mainnet and sepolia) for AttestationCenter events
type PermanentPoller struct {
	logger            observability.Logger
	tracer            observability.Tracer
	taskMonitorClient *taskmonitor.Client
	ctx               context.Context
	cancel            context.CancelFunc
	wg                sync.WaitGroup
	mu                sync.RWMutex
	isRunning         bool
	lastBlocks        map[string]uint64 // chainID -> lastBlock
}

// NewPermanentPoller creates a new permanent poller for Base networks
func NewPermanentPoller(ctx context.Context, logger observability.Logger, tracer observability.Tracer) (*PermanentPoller, error) {
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
		ctx:               ctx,
		cancel:            cancel,
		lastBlocks:        make(map[string]uint64),
	}, nil
}

// Start starts the permanent poller for Base networks
func (p *PermanentPoller) Start() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.isRunning {
		return fmt.Errorf("permanent poller is already running")
	}

	// Start polling for Base mainnet (8453)
	p.wg.Add(1)
	go func() {
		defer p.wg.Done()
		p.startChainPoller("8453", "Base Mainnet", config.GetAttestationCenterAddress())
	}()

	// Start polling for Base sepolia (84532)
	p.wg.Add(1)
	go func() {
		defer p.wg.Done()
		p.startChainPoller("84532", "Base Sepolia", config.GetTestAttestationCenterAddress())
	}()

	p.isRunning = true
	p.logger.Info(p.ctx, "Permanent Base network poller started",
		observability.String("chain_8453", "Base Mainnet"),
		observability.String("chain_84532", "Base Sepolia"))

	return nil
}

// Stop stops the permanent poller
func (p *PermanentPoller) Stop() {
	p.mu.Lock()
	defer p.mu.Unlock()

	if !p.isRunning {
		return
	}

	p.logger.Info(p.ctx, "Stopping permanent Base network poller")
	p.cancel()
	p.wg.Wait()
	p.isRunning = false
	p.logger.Info(p.ctx, "Permanent Base network poller stopped")
}

// startChainPoller starts polling for a specific chain
func (p *PermanentPoller) startChainPoller(chainID, chainName, contractAddr string) {
	// Create node client
	rpcURLs := config.GetChainRPCUrls()
	rpcURL, exists := rpcURLs[chainID]
	if !exists {
		p.logger.Error(p.ctx, "RPC URL not found for chain", observability.String("chain_id", chainID))
		return
	}

	nodeCfg := &nodeclient.Config{
		APIKey:         "",
		BaseURL:        rpcURL,
		RequestTimeout: 30 * time.Second,
		Logger:         p.logger,
	}

	client, err := nodeclient.NewNodeClient(nodeCfg)
	if err != nil {
		p.logger.Error(p.ctx, "Failed to create node client", observability.String("chain_id", chainID), observability.Error(err))
		return
	}
	defer client.Close()

	// Get ABI for AttestationCenter
	attABI, err := contractAttestationCenter.ContractAttestationCenterMetaData.GetAbi()
	if err != nil {
		p.logger.Error(p.ctx, "Failed to load AttestationCenter ABI", observability.Error(err))
		return
	}

	// Get event signatures
	taskSubmittedEvent, exists := attABI.Events["TaskSubmitted"]
	if !exists {
		p.logger.Error(p.ctx, "TaskSubmitted event not found in ABI")
		return
	}

	taskRejectedEvent, exists := attABI.Events["TaskRejected"]
	if !exists {
		p.logger.Error(p.ctx, "TaskRejected event not found in ABI")
		return
	}

	contractAddress := common.HexToAddress(contractAddr)

	// Initialize last block
	blockNumberHex, err := client.EthBlockNumber(p.ctx)
	if err != nil {
		p.logger.Error(p.ctx, "Failed to get current block number", observability.String("chain_id", chainID), observability.Error(err))
		return
	}
	lastBlock, err := hexToUint64(blockNumberHex)
	if err != nil {
		p.logger.Error(p.ctx, "Failed to parse block number", observability.String("chain_id", chainID), observability.Error(err))
		return
	}

	// Look back a few blocks on startup
	lookback := config.GetLookbackBlocks()
	if lastBlock > lookback {
		lastBlock = lastBlock - lookback
	} else {
		lastBlock = 0
	}

	p.mu.Lock()
	p.lastBlocks[chainID] = lastBlock
	p.mu.Unlock()

	p.logger.Info(p.ctx, "Starting poller",
		observability.String("chain_id", chainID),
		observability.String("chain_name", chainName),
		observability.String("contract_address", contractAddr),
		observability.Uint64("start_block", lastBlock))

	// Poll every minute
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-p.ctx.Done():
			return
		case <-ticker.C:
			// Get current block
			blockNumberHex, err := client.EthBlockNumber(p.ctx)
			if err != nil {
				p.logger.Error(p.ctx, "Failed to get current block number", observability.String("chain_id", chainID), observability.Error(err))
				continue
			}
			currentBlock, err := hexToUint64(blockNumberHex)
			if err != nil {
				p.logger.Error(p.ctx, "Failed to parse block number", observability.String("chain_id", chainID), observability.Error(err))
				continue
			}

			p.mu.RLock()
			fromBlock := p.lastBlocks[chainID]
			p.mu.RUnlock()

			if currentBlock <= fromBlock {
				continue
			}

			// Poll for TaskSubmitted events
			if err := p.pollEvent(client, chainID, chainName, contractAddress, taskSubmittedEvent, "TaskSubmitted", fromBlock+1, currentBlock); err != nil {
				p.logger.Error(p.ctx, "Failed to poll TaskSubmitted events", observability.String("chain_id", chainID), observability.Error(err))
			}

			// Poll for TaskRejected events
			if err := p.pollEvent(client, chainID, chainName, contractAddress, taskRejectedEvent, "TaskRejected", fromBlock+1, currentBlock); err != nil {
				p.logger.Error(p.ctx, "Failed to poll TaskRejected events", observability.String("chain_id", chainID), observability.Error(err))
			}

			// Update last block
			p.mu.Lock()
			p.lastBlocks[chainID] = currentBlock
			p.mu.Unlock()
		}
	}
}

// pollEvent polls for a specific event
func (p *PermanentPoller) pollEvent(client *nodeclient.NodeClient, chainID, chainName string, contractAddr common.Address, event abi.Event, eventName string, fromBlock, toBlock uint64) error {
	const maxRange uint64 = 10

	for cur := fromBlock; cur <= toBlock; {
		chunkEnd := cur + maxRange - 1
		if chunkEnd > toBlock {
			chunkEnd = toBlock
		}

		// Build filter query
		fq := ethereum.FilterQuery{
			Addresses: []common.Address{contractAddr},
			Topics:    [][]common.Hash{{event.ID}},
			FromBlock: new(big.Int).SetUint64(cur),
			ToBlock:   new(big.Int).SetUint64(chunkEnd),
		}

		// Convert to EthGetLogsParams
		params := convertFilterQueryToEthGetLogsParams(fq, cur, chunkEnd)

		// Query logs
		logs, err := client.EthGetLogs(p.ctx, params)
		if err != nil {
			if p.ctx.Err() != nil {
				return nil
			}
			p.logger.Error(p.ctx, "EthGetLogs failed",
				observability.String("chain_id", chainID),
				observability.String("event_name", eventName),
				observability.Uint64("from_block", cur),
				observability.Uint64("to_block", chunkEnd),
				observability.Error(err))
			cur = chunkEnd + 1
			continue
		}

		// Process logs
		for _, nodeLog := range logs {
			if err := p.processLog(chainID, chainName, event, eventName, nodeLog); err != nil {
				p.logger.Error(p.ctx, "Failed to process log",
					observability.String("chain_id", chainID),
					observability.String("event_name", eventName),
					observability.String("tx_hash", nodeLog.TransactionHash),
					observability.Error(err))
			}
		}

		cur = chunkEnd + 1
	}

	return nil
}

// processLog processes a log and sends it to TaskMonitor
func (p *PermanentPoller) processLog(chainID, chainName string, event abi.Event, eventName string, nodeLog nodeclient.Log) error {
	// Convert nodeclient.Log to ethtypes.Log
	lg, err := convertNodeLogToTypesLog(nodeLog)
	if err != nil {
		return fmt.Errorf("failed to convert log: %w", err)
	}

	// Parse event data
	parsedData, err := p.parseEventData(event, lg)
	if err != nil {
		return fmt.Errorf("failed to parse event data: %w", err)
	}

	// Parse into TaskSubmissionData
	taskData, err := p.parseTaskSubmissionData(p.ctx, parsedData, lg.TxHash.Hex())
	if err != nil {
		return fmt.Errorf("failed to parse task submission data: %w", err)
	}

	// Create trace span
	ctx, span := p.tracer.Start(p.ctx, "attestation.event.detected",
		observability.WithSpanKind(trace.SpanKindProducer),
		observability.WithAttributes(
			attribute.String("chain.id", chainID),
			attribute.String("chain.name", chainName),
			attribute.String("event.name", eventName),
			attribute.String("tx.hash", lg.TxHash.Hex()),
			attribute.Int64("block.number", int64(lg.BlockNumber)),
		),
	)
	defer span.End()

	// Send to TaskMonitor via RPC
	if err := p.taskMonitorClient.ReportConsensusEvent(ctx, chainID, eventName, lg.TxHash.Hex(), taskData); err != nil {
		span.RecordError(err, observability.WithErrorAttributes(
			attribute.String("error.type", "rpc_call_failed"),
		))
		span.SetStatus(codes.Error, "failed to send consensus event to task monitor")
		return fmt.Errorf("failed to send to task monitor: %w", err)
	}

	span.AddEvent("consensus.event.sent")
	span.SetStatus(codes.Ok, "consensus event sent successfully")

	return nil
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

// parseTaskSubmissionData parses the event data into TaskSubmissionData
func (p *PermanentPoller) parseTaskSubmissionData(ctx context.Context, parsedData map[string]interface{}, txHash string) (*tmTypes.TaskSubmissionData, error) {
	// Extract taskDefinitionId - it's indexed, so it comes as a string (hex-encoded)
	taskDefinitionIdStr, ok := parsedData["taskDefinitionId"].(string)
	if !ok {
		return nil, fmt.Errorf("taskDefinitionId not found or invalid type")
	}

	// Convert hex string to integer
	taskDefinitionIdInt64, err := strconv.ParseInt(taskDefinitionIdStr, 0, 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse taskDefinitionId: %v", err)
	}
	taskDefinitionId := int(taskDefinitionIdInt64)

	if taskDefinitionId == 10001 || taskDefinitionId == 10002 {
		taskData := &tmTypes.TaskSubmissionData{
			TaskID: 0,
		}
		return taskData, nil
	}

	// Extract task number - it's already parsed as uint32, so we need to handle it as a number
	var taskNumber int64
	switch v := parsedData["taskNumber"].(type) {
	case string:
		// If it's a string, parse it
		var err error
		taskNumber, err = strconv.ParseInt(v, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("failed to parse taskNumber: %v", err)
		}
	case float64:
		// If it's a float64 (from JSON unmarshaling), convert to int64
		taskNumber = int64(v)
	case int64:
		taskNumber = v
	case int:
		taskNumber = int64(v)
	case uint32:
		taskNumber = int64(v)
	case uint64:
		taskNumber = int64(v)
	default:
		return nil, fmt.Errorf("taskNumber has unexpected type: %T", v)
	}

	// Extract proof of task
	proofOfTask, ok := parsedData["proofOfTask"].(string)
	if !ok {
		return nil, fmt.Errorf("proofOfTask not found or invalid type")
	}

	// Extract data field - it's bytes, so it could be []byte or string
	var data string
	switch v := parsedData["data"].(type) {
	case []byte:
		data = hex.EncodeToString(v)
	default:
		return nil, fmt.Errorf("data field has unexpected type: %T", v)
	}

	// Extract operator address
	performerAddress, ok := parsedData["operator"].(string)
	if !ok {
		return nil, fmt.Errorf("operator not found or invalid type")
	}

	// Extract attesters IDs
	attestersIdsInterface, ok := parsedData["attestersIds"]
	if !ok {
		// Try alternative field names that might be used
		if altInterface, altOk := parsedData["attesterIds"]; altOk {
			p.logger.Info(ctx, "Found attesterIds with alternative spelling")
			attestersIdsInterface = altInterface
		} else if altInterface, altOk := parsedData["attesters"]; altOk {
			p.logger.Info(ctx, "Found attesters field")
			attestersIdsInterface = altInterface
		} else {
			// Don't return error, just log and continue with empty slice
			attestersIdsInterface = []interface{}{}
		}
	}

	// Convert attestersIds to int64 slice
	var attestersIds []int64
	switch v := attestersIdsInterface.(type) {
	case []string:
		// Handle the corrected format from formatValue ([]*big.Int -> []string)
		for _, av := range v {
			if n, err := strconv.ParseInt(av, 10, 64); err == nil {
				attestersIds = append(attestersIds, n)
			} else {
				p.logger.Warn(ctx, "Failed to parse attester ID as string", observability.String("value", av), observability.Error(err))
			}
		}
	case []interface{}:
		// Fallback for legacy format
		for i, av := range v {
			switch vv := av.(type) {
			case float64:
				attestersIds = append(attestersIds, int64(vv))
			case string:
				// attempt parse decimal
				if n, err := strconv.ParseInt(vv, 10, 64); err == nil {
					attestersIds = append(attestersIds, n)
				} else {
					p.logger.Warn(ctx, "Failed to parse attester ID as string", observability.Int("index", i), observability.String("value", vv), observability.Error(err))
				}
			case *big.Int:
				attestersIds = append(attestersIds, vv.Int64())
			default:
				p.logger.Warn(ctx, "Unknown attester ID type", observability.Int("index", i), observability.String("type", fmt.Sprintf("%T", vv)), observability.String("value", fmt.Sprintf("%v", vv)))
			}
		}
	case []*big.Int:
		// Direct handling of []*big.Int
		for _, id := range v {
			attestersIds = append(attestersIds, id.Int64())
		}
	default:
		// Try to manually parse if it's a slice of unknown interface{}
		if slice, ok := v.([]interface{}); ok {
			for i, item := range slice {
				switch itemVal := item.(type) {
				case *big.Int:
					attestersIds = append(attestersIds, itemVal.Int64())
				case string:
					if n, err := strconv.ParseInt(itemVal, 10, 64); err == nil {
						attestersIds = append(attestersIds, n)
					} else {
						p.logger.Warn(ctx, "Failed to parse attester ID from string", observability.Int("index", i), observability.String("value", itemVal), observability.Error(err))
					}
				case float64:
					attestersIds = append(attestersIds, int64(itemVal))
				default:
					p.logger.Warn(ctx, "Unknown attester ID type in slice", observability.Int("index", i), observability.String("type", fmt.Sprintf("%T", itemVal)), observability.String("value", fmt.Sprintf("%v", itemVal)))
				}
			}
		}
	}

	// Create task submission data
	return &tmTypes.TaskSubmissionData{
		TaskID:               0,
		TaskNumber:           taskNumber,
		TaskDefinitionID:     taskDefinitionId,
		IsAccepted:           true, // Will be set based on event name in client
		TaskSubmissionTxHash: txHash,
		PerformerAddress:     performerAddress,
		AttesterIds:          attestersIds,
		ProofOfTask:          proofOfTask,
		Data:                 data,
	}, nil
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

func convertFilterQueryToEthGetLogsParams(fq ethereum.FilterQuery, fromBlock, toBlock uint64) nodeclient.EthGetLogsParams {
	fromHex := uint64ToHex(fromBlock)
	toHex := uint64ToHex(toBlock)
	fromBlockNum := nodeclient.BlockNumber(fromHex)
	toBlockNum := nodeclient.BlockNumber(toHex)

	params := nodeclient.EthGetLogsParams{
		FromBlock: &fromBlockNum,
		ToBlock:   &toBlockNum,
	}

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

func uint64ToHex(val uint64) string {
	return fmt.Sprintf("0x%x", val)
}

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
