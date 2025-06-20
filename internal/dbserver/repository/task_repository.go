package repository

import (
	"errors"
	"time"

	"github.com/trigg3rX/triggerx-backend/internal/dbserver/repository/queries"
	"github.com/trigg3rX/triggerx-backend/internal/dbserver/types"
	"github.com/trigg3rX/triggerx-backend/pkg/database"
)

type TaskRepository interface {
	CreateTaskDataInDB(task *types.CreateTaskDataRequest) (int64, error)
	AddTaskPerformerID(taskID int64, performerID int64) error
	UpdateTaskExecutionDataInDB(task *types.UpdateTaskExecutionDataRequest) error
	UpdateTaskAttestationDataInDB(task *types.UpdateTaskAttestationDataRequest) error
	GetTaskDataByID(taskID int64) (types.TaskData, error)
	GetTasksByJobID(jobID int64) ([]types.TasksByJobIDResponse, error)
	UpdateTaskFee(taskID int64, fee float64) error
	GetTaskFee(taskID int64) (float64, error)
}

type taskRepository struct {
	db *database.Connection
}

func NewTaskRepository(db *database.Connection) TaskRepository {
	return &taskRepository{
		db: db,
	}
}

func (r *taskRepository) CreateTaskDataInDB(task *types.CreateTaskDataRequest) (int64, error) {
	var maxTaskID int64
	err := r.db.Session().Query(queries.GetMaxTaskIDQuery).Scan(&maxTaskID)
	if err != nil {
		return -1, errors.New("error getting max task ID")
	}
	err = r.db.Session().Query(queries.CreateTaskDataQuery, maxTaskID+1, task.JobID, task.TaskDefinitionID, time.Now()).Exec()
	if err != nil {
		return -1, errors.New("error creating task data")
	}
	return maxTaskID + 1, nil
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
	return nil
}

func (r *taskRepository) UpdateTaskAttestationDataInDB(task *types.UpdateTaskAttestationDataRequest) error {
	err := r.db.Session().Query(queries.UpdateTaskAttestationDataQuery, task.TaskNumber, task.TaskAttesterIDs, task.TpSignature, task.TaSignature, task.TaskSubmissionTxHash, task.IsSuccessful, task.TaskID).Exec()
	if err != nil {
		return errors.New("error updating task attestation data")
	}
	return nil
}

func (r *taskRepository) GetTaskDataByID(taskID int64) (types.TaskData, error) {
	var task types.TaskData
	err := r.db.Session().Query(queries.GetTaskDataByIDQuery, taskID).Scan(&task.TaskID, &task.TaskNumber, &task.JobID, &task.TaskDefinitionID, &task.CreatedAt, &task.TaskOpXCost, &task.ExecutionTimestamp, &task.ExecutionTxHash, &task.TaskPerformerID, &task.ProofOfTask, &task.TaskAttesterIDs, &task.TpSignature, &task.TaSignature, &task.TaskSubmissionTxHash, &task.IsSuccessful)
	if err != nil {
		return types.TaskData{}, errors.New("error getting task data by ID")
	}
	return task, nil
}

func (r *taskRepository) GetTasksByJobID(jobID int64) ([]types.TasksByJobIDResponse, error) {
	iter := r.db.Session().Query(queries.GetTasksByJobIDQuery, jobID).Iter()
	var tasks []types.TasksByJobIDResponse
	var task types.TasksByJobIDResponse

	for iter.Scan(
		&task.TaskID,
		&task.TaskNumber,
		&task.TaskOpXCost,
		&task.ExecutionTimestamp,
		&task.ExecutionTxHash,
		&task.TaskPerformerID,
		&task.TaskAttesterIDs,
		&task.IsSuccessful,
	) {
		tasks = append(tasks, task)
	}

	if err := iter.Close(); err != nil {
		return []types.TasksByJobIDResponse{}, errors.New("error getting tasks by job ID")
	}

	return tasks, nil
}

func (r *taskRepository) UpdateTaskFee(taskID int64, fee float64) error {
	err := r.db.Session().Query(queries.UpdateTaskFeeQuery, fee, taskID).Exec()
	if err != nil {
		return errors.New("error updating task fee")
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
