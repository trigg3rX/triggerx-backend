package core

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"

	"github.com/trigg3rX/triggerx-backend/internal/aggregatorlogwatcher/config"
	"github.com/trigg3rX/triggerx-backend/internal/aggregatorlogwatcher/rpc/clients/eventmonitor"
	"github.com/trigg3rX/triggerx-backend/pkg/observability"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

// LogWatcher watches aggregator logs for task submission events
type LogWatcher struct {
	logDir          string
	eventMonitorClient *eventmonitor.Client
	logger          observability.Logger
	tracer          observability.Tracer
	ctx             context.Context
	cancel          context.CancelFunc
	wg              sync.WaitGroup
	processedTxs    map[string]bool // Track processed transactions to avoid duplicates
	mu              sync.RWMutex
}

// NewLogWatcher creates a new log watcher
func NewLogWatcher(ctx context.Context, logger observability.Logger, tracer observability.Tracer) (*LogWatcher, error) {
	ctx, cancel := context.WithCancel(ctx)

	// Create eventmonitor client
	eventMonitorURL := config.GetEventMonitorRPCUrl()
	if eventMonitorURL == "" {
		cancel()
		return nil, fmt.Errorf("event monitor RPC URL is not configured")
	}

	client, err := eventmonitor.NewClient(eventMonitorURL, logger, tracer)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to create event monitor client: %w", err)
	}

	logDir := config.GetAggregatorLogDir()
	if logDir == "" {
		logDir = "othentic/data/logs/aggregator"
	}

	return &LogWatcher{
		logDir:            logDir,
		eventMonitorClient: client,
		logger:            logger,
		tracer:            tracer,
		ctx:               ctx,
		cancel:            cancel,
		processedTxs:     make(map[string]bool),
	}, nil
}

// Start starts watching logs
func (w *LogWatcher) Start() error {
	w.logger.Info(w.ctx, "Starting aggregator log watcher",
		observability.String("log_dir", w.logDir))

	// Start watching for new log entries
	w.wg.Add(1)
	go func() {
		defer w.wg.Done()
		w.watchLogs()
	}()

	return nil
}

// Stop stops watching logs
func (w *LogWatcher) Stop() {
	w.logger.Info(w.ctx, "Stopping aggregator log watcher")
	w.cancel()
	w.wg.Wait()
	
	if w.eventMonitorClient != nil {
		_ = w.eventMonitorClient.Close(w.ctx)
	}
	
	w.logger.Info(w.ctx, "Aggregator log watcher stopped")
}

// watchLogs continuously watches log files for new entries
func (w *LogWatcher) watchLogs() {
	ticker := time.NewTicker(2 * time.Second) // Check every 2 seconds
	defer ticker.Stop()

	// Track last file positions
	filePositions := make(map[string]int64)

	for {
		select {
		case <-w.ctx.Done():
			return
		case <-ticker.C:
			w.scanLogFiles(filePositions)
		}
	}
}

// scanLogFiles scans log files for new entries
func (w *LogWatcher) scanLogFiles(filePositions map[string]int64) {
	err := filepath.Walk(w.logDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Only process .log files
		if !info.IsDir() && strings.HasSuffix(path, ".log") {
			w.processLogFile(path, filePositions)
		}

		return nil
	})

	if err != nil {
		w.logger.Error(w.ctx, "Error scanning log directory",
			observability.Error(err),
			observability.String("log_dir", w.logDir))
	}
}

// processLogFile processes a log file for new entries
func (w *LogWatcher) processLogFile(filePath string, filePositions map[string]int64) {
	file, err := os.Open(filePath)
	if err != nil {
		return
	}
	defer file.Close()

	// Get current file size
	stat, err := file.Stat()
	if err != nil {
		return
	}

	// Get last known position
	lastPos, exists := filePositions[filePath]
	if !exists {
		// First time seeing this file, start from end
		filePositions[filePath] = stat.Size()
		return
	}

	// If file hasn't grown, nothing to do
	if stat.Size() <= lastPos {
		return
	}

	// Seek to last known position
	if _, err := file.Seek(lastPos, 0); err != nil {
		return
	}

	// Read new lines
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		// Parse log entry
		var entry LogEntry
		if err := json.Unmarshal([]byte(line), &entry); err != nil {
			continue
		}

		// Check for task submission messages
		w.processLogEntry(&entry)
	}

	// Update position
	filePositions[filePath] = stat.Size()
}

// LogEntry represents a log entry from aggregator
type LogEntry struct {
	Timestamp string `json:"timestamp"`
	Version   string `json:"version"`
	Level     string `json:"level"`
	Message   string `json:"message"`
}

// processLogEntry processes a log entry looking for task submission events
func (w *LogWatcher) processLogEntry(entry *LogEntry) {
	// Pattern 1: "has been rejected and submitted"
	// Pattern 2: "has been approved and submitted"
	// Both contain: "on chain <chainID>, tx: <txHash>"

	rejectedPattern := regexp.MustCompile(`has been rejected and submitted.*on chain (\d+).*tx: (0x[a-fA-F0-9]+)`)
	approvedPattern := regexp.MustCompile(`has been approved and submitted.*on chain (\d+).*tx: (0x[a-fA-F0-9]+)`)

	var chainID, txHash string
	var found bool

	if matches := rejectedPattern.FindStringSubmatch(entry.Message); len(matches) == 3 {
		chainID = matches[1]
		txHash = matches[2]
		found = true
	} else if matches := approvedPattern.FindStringSubmatch(entry.Message); len(matches) == 3 {
		chainID = matches[1]
		txHash = matches[2]
		found = true
	}

	if !found {
		return
	}

	// Check if we've already processed this transaction
	w.mu.RLock()
	if w.processedTxs[txHash] {
		w.mu.RUnlock()
		return
	}
	w.mu.RUnlock()

	// Mark as processed
	w.mu.Lock()
	w.processedTxs[txHash] = true
	w.mu.Unlock()

	// Process transaction
	w.wg.Add(1)
	go func() {
		defer w.wg.Done()
		w.processTransaction(txHash, chainID)
	}()
}

// processTransaction sends transaction to eventmonitor for processing
func (w *LogWatcher) processTransaction(txHash, chainID string) {
	ctx, span := w.tracer.Start(w.ctx, "aggregator.log.transaction.processed",
		observability.WithAttributes(
			attribute.String("tx.hash", txHash),
			attribute.String("chain.id", chainID),
		),
	)
	defer span.End()

	w.logger.Info(ctx, "Processing transaction from aggregator log",
		observability.String("tx_hash", txHash),
		observability.String("chain_id", chainID))

	req := &types.ProcessTransactionRequest{
		TxHash:  txHash,
		ChainID: chainID,
	}

	resp, err := w.eventMonitorClient.ProcessTransaction(ctx, req)
	if err != nil {
		span.RecordError(err)
		w.logger.Error(ctx, "Failed to process transaction",
			observability.Error(err),
			observability.String("tx_hash", txHash),
			observability.String("chain_id", chainID))
		return
	}

	w.logger.Info(ctx, "Transaction processed successfully",
		observability.String("tx_hash", txHash),
		observability.String("chain_id", chainID),
		observability.String("event_name", resp.EventName))

	span.SetStatus(codes.Ok, "transaction processed successfully")
}
