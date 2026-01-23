package repository

import (
	"errors"
	"fmt"
	"math/big"

	"github.com/trigg3rX/triggerx-backend/internal/dbserver/events"
	"github.com/trigg3rX/triggerx-backend/internal/dbserver/repository/queries"
	"github.com/trigg3rX/triggerx-backend/internal/dbserver/types"
	"github.com/trigg3rX/triggerx-backend/pkg/database"
	commonTypes "github.com/trigg3rX/triggerx-backend/pkg/types"
)

type TaskRepository interface {
	GetTaskDataByID(taskID int64) (commonTypes.TaskData, error)
	GetTasksByJobID(jobID *big.Int) ([]types.GetTasksByJobID, error)
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
