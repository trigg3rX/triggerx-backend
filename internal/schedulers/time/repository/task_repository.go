package repository

import (
	"context"
	"errors"
	"math/big"
	"time"

	"github.com/gocql/gocql"
	"github.com/trigg3rX/triggerx-backend/pkg/database"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

// TaskRepository defines the interface for task data operations.
type TaskRepository interface {
	// CreateTaskDataInDB creates a new task record in the database.
	CreateTaskDataInDB(ctx context.Context, task *types.CreateTaskDataRequest) (int64, error)

	// AddTaskIDToJob adds a task ID to the job's task_ids list.
	AddTaskIDToJob(jobID *big.Int, taskID int64) error
}

type taskRepository struct {
	db *database.Connection
}

// NewTaskRepository creates a new task repository instance.
func NewTaskRepository(db *database.Connection) TaskRepository {
	return &taskRepository{
		db: db,
	}
}

// CreateTaskDataInDB creates a new task record in the database.
// It generates a new task ID by getting the max task ID and incrementing it.
func (r *taskRepository) CreateTaskDataInDB(ctx context.Context, task *types.CreateTaskDataRequest) (int64, error) {
	var maxTaskID int64
	err := r.db.Session().Query(getMaxTaskIDQuery).Scan(&maxTaskID)
	if err != nil {
		return -1, errors.New("error getting max task ID")
	}

	taskID := maxTaskID + 1
	err = r.db.Session().Query(createTaskDataQuery, taskID, task.JobID, task.TaskDefinitionID, time.Now(), task.IsImua).Exec()
	if err != nil {
		return -1, errors.New("error creating task data")
	}

	return taskID, nil
}

// AddTaskIDToJob adds a task ID to the job's task_ids list.
// It first retrieves existing task IDs, appends the new one, and updates the job.
func (r *taskRepository) AddTaskIDToJob(jobID *big.Int, taskID int64) error {
	var existingTaskIDs []int64
	err := r.db.Session().Query(getTaskIDsByJobIDQuery, jobID).Scan(&existingTaskIDs)
	if err != nil {
		// If no task IDs exist yet (ErrNotFound) or if the field is null, start with an empty slice
		if err == gocql.ErrNotFound {
			existingTaskIDs = []int64{}
		} else {
			// For other errors, return the error
			return errors.New("error getting task IDs by job ID")
		}
	}

	// Handle case where existingTaskIDs might be nil (null in database)
	if existingTaskIDs == nil {
		existingTaskIDs = []int64{}
	}

	// Append the new task ID
	taskIDs := append(existingTaskIDs, taskID)
	err = r.db.Session().Query(addTaskIDToJobQuery, taskIDs, jobID).Exec()
	if err != nil {
		return errors.New("error adding task ID to job")
	}

	return nil
}

// Query constants
const (
	getMaxTaskIDQuery = `SELECT MAX(task_id) FROM triggerx.task_data`

	createTaskDataQuery = `
        INSERT INTO triggerx.task_data (
            task_id, job_id, task_definition_id, created_at, is_imua, task_status
        ) VALUES (?, ?, ?, ?, ?, 'processing')`

	getTaskIDsByJobIDQuery = `
		SELECT task_ids FROM triggerx.job_data 
		WHERE job_id = ?`

	addTaskIDToJobQuery = `
		UPDATE triggerx.job_data
		SET task_ids = ?
		WHERE job_id = ?`
)
