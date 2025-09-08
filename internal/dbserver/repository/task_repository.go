package repository

import (
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
	CreateTaskDataInDB(task *types.CreateTaskDataRequest) (int64, error)
	AddTaskPerformerID(taskID int64, performerID int64) error
	UpdateTaskExecutionDataInDB(task *types.UpdateTaskExecutionDataRequest) error
	UpdateTaskAttestationDataInDB(task *types.UpdateTaskAttestationDataRequest) error
	UpdateTaskNumberAndStatus(taskID int64, taskNumber int64, status string, txHash string) error
	GetTaskDataByID(taskID int64) (commonTypes.TaskData, error)
	GetTasksByJobID(jobID *big.Int) ([]types.GetTasksByJobID, error)
	AddTaskIDToJob(jobID *big.Int, taskID int64) error
	UpdateTaskFee(taskID int64, fee float64) error
	GetTaskFee(taskID int64) (float64, error)
	GetCreatedChainIDByJobID(jobID *big.Int) (string, error)
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

func (r *taskRepository) CreateTaskDataInDB(task *types.CreateTaskDataRequest) (int64, error) {
	var maxTaskID int64
	err := r.db.Session().Query(queries.GetMaxTaskIDQuery).Scan(&maxTaskID)
	if err != nil {
		return -1, errors.New("error getting max task ID")
	}

	taskID := maxTaskID + 1
	err = r.db.Session().Query(queries.CreateTaskDataQuery, taskID, task.JobID, task.TaskDefinitionID, time.Now(), task.IsImua).Exec()
	if err != nil {
		return -1, errors.New("error creating task data")
	}

	// Emit WebSocket event for task creation
	if r.publisher != nil {
		// Extract user ID from job data if available
		userID := r.getUserIDFromJobID(task.JobID)
		r.publisher.PublishTaskCreated(taskID, task.JobID.String(), int64(task.TaskDefinitionID), task.IsImua, userID)
	}

	return taskID, nil
}

func (r *taskRepository) AddTaskPerformerID(taskID int64, performerID int64) error {
	err := r.db.Session().Query(queries.AddTaskPerformerIDQuery, taskID, performerID).Exec()
	if err != nil {
		return errors.New("error adding task performer ID")
	}
	return nil
}

func (r *taskRepository) UpdateTaskExecutionDataInDB(task *types.UpdateTaskExecutionDataRequest) error {
	err := r.db.Session().Query(queries.UpdateTaskExecutionDataQuery, task.TaskPerformerID, task.ExecutionTimestamp, task.ExecutionTxHash, task.ProofOfTask, task.TaskOpXCost, task.TaskID).Exec()
	if err != nil {
		return errors.New("error updating task execution data")
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
		r.publisher.PublishTaskUpdated(task.TaskID, jobID.String(), userID, updateEvent)
	}

	return nil
}

func (r *taskRepository) UpdateTaskAttestationDataInDB(task *types.UpdateTaskAttestationDataRequest) error {
	err := r.db.Session().Query(queries.UpdateTaskAttestationDataQuery, task.TaskNumber, task.TaskAttesterIDs, task.TpSignature, task.TaSignature, task.TaskSubmissionTxHash, task.IsSuccessful, task.TaskID).Exec()
	if err != nil {
		return errors.New("error updating task attestation data")
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
		r.publisher.PublishTaskUpdated(task.TaskID, jobID.String(), userID, updateEvent)
	}

	return nil
}

func (r *taskRepository) UpdateTaskNumberAndStatus(taskID int64, taskNumber int64, status string, txHash string) error {
	// Get old status for comparison
	oldStatus := r.getTaskStatus(taskID)

	err := r.db.Session().Query(queries.UpdateTaskNumberAndStatusQuery, taskNumber, status, txHash, taskID).Exec()
	if err != nil {
		return errors.New("error updating task number and status")
	}

	// Emit WebSocket event for task status change
	if r.publisher != nil {
		jobID := r.getJobIDFromTaskID(taskID)
		userID := r.getUserIDFromJobID(jobID)

		r.publisher.PublishTaskStatusChanged(taskID, jobID.String(), oldStatus, status, userID, &taskNumber, &txHash)
	}

	return nil
}

func (r *taskRepository) GetTaskDataByID(taskID int64) (commonTypes.TaskData, error) {
	var task commonTypes.TaskData
	var jobIDBigInt *big.Int
	err := r.db.Session().Query(queries.GetTaskDataByIDQuery, taskID).Scan(&task.TaskID, &task.TaskNumber, &jobIDBigInt, &task.TaskDefinitionID, &task.CreatedAt, &task.TaskOpxCost, &task.ExecutionTimestamp, &task.ExecutionTxHash, &task.TaskPerformerID, &task.ProofOfTask, &task.TaskAttesterIDs, &task.TpSignature, &task.TaSignature, &task.TaskSubmissionTxHash, &task.IsAccepted, &task.TaskStatus, &task.IsImua)
	if err != nil {
		return commonTypes.TaskData{}, errors.New("error getting task data by ID")
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
	) {
		tasks = append(tasks, task)
	}

	if err := iter.Close(); err != nil {
		return []types.GetTasksByJobID{}, errors.New("error getting tasks by job ID: " + err.Error())
	}

	return tasks, nil
}

func (r *taskRepository) AddTaskIDToJob(jobID *big.Int, taskID int64) error {
	var taskIDs []int64
	iter := r.db.Session().Query(queries.GetTaskIDsByJobIDQuery, jobID).Iter()
	for iter.Scan(&taskIDs) {
		taskIDs = append(taskIDs, taskID)
	}
	if err := iter.Close(); err != nil {
		return errors.New("error getting task IDs by job ID")
	}
	taskIDs = append(taskIDs, taskID)
	err := r.db.Session().Query(queries.AddTaskIDToJobQuery, taskIDs, jobID).Exec()
	if err != nil {
		return errors.New("error adding task ID to job")
	}
	return nil
}

func (r *taskRepository) UpdateTaskFee(taskID int64, fee float64) error {
	// Get old fee for comparison
	oldFee, _ := r.GetTaskFee(taskID)

	err := r.db.Session().Query(queries.UpdateTaskFeeQuery, fee, taskID).Exec()
	if err != nil {
		return errors.New("error updating task fee")
	}

	// Emit WebSocket event for task fee update
	if r.publisher != nil {
		jobID := r.getJobIDFromTaskID(taskID)
		userID := r.getUserIDFromJobID(jobID)

		r.publisher.PublishTaskFeeUpdated(taskID, jobID.String(), oldFee, fee, userID)
	}

	return nil
}

func (r *taskRepository) GetTaskFee(taskID int64) (float64, error) {
	var fee float64
	err := r.db.Session().Query(queries.GetTaskFeeQuery, taskID).Scan(&fee)
	if err != nil {
		return 0, errors.New("error getting task fee")
	}
	return fee, nil
}

func (r *taskRepository) GetCreatedChainIDByJobID(jobID *big.Int) (string, error) {
	var createdChainID string
	err := r.db.Session().Query(queries.GetCreatedChainIDByJobIDQuery, jobID).Scan(&createdChainID)
	if err != nil {
		return "", errors.New("error getting created chain ID by job ID")
	}
	return createdChainID, nil
}

// Helper methods for WebSocket event emission

// getJobIDFromTaskID retrieves job ID for a given task ID
func (r *taskRepository) getJobIDFromTaskID(taskID int64) *big.Int {
	taskData, err := r.GetTaskDataByID(taskID)
	if err != nil {
		return big.NewInt(0)
	}
	return taskData.JobID.ToBigInt()
}

// getUserIDFromJobID retrieves user ID for a given job ID
func (r *taskRepository) getUserIDFromJobID(jobID *big.Int) string {
	// This is a simplified implementation
	// In production, you would query the job_data table to get the user_id
	// For now, we'll return a default value
	return "system"
}

// getTaskStatus retrieves the current status of a task
func (r *taskRepository) getTaskStatus(taskID int64) string {
	taskData, err := r.GetTaskDataByID(taskID)
	if err != nil {
		return ""
	}
	return taskData.TaskStatus
}
