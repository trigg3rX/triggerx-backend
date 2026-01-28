package repository

import (
	"errors"
	"time"

	"github.com/gocql/gocql"
	"github.com/trigg3rX/triggerx-backend/pkg/database"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

// TaskRepository defines the interface for task data operations.
type TaskRepository interface {
	// CreateTaskDataInDB creates a new task record in the database.
	CreateTaskDataInDB(task *types.CreateTaskDataRequest) (int64, error)

	// AddTaskIDToJob adds a task ID to the job's task_ids list.
	AddTaskIDToJob(jobID string, taskID int64) error
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
// It also fetches the job_cost_prediction from job_data and sets it as task_opx_predicted_cost.
// After creating the task, it increments the user's total_tasks counter.
func (r *taskRepository) CreateTaskDataInDB(task *types.CreateTaskDataRequest) (int64, error) {
	var maxTaskID int64
	err := r.db.Session().Query(getMaxTaskIDQuery).Scan(&maxTaskID)
	if err != nil {
		return -1, errors.New("error getting max task ID")
	}

	// Fetch job_cost_prediction from job_data
	var jobCostPrediction string
	err = r.db.Session().Query(getJobCostPredictionQuery, task.JobID).Scan(&jobCostPrediction)
	if err != nil {
		return -1, errors.New("error getting job cost prediction")
	}

	taskID := maxTaskID + 1
	err = r.db.Session().Query(createTaskDataQuery, taskID, task.JobID, task.TaskDefinitionID, task.Network, string(types.TaskStatusCreated), time.Now().UTC(), jobCostPrediction).Exec()
	if err != nil {
		return -1, errors.New("error creating task data")
	}

	// Increment user total_tasks after task creation
	var userAddress string
	err = r.db.Session().Query(getUserAddressByJobIDQuery, task.JobID).Scan(&userAddress)
	if err != nil {
		// Log error but don't fail task creation if user address lookup fails
		// This is a non-critical operation
		return taskID, nil
	}

	var userTotalTasks int64
	err = r.db.Session().Query(getUserTotalTasksQuery, userAddress).Scan(&userTotalTasks)
	if err != nil {
		// Log error but don't fail task creation if user total tasks lookup fails
		// This is a non-critical operation
		return taskID, nil
	}

	// Increment total_tasks for the user
	err = r.db.Session().Query(incrementUserTotalTasksQuery, userTotalTasks + 1, time.Now().UTC(), userAddress).Exec()
	if err != nil {
		// Log error but don't fail task creation if increment fails
		// This is a non-critical operation
		return taskID, nil
	}

	return taskID, nil
}

// AddTaskIDToJob adds a task ID to the job's task_ids list.
// It first retrieves existing task IDs, appends the new one, and updates the job.
func (r *taskRepository) AddTaskIDToJob(jobID string, taskID int64) error {
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

	// Append the new task ID
	taskIDs := append(existingTaskIDs, taskID)
	err = r.db.Session().Query(addTaskIDToJobQuery, taskIDs, jobID).Exec()
	if err != nil {
		return errors.New("error adding task ID to job")
	}

	return nil
}
