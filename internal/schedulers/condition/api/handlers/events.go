package handlers

import (
	"math/big"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"

	eventmonitorTypes "github.com/trigg3rX/triggerx-backend/internal/eventmonitor/types"
	"github.com/trigg3rX/triggerx-backend/internal/schedulers/condition/scheduler"
	"github.com/trigg3rX/triggerx-backend/internal/schedulers/condition/scheduler/worker"
	"github.com/trigg3rX/triggerx-backend/pkg/observability"
	pkgTypes "github.com/trigg3rX/triggerx-backend/pkg/types"
)

// HandleEventNotification handles event notifications from Event Monitor Service
func HandleEventNotification(logger observability.Logger, scheduler *scheduler.ConditionBasedScheduler) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract trace context from HTTP headers
		ctx := c.Request.Context()
		propagator := otel.GetTextMapPropagator()
		ctx = propagator.Extract(ctx, propagation.HeaderCarrier(c.Request.Header))

		var notification eventmonitorTypes.EventNotification
		if err := c.ShouldBindJSON(&notification); err != nil {
			logger.Warn(ctx, "Invalid event notification request", observability.Error(err))
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   err.Error(),
			})
			return
		}

		logger.Info(ctx, "Received event notification from Event Monitor Service",
			observability.String("request_id", notification.RequestID),
			observability.String("chain_id", notification.ChainID),
			observability.String("contract_address", notification.ContractAddr),
			observability.String("event_signature", notification.EventSig),
			observability.String("tx_hash", notification.TxHash),
			observability.Uint64("block_number", notification.BlockNumber))

		// Convert request ID to BigInt
		jobIDBigInt := new(big.Int)
		jobIDBigInt, ok := jobIDBigInt.SetString(notification.RequestID, 10)
		if !ok {
			logger.Error(ctx, "Invalid job ID in notification", observability.String("request_id", notification.RequestID))
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   "invalid job ID",
			})
			return
		}
		jobID := pkgTypes.NewBigInt(jobIDBigInt)

		// Check if job has expired
		jobData, err := scheduler.GetJobData(jobIDBigInt)
		if err != nil {
			logger.Warn(ctx, "Job data not found, may have been cleaned up",
				observability.String("job_id", jobID.String()),
				observability.String("request_id", notification.RequestID))
			c.JSON(http.StatusOK, gin.H{
				"success": true,
				"message": "Job not found, ignoring notification",
			})
			return
		}

		// Check if job has expired
		now := time.Now()
		if jobData.EventWorkerData.ExpirationTime.Before(now) {
			logger.Info(ctx, "Job has expired, unregistering from Event Monitor Service",
				observability.String("job_id", jobID.String()),
				observability.Time("expiration_time", jobData.EventWorkerData.ExpirationTime))

			// Unregister from Event Monitor Service
			if err := scheduler.UnregisterEventJob(ctx, jobIDBigInt); err != nil {
				logger.Error(ctx, "Failed to unregister expired job from Event Monitor Service",
					observability.String("job_id", jobID.String()),
					observability.Error(err))
			}

			c.JSON(http.StatusOK, gin.H{
				"success": true,
				"message": "Job expired, unregistered",
			})
			return
		}

		// Create trigger notification
		triggerNotification := &worker.TriggerNotification{
			JobID:         jobIDBigInt,
			TriggerTxHash: notification.TxHash,
			TriggeredAt:   notification.Timestamp,
		}

		// Process the trigger notification with trace context
		if err := scheduler.HandleTriggerNotification(ctx, triggerNotification); err != nil {
			logger.Error(ctx, "Failed to process event notification",
				observability.String("job_id", jobID.String()),
				observability.String("tx_hash", notification.TxHash),
				observability.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error":   err.Error(),
			})
			return
		}

		// For non-recurring jobs, unregister from Event Monitor Service after processing
		if !jobData.EventWorkerData.Recurring {
			logger.Info(ctx, "Non-recurring job triggered, unregistering from Event Monitor Service",
				observability.String("job_id", jobID.String()))

			// Unregister from Event Monitor Service
			if err := scheduler.UnregisterEventJob(ctx, jobIDBigInt); err != nil {
				logger.Error(ctx, "Failed to unregister non-recurring job from Event Monitor Service",
					observability.String("job_id", jobID.String()),
					observability.Error(err))
				// Don't fail the request, just log the error
			}
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "Event notification processed successfully",
		})
	}
}
