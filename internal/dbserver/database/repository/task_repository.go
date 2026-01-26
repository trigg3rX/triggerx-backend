package repository

import (
	"errors"
	"fmt"

	"github.com/trigg3rX/triggerx-backend/internal/dbserver/events"

	"github.com/trigg3rX/triggerx-backend/pkg/database"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

type TaskRepository interface {
	GetTaskDataByID(taskID int64) (*types.TaskDataDTO, error)
	GetTasksByJobID(jobID string) ([]types.TasksByJobIDResponse, error)
	GetRecentTasks(limit int) ([]types.RecentTaskResponse, error)
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
		&entity.ProofOfTask, &entity.IsSuccessful, &entity.IsAccepted, &entity.Network)
	if err != nil {
		return nil, fmt.Errorf("error getting task data by ID: %w", err)
	}

	dto := types.TaskDataEntityToDTO(&entity)
	return dto, nil
}

func (r *taskRepository) GetTasksByJobID(jobID string) ([]types.TasksByJobIDResponse, error) {
	iter := r.db.Session().Query(GetTasksByJobIDQuery, jobID).Iter()
	var tasks []types.TasksByJobIDResponse
	var task types.TasksByJobIDResponse

	for iter.Scan(
		&task.TaskID, &task.TaskNumber, &task.TaskStatus, &task.TaskError,
		&task.CreatedAt, &task.ExecutedAt, &task.SubmittedAt,
		&task.TaskOpxPredictedCost, &task.TaskOpxActualCost, &task.ExecutionTxHash,
		&task.ConvertedArguments, &task.TaskPerformerAddress,
		&task.TaskAttesterAddress, &task.IsSuccessful, &task.IsAccepted) {
		// Get chain ID and generate TxURL if execution tx hash exists
		if task.ExecutionTxHash != "" {
			createdChainID, err := r.GetCreatedChainIDByJobID(jobID)
			if err == nil {
				task.TxURL = getExplorerBaseURL(createdChainID) + task.ExecutionTxHash
			}
		}
		tasks = append(tasks, task)
	}

	if err := iter.Close(); err != nil {
		return []types.TasksByJobIDResponse{}, errors.New("error getting tasks by job ID: " + err.Error())
	}

	return tasks, nil
}

func (r *taskRepository) GetRecentTasks(limit int) ([]types.RecentTaskResponse, error) {
	iter := r.db.Session().Query(GetRecentTasksQuery, limit).Iter()
	var tasks []types.RecentTaskResponse

	for {
		var task types.RecentTaskResponse
		var performerAddress []string

		if !iter.Scan(
			&task.TaskID, &task.TaskNumber, &task.JobID, &task.TaskDefinitionID,
			&task.CreatedAt, &task.TaskOpXCost, &task.ExecutionTimestamp,
			&task.ExecutionTxHash, &performerAddress, &task.TaskAttesterAddresses,
			&task.TaskStatus, &task.TaskError, &task.Network,
		) {
			break
		}
		// Convert list to single string (take first if exists)
		if len(performerAddress) > 0 {
			task.TaskPerformerAddress = performerAddress[0]
		}
		// Get chain ID and generate TxURL if execution tx hash exists
		if task.ExecutionTxHash != "" {
			createdChainID, err := r.GetCreatedChainIDByJobID(task.JobID)
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
