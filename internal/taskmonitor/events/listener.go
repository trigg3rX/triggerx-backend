package events

import (
	"context"
	"fmt"
	"math/big"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/trigg3rX/triggerx-backend/internal/taskmonitor/clients/database"
	"github.com/trigg3rX/triggerx-backend/internal/taskmonitor/clients/notify"
	"github.com/trigg3rX/triggerx-backend/internal/taskmonitor/config"
	"github.com/trigg3rX/triggerx-backend/internal/taskmonitor/tasks"
	"github.com/trigg3rX/triggerx-backend/pkg/ipfs"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"

	// Contract bindings
	contractAttestationCenter "github.com/trigg3rX/triggerx-contracts/bindings/contracts/AttestationCenter"
)

// ContractType represents the type of contract
// Kept as string to minimize coupling

type ContractType string

const (
	ContractTypeAttestationCenter ContractType = "attestation_center"
)

// ContractEventData represents parsed contract event data used downstream

type ContractEventData struct {
	EventType    string                 `json:"event_type"`
	ContractType ContractType           `json:"contract_type"`
	ParsedData   map[string]interface{} `json:"parsed_data"`
	RawData      []byte                 `json:"raw_data"`
	Topics       []string               `json:"topics"`
	BlockNumber  uint64                 `json:"block_number"`
	TxHash       string                 `json:"tx_hash"`
	LogIndex     uint                   `json:"log_index"`
}

// ChainEvent represents an event from any blockchain

type ChainEvent struct {
	ChainID      string       `json:"chain_id"`
	ChainName    string       `json:"chain_name"`
	ContractAddr string       `json:"contract_address"`
	ContractType ContractType `json:"contract_type"`
	EventName    string       `json:"event_name"`
	BlockNumber  uint64       `json:"block_number"`
	TxHash       string       `json:"tx_hash"`
	LogIndex     uint         `json:"log_index"`
	Data         interface{}  `json:"data"`
	RawLog       types.Log    `json:"raw_log"`
	ProcessedAt  time.Time    `json:"processed_at"`
}

// ContractEventListener handles listening to contract events across multiple chains
type ContractEventListener struct {
	logger            logging.Logger
	config            *ListenerConfig
	ctx               context.Context
	cancel            context.CancelFunc
	wg                sync.WaitGroup
	isRunning         bool
	mu                sync.RWMutex
	eventChan         chan *ChainEvent
	processingWg      sync.WaitGroup
	dbClient          *database.DatabaseClient
	ipfsClient        ipfs.IPFSClient
	taskStreamManager *tasks.TaskStreamManager
}

// ListenerConfig holds configuration for the event listener
type ListenerConfig struct {
	Chains            []ChainConfig                `json:"chains"`
	ReconnectConfig   ReconnectConfig              `json:"reconnect"`
	ProcessingWorkers int                          `json:"processing_workers"`
	EventBufferSize   int                          `json:"event_buffer_size"`
	ProcessingTimeout time.Duration                `json:"processing_timeout"`
	ContractAddresses map[string]map[string]string `json:"contract_addresses"` // chainID -> contractType -> address
}

// ChainConfig represents blockchain configuration for event listening
type ChainConfig struct {
	ChainID string `json:"chain_id"`
	Name    string `json:"name"`
	RPCURL  string `json:"rpc_url"`
	Enabled bool   `json:"enabled"`
}

// ReconnectConfig holds reconnection settings
type ReconnectConfig struct {
	MaxRetries    int           `json:"max_retries"`
	BaseDelay     time.Duration `json:"base_delay"`
	MaxDelay      time.Duration `json:"max_delay"`
	BackoffFactor float64       `json:"backoff_factor"`
}

// EventProcessor handles individual event processing
type EventProcessor struct {
	logger          logging.Logger
	operatorHandler *OperatorEventHandler
	taskHandler     *TaskEventHandler
}

// OperatorEventHandler handles operator-related events
type OperatorEventHandler struct {
	logger logging.Logger
}

// TaskEventHandler handles task-related events
type TaskEventHandler struct {
	logger            logging.Logger
	db                *database.DatabaseClient
	ipfsClient        ipfs.IPFSClient
	taskStreamManager *tasks.TaskStreamManager
	notifier          notify.Notifier
}

