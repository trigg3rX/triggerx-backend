package tasks

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/redis/go-redis/v9"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"

	"github.com/trigg3rX/triggerx-backend/internal/taskdispatcher/config"
	"github.com/trigg3rX/triggerx-backend/internal/taskdispatcher/metrics"
	"github.com/trigg3rX/triggerx-backend/pkg/observability"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

func (tsm *TaskStreamManager) AddTaskToDispatchedStream(ctx context.Context, task types.TaskStreamData) (bool, error) {
	// Prepare payload identical to previous implementation
	jsonData, err := json.Marshal(task.SendTaskDataToKeeper)
	if err != nil {
		tsm.logger.Error(ctx, "Failed to marshal scheduler task data", observability.Int64("task_id", task.SendTaskDataToKeeper.TaskID[0]), observability.Error(err))
		return false, fmt.Errorf("failed to marshal task data: %w", err)
	}

	// COMMENTED OUT: Aggregator send logic - now sending directly to performer
	// broadcast := types.BroadcastDataForPerformer{
	// 	TaskID:           task.SendTaskDataToKeeper.TaskID[0],
	// 	TaskDefinitionID: task.SendTaskDataToKeeper.TargetData[0].TaskDefinitionID,
	// 	PerformerAddress: task.SendTaskDataToKeeper.PerformerData.KeeperAddress,
	// 	Data:             []byte(jsonData),
	// }

	// var success bool
	// if task.IsMainnet {
	// 	success, err = tsm.aggregatorClient.SendTaskToPerformer(ctx, &broadcast)
	// } else {
	// 	success, err = tsm.testAggregatorClient.SendTaskToPerformer(ctx, &broadcast)
	// }
	// if err != nil {
	// 	tsm.logger.Error("Failed to send task to aggregator", "task_id", task.SendTaskDataToKeeper.TaskID[0], "error", err)
	// 	return false, err
	// }
	// if !success {
	// 	tsm.logger.Warn("Aggregator send returned unsuccessful", "task_id", task.SendTaskDataToKeeper.TaskID[0])
	// 	return false, fmt.Errorf("aggregator send unsuccessful")
	// }

	// Send directly to performer's /p2p/message endpoint
	// The format is: POST with JSON body { "data": "0x<hex-encoded-json>" }
	hexEncodedData := "0x" + hex.EncodeToString(jsonData)
	requestBody := map[string]string{
		"data": hexEncodedData,
	}
	requestBodyJSON, err := json.Marshal(requestBody)
	if err != nil {
		tsm.logger.Error(ctx, "Failed to marshal request body for performer", observability.Int64("task_id", task.SendTaskDataToKeeper.TaskID[0]), observability.Error(err))
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
		observability.Int64("task_id", task.SendTaskDataToKeeper.TaskID[0]),
		observability.String("performer_url", performerURL),
		observability.String("network", string(task.Network)))

	// Create HTTP request with timeout
	httpCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(httpCtx, http.MethodPost, performerURL, bytes.NewBuffer(requestBodyJSON))
	if err != nil {
		tsm.logger.Error(ctx, "Failed to create HTTP request for performer", observability.Int64("task_id", task.SendTaskDataToKeeper.TaskID[0]), observability.Error(err))
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
		tsm.logger.Error(ctx, "Failed to send task to performer", observability.Int64("task_id", task.SendTaskDataToKeeper.TaskID[0]), observability.Error(err))
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
		tsm.logger.Error(ctx, "Performer returned error status", observability.Int64("task_id", task.SendTaskDataToKeeper.TaskID[0]), observability.Int("status", resp.StatusCode))
		return false, fmt.Errorf("performer returned status %d", resp.StatusCode)
	}

	// Track successful dispatch
	metrics.TrackKeeperDispatch(ctx, true, task.Network, dispatchDuration)

	tsm.logger.Info(ctx, "Task accepted by performer",
		observability.Int64("task_id", task.SendTaskDataToKeeper.TaskID[0]),
		observability.String("performer_url", performerURL),
		observability.Int("status_code", resp.StatusCode),
		observability.Duration("dispatch_duration", dispatchDuration))

	success, err := tsm.addTaskToStream(ctx, types.StreamTaskDispatched, &task)
	if err != nil {
		return false, err
	}

	// Update keeper data with task execution
	err = tsm.databaseClient.UpdateTaskStatusToDispatched(ctx, task.SendTaskDataToKeeper.TaskID[0])
	if err != nil {
		tsm.logger.Error(ctx, "Failed to update task status to dispatched", observability.Error(err))
		return false, fmt.Errorf("failed to update task status to dispatched: %w", err)
	}

	if !success {
		return false, fmt.Errorf("failed to add task to stream")
	}
	return true, nil
}

