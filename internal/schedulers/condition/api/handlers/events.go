package handlers

import (
	"math/big"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	eventmonitorTypes "github.com/trigg3rX/triggerx-backend/internal/eventmonitor/types"
	"github.com/trigg3rX/triggerx-backend/internal/schedulers/condition/scheduler"
	"github.com/trigg3rX/triggerx-backend/internal/schedulers/condition/scheduler/worker"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
	pkgTypes "github.com/trigg3rX/triggerx-backend/pkg/types"
)

// HandleEventNotification handles event notifications from Event Monitor Service
func HandleEventNotification(logger logging.Logger, scheduler *scheduler.ConditionBasedScheduler) gin.HandlerFunc {
	return func(c *gin.Context) {
		var notification eventmonitorTypes.EventNotification
		if err := c.ShouldBindJSON(&notification); err != nil {
			logger.Warn("Invalid event notification request", "error", err)
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   err.Error(),
			})
			return
		}

		logger.Info("Received event notification from Event Monitor Service",
			"request_id", notification.RequestID,
			"chain_id", notification.ChainID,
			"contract_address", notification.ContractAddr,
			"event_signature", notification.EventSig,
			"tx_hash", notification.TxHash,
			"block_number", notification.BlockNumber)

		// Convert request ID to BigInt
		jobIDBigInt := new(big.Int)
		jobIDBigInt, ok := jobIDBigInt.SetString(notification.RequestID, 10)
		if !ok {
			logger.Error("Invalid job ID in notification", "request_id", notification.RequestID)
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
			logger.Warn("Job data not found, may have been cleaned up",
				"job_id", jobID,
				"request_id", notification.RequestID)
			c.JSON(http.StatusOK, gin.H{
				"success": true,
				"message": "Job not found, ignoring notification",
			})
			return
		}

		// Check if job has expired
		now := time.Now()
		if jobData.EventWorkerData.ExpirationTime.Before(now) {
			logger.Info("Job has expired, unregistering from Event Monitor Service",
				"job_id", jobID,
				"expiration_time", jobData.EventWorkerData.ExpirationTime)

			// Unregister from Event Monitor Service
			if err := scheduler.UnregisterEventJob(jobIDBigInt); err != nil {
				logger.Error("Failed to unregister expired job from Event Monitor Service",
					"job_id", jobID,
					"error", err)
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

		// Process the trigger notification (same as current processEvent logic)
		if err := scheduler.HandleTriggerNotification(triggerNotification); err != nil {
			logger.Error("Failed to process event notification",
				"job_id", jobID,
				"tx_hash", notification.TxHash,
				"error", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error":   err.Error(),
			})
			return
		}

		// For non-recurring jobs, unregister from Event Monitor Service after processing
		if !jobData.EventWorkerData.Recurring {
			logger.Info("Non-recurring job triggered, unregistering from Event Monitor Service",
				"job_id", jobID)

			// Unregister from Event Monitor Service
			if err := scheduler.UnregisterEventJob(jobIDBigInt); err != nil {
				logger.Error("Failed to unregister non-recurring job from Event Monitor Service",
					"job_id", jobID,
					"error", err)
				// Don't fail the request, just log the error
			}
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "Event notification processed successfully",
		})
	}
}
