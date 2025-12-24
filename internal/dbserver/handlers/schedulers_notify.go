package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"net/http"

	"github.com/trigg3rX/triggerx-backend/internal/dbserver/config"
	httppkg "github.com/trigg3rX/triggerx-backend/pkg/http"
	commonTypes "github.com/trigg3rX/triggerx-backend/pkg/types"
	"github.com/trigg3rX/triggerx-backend/pkg/observability"
)

// notifyConditionScheduler sends a notification to the condition scheduler
func (h *Handler) notifyConditionScheduler(ctx context.Context, jobID *big.Int, scheduleConditionJobData commonTypes.ScheduleConditionJobData) (bool, error) {
	success, err := h.sendDataToScheduler(ctx, "/api/v1/job/schedule", scheduleConditionJobData)
	if err != nil {
		h.logger.Error(ctx, "[NotifyConditionScheduler] Failed to notify condition scheduler for job", observability.Int64("job_id", jobID.Int64()), observability.Error(err))
		return false, err
	}
	if !success {
		h.logger.Error(ctx, "[NotifyConditionScheduler] Failed to notify condition scheduler for job", observability.Int64("job_id", jobID.Int64()))
		return false, fmt.Errorf("failed to notify condition scheduler for job %d", jobID)
	}
	h.logger.Info(ctx, "Successfully sent data to condition scheduler", observability.String("job_id", jobID.String()))
	return true, nil
}

// SendPauseToEventScheduler sends a DELETE request to the event scheduler
func (h *Handler) notifyPauseToConditionScheduler(ctx context.Context, jobID *big.Int) (bool, error) {
	success, err := h.sendDataToScheduler(ctx, "/api/v1/job/pause", commonTypes.ScheduleConditionJobData{JobID: commonTypes.NewBigInt(jobID)})
	if err != nil {
		h.logger.Error(ctx, "[NotifyEventScheduler] Failed to notify event scheduler for job", observability.Int64("job_id", jobID.Int64()), observability.Error(err))
		return false, err
	}
	if !success {
		h.logger.Error(ctx, "[NotifyEventScheduler] Failed to notify event scheduler for job", observability.Int64("job_id", jobID.Int64()))
		return false, fmt.Errorf("failed to notify event scheduler for job %d", jobID)
	}
	h.logger.Info(ctx, "Successfully sent data to event scheduler", observability.String("job_id", jobID.String()))
	return true, nil
}

// sendDataToScheduler is a generic function to send data to any scheduler
func (h *Handler) sendDataToScheduler(ctx context.Context, route string, data commonTypes.ScheduleConditionJobData) (bool, error) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return false, fmt.Errorf("error marshaling data: %v", err)
	}

	apiURL := fmt.Sprintf("%s%s", config.GetConditionSchedulerRPCUrl(), route)

	client, err := httppkg.NewHTTPClient(httppkg.DefaultHTTPRetryConfig())
	if err != nil {
		return false, fmt.Errorf("error creating HTTP client: %v", err)
	}

	req, err := http.NewRequestWithContext(context.Background(), "POST", apiURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return false, fmt.Errorf("error creating request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Close = true

	resp, err := client.DoWithRetry(context.Background(), req)
	if err != nil {
		return false, fmt.Errorf("error sending data to condition scheduler: %v", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			h.logger.Error(ctx, "Error closing response body", observability.Error(err))
		}
	}()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return false, fmt.Errorf("condition scheduler service error (status=%d): %s", resp.StatusCode, string(body))
	}

	return true, nil
}