// NewContractEventListener creates a new contract event listener
func NewContractEventListener(logger logging.Logger, config *ListenerConfig, dbClient *database.DatabaseClient, ipfsClient ipfs.IPFSClient, taskStreamManager *tasks.TaskStreamManager) *ContractEventListener {
	ctx, cancel := context.WithCancel(context.Background())

	return &ContractEventListener{
		logger:            logger,
		config:            config,
		ctx:               ctx,
		cancel:            cancel,
		eventChan:         make(chan *ChainEvent, config.EventBufferSize),
		dbClient:          dbClient,
		ipfsClient:        ipfsClient,
		taskStreamManager: taskStreamManager,
	}
}

// Start begins listening for contract events
func (l *ContractEventListener) Start() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.isRunning {
		return fmt.Errorf("event listener is already running")
	}

	// Set up chain polling workers
	if err := l.setupChainConnections(); err != nil {
		return fmt.Errorf("failed to setup chain connections: %w", err)
	}

	// Start event processing workers
	l.startEventProcessors()

	l.isRunning = true
	l.logger.Info("Contract event listener (polling) started successfully")

	return nil
}

// Stop gracefully stops the event listener
func (l *ContractEventListener) Stop() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if !l.isRunning {
		return fmt.Errorf("event listener is not running")
	}

	l.logger.Info("Stopping contract event listener")

	// Cancel context to stop all goroutines
	l.cancel()

	// Wait for all goroutines to finish
	l.wg.Wait()
	l.processingWg.Wait()

	l.isRunning = false
	l.logger.Info("Contract event listener stopped")

	return nil
}

// setupChainConnections sets up blockchain connections and subscriptions
func (l *ContractEventListener) setupChainConnections() error {
	for _, chainConfig := range l.config.Chains {
		if !chainConfig.Enabled {
			continue
		}

		// Launch a poller per chain
		cc := chainConfig
		l.wg.Add(1)
		go func() {
			defer l.wg.Done()
			if err := l.startChainPoller(cc); err != nil {
				l.logger.Errorf("Chain poller error for %s: %v", cc.Name, err)
			}
		}()
	}

	return nil
}

// startEventProcessors starts worker goroutines for event processing
func (l *ContractEventListener) startEventProcessors() {
	processor := &EventProcessor{
		logger:          l.logger,
		operatorHandler: &OperatorEventHandler{logger: l.logger},
		taskHandler:     &TaskEventHandler{logger: l.logger, db: l.dbClient, ipfsClient: l.ipfsClient, taskStreamManager: l.taskStreamManager, notifier: notify.NewCompositeNotifier(l.logger, notify.NewWebhookNotifier(l.logger), notify.NewSMTPNotifier(l.logger))},
	}

	// Start multiple processing workers
	for i := 0; i < l.config.ProcessingWorkers; i++ {
		l.processingWg.Add(1)
		go l.eventProcessorWorker(processor, i)
	}
}

// eventProcessorWorker processes events from the event channel
func (l *ContractEventListener) eventProcessorWorker(processor *EventProcessor, workerID int) {
	defer l.processingWg.Done()

	l.logger.Debugf("Event processor worker %d started", workerID)

	for {
		select {
		case <-l.ctx.Done():
			return
		case event := <-l.eventChan:
			l.processEvent(processor, event)
		}
	}
}

// processEvent processes a single contract event
func (l *ContractEventListener) processEvent(processor *EventProcessor, event *ChainEvent) {
	switch event.ContractType {
	case ContractTypeAttestationCenter:
		l.logger.Debugf("Processing %s event from AttestationCenter contract on chain %s", event.EventName, event.ChainID)
		processor.taskHandler.ProcessTaskEvent(event)
	default:
		l.logger.Warnf("Unknown contract type: %s for event %s from contract %s on chain %s",
			event.ContractType, event.EventName, event.ContractAddr, event.ChainID)
	}
}

// GetStatus returns the current status of the event listener
func (l *ContractEventListener) GetStatus() map[string]interface{} {
	l.mu.RLock()
	defer l.mu.RUnlock()

	status := map[string]interface{}{
		"running":            l.isRunning,
		"processing_workers": l.config.ProcessingWorkers,
		"event_buffer_size":  l.config.EventBufferSize,
		"event_buffer_used":  len(l.eventChan),
	}

	return status
}

