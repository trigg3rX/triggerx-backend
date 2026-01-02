package repository

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/trigg3rX/triggerx-backend/internal/dbserver/events"
	"github.com/trigg3rX/triggerx-backend/internal/dbserver/repository/queries"
	"github.com/trigg3rX/triggerx-backend/internal/dbserver/types"
	"github.com/trigg3rX/triggerx-backend/pkg/database"
	commonTypes "github.com/trigg3rX/triggerx-backend/pkg/types"
)

type TaskRepository interface {
	CreateTaskDataInDB(ctx context.Context, task *types.CreateTaskDataRequest) (int64, error)
	AddTaskPerformerID(taskID int64, performerID int64) error
	UpdateTaskExecutionDataInDB(ctx context.Context, task *types.UpdateTaskExecutionDataRequest) error
	UpdateTaskAttestationDataInDB(ctx context.Context, task *types.UpdateTaskAttestationDataRequest) error
	UpdateTaskNumberAndStatus(ctx context.Context, taskID int64, taskNumber int64, status string, txHash string) error
	GetTaskDataByID(taskID int64) (commonTypes.TaskData, error)
	GetTasksByJobID(jobID *big.Int) ([]types.GetTasksByJobID, error)
	AddTaskIDToJob(jobID *big.Int, taskID int64) error
	UpdateTaskFee(ctx context.Context, taskID int64, fee float64) error
	GetTaskFee(taskID int64) (float64, error)
	GetCreatedChainIDByJobID(jobID *big.Int) (string, error)
	GetRecentTasks(limit int) ([]types.RecentTaskResponse, error)
}

type taskRepository struct {
	db        *database.Connection
	publisher *events.Publisher
}

func NewTaskRepository(db *database.Connection) TaskRepository {
	return &taskRepository{
		db: db,
	}
}

// NewTaskRepositoryWithPublisher creates a new task repository with WebSocket publisher
func NewTaskRepositoryWithPublisher(db *database.Connection, publisher *events.Publisher) TaskRepository {
	return &taskRepository{
		db:        db,
		publisher: publisher,
	}
}

func (r *taskRepository) CreateTaskDataInDB(ctx context.Context, task *types.CreateTaskDataRequest) (int64, error) {
	var maxTaskID int64
	err := r.db.Session().Query(queries.GetMaxTaskIDQuery).Scan(&maxTaskID)
	if err != nil {
		return -1, fmt.Errorf("error getting max task ID: %w", err)
	}

	taskID := maxTaskID + 1
	err = r.db.Session().Query(queries.CreateTaskDataQuery, taskID, task.JobID, task.TaskDefinitionID, time.Now(), task.IsImua).Exec()
	if err != nil {
		return -1, fmt.Errorf("error creating task data: %w", err)
	}

	// Emit WebSocket event for task creation
	if r.publisher != nil {
		// Extract user ID from job data if available
		userID := r.getUserIDFromJobID(task.JobID)
		r.publisher.PublishTaskCreated(ctx, taskID, task.JobID.String(), int64(task.TaskDefinitionID), task.IsImua, userID)
	}

	return taskID, nil
}

func (r *taskRepository) AddTaskPerformerID(taskID int64, performerID int64) error {
	err := r.db.Session().Query(queries.AddTaskPerformerIDQuery, taskID, performerID).Exec()
	if err != nil {
		return fmt.Errorf("error adding task performer ID: %w", err)
	}
	return nil
}

func (r *taskRepository) UpdateTaskExecutionDataInDB(ctx context.Context, task *types.UpdateTaskExecutionDataRequest) error {
	err := r.db.Session().Query(queries.UpdateTaskExecutionDataQuery, task.TaskPerformerID, task.ExecutionTimestamp, task.ExecutionTxHash, task.ProofOfTask, task.TaskOpXCost, task.TaskID).Exec()
	if err != nil {
		return fmt.Errorf("error updating task execution data: %w", err)
	}

	// Emit WebSocket event for task update
	if r.publisher != nil {
		jobID := r.getJobIDFromTaskID(task.TaskID)
		userID := r.getUserIDFromJobID(jobID)

		updateEvent := &events.TaskUpdatedEvent{
			TaskPerformerID:    &task.TaskPerformerID,
			ExecutionTimestamp: &task.ExecutionTimestamp,
			ExecutionTxHash:    &task.ExecutionTxHash,
			ProofOfTask:        &task.ProofOfTask,
			TaskOpXCost:        &task.TaskOpXCost,
		}
		r.publisher.PublishTaskUpdated(ctx, task.TaskID, jobID.String(), userID, updateEvent)
	}

	return nil
}

