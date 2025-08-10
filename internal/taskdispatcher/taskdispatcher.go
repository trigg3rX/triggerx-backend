package taskdispatcher

import (
	"context"
	"fmt"
	"time"

	"github.com/trigg3rX/triggerx-backend/internal/taskdispatcher/tasks"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
	"github.com/trigg3rX/triggerx-backend/pkg/cryptography"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

// TaskDispatcher encapsulates dependencies for handling scheduler submissions
// and forwarding them to the aggregator.
type TaskDispatcher struct {
	logger           logging.Logger
	taskStreamManager *tasks.TaskStreamManager
	healthClient *HealthClient
	signingKey string
	signingAddress string
}

// NewTaskDispatcher constructs a new dispatcher with an initialized aggregator client.
func NewTaskDispatcher(
	logger logging.Logger,
	taskStreamManager *tasks.TaskStreamManager,
	healthClient *HealthClient,
	signingKey string,
	signingAddress string) (*TaskDispatcher, error) {
	
	return &TaskDispatcher{
		logger:           logger,
		taskStreamManager: taskStreamManager,
		healthClient: healthClient,
		signingKey: signingKey,
		signingAddress: signingAddress,
	}, nil
}

// SubmitTaskFromScheduler is the core business method used by the RPC handler.
// It receives the scheduler request, optionally enqueues to Redis (skipped for now),
// and forwards the data to the aggregator using the same approach as send_task.go.
func (d *TaskDispatcher) SubmitTaskFromScheduler(ctx context.Context, req *types.SchedulerTaskRequest) (*types.TaskManagerAPIResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("nil request")
	}
	if len(req.SendTaskDataToKeeper.TaskID) == 0 {
		return nil, fmt.Errorf("missing task id")
	}
	if len(req.SendTaskDataToKeeper.TriggerData) == 0 {
		return nil, fmt.Errorf("missing trigger data")
	}
	if len(req.SendTaskDataToKeeper.TargetData) == 0 {
		return nil, fmt.Errorf("missing target data")
	}

	taskCount := len(req.SendTaskDataToKeeper.TaskID)
	d.logger.Info("Receiving task from scheduler",
		"task_ids", req.SendTaskDataToKeeper.TaskID,
		"task_count", taskCount,
		"scheduler_id", req.SendTaskDataToKeeper.SchedulerID,
		"source", req.Source)

	// Use dynamic performer selection instead of hardcoded selection
	performer, err := d.healthClient.GetPerformerData(req.SendTaskDataToKeeper.TargetData[0].IsImua)
	if err != nil {
		d.logger.Error("Failed to get performer data dynamically",
			"task_id", req.SendTaskDataToKeeper.TaskID[0],
			"error", err)
		return nil, fmt.Errorf("failed to get performer: %w", err)
	}
	// Update task with performer information
	req.SendTaskDataToKeeper.PerformerData = performer

	// Sign the task data with improved error handling
	signature, err := cryptography.SignJSONMessage(req.SendTaskDataToKeeper, d.signingKey)
	if err != nil {
		d.logger.Error("Failed to sign batch task data",
			"task_id", req.SendTaskDataToKeeper.TaskID[0],
			"error", err)
		return nil, fmt.Errorf("failed to sign task data: %w", err)
	}
	req.SendTaskDataToKeeper.ManagerSignature = signature

	// Handle batch requests by creating individual task stream data for each task
	if taskCount > 1 {
		// This is a batch request (likely from time scheduler)
		d.logger.Info("Processing batch request", "task_count", taskCount)

		for i := 0; i < taskCount; i++ {
			// Create individual task data for each task in the batch
			individualTaskData := types.SendTaskDataToKeeper{
				TaskID:           []int64{req.SendTaskDataToKeeper.TaskID[i]},
				PerformerData:    req.SendTaskDataToKeeper.PerformerData,
				TargetData:       []types.TaskTargetData{req.SendTaskDataToKeeper.TargetData[i]},
				TriggerData:      []types.TaskTriggerData{req.SendTaskDataToKeeper.TriggerData[i]},
				SchedulerID:      req.SendTaskDataToKeeper.SchedulerID,
				ManagerSignature: req.SendTaskDataToKeeper.ManagerSignature,
			}

			taskStreamData := tasks.TaskStreamData{
				JobID:                individualTaskData.TargetData[0].JobID,
				TaskDefinitionID:     individualTaskData.TargetData[0].TaskDefinitionID,
				CreatedAt:            time.Now(),
				RetryCount:           0,
				SendTaskDataToKeeper: individualTaskData,
			}

			// Add individual task to batch processor
			_, err := d.taskStreamManager.AddTaskToDispatchedStream(ctx, taskStreamData)
			if err != nil {
				d.logger.Error("Failed to add individual task to batch processor",
					"task_id", individualTaskData.TaskID[0],
					"batch_index", i,
					"source", req.Source,
					"error", err)
				// Continue processing other tasks in the batch
				continue
			}

			d.logger.Debug("Individual task added to batch processor",
				"task_id", individualTaskData.TaskID[0],
				"batch_index", i)
		}
	} else {
		// This is a single task request (likely from condition scheduler)
		taskStreamData := tasks.TaskStreamData{
			JobID:                req.SendTaskDataToKeeper.TargetData[0].JobID,
			TaskDefinitionID:     req.SendTaskDataToKeeper.TargetData[0].TaskDefinitionID,
			CreatedAt:            time.Now(),
			RetryCount:           0,
			SendTaskDataToKeeper: req.SendTaskDataToKeeper,
		}

		// Add task to batch processor for improved performance
		_, err := d.taskStreamManager.AddTaskToDispatchedStream(ctx, taskStreamData)
		if err != nil {
			d.logger.Error("Failed to add task to batch processor",
				"task_id", req.SendTaskDataToKeeper.TaskID[0],
				"source", req.Source,
				"error", err)
			return nil, fmt.Errorf("failed to add task to batch processor: %w", err)
		}
	}

	d.logger.Info("[Dispatcher] Task forwarded to aggregator", "task_id", req.SendTaskDataToKeeper.TaskID[0])
	return &types.TaskManagerAPIResponse{
		Success:   true,
		TaskID:    []int64{req.SendTaskDataToKeeper.TaskID[0]},
		Message:   "Task submitted successfully",
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}, nil
}

func (d *TaskDispatcher) Close() error {
	return d.taskStreamManager.Close()
}
