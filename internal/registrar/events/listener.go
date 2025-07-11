package events

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/trigg3rX/triggerx-backend-imua/internal/registrar/config"
	"github.com/trigg3rX/triggerx-backend-imua/internal/registrar/events/websocket"
	"github.com/trigg3rX/triggerx-backend-imua/pkg/logging"
)

// ContractEventListener handles listening to contract events across multiple chains
type ContractEventListener struct {
	logger       logging.Logger
	client       *websocket.Client
	config       *ListenerConfig
	ctx          context.Context
	cancel       context.CancelFunc
	wg           sync.WaitGroup
	isRunning    bool
	mu           sync.RWMutex
	eventChan    chan *websocket.ChainEvent
	processingWg sync.WaitGroup
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
	ChainID      string `json:"chain_id"`
	Name         string `json:"name"`
	RPCURL       string `json:"rpc_url"`
	WebSocketURL string `json:"websocket_url"`
	Enabled      bool   `json:"enabled"`
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
	logger logging.Logger
}

// NewContractEventListener creates a new contract event listener
func NewContractEventListener(logger logging.Logger, config *ListenerConfig) *ContractEventListener {
	ctx, cancel := context.WithCancel(context.Background())

	client := websocket.NewClient(logger)

	return &ContractEventListener{
		logger:    logger,
		client:    client,
		config:    config,
		ctx:       ctx,
		cancel:    cancel,
		eventChan: make(chan *websocket.ChainEvent, config.EventBufferSize),
	}
}

// Start begins listening for contract events
func (l *ContractEventListener) Start() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.isRunning {
		return fmt.Errorf("event listener is already running")
	}

	l.logger.Info("Starting contract event listener")

	// Set up chain connections and subscriptions
	if err := l.setupChainConnections(); err != nil {
		return fmt.Errorf("failed to setup chain connections: %w", err)
	}

	// Start the websocket client
	if err := l.client.Start(); err != nil {
		return fmt.Errorf("failed to start websocket client: %w", err)
	}

	// Start event processing workers
	l.startEventProcessors()

	// Start the main event listening loop
	l.wg.Add(1)
	go l.eventListeningLoop()

	l.isRunning = true
	l.logger.Info("Contract event listener started successfully")

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

	// Stop the websocket client
	if err := l.client.Stop(); err != nil {
		l.logger.Errorf("Error stopping websocket client: %v", err)
	}

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
			l.logger.Infof("Skipping disabled chain: %s", chainConfig.Name)
			continue
		}

		// Add chain to websocket client
		wsConfig := websocket.ChainConfig{
			ChainID:      chainConfig.ChainID,
			Name:         chainConfig.Name,
			RPCURL:       chainConfig.RPCURL,
			WebSocketURL: chainConfig.WebSocketURL,
			Contracts:    l.getContractConfigsForChain(chainConfig.ChainID),
		}

		if err := l.client.AddChain(wsConfig); err != nil {
			return fmt.Errorf("failed to add chain %s: %w", chainConfig.Name, err)
		}

		// Set up specific contract subscriptions
		if err := l.setupContractSubscriptions(chainConfig.ChainID); err != nil {
			return fmt.Errorf("failed to setup subscriptions for chain %s: %w", chainConfig.Name, err)
		}

		l.logger.Infof("Successfully configured chain: %s (%s)", chainConfig.Name, chainConfig.ChainID)
	}

	return nil
}

// getContractConfigsForChain returns contract configurations for a specific chain
func (l *ContractEventListener) getContractConfigsForChain(chainID string) []websocket.ContractConfig {
	var configs []websocket.ContractConfig

	chainAddresses, exists := l.config.ContractAddresses[chainID]
	if !exists {
		l.logger.Warnf("No contract addresses configured for chain %s", chainID)
		return configs
	}

	// AvsGovernance contract
	if addr, exists := chainAddresses["avs_governance"]; exists {
		configs = append(configs, websocket.ContractConfig{
			Address:      addr,
			ContractType: websocket.ContractTypeAvsGovernance,
			Events:       []string{"OperatorRegistered", "OperatorUnregistered"},
		})
	}

	// AttestationCenter contract
	if addr, exists := chainAddresses["attestation_center"]; exists {
		configs = append(configs, websocket.ContractConfig{
			Address:      addr,
			ContractType: websocket.ContractTypeAttestationCenter,
			Events:       []string{"TaskSubmitted", "TaskRejected"},
		})
	}

	return configs
}

// setupContractSubscriptions sets up specific contract event subscriptions
func (l *ContractEventListener) setupContractSubscriptions(chainID string) error {
	chainAddresses, exists := l.config.ContractAddresses[chainID]
	if !exists {
		return fmt.Errorf("no contract addresses configured for chain %s", chainID)
	}

	// Subscribe to AvsGovernance events
	if addr, exists := chainAddresses["avs_governance"]; exists {
		if err := l.client.SubscribeToContract(
			chainID,
			addr,
			websocket.ContractTypeAvsGovernance,
			[]string{"OperatorRegistered", "OperatorUnregistered"},
		); err != nil {
			return fmt.Errorf("failed to subscribe to AvsGovernance events: %w", err)
		}
		l.logger.Infof("Subscribed to AvsGovernance events on chain %s", chainID)
	}

	// Subscribe to AttestationCenter events
	if addr, exists := chainAddresses["attestation_center"]; exists {
		if err := l.client.SubscribeToContract(
			chainID,
			addr,
			websocket.ContractTypeAttestationCenter,
			[]string{"TaskSubmitted", "TaskRejected"},
		); err != nil {
			return fmt.Errorf("failed to subscribe to AttestationCenter events: %w", err)
		}
		l.logger.Infof("Subscribed to AttestationCenter events on chain %s", chainID)
	}

	return nil
}

