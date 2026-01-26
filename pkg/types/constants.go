package types

// Network for Jobs, Tasks and Keepers
type Network string

const (
	NetworkMainnet Network = "mainnet"
	NetworkSepolia Network = "sepolia"
	NetworkImua    Network = "imua"
	NetworkHolesky Network = "holesky" // DEPRECATED, only used for Leaderboard Data
)

func GetNetwork(chainID string) Network {
	switch chainID {
	case "1", "10", "8453", "42161":
		return NetworkMainnet
	case "11155111", "11155420", "84532", "421614":
		return NetworkSepolia
	case "17000":
		return NetworkHolesky
	case "imua":
		return NetworkImua
	default:
		return NetworkSepolia
	}
}

// Job Types
type JobType string

const (
	JobTypeFrontend JobType = "frontend" // Scheduled from Web UI
	JobTypeSdk      JobType = "sdk"      // Scheduled from SDK
	JobTypeTemplate JobType = "template" // Template from frontend
	// JobTypeContract JobType = "contract" // Created from contract call (not yet implemented)
)

// Job Status
type JobStatus string

const (
	JobStatusRunning   JobStatus = "running"   // Job created and actively monitored by scheduler
	JobStatusCompleted JobStatus = "completed" // Job finished (atleast 1 task completed)
	JobStatusExpired   JobStatus = "expired"   // Job past expiration time with no tasks completed
	JobStatusFailed    JobStatus = "failed"    // Job failed permanently (≥1/3rd tasks failed)
	JobStatusDeleted   JobStatus = "deleted"   // Job deleted by user (was active when deleted)
)

// Chain Status (for job chaining)
type ChainStatus int

const (
	ChainStatusNone  ChainStatus = 0 // Not part of a chain
	ChainStatusHead  ChainStatus = 1 // First job in chain
	ChainStatusBlock ChainStatus = 2 // Middle or last job in chain
)

// Argument Types (for job arguments)
type ArgType int

const (
	ArgTypeNone    ArgType = 0 // No arguments
	ArgTypeStatic  ArgType = 1 // Static arguments
	ArgTypeDynamic ArgType = 2 // Dynamic arguments (from script)
)

// Schedule Types (for time-based jobs)
type ScheduleType string

const (
	ScheduleTypeCron     ScheduleType = "cron"     // Cron expression based scheduling
	ScheduleTypeInterval ScheduleType = "interval" // Interval based scheduling (seconds)
	ScheduleTypeSpecific ScheduleType = "specific" // Specific timestamp scheduling
)

// Condition Types (for condition-based jobs)
type ConditionType string

const (
	ConditionTypeGreaterThan  ConditionType = "greater_than"
	ConditionTypeLessThan     ConditionType = "less_than"
	ConditionTypeBetween      ConditionType = "between"
	ConditionTypeEquals       ConditionType = "equals"
	ConditionTypeNotEquals    ConditionType = "not_equals"
	ConditionTypeGreaterEqual ConditionType = "greater_equal"
	ConditionTypeLessEqual    ConditionType = "less_equal"
)

// Value Source Types (for condition-based jobs)
type ValueSourceType string

const (
	ValueSourceTypeAPI       ValueSourceType = "api"       // API endpoint as value source
	ValueSourceTypeOracle    ValueSourceType = "oracle"    // Oracle as value source
	ValueSourceTypeWebSocket ValueSourceType = "websocket" // Websocket as value source
)

// Script Languages
type ScriptLanguage string

const (
	ScriptLanguageTypeScript ScriptLanguage = "ts"
	ScriptLanguageGo         ScriptLanguage = "go"
	ScriptLanguagePython     ScriptLanguage = "python"
	ScriptLanguageJavaScript ScriptLanguage = "javascript"
)

// Execution Status (for agent script executions)
type ExecutionStatus string

const (
	ExecutionStatusSuccess     ExecutionStatus = "success"      // Execution succeeded
	ExecutionStatusFailed      ExecutionStatus = "failed"       // Execution failed
	ExecutionStatusNoExecution ExecutionStatus = "no_execution" // Script decided not to execute
)

// Verification Status (for agent script executions)
type VerificationStatus string

const (
	VerificationStatusPending    VerificationStatus = "pending"    // Awaiting verification
	VerificationStatusVerified   VerificationStatus = "verified"   // Verified as correct
	VerificationStatusChallenged VerificationStatus = "challenged" // Challenged by validator
	VerificationStatusSlashed    VerificationStatus = "slashed"    // Performer slashed
)

// Challenge Reasons (for execution challenges)
type ChallengeReason string

const (
	ChallengeReasonWrongOutput      ChallengeReason = "wrong_output"      // Output hash doesn't match expected result
	ChallengeReasonMissingExecution ChallengeReason = "missing_execution" // Execution should have happened but didn't
	ChallengeReasonInvalidCalldata  ChallengeReason = "invalid_calldata"  // Calldata is malformed or invalid
	ChallengeReasonWrongTrigger     ChallengeReason = "wrong_trigger"     // Trigger condition was incorrectly evaluated
)

// Resolution Status (for execution challenges)
type ResolutionStatus string

const (
	ResolutionStatusPending      ResolutionStatus = "pending"      // Challenge pending resolution
	ResolutionStatusApproved     ResolutionStatus = "approved"     // Challenge approved (performer was wrong)
	ResolutionStatusRejected     ResolutionStatus = "rejected"     // Challenge rejected (performer was correct)
	ResolutionStatusInconclusive ResolutionStatus = "inconclusive" // Resolution inconclusive
)

// Task Definition IDs
type TaskDefinitionID int

const (
	TaskDefTimeBasedStatic       TaskDefinitionID = 1 // Time based job with static arguments
	TaskDefTimeBasedDynamic      TaskDefinitionID = 2 // Time based job with dynamic arguments
	TaskDefTimeBasedAgent        TaskDefinitionID = 7 // Time based job with agentic script

	TaskDefEventBasedStatic      TaskDefinitionID = 3 // Event based job with static arguments
	TaskDefEventBasedDynamic     TaskDefinitionID = 4 // Event based job with dynamic arguments
	TaskDefEventBasedAgent       TaskDefinitionID = 8 // Event based job with agentic script

	TaskDefConditionBasedStatic  TaskDefinitionID = 5 // Condition based job with static arguments
	TaskDefConditionBasedDynamic TaskDefinitionID = 6 // Condition based job with dynamic arguments
	TaskDefConditionBasedAgent   TaskDefinitionID = 9 // Condition based job with agentic script
)

// Task Status
type TaskStatus string

const (
	TaskStatusCreated    TaskStatus = "created"   // Task created by schedulers
	TaskStatusDispatched TaskStatus = "dispatched" // Task dispatched to performer
	TaskStatusExecuted   TaskStatus = "executed"   // Task executed by performer
	TaskStatusCompleted  TaskStatus = "completed"  // Task completed successfully
	TaskStatusFailed     TaskStatus = "failed"     // Task failed execution
)

// Default Values
const (
	// Agent job defaults
	DefaultMaxExecutionTime = 50    // Default script timeout in seconds
	DefaultChallengePeriod  = 21600 // Default challenge period in seconds (6 hours)
)
