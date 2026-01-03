package handlers

import (
	"context"
	"fmt"
	"math/big"

	"github.com/trigg3rX/triggerx-backend/pkg/observability"
	commonTypes "github.com/trigg3rX/triggerx-backend/pkg/types"
)

// notifyConditionScheduler sends a notification to the condition scheduler via gRPC
func (h *Handler) notifyConditionScheduler(ctx context.Context, jobID *big.Int, scheduleConditionJobData commonTypes.ScheduleConditionJobData) (bool, error) {
	if h.conditionSchedulerClient == nil {
		return false, fmt.Errorf("condition scheduler gRPC client not initialized")
	}

	if err := h.conditionSchedulerClient.ScheduleJob(ctx, &scheduleConditionJobData); err != nil {
		h.logger.Error(ctx, "[NotifyConditionScheduler] Failed to notify condition scheduler for job", observability.Int64("job_id", jobID.Int64()), observability.Error(err))
		return false, err
	}

	h.logger.Info(ctx, "Successfully sent data to condition scheduler via gRPC", observability.String("job_id", jobID.String()))
	return true, nil
}

// notifyPauseToConditionScheduler sends an unschedule notification to the condition scheduler via gRPC
func (h *Handler) notifyPauseToConditionScheduler(ctx context.Context, jobID *big.Int) (bool, error) {
	if h.conditionSchedulerClient == nil {
		return false, fmt.Errorf("condition scheduler gRPC client not initialized")
	}

	jobIDBigInt := commonTypes.NewBigInt(jobID)
	if err := h.conditionSchedulerClient.UnscheduleJob(ctx, jobIDBigInt); err != nil {
		h.logger.Error(ctx, "[NotifyConditionScheduler] Failed to unschedule job in condition scheduler", observability.Int64("job_id", jobID.Int64()), observability.Error(err))
		return false, err
	}

	h.logger.Info(ctx, "Successfully unscheduled job in condition scheduler via gRPC", observability.String("job_id", jobID.String()))
	return true, nil
}