// startEventProcessors starts worker goroutines for event processing
func (l *ContractEventListener) startEventProcessors() {
	processor := &EventProcessor{
		logger:          l.logger,
		operatorHandler: &OperatorEventHandler{logger: l.logger},
		taskHandler:     &TaskEventHandler{logger: l.logger},
	}

	// Start multiple processing workers
	for i := 0; i < l.config.ProcessingWorkers; i++ {
		l.processingWg.Add(1)
		go l.eventProcessorWorker(processor, i)
	}

	l.logger.Infof("Started %d event processing workers", l.config.ProcessingWorkers)
}

// eventListeningLoop is the main event listening loop
func (l *ContractEventListener) eventListeningLoop() {
	defer l.wg.Done()

	l.logger.Info("Starting event listening loop")

	for {
		select {
		case <-l.ctx.Done():
			l.logger.Info("Event listening loop stopped")
			return
		case event := <-l.client.EventChannel():
			// Forward event to processing workers
			select {
			case l.eventChan <- event:
				// Event queued successfully
			default:
				// Event channel is full, log warning
				l.logger.Warnf("Event channel full, dropping event: %s from %s",
					event.EventName, event.ContractAddr)
			}
		}
	}
}

// eventProcessorWorker processes events from the event channel
func (l *ContractEventListener) eventProcessorWorker(processor *EventProcessor, workerID int) {
	defer l.processingWg.Done()

	l.logger.Infof("Event processor worker %d started", workerID)

	for {
		select {
		case <-l.ctx.Done():
			l.logger.Infof("Event processor worker %d stopped", workerID)
			return
		case event := <-l.eventChan:
			l.processEvent(processor, event, workerID)
		}
	}
}

// processEvent processes a single contract event
func (l *ContractEventListener) processEvent(processor *EventProcessor, event *websocket.ChainEvent, workerID int) {
	// Set processing timeout
	ctx, cancel := context.WithTimeout(l.ctx, l.config.ProcessingTimeout)
	defer cancel()

	l.logger.Debugf("Worker %d processing %s event from %s",
		workerID, event.EventName, event.ContractType)

	// Process event based on contract type
	switch event.ContractType {
	case websocket.ContractTypeAvsGovernance:
		processor.operatorHandler.ProcessOperatorEvent(ctx, event)
	case websocket.ContractTypeAttestationCenter:
		processor.taskHandler.ProcessTaskEvent(ctx, event)
	default:
		l.logger.Warnf("Unknown contract type: %s", event.ContractType)
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
		"chains":             make(map[string]interface{}),
	}

	// Get chain-specific status
	chainStatus := l.client.GetChainStatus()
	status["chains"] = chainStatus

	return status
}

// GetDefaultConfig returns a default configuration for the event listener
func GetDefaultConfig() *ListenerConfig {
	return &ListenerConfig{
		Chains: []ChainConfig{
			{
				ChainID:      "17000",
				Name:         "Ethereum Holesky",
				RPCURL:       config.GetChainRPCUrl(false, "17000"),
				WebSocketURL: config.GetChainRPCUrl(true, "17000"),
				Enabled:      true,
			},
			{
				ChainID:      "84532",
				Name:         "Base Sepolia",
				RPCURL:       config.GetChainRPCUrl(false, "84532"),
				WebSocketURL: config.GetChainRPCUrl(true, "84532"),
				Enabled:      true,
			},
			{
				ChainID:      "11155420",
				Name:         "Optimism Sepolia",
				RPCURL:       config.GetChainRPCUrl(false, "11155420"),
				WebSocketURL: config.GetChainRPCUrl(true, "11155420"),
				Enabled:      true,
			},
		},
		ReconnectConfig: ReconnectConfig{
			MaxRetries:    10,
			BaseDelay:     5 * time.Second,
			MaxDelay:      5 * time.Minute,
			BackoffFactor: 2.0,
		},
		ProcessingWorkers: 4,
		EventBufferSize:   1000,
		ProcessingTimeout: 30 * time.Second,
		ContractAddresses: map[string]map[string]string{
			"17000": { // Ethereum Holesky
				"avs_governance":     config.GetAvsGovernanceAddress(),
				"avs_governance_logic": config.GetAvsGovernanceLogicAddress(),
			},
			"84532": { // Base Sepolia
				"attestation_center": config.GetAttestationCenterAddress(),
				"obls": config.GetOBLSAddress(),
				"trigger_gas_registry": config.GetTriggerGasRegistryAddress(),
			},
			"11155420": { // Optimism Sepolia
				"trigger_gas_registry": config.GetTriggerGasRegistryAddress(),
			},
		},
	}
}
