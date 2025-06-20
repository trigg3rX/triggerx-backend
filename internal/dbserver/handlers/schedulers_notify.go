package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/trigg3rX/triggerx-backend/internal/dbserver/config"
	"github.com/trigg3rX/triggerx-backend/pkg/retry"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

// notifyConditionScheduler sends a notification to the condition scheduler
func (h *Handler) notifyConditionScheduler(jobID int64, scheduleConditionJobData types.ScheduleConditionJobData) (bool, error) {
	success, err := h.sendDataToScheduler("/api/v1/job/schedule", scheduleConditionJobData)
	if err != nil {
		h.logger.Errorf("[NotifyConditionScheduler] Failed to notify condition scheduler for job %d: %v", jobID, err)
		return false, err
	}
	if !success {
		h.logger.Errorf("[NotifyConditionScheduler] Failed to notify condition scheduler for job %d", jobID)
		return false, fmt.Errorf("failed to notify condition scheduler for job %d", jobID)
	}
	return true, nil
}

// SendPauseToEventScheduler sends a DELETE request to the event scheduler
func (h *Handler) notifyPauseToConditionScheduler(jobID int64) (bool, error) {
	success, err := h.sendDataToScheduler("/api/v1/job/pause", types.ScheduleConditionJobData{JobID: jobID})
	if err != nil {
		h.logger.Errorf("[NotifyEventScheduler] Failed to notify event scheduler for job %d: %v", jobID, err)
		return false, err
	}
	if !success {
		h.logger.Errorf("[NotifyEventScheduler] Failed to notify event scheduler for job %d", jobID)
		return false, fmt.Errorf("failed to notify event scheduler for job %d", jobID)
	}

	return true, nil
}

// sendDataToScheduler is a generic function to send data to any scheduler
func (h *Handler) sendDataToScheduler(route string, data types.ScheduleConditionJobData) (bool, error) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return false, fmt.Errorf("error marshaling data: %v", err)
	}

	apiURL := fmt.Sprintf("%s%s", config.GetConditionSchedulerRPCUrl(), route)

	// Create a client with aggressive timeouts and connection pooling
	httpConfig := &retry.HTTPRetryConfig{
		RetryConfig: &retry.RetryConfig{
			MaxRetries:      3,
			InitialDelay:    1 * time.Second,
			MaxDelay:        10 * time.Second,
			BackoffFactor:   2.0,
			JitterFactor:    0.5,
			LogRetryAttempt: true,
			StatusCodes: []int{
				http.StatusInternalServerError,
				http.StatusBadGateway,
				http.StatusServiceUnavailable,
				http.StatusGatewayTimeout,
			},
			ShouldRetry: func(err error) bool {
				return err != nil
			},
		},
		Timeout:         3 * time.Second,
		IdleConnTimeout: 30 * time.Second,
	}
	client, err := retry.NewHTTPClient(httpConfig, h.logger)
	if err != nil {
		return false, fmt.Errorf("error creating HTTP client: %v", err)
	}

	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return false, fmt.Errorf("error creating request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Close = true

	resp, err := client.DoWithRetry(req)
	if err != nil {
		return false, fmt.Errorf("error sending data to condition scheduler: %v", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			h.logger.Errorf("Error closing response body: %v", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return false, fmt.Errorf("condition scheduler service error (status=%d): %s", resp.StatusCode, string(body))
	}

	h.logger.Infof("Successfully sent data to condition scheduler")
	return true, nil
}
