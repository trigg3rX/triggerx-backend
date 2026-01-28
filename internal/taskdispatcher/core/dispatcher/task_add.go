package dispatcher

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"

	"github.com/trigg3rX/triggerx-backend/internal/taskdispatcher/config"
	"github.com/trigg3rX/triggerx-backend/internal/taskdispatcher/metrics"
	"github.com/trigg3rX/triggerx-backend/pkg/observability"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

// AddTaskToDispatchedStream adds a task to the dispatched stream after sending to performer
func (tsm *TaskStreamManager) AddTaskToDispatchedStream(ctx context.Context, task types.TaskStreamData) (bool, error) {
	// Prepare payload identical to previous implementation
	jsonData, err := json.Marshal(task.SendTaskDataToKeeper)
	if err != nil {
		tsm.logger.Error(ctx, "Failed to marshal scheduler task data", observability.Int64("task_id", task.SendTaskDataToKeeper.TaskID), observability.Error(err))
		return false, fmt.Errorf("failed to marshal task data: %w", err)
	}

	// Send directly to performer's /p2p/message endpoint
	// The format is: POST with JSON body { "data": "0x<hex-encoded-json>" }
	hexEncodedData := "0x" + hex.EncodeToString(jsonData)
	requestBody := map[string]string{
		"data": hexEncodedData,
	}
	requestBodyJSON, err := json.Marshal(requestBody)
	if err != nil {
		tsm.logger.Error(ctx, "Failed to marshal request body for performer", observability.Int64("task_id", task.SendTaskDataToKeeper.TaskID), observability.Error(err))
		return false, fmt.Errorf("failed to marshal request body: %w", err)
	}

	// Select performer URL based on mainnet/testnet
	var performerURL string
	switch task.Network {
	case types.NetworkMainnet:
		performerURL = config.GetPerformerAPIUrl() + "/p2p/message"
	case types.NetworkSepolia:
		performerURL = config.GetTestPerformerAPIUrl() + "/p2p/message"
	case types.NetworkImua:
		performerURL = config.GetTestPerformerAPIUrl() + "/p2p/message"
	default:
		return false, fmt.Errorf("invalid network: %s", task.Network)
	}

	tsm.logger.Info(ctx, "Sending task directly to performer",
		observability.Int64("task_id", task.SendTaskDataToKeeper.TaskID),
		observability.String("performer_url", performerURL),
		observability.String("network", string(task.Network)))

	// Create HTTP request with timeout
	httpCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(httpCtx, http.MethodPost, performerURL, bytes.NewBuffer(requestBodyJSON))
	if err != nil {
		tsm.logger.Error(ctx, "Failed to create HTTP request for performer", observability.Int64("task_id", task.SendTaskDataToKeeper.TaskID), observability.Error(err))
		return false, fmt.Errorf("failed to create HTTP request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	// Inject trace context into HTTP headers
	propagator := otel.GetTextMapPropagator()
	propagator.Inject(httpCtx, propagation.HeaderCarrier(req.Header))

	// Send request to performer and track dispatch metrics
	dispatchStart := time.Now()
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	dispatchDuration := time.Since(dispatchStart)

	if err != nil {
		metrics.TrackKeeperDispatch(ctx, false, task.Network, dispatchDuration)
		tsm.logger.Error(ctx, "Failed to send task to performer", observability.Int64("task_id", task.SendTaskDataToKeeper.TaskID), observability.Error(err))
		return false, fmt.Errorf("failed to send task to performer: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			tsm.logger.Warn(ctx, "Failed to close response body", observability.Error(err))
		}
	}()

	// Accept both 200 OK (legacy) and 202 Accepted (async acknowledgement)
	// 202 means the performer accepted the task and will process it asynchronously
	// Task completion will be reported to TaskMonitor, not back to this caller
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
		metrics.TrackKeeperDispatch(ctx, false, task.Network, dispatchDuration)
		tsm.logger.Error(ctx, "Performer returned error status", observability.Int64("task_id", task.SendTaskDataToKeeper.TaskID), observability.Int("status", resp.StatusCode))
		return false, fmt.Errorf("performer returned status %d", resp.StatusCode)
	}

	// Track successful dispatch
	metrics.TrackKeeperDispatch(ctx, true, task.Network, dispatchDuration)

	tsm.logger.Info(ctx, "Task accepted by performer",
		observability.Int64("task_id", task.SendTaskDataToKeeper.TaskID),
		observability.String("performer_url", performerURL),
		observability.Int("status_code", resp.StatusCode),
		observability.Duration("dispatch_duration", dispatchDuration))

	success, err := tsm.addTaskToStream(ctx, types.StreamTaskDispatched, &task)
	if err != nil {
		return false, err
	}

	// Update keeper data with task execution
	err = tsm.taskRepo.UpdateTaskStatusToDispatched(ctx, task.SendTaskDataToKeeper.TaskID)
	if err != nil {
		tsm.logger.Error(ctx, "Failed to update task status to dispatched", observability.Error(err))
		return false, fmt.Errorf("failed to update task status to dispatched: %w", err)
	}

	if !success {
		return false, fmt.Errorf("failed to add task to stream")
	}
	return true, nil
}