func (r *taskRepository) UpdateTaskAttestationDataInDB(ctx context.Context, task *types.UpdateTaskAttestationDataRequest) error {
	err := r.db.Session().Query(queries.UpdateTaskAttestationDataQuery, task.TaskNumber, task.TaskAttesterIDs, task.TpSignature, task.TaSignature, task.TaskSubmissionTxHash, task.IsSuccessful, task.TaskID).Exec()
	if err != nil {
		return fmt.Errorf("error updating task attestation data: %w", err)
	}

	// Emit WebSocket event for task attestation update
	if r.publisher != nil {
		jobID := r.getJobIDFromTaskID(task.TaskID)
		userID := r.getUserIDFromJobID(jobID)

		// Convert types for WebSocket event
		taskAttesterIDsStr := ""
		if len(task.TaskAttesterIDs) > 0 {
			// Convert []int64 to string representation
			taskAttesterIDsStr = fmt.Sprintf("%v", task.TaskAttesterIDs)
		}
		tpSignatureStr := ""
		if len(task.TpSignature) > 0 {
			tpSignatureStr = string(task.TpSignature)
		}
		taSignatureStr := ""
		if len(task.TaSignature) > 0 {
			taSignatureStr = string(task.TaSignature)
		}

		updateEvent := &events.TaskUpdatedEvent{
			TaskNumber:           &task.TaskNumber,
			TaskAttesterIDs:      &taskAttesterIDsStr,
			TpSignature:          &tpSignatureStr,
			TaSignature:          &taSignatureStr,
			TaskSubmissionTxHash: &task.TaskSubmissionTxHash,
			IsSuccessful:         &task.IsSuccessful,
		}
		r.publisher.PublishTaskUpdated(ctx, task.TaskID, jobID.String(), userID, updateEvent)
	}

	return nil
}

func (r *taskRepository) UpdateTaskNumberAndStatus(ctx context.Context, taskID int64, taskNumber int64, status string, txHash string) error {
	// Get old status for comparison
	oldStatus := r.getTaskStatus(taskID)

	err := r.db.Session().Query(queries.UpdateTaskNumberAndStatusQuery, taskNumber, status, txHash, taskID).Exec()
	if err != nil {
		return fmt.Errorf("error updating task number and status: %w", err)
	}

	// Emit WebSocket event for task status change
	if r.publisher != nil {
		jobID := r.getJobIDFromTaskID(taskID)
		userID := r.getUserIDFromJobID(jobID)

		r.publisher.PublishTaskStatusChanged(ctx, taskID, jobID.String(), oldStatus, status, userID, &taskNumber, &txHash)
	}

	return nil
}

func (r *taskRepository) GetTaskDataByID(taskID int64) (commonTypes.TaskData, error) {
	var task commonTypes.TaskData
	var jobIDBigInt *big.Int
	err := r.db.Session().Query(queries.GetTaskDataByIDQuery, taskID).Scan(&task.TaskID, &task.TaskNumber, &jobIDBigInt, &task.TaskDefinitionID, &task.CreatedAt, &task.TaskOpxCost, &task.ExecutionTimestamp, &task.ExecutionTxHash, &task.TaskPerformerID, &task.ProofOfTask, &task.ConvertedArguments, &task.TaskAttesterIDs, &task.TpSignature, &task.TaSignature, &task.TaskSubmissionTxHash, &task.IsAccepted, &task.TaskStatus, &task.TaskError, &task.IsImua)
	if err != nil {
		return commonTypes.TaskData{}, fmt.Errorf("error getting task data by ID: %w", err)
	}
	task.JobID = commonTypes.NewBigInt(jobIDBigInt)
	return task, nil
}

func (r *taskRepository) GetTasksByJobID(jobID *big.Int) ([]types.GetTasksByJobID, error) {
	iter := r.db.Session().Query(queries.GetTasksByJobIDQuery, jobID).Iter()
	var tasks []types.GetTasksByJobID
	var task types.GetTasksByJobID

	for iter.Scan(
		&task.TaskID,
		&task.TaskNumber,
		&task.TaskOpXCost,
		&task.ExecutionTimestamp,
		&task.ExecutionTxHash,
		&task.TaskPerformerID,
		&task.TaskAttesterIDs,
		&task.IsAccepted,
		&task.TaskStatus,
		&task.TaskError,
		&task.ConvertedArguments,
	) {
		tasks = append(tasks, task)
	}

	if err := iter.Close(); err != nil {
		return []types.GetTasksByJobID{}, errors.New("error getting tasks by job ID: " + err.Error())
	}

	return tasks, nil
}

func (r *taskRepository) AddTaskIDToJob(jobID *big.Int, taskID int64) error {
	var existingTaskIDs []int64
	// First, get existing task IDs
	err := r.db.Session().Query(queries.GetTaskIDsByJobIDQuery, jobID).Scan(&existingTaskIDs)
	if err != nil {
		// If no existing tasks, start with empty slice
		existingTaskIDs = []int64{}
	}

	// Append the new task ID
	existingTaskIDs = append(existingTaskIDs, taskID)

	// Update the job with the new task IDs list
	err = r.db.Session().Query(queries.AddTaskIDToJobQuery, existingTaskIDs, jobID).Exec()
	if err != nil {
		return fmt.Errorf("error adding task ID to job: %w", err)
	}
	return nil
}