// pollSubscription describes one event filter to poll
// It intentionally avoids any websocket client types
// and only carries what's needed for RPC log filtering

type pollSubscription struct {
	ContractAddr common.Address
	ContractType ContractType
	EventName    string
	EventSig     common.Hash
	FilterQuery  ethereum.FilterQuery
}

// startChainPoller starts a polling loop for a given chain
func (l *ContractEventListener) startChainPoller(chainConfig ChainConfig) error {
	client, err := ethclient.Dial(chainConfig.RPCURL)
	if err != nil {
		return fmt.Errorf("failed to connect to %s RPC: %w", chainConfig.Name, err)
	}
	defer client.Close()

	// Build polling subscriptions from configured addresses
	subs := make([]pollSubscription, 0)
	chainAddresses := l.config.ContractAddresses[chainConfig.ChainID]
	if addr, ok := chainAddresses["attestation_center"]; ok {
		attABI, err := contractAttestationCenter.ContractAttestationCenterMetaData.GetAbi()
		if err != nil {
			return fmt.Errorf("failed to load AttestationCenter ABI: %w", err)
		}
		addrHex := common.HexToAddress(addr)
		for _, evName := range []string{"TaskSubmitted", "TaskRejected"} {
			ev, exists := attABI.Events[evName]
			if !exists {
				l.logger.Errorf("Event %s not in AttestationCenter ABI", evName)
				continue
			}
			fq := ethereum.FilterQuery{
				Addresses: []common.Address{addrHex},
				Topics:    [][]common.Hash{{ev.ID}},
			}
			subs = append(subs, pollSubscription{
				ContractAddr: addrHex,
				ContractType: ContractTypeAttestationCenter,
				EventName:    evName,
				EventSig:     ev.ID,
				FilterQuery:  fq,
			})
		}
	}

	if len(subs) == 0 {
		l.logger.Warnf("[%s] No subscriptions configured for polling (chainID=%s)", chainConfig.Name, chainConfig.ChainID)
	} else {
		for _, s := range subs {
			l.logger.Infof("[%s] Polling subscription added: %s.%s at %s", chainConfig.Name, s.ContractType, s.EventName, s.ContractAddr.Hex())
		}
	}

	// Initialize from the current block
	lastBlock, err := client.BlockNumber(l.ctx)
	if err != nil {
		return fmt.Errorf("failed to get current block number: %w", err)
	}
	l.logger.Infof("[%s] Starting poller at block %d", chainConfig.Name, lastBlock)

	// poll every 1 minute
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-l.ctx.Done():
			return nil
		case <-ticker.C:
			currentBlock, err := client.BlockNumber(l.ctx)
			if err != nil {
				l.logger.Errorf("[%s] Failed to get current block number: %v", chainConfig.Name, err)
				continue
			}
			if currentBlock <= lastBlock {
				l.logger.Debugf("[%s] No new blocks to process (last=%d, current=%d)", chainConfig.Name, lastBlock, currentBlock)
				continue
			}
			// l.logger.Infof("[%s] Polling block range [%d, %d] (span=%d)", chainConfig.Name, lastBlock+1, currentBlock, currentBlock-(lastBlock+1)+1)

			from := new(big.Int).SetUint64(lastBlock + 1)
			to := new(big.Int).SetUint64(currentBlock)

			for _, sub := range subs {
				start := from.Uint64()
				end := to.Uint64()
				const maxRange uint64 = 10
				for cur := start; cur <= end; {
					chunkEnd := cur + maxRange - 1
					if chunkEnd > end {
						chunkEnd = end
					}

					// l.logger.Debugf("[%s] Querying %s.%s logs in chunk [%d, %d]", chainConfig.Name, sub.ContractType, sub.EventName, cur, chunkEnd)

					fq := sub.FilterQuery
					fq.FromBlock = new(big.Int).SetUint64(cur)
					fq.ToBlock = new(big.Int).SetUint64(chunkEnd)

					logs, err := client.FilterLogs(l.ctx, fq)
					if err != nil {
						l.logger.Errorf("[%s] FilterLogs failed for %s.%s range [%#x, %#x]: %v", chainConfig.Name, sub.ContractType, sub.EventName, cur, chunkEnd, err)
						// proceed to next chunk to avoid blocking entire range
						cur = chunkEnd + 1
						continue
					}
					// if len(logs) == 0 {
					// 	l.logger.Debugf("[%s] No logs for %s.%s in chunk [%d, %d]", chainConfig.Name, sub.ContractType, sub.EventName, cur, chunkEnd)
					// } else {
					// l.logger.Infof("[%s] Fetched %d logs for %s.%s in chunk [%d, %d]", chainConfig.Name, len(logs), sub.ContractType, sub.EventName, cur, chunkEnd)
					// }
					for _, lg := range logs {
						if err := l.emitChainEventFromLog(chainConfig, sub, lg); err != nil {
							l.logger.Errorf("[%s] Failed to emit event for %s.%s: %v", chainConfig.Name, sub.ContractType, sub.EventName, err)
						}
					}

					cur = chunkEnd + 1
				}
			}

			lastBlock = currentBlock
			// l.logger.Debugf("[%s] Updated last processed block to %d", chainConfig.Name, lastBlock)
		}
	}
}

