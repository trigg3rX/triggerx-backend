package types

import (
	"math/big"
	"time"
)

// User represents a user in the system with their associated data and metrics
type UserData struct {
	UserID      int64
	UserAddress string
	EmailID     string
	JobIDs      []*big.Int
	// TGConsumed is the current TG consumed by the user from the task data (in Wei)
	TGConsumed    *big.Int
	TotalJobs     int64
	TotalTasks    int64
	CreatedAt     time.Time
	LastUpdatedAt time.Time
}

// Job represents a job in the system with its associated data and metrics
type JobData struct {
	JobID            *big.Int
	JobTitle         string
	TaskDefinitionID int
	CreatedChainID   string
	UserID           int64
	LinkJobID        *big.Int
	ChainStatus      int
	Timezone         string
	IsImua           bool
	JobType          string // sdk, frontend, contract, template
	TimeFrame        int64
	Recurring        bool
	Status           string // created, running, completed, failed, paused
	// Intial cost prediction for the job for a single task execution (in Wei)
	JobCostPrediction big.Int
	// Actual cost of the job, updated after each task execution
	JobCostActual big.Int
	// Task IDs for the job
	TaskIDs        []int64
	CreatedAt      time.Time
	UpdatedAt      time.Time
	LastExecutedAt time.Time
}

type TimeJobData struct {
	JobID            *big.Int
	TaskDefinitionID int
	// Time based job specific fields
	ScheduleType           string
	TimeInterval           int64
	CronExpression         string
	SpecificSchedule       string
	NextExecutionTimestamp time.Time
	// Target fields (common for all job types)
	TargetChainID             string
	TargetContractAddress     string
	TargetFunction            string
	ABI                       string
	ArgType                   int
	Arguments                 []string
	DynamicArgumentsScriptUrl string
	IsCompleted               bool
	LastExecutedAt            time.Time
	// ExpirationTime = CreatedAt + TimeFrame, easier than using TimeFrame in schedulers
	ExpirationTime time.Time
}

type EventJobData struct {
	JobID            *big.Int
	TaskDefinitionID int
	Recurring        bool
	// Event based job specific fields
	TriggerChainID             string
	TriggerContractAddress     string
	TriggerEvent               string
	TriggerEventFilterParaName string
	TriggerEventFilterValue    string
	// Target fields (common for all job types)
	TargetChainID             string
	TargetContractAddress     string
	TargetFunction            string
	ABI                       string
	ArgType                   int
	Arguments                 []string
	DynamicArgumentsScriptUrl string
	IsCompleted               bool
	LastExecutedAt            time.Time
	ExpirationTime            time.Time
}

type ConditionJobData struct {
	JobID            *big.Int
	TaskDefinitionID int
	Recurring        bool
	// Condition based job specific fields
	ConditionType    string
	UpperLimit       float64
	LowerLimit       float64
	ValueSourceType  string
	ValueSourceUrl   string
	SelectedKeyRoute string
	// Target fields (common for all job types)
	TargetChainID             string
	TargetContractAddress     string
	TargetFunction            string
	ABI                       string
	ArgType                   int
	Arguments                 []string
	DynamicArgumentsScriptUrl string
	IsCompleted               bool
	LastExecutedAt            time.Time
	ExpirationTime            time.Time
}

type TaskData struct {
	TaskID               int64
	TaskNumber           int64
	JobID                *big.Int
	TaskDefinitionID     int
	CreatedAt            time.Time
	TaskOpXPredictedCost big.Int
	TaskOpXActualCost    big.Int
	ExecutionTimestamp   time.Time
	ExecutionTxHash      string
	TaskPerformerID      int64
	TaskAttesterIDs      []int64
	ProofOfTask          string
	SubmissionTxHash     string
	IsSuccessful         bool
	IsAccepted           bool
	IsImua               bool
}

// Keeper represents a keeper in the system with their associated data and metrics
type KeeperData struct {
	KeeperID         int64
	KeeperName       string
	KeeperAddress    string
	RewardsAddress   string
	ConsensusAddress string
	RegisteredTx     string
	OperatorID       int64
	VotingPower      big.Int
	Whitelisted      bool
	Registered       bool
	Online           bool
	Version          string
	OnImua           bool
	PublicIP         string
	PeerID           string
	ChatID           int64
	EmailID          string
	RewardsBooster   big.Int
	NoExecutedTasks  int64
	NoAttestedTasks  int64
	Uptime           int64
	KeeperPoints     big.Int
	LastCheckedIn    time.Time
}

// ApiKey represents an API key in the system with its associated data and metrics
type ApiKeyData struct {
	Key          string
	Owner        string
	IsActive     bool
	RateLimit    int
	SuccessCount int64
	FailedCount  int64
	LastUsed     time.Time
	CreatedAt    time.Time
}
