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

// SendDataToEventScheduler sends data to the event scheduler
func (h *Handler) SendDataToEventScheduler(route string, data interface{}) (bool, error) {
	apiURL := fmt.Sprintf("%s%s", config.GetEventSchedulerRPCUrl(), route)
	return h.sendDataToScheduler(apiURL, data, "event scheduler")
}

// SendDataToConditionScheduler sends data to the condition scheduler
func (h *Handler) SendDataToConditionScheduler(route string, data interface{}) (bool, error) {
	apiURL := fmt.Sprintf("%s%s", config.GetConditionSchedulerRPCUrl(), route)
	return h.sendDataToScheduler(apiURL, data, "condition scheduler")
}

// sendDataToScheduler is a generic function to send data to any scheduler
func (h *Handler) sendDataToScheduler(apiURL string, data interface{}, schedulerName string) (bool, error) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return false, fmt.Errorf("error marshaling data: %v", err)
	}

	// Create a client with aggressive timeouts and connection pooling
	retryConfig := &retry.HTTPRetryConfig{
		MaxRetries:      3,
		InitialDelay:    200 * time.Millisecond,
		MaxDelay:        2 * time.Second,
		BackoffFactor:   2.0,
		LogRetryAttempt: true,
		RetryStatusCodes: []int{
			http.StatusInternalServerError,
			http.StatusBadGateway,
			http.StatusServiceUnavailable,
			http.StatusGatewayTimeout,
		},
		Timeout:             3 * time.Second,
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 100,
		IdleConnTimeout:     30 * time.Second,
	}

	client := retry.NewHTTPClient(retryConfig, h.logger)

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
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return false, fmt.Errorf("%s service error (status=%d): %s", schedulerName, resp.StatusCode, string(body))
	}

	h.logger.Infof("Successfully sent data to %s", schedulerName)
	return true, nil
}

// SendPauseToEventScheduler sends a DELETE request to the event scheduler
func (h *Handler) SendPauseToEventScheduler(route string) (bool, error) {
	apiURL := fmt.Sprintf("%s%s", config.GetEventSchedulerRPCUrl(), route)
	return h.sendPauseToScheduler(apiURL, "event scheduler")
}

// SendPauseToConditionScheduler sends a DELETE request to the condition scheduler
func (h *Handler) SendPauseToConditionScheduler(route string) (bool, error) {
	apiURL := fmt.Sprintf("%s%s", config.GetConditionSchedulerRPCUrl(), route)
	return h.sendPauseToScheduler(apiURL, "condition scheduler")
}

// sendPauseToScheduler sends a DELETE request to any scheduler
func (h *Handler) sendPauseToScheduler(apiURL string, schedulerName string) (bool, error) {
	// Create a client with aggressive timeouts and connection pooling
	retryConfig := &retry.HTTPRetryConfig{
		MaxRetries:      3,
		InitialDelay:    200 * time.Millisecond,
		MaxDelay:        2 * time.Second,
		BackoffFactor:   2.0,
		LogRetryAttempt: true,
		RetryStatusCodes: []int{
			http.StatusInternalServerError,
			http.StatusBadGateway,
			http.StatusServiceUnavailable,
			http.StatusGatewayTimeout,
		},
		Timeout:             3 * time.Second,
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 100,
		IdleConnTimeout:     30 * time.Second,
	}

	client := retry.NewHTTPClient(retryConfig, h.logger)

	req, err := http.NewRequest("POST", apiURL, nil)
	if err != nil {
		return false, fmt.Errorf("error creating request: %v", err)
	}
	req.Close = true

	resp, err := client.DoWithRetry(req)
	if err != nil {
		return false, fmt.Errorf("error sending DELETE to %s: %v", schedulerName, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return false, fmt.Errorf("%s service error (status=%d): %s", schedulerName, resp.StatusCode, string(body))
	}

	h.logger.Infof("Successfully sent DELETE to %s", schedulerName)
	return true, nil
}



// notifyEventScheduler sends a notification to the event scheduler
func (h *Handler) notifyEventScheduler(jobID int64, job types.EventJobData) {
	success, err := h.SendDataToEventScheduler("/api/v1/job/schedule", job)
	if err != nil {
		h.logger.Errorf("[NotifyEventScheduler] Failed to notify event scheduler for job %d: %v", jobID, err)
	} else if success {
		h.logger.Infof("[NotifyEventScheduler] Successfully notified event scheduler for job %d", jobID)
	}
}

// notifyConditionScheduler sends a notification to the condition scheduler
func (h *Handler) notifyConditionScheduler(jobID int64, job types.ConditionJobData) {
	success, err := h.SendDataToConditionScheduler("/api/v1/job/schedule", job)
	if err != nil {
		h.logger.Errorf("[NotifyConditionScheduler] Failed to notify condition scheduler for job %d: %v", jobID, err)
	} else if success {
		h.logger.Infof("[NotifyConditionScheduler] Successfully notified condition scheduler for job %d", jobID)
	}
}
