package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"net/http"

	"github.com/trigg3rX/triggerx-backend/internal/dbserver/config"
	httppkg "github.com/trigg3rX/triggerx-backend/pkg/http"
	commonTypes "github.com/trigg3rX/triggerx-backend/pkg/types"
)

// notifyConditionScheduler sends a notification to the condition scheduler
func (h *Handler) notifyConditionScheduler(jobID *big.Int, scheduleConditionJobData commonTypes.ScheduleConditionJobData) (bool, error) {
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
func (h *Handler) notifyPauseToConditionScheduler(jobID *big.Int) (bool, error) {
	success, err := h.sendDataToScheduler("/api/v1/job/pause", commonTypes.ScheduleConditionJobData{JobID: commonTypes.NewBigInt(jobID)})
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
func (h *Handler) sendDataToScheduler(route string, data commonTypes.ScheduleConditionJobData) (bool, error) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return false, fmt.Errorf("error marshaling data: %v", err)
	}

	apiURL := fmt.Sprintf("%s%s", config.GetConditionSchedulerRPCUrl(), route)

	client, err := httppkg.NewHTTPClient(httppkg.DefaultHTTPRetryConfig(), h.logger)
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