func (tsm *TaskStreamManager) addTaskToStream(ctx context.Context, stream string, task *types.TaskStreamData) (bool, error) {
	start := time.Now()
	ctx, cancel := context.WithTimeout(ctx, config.GetRequestTimeout())
	defer cancel()

	taskJSON, err := json.Marshal(task)
	if err != nil {
		tsm.logger.Error(ctx, "Failed to marshal task data",
			observability.Int64("task_id", task.SendTaskDataToKeeper.TaskID[0]),
			observability.String("stream", stream),
			observability.Error(err))
		return false, fmt.Errorf("failed to marshal task data: %w", err)
	}

	res, err := tsm.client.XAdd(ctx, &redis.XAddArgs{
		Stream: stream,
		MaxLen: int64(10000),
		Approx: true,
		Values: map[string]interface{}{
			"task":       taskJSON,
			"created_at": time.Now().Unix(),
		},
	})
	duration := time.Since(start)

	if err != nil {
		if metrics.TasksAddedToStreamTotal != nil {
			metrics.TasksAddedToStreamTotal.WithLabelValues(stream, "failure").Inc(ctx)
		}
		tsm.logger.Error(ctx, "Failed to add task to stream",
			observability.Int64("task_id", task.SendTaskDataToKeeper.TaskID[0]),
			observability.String("stream", stream),
			observability.Duration("duration", duration),
			observability.Error(err))
		return false, fmt.Errorf("failed to add task to stream: %w", err)
	}

	// Store the task index mapping for efficient lookup
	if stream == types.StreamTaskDispatched {
		taskID := task.SendTaskDataToKeeper.TaskID[0]
		err = tsm.storeTaskIndex(ctx, taskID, res)
		if err != nil {
			tsm.logger.Warn(ctx, "Failed to store task index, but task was added to stream",
				observability.Int64("task_id", taskID),
				observability.String("message_id", res),
				observability.Error(err))
			// Don't fail the entire operation if index storage fails
		}

		// Add task to timeout tracking
		err = tsm.addTaskToTimeoutTracking(ctx, taskID)
		if err != nil {
			tsm.logger.Warn(ctx, "Failed to add task to timeout tracking, but task was added to stream",
				observability.Int64("task_id", taskID),
				observability.Error(err))
			// Don't fail the entire operation if timeout tracking fails
		}
	}

	// Track expiration for this stream entry with stream-specific TTL
	var entryTTL time.Duration
	switch stream {
	case types.StreamTaskDispatched:
		entryTTL = types.TasksProcessingTTL
	case types.StreamTaskCompleted:
		entryTTL = types.TasksCompletedTTL
	case types.StreamTaskFailed:
		entryTTL = types.TasksFailedTTL
	case types.StreamTaskRetry:
		entryTTL = types.TasksRetryTTL
	default:
		entryTTL = types.TasksProcessingTTL // Default fallback
	}

	// Track expiration using sorted set (similar to timeout tracking)
	expirationKey := fmt.Sprintf("stream:expiration:%s", stream)
	expirationTimestamp := float64(time.Now().Add(entryTTL).Unix())
	member := fmt.Sprintf("%s:%s", stream, res)
	_, err = tsm.client.ZAdd(ctx, expirationKey, redis.Z{
		Score:  expirationTimestamp,
		Member: member,
	})
	if err != nil {
		tsm.logger.Warn(ctx, "Failed to add stream entry expiration, but task was added to stream",
			observability.Int64("task_id", task.SendTaskDataToKeeper.TaskID[0]),
			observability.String("stream", stream),
			observability.String("message_id", res),
			observability.Error(err))
		// Don't fail the entire operation if expiration tracking fails
	} else {
		// Set TTL on the sorted set (should be longer than max entry TTL)
		_ = tsm.client.SetTTL(ctx, expirationKey, 48*time.Hour)
	}

	if metrics.TasksAddedToStreamTotal != nil {
		metrics.TasksAddedToStreamTotal.WithLabelValues(stream, "success").Inc(ctx)
	}
	tsm.logger.Debug(ctx, "Task added to stream successfully",
		observability.Int64("task_id", task.SendTaskDataToKeeper.TaskID[0]),
		observability.String("stream", stream),
		observability.String("stream_id", res),
		observability.Duration("duration", duration),
		observability.Int("task_json_size", len(taskJSON)),
		observability.Duration("entry_ttl", entryTTL))

	return true, nil
}