// emitChainEventFromLog parses a log and emits a ChainEvent into eventChan
func (l *ContractEventListener) emitChainEventFromLog(chainConfig ChainConfig, sub pollSubscription, lg types.Log) error {
	var parsed interface{}
	var err error

	switch sub.ContractType {
	case ContractTypeAttestationCenter:
		parsed, err = l.parseAttestationCenterEvent(sub.EventName, lg)
	default:
		return nil
	}
	if err != nil {
		return err
	}

	evt := &ChainEvent{
		ChainID:      chainConfig.ChainID,
		ChainName:    chainConfig.Name,
		ContractAddr: lg.Address.Hex(),
		ContractType: sub.ContractType,
		EventName:    sub.EventName,
		BlockNumber:  lg.BlockNumber,
		TxHash:       lg.TxHash.Hex(),
		LogIndex:     lg.Index,
		Data:         parsed,
		RawLog:       lg,
		ProcessedAt:  time.Now(),
	}

	l.logger.Debugf("[%s] Emitting event %s.%s block=%d tx=%s idx=%d", chainConfig.Name, sub.ContractType, sub.EventName, lg.BlockNumber, lg.TxHash.Hex(), lg.Index)

	select {
	case l.eventChan <- evt:
		return nil
	default:
		l.logger.Warnf("[%s] Event channel full, dropping event %s.%s at block %d", chainConfig.Name, sub.ContractType, sub.EventName, lg.BlockNumber)
		return fmt.Errorf("event channel full")
	}
}

// parseAttestationCenterEvent parses AttestationCenter events into ContractEventData
func (l *ContractEventListener) parseAttestationCenterEvent(eventName string, lg types.Log) (interface{}, error) {
	attABI, err := contractAttestationCenter.ContractAttestationCenterMetaData.GetAbi()
	if err != nil {
		return nil, fmt.Errorf("failed to load AttestationCenter ABI: %w", err)
	}
	ev, ok := attABI.Events[eventName]
	if !ok {
		return nil, fmt.Errorf("event %s not found in AttestationCenter ABI", eventName)
	}

	// Debug: Log event structure and raw data
	// l.logger.Debug("Event structure", "event_name", eventName, "input_count", len(ev.Inputs), "data_length", len(lg.Data))
	// for i, input := range ev.Inputs {
	// l.logger.Debug("Event input", "index", i, "name", input.Name, "type", input.Type.String(), "indexed", input.Indexed)
	// }
	// l.logger.Debug("Raw event data", "data_hex", fmt.Sprintf("%x", lg.Data))

	// Decode non-indexed fields
	nonIndexedArgs := make(abi.Arguments, 0)
	for _, input := range ev.Inputs {
		if !input.Indexed {
			nonIndexedArgs = append(nonIndexedArgs, input)
		}
	}
	nonIndexed := make(map[string]interface{})
	if len(nonIndexedArgs) > 0 {
		// Try to unpack all inputs first
		allUnpacked := make(map[string]interface{})
		if err := ev.Inputs.UnpackIntoMap(allUnpacked, lg.Data); err != nil {
			return nil, fmt.Errorf("failed to unpack event data: %w", err)
		}

		// Filter to only non-indexed fields
		for _, input := range ev.Inputs {
			if !input.Indexed {
				if value, exists := allUnpacked[input.Name]; exists {
					nonIndexed[input.Name] = value
				}
			}
		}

		// Debug: Log what was unpacked from non-indexed data
		// l.logger.Debug("Unpacked non-indexed data", "count", len(nonIndexed))
		// for k, v := range nonIndexed {
		// 	l.logger.Debug("Non-indexed field", "key", k, "type", fmt.Sprintf("%T", v), "value", v)
		// }

		// Debug: Log all unpacked data for comparison
		// l.logger.Debug("All unpacked data", "count", len(allUnpacked))
		// for k, v := range allUnpacked {
		// 	l.logger.Debug("All field", "key", k, "type", fmt.Sprintf("%T", v), "value", v)
		// }
	}

	// Parse indexed parameters from topics
	parsedData := make(map[string]interface{})
	// Start with non-indexed values
	for k, v := range nonIndexed {
		parsedData[k] = v
	}

	// Parse indexed parameters from topics (skip topics[0] which is the signature)
	topicIndex := 1
	for _, input := range ev.Inputs {
		if input.Indexed {
			if topicIndex < len(lg.Topics) {
				parsedData[input.Name] = l.parseTopicData(input, lg.Topics[topicIndex])
				topicIndex++
			}
		}
	}

	// Build ContractEventData compatible with downstream handler
	ced := &ContractEventData{
		EventType:    eventName,
		ContractType: ContractTypeAttestationCenter,
		ParsedData:   parsedData,
		RawData:      lg.Data,
		Topics:       topicsToHex(lg.Topics),
		BlockNumber:  lg.BlockNumber,
		TxHash:       lg.TxHash.Hex(),
		LogIndex:     lg.Index,
	}
	return ced, nil
}