func (r *taskRepository) UpdateTaskFee(ctx context.Context, taskID int64, fee float64) error {
	// Get old fee for comparison
	oldFee, _ := r.GetTaskFee(taskID)

	err := r.db.Session().Query(queries.UpdateTaskFeeQuery, fee, taskID).Exec()
	if err != nil {
		return fmt.Errorf("error updating task fee: %w", err)
	}

	// Emit WebSocket event for task fee update
	if r.publisher != nil {
		jobID := r.getJobIDFromTaskID(taskID)
		userID := r.getUserIDFromJobID(jobID)

		r.publisher.PublishTaskFeeUpdated(ctx, taskID, jobID.String(), oldFee, fee, userID)
	}

	return nil
}

func (r *taskRepository) GetTaskFee(taskID int64) (float64, error) {
	var fee float64
	err := r.db.Session().Query(queries.GetTaskFeeQuery, taskID).Scan(&fee)
	if err != nil {
		return 0, fmt.Errorf("error getting task fee: %w", err)
	}
	return fee, nil
}

func (r *taskRepository) GetCreatedChainIDByJobID(jobID *big.Int) (string, error) {
	var createdChainID string
	err := r.db.Session().Query(queries.GetCreatedChainIDByJobIDQuery, jobID).Scan(&createdChainID)
	if err != nil {
		return "", fmt.Errorf("error getting created chain ID by job ID: %w", err)
	}
	return createdChainID, nil
}

func (r *taskRepository) GetRecentTasks(limit int) ([]types.RecentTaskResponse, error) {
	iter := r.db.Session().Query(queries.GetRecentTasksQuery, limit).Iter()
	var tasks []types.RecentTaskResponse

	for {
		var task types.RecentTaskResponse
		var jobIDBigInt *big.Int

		if !iter.Scan(
			&task.TaskID,
			&task.TaskNumber,
			&jobIDBigInt,
			&task.TaskDefinitionID,
			&task.CreatedAt,
			&task.TaskOpXCost,
			&task.ExecutionTimestamp,
			&task.ExecutionTxHash,
			&task.TaskPerformerID,
			&task.TaskAttesterIDs,
			&task.TaskStatus,
			&task.TaskError,
			&task.IsImua,
		) {
			break
		}

		// Convert job ID to string
		if jobIDBigInt != nil {
			task.JobID = jobIDBigInt.String()

			// Get chain ID and generate TxURL if execution tx hash exists
			if task.ExecutionTxHash != "" {
				createdChainID, err := r.GetCreatedChainIDByJobID(jobIDBigInt)
				if err == nil {
					task.TxURL = getExplorerBaseURL(createdChainID) + task.ExecutionTxHash
				}
			}
		}

		tasks = append(tasks, task)
	}

	if err := iter.Close(); err != nil {
		return []types.RecentTaskResponse{}, errors.New("error getting recent tasks: " + err.Error())
	}

	return tasks, nil
}

// getExplorerBaseURL is a helper to get explorer URL - mirrors the one in handlers
func getExplorerBaseURL(chainID string) string {
	switch chainID {
	// Testnets
	case "11155111":
		return "https://eth-sepolia.blockscout.com/tx/"
	case "11155420": // OP Sepolia
		return "https://testnet-explorer.optimism.io/tx/"
	case "84532": // Base Sepolia
		return "https://base-sepolia.blockscout.com/tx/"
	case "421614": // Arbitrum Sepolia
		return "https://arbitrum-sepolia.blockscout.com/tx/"
	// Mainnets
	case "1": // Ethereum Mainnet
		return "https://eth.blockscout.com/tx/"
	case "10": // Optimism Mainnet
		return "https://explorer.optimism.io/tx/"
	case "8453": // Base Mainnet
		return "https://base.blockscout.com/tx/"
	case "42161": // Arbitrum Mainnet
		return "https://arbitrum.blockscout.com/tx/"
	default:
		return "https://sepolia.etherscan.io/tx/"
	}
}

// Helper methods for WebSocket event emission

// getJobIDFromTaskID retrieves job ID for a given task ID
func (r *taskRepository) getJobIDFromTaskID(taskID int64) *big.Int {
	taskData, err := r.GetTaskDataByID(taskID)
	if err != nil {
		return nil
	}
	return taskData.JobID.ToBigInt()
}

// getUserIDFromJobID retrieves user ID for a given job ID
func (r *taskRepository) getUserIDFromJobID(jobID *big.Int) string {
	if jobID == nil || jobID.Cmp(big.NewInt(0)) == 0 {
		return ""
	}

	// Query the job_data table to get the user_id
	var userID int64
	query := "SELECT user_id FROM triggerx.job_data WHERE job_id = ?"
	err := r.db.Session().Query(query, jobID).Scan(&userID)
	if err != nil {
		return ""
	}
	return fmt.Sprintf("%d", userID)
}

// getTaskStatus retrieves the current status of a task
func (r *taskRepository) getTaskStatus(taskID int64) string {
	taskData, err := r.GetTaskDataByID(taskID)
	if err != nil {
		return ""
	}
	return taskData.TaskStatus
}
