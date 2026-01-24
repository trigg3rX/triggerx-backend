package repository

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/trigg3rX/triggerx-backend/internal/dbserver/events"

	"github.com/trigg3rX/triggerx-backend/pkg/database"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

type TaskRepository interface {
	GetTaskDataByID(taskID int64) (*types.TaskDataDTO, error)
	GetTasksByJobID(jobID string) ([]types.GetTasksByJobID, error)
	GetRecentTasks(limit int) ([]types.RecentTaskResponse, error)
	GetTaskFee(taskID int64) (string, error)
	GetCreatedChainIDByJobID(jobID string) (string, error)
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

func (r *taskRepository) GetTaskDataByID(taskID int64) (*types.TaskDataDTO, error) {
	var entity types.TaskDataEntity
	err := r.db.Session().Query(GetTaskDataByIDQuery, taskID).Scan(
		&entity.TaskID, &entity.TaskNumber, &entity.TaskStatus, &entity.TaskError,
		&entity.JobID, &entity.TaskDefinitionID, &entity.CreatedAt,
		&entity.ExecutedAt, &entity.SubmittedAt, &entity.TaskOpxPredictedCost,
		&entity.TaskOpxActualCost, &entity.ExecutionTxHash, &entity.SubmissionTxHash,
		&entity.ConvertedArguments, &entity.TaskPerformerAddress, &entity.TaskAttesterAddress,
		&entity.ProofOfTask, &entity.IsSuccessful, &entity.IsAccepted, &entity.IsImua)
	if err != nil {
		return nil, fmt.Errorf("error getting task data by ID: %w", err)
	}

	dto := types.TaskDataEntityToDTO(&entity)
	return dto, nil
}

func (r *taskRepository) GetTasksByJobID(jobID string) ([]types.GetTasksByJobID, error) {
	iter := r.db.Session().Query(GetTasksByJobIDQuery, jobID).Iter()
	var tasks []types.GetTasksByJobID
	var task types.GetTasksByJobID
	var entity types.TaskDataEntity

	for iter.Scan(
		&entity.TaskID, &entity.TaskNumber, &entity.TaskStatus, &entity.TaskError,
		&entity.TaskDefinitionID, &entity.CreatedAt, &entity.ExecutedAt, &entity.SubmittedAt,
		&entity.TaskOpxPredictedCost, &entity.TaskOpxActualCost, &entity.ExecutionTxHash,
		&entity.SubmissionTxHash, &entity.ConvertedArguments, &entity.TaskPerformerAddress,
		&entity.TaskAttesterAddress, &entity.IsSuccessful, &entity.IsAccepted) {

		// Convert entity to response type
		task.TaskID = entity.TaskID
		task.TaskNumber = entity.TaskNumber
		// Convert TaskOpxActualCost from string to float64 for response
		if costFloat, err := strconv.ParseFloat(entity.TaskOpxActualCost, 64); err == nil {
			task.TaskOpXCost = costFloat
		} else {
			task.TaskOpXCost = 0
		}
		task.ExecutionTimestamp = entity.ExecutedAt
		task.ExecutionTxHash = entity.ExecutionTxHash
		// TaskPerformerAddress is []string in GetTasksByJobID
		task.TaskPerformerAddress = entity.TaskPerformerAddress
		// TaskAttesterAddress (singular) is []string in GetTasksByJobID
		task.TaskAttesterAddress = entity.TaskAttesterAddress
		task.IsAccepted = entity.IsAccepted
		task.TaskStatus = entity.TaskStatus
		task.TaskError = entity.TaskError
		task.ConvertedArguments = entity.ConvertedArguments
		tasks = append(tasks, task)
	}

	if err := iter.Close(); err != nil {
		return []types.GetTasksByJobID{}, errors.New("error getting tasks by job ID: " + err.Error())
	}

	return tasks, nil
}

func (r *taskRepository) GetRecentTasks(limit int) ([]types.RecentTaskResponse, error) {
	iter := r.db.Session().Query(GetRecentTasksQuery, limit).Iter()
	var tasks []types.RecentTaskResponse

	for {
		var task types.RecentTaskResponse
		var entity types.TaskDataEntity

		if !iter.Scan(
			&entity.TaskID, &entity.TaskNumber, &entity.JobID, &entity.TaskDefinitionID,
			&entity.CreatedAt, &entity.TaskOpxActualCost, &entity.ExecutedAt,
			&entity.ExecutionTxHash, &entity.TaskPerformerAddress, &entity.TaskAttesterAddress,
			&entity.TaskStatus, &entity.TaskError, &entity.IsImua,
		) {
			break
		}

		// Convert entity to response type
		task.TaskID = entity.TaskID
		task.TaskNumber = entity.TaskNumber
		task.JobID = entity.JobID
		task.TaskDefinitionID = entity.TaskDefinitionID
		task.CreatedAt = entity.CreatedAt
		// Convert TaskOpxActualCost from string to float64
		if costFloat, err := strconv.ParseFloat(entity.TaskOpxActualCost, 64); err == nil {
			task.TaskOpXCost = costFloat
		} else {
			task.TaskOpXCost = 0
		}
		task.ExecutionTimestamp = entity.ExecutedAt
		task.ExecutionTxHash = entity.ExecutionTxHash
		// Convert list to single string (take first if exists)
		if len(entity.TaskPerformerAddress) > 0 {
			task.TaskPerformerAddress = entity.TaskPerformerAddress[0]
		}
		task.TaskAttesterAddresses = entity.TaskAttesterAddress
		task.TaskStatus = entity.TaskStatus
		task.TaskError = entity.TaskError
		task.IsImua = entity.IsImua

		// Get chain ID and generate TxURL if execution tx hash exists
		if task.ExecutionTxHash != "" {
			createdChainID, err := r.GetCreatedChainIDByJobID(entity.JobID)
			if err == nil {
				task.TxURL = getExplorerBaseURL(createdChainID) + task.ExecutionTxHash
			}
		}

		tasks = append(tasks, task)
	}

	if err := iter.Close(); err != nil {
		return []types.RecentTaskResponse{}, errors.New("error getting recent tasks: " + err.Error())
	}

	return tasks, nil
}

func (r *taskRepository) GetTaskFee(taskID int64) (string, error) {
	var fee string
	err := r.db.Session().Query(GetTaskFeeQuery, taskID).Scan(&fee)
	if err != nil {
		return "0", fmt.Errorf("error getting task fee: %w", err)
	}
	return fee, nil
}

func (r *taskRepository) GetCreatedChainIDByJobID(jobID string) (string, error) {
	var createdChainID string
	err := r.db.Session().Query(GetCreatedChainIDByJobIDQuery, jobID).Scan(&createdChainID)
	if err != nil {
		return "", fmt.Errorf("error getting created chain ID by job ID: %w", err)
	}
	return createdChainID, nil
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
