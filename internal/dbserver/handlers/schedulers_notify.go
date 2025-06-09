package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/trigg3rX/triggerx-backend/internal/dbserver/config"
	"github.com/trigg3rX/triggerx-backend/internal/dbserver/types"
	"github.com/trigg3rX/triggerx-backend/pkg/retry"
)

// notifyEventScheduler sends a notification to the event scheduler
func (h *Handler) notifyEventScheduler(jobID int64, job types.EventJobData) (bool, error) {
	success, err := h.sendDataToScheduler("/job/schedule", job, "event scheduler")
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

// notifyConditionScheduler sends a notification to the condition scheduler
func (h *Handler) notifyConditionScheduler(jobID int64, job types.ConditionJobData) (bool, error) {
	success, err := h.sendDataToScheduler("/job/schedule", job, "condition scheduler")
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
func (h *Handler) notifyPauseToEventScheduler(jobID int64) (bool, error) {
	success, err := h.sendDataToScheduler("/job/pause", types.EventJobData{JobID: jobID}, "event scheduler")
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

// SendPauseToConditionScheduler sends a DELETE request to the condition scheduler
func (h *Handler) notifyPauseToConditionScheduler(jobID int64) (bool, error) {
	success, err := h.sendDataToScheduler("/job/pause", types.ConditionJobData{JobID: jobID}, "condition scheduler")
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

// sendDataToScheduler is a generic function to send data to any scheduler
func (h *Handler) sendDataToScheduler(route string, data interface{}, schedulerName string) (bool, error) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return false, fmt.Errorf("error marshaling data: %v", err)
	}

	var apiURL string
	if schedulerName == "event scheduler" {
		apiURL = fmt.Sprintf("%s%s", config.GetEventSchedulerRPCUrl(), route)
	} else if schedulerName == "condition scheduler" {
		apiURL = fmt.Sprintf("%s%s", config.GetConditionSchedulerRPCUrl(), route)
	} else {
		return false, fmt.Errorf("invalid scheduler name: %s", schedulerName)
	}

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
		return false, fmt.Errorf("error sending data to %s: %v", schedulerName, err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			h.logger.Errorf("Error closing response body: %v", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return false, fmt.Errorf("%s service error (status=%d): %s", schedulerName, resp.StatusCode, string(body))
	}

	h.logger.Infof("Successfully sent data to %s", schedulerName)
	return true, nil
}