func topicsToHex(topics []common.Hash) []string {
	out := make([]string, len(topics))
	for i, t := range topics {
		out[i] = t.Hex()
	}
	return out
}

// GetTestnetConfig returns a testnet configuration for the event listener
func GetTestnetConfig() *ListenerConfig {
	return &ListenerConfig{
		Chains: []ChainConfig{
			{
				ChainID: "84532",
				Name:    "Base Sepolia",
				RPCURL:  config.GetChainRPCUrl(true, "84532"),
				Enabled: true,
			},
		},
		ReconnectConfig: ReconnectConfig{
			MaxRetries:    10,
			BaseDelay:     1 * time.Second,
			MaxDelay:      5 * time.Minute,
			BackoffFactor: 2.0,
		},
		ProcessingWorkers: 4,
		EventBufferSize:   1000,
		ProcessingTimeout: 30 * time.Second,
		ContractAddresses: map[string]map[string]string{
			"84532": { // Base Sepolia
				"attestation_center": config.GetTestAttestationCenterAddress(),
			},
		},
	}
}

func GetMainnetConfig() *ListenerConfig {
	return &ListenerConfig{
		Chains: []ChainConfig{
			{
				ChainID: "8453",
				Name:    "Base",
				RPCURL:  config.GetChainRPCUrl(true, "8453"),
				Enabled: true,
			},
		},
		ReconnectConfig: ReconnectConfig{
			MaxRetries:    10,
			BaseDelay:     1 * time.Second,
			MaxDelay:      5 * time.Minute,
			BackoffFactor: 2.0,
		},
		ProcessingWorkers: 4,
		EventBufferSize:   1000,
		ProcessingTimeout: 30 * time.Second,
		ContractAddresses: map[string]map[string]string{
			"8453": { // Base Mainnet
				"attestation_center": config.GetAttestationCenterAddress(),
			},
		},
	}
}

// parseTopicData parses topic data based on the input type
func (l *ContractEventListener) parseTopicData(input abi.Argument, topic common.Hash) interface{} {
	switch input.Type.String() {
	case "address":
		return common.HexToAddress(topic.Hex()).Hex()
	case "uint256", "uint128", "uint64", "uint32", "uint16", "uint8":
		return new(big.Int).SetBytes(topic.Bytes()).String()
	case "int256", "int128", "int64", "int32", "int16", "int8":
		// For signed integers, we need to handle two's complement
		value := new(big.Int).SetBytes(topic.Bytes())
		if value.Bit(255) == 1 { // Check if the sign bit is set
			// Convert from two's complement
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
