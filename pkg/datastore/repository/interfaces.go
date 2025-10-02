package repository

import (
	"math/big"
	"time"

	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

// JobRepository defines the interface for job-related database operations
type JobRepository interface {
	CreateNewJob(job *types.JobData) (*big.Int, error)
	UpdateJobFromUserInDB(jobID *big.Int, job *types.UpdateJobDataFromUserRequest) error
	UpdateJobLastExecutedAt(jobID *big.Int, taskID int64, jobCostActual float64, lastExecutedAt time.Time) error
	UpdateJobStatus(jobID *big.Int, status string) error
	GetJobByID(jobID *big.Int) (*types.JobData, error)
	GetTaskDefinitionIDByJobID(jobID *big.Int) (int, error)
	GetTaskFeesByJobID(jobID *big.Int) ([]types.TaskFeeResponse, error)
	GetJobsByUserIDAndChainID(userID int64, createdChainID string) ([]types.JobData, error)
}

// TimeJobRepository defines the interface for time-based job operations
type TimeJobRepository interface {
	CreateTimeJob(timeJob *types.TimeJobData) error
	GetTimeJobByJobID(jobID *big.Int) (types.TimeJobData, error)
	CompleteTimeJob(jobID *big.Int) error
	UpdateTimeJobStatus(jobID *big.Int, isActive bool) error
	GetTimeJobsByNextExecutionTimestamp(lookAheadTime time.Time) ([]types.ScheduleTimeTaskData, error)
	UpdateTimeJobNextExecutionTimestamp(jobID *big.Int, nextExecutionTimestamp time.Time) error
	UpdateTimeJobInterval(jobID *big.Int, timeInterval int64) error
	GetActiveTimeJobs() ([]types.TimeJobData, error)
}

// EventJobRepository defines the interface for event-based job operations
type EventJobRepository interface {
	CreateEventJob(eventJob *types.EventJobData) error
	GetEventJobByJobID(jobID *big.Int) (types.EventJobData, error)
	CompleteEventJob(jobID *big.Int) error
	UpdateEventJobStatus(jobID *big.Int, isActive bool) error
	GetActiveEventJobs() ([]types.EventJobData, error)
}

// ConditionJobRepository defines the interface for condition-based job operations
type ConditionJobRepository interface {
	CreateConditionJob(conditionJob *types.ConditionJobData) error
	GetConditionJobByJobID(jobID *big.Int) (types.ConditionJobData, error)
	CompleteConditionJob(jobID *big.Int) error
	UpdateConditionJobStatus(jobID *big.Int, isActive bool) error
	GetActiveConditionJobs() ([]types.ConditionJobData, error)
}

// TaskRepository defines the interface for task-related database operations
type TaskRepository interface {
	GetMaxTaskID() (int64, error)
	CreateTaskDataInDB(task *types.CreateTaskDataRequest) (int64, error)
	AddTaskPerformerID(taskID int64, performerID int64) error
	UpdateTaskExecutionDataInDB(task *types.UpdateTaskExecutionDataRequest) error
	UpdateTaskAttestationDataInDB(task *types.UpdateTaskAttestationDataRequest) error
	UpdateTaskNumberAndStatus(taskID int64, taskNumber int64, status string, txHash string) error
	GetTaskDataByID(taskID int64) (types.TaskData, error)
	GetTasksByJobID(jobID *big.Int) ([]types.GetTasksByJobID, error)
	AddTaskIDToJob(jobID *big.Int, taskID int64) error
	UpdateTaskFee(taskID int64, fee float64) error
	GetTaskFee(taskID int64) (float64, error)
	GetCreatedChainIDByJobID(jobID *big.Int) (string, error)
}

// KeeperRepository defines the interface for keeper-related database operations
type KeeperRepository interface {
	CheckKeeperExists(address string) (int64, error)
	CreateKeeper(keeperData types.CreateKeeperData) (int64, error)
	GetKeeperAsPerformer() ([]types.GetPerformerData, error)
	GetKeeperDataByID(id int64) (types.KeeperData, error)
	IncrementKeeperTaskCount(id int64) (int64, error)
	GetKeeperTaskCount(id int64) (int64, error)
	UpdateKeeperPoints(id int64, taskFee float64) (float64, error)
	UpdateKeeperChatID(address string, chatID int64) error
	GetKeeperPointsByIDInDB(id int64) (float64, error)
	GetKeeperCommunicationInfo(id int64) (types.KeeperCommunicationInfo, error)
	GetKeeperLeaderboard() ([]types.KeeperLeaderboardEntry, error)
	GetKeeperLeaderboardByOnImua(onImua bool) ([]types.KeeperLeaderboardEntry, error)
	GetKeeperLeaderboardByIdentifierInDB(address string, name string) (types.KeeperLeaderboardEntry, error)
	CreateOrUpdateKeeperFromGoogleForm(keeperData types.GoogleFormCreateKeeperData) (int64, error)
}

// UserRepository defines the interface for user-related database operations
type UserRepository interface {
	CheckUserExists(address string) (int64, error)
	CreateNewUser(user *types.CreateUserDataRequest) (types.CreateUserDataRequest, error)
	UpdateUserBalance(user *types.UpdateUserBalanceRequest) error
	UpdateUserJobIDs(userID int64, jobIDs []*big.Int) error
	UpdateUserTasksAndPoints(userID int64, tasksCompleted int64, userPoints float64) error
	GetUserDataByAddress(address string) (int64, types.UserData, error)
	GetUserPointsByID(id int64) (float64, error)
	GetUserPointsByAddress(address string) (float64, error)
	GetUserJobIDsByAddress(address string) (int64, []*big.Int, error)
	GetUserLeaderboard() ([]types.UserLeaderboardEntry, error)
	GetUserLeaderboardByAddress(address string) (types.UserLeaderboardEntry, error)
	UpdateUserEmail(address string, email string) error
}

// ApiKeysRepository defines the interface for API key-related database operations
type ApiKeysRepository interface {
	CreateApiKey(apiKey *types.ApiKeyData) error
	GetApiKeyDataByOwner(owner string) ([]*types.ApiKeyData, error)
	GetApiKeyDataByKey(key string) (*types.ApiKeyData, error)
	GetApiKeyCounters(key string) (*types.ApiKeyCounters, error)
	GetApiKeyByOwner(owner string) (key string, err error)
	GetApiOwnerByApiKey(key string) (owner string, err error)
	UpdateApiKey(apiKey *types.UpdateApiKeyRequest) error
	UpdateApiKeyStatus(apiKey *types.UpdateApiKeyStatusRequest) error
	UpdateApiKeyLastUsed(key string, isSuccess bool) error
}
