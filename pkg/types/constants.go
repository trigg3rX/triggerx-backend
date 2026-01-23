package types

// Task Definition IDs
const (
	TaskDefTimeBasedStatic       = 1 // Time based job with static arguments
	TaskDefTimeBasedDynamic      = 2 // Time based job with dynamic arguments
	TaskDefTimeBasedAgent        = 7 // Time based job with agentic script

	TaskDefEventBasedStatic      = 3 // Event based job with static arguments
	TaskDefEventBasedDynamic     = 4 // Event based job with dynamic arguments
	TaskDefEventBasedAgent       = 8 // Event based job with agentic script

	TaskDefConditionBasedStatic  = 5 // Condition based job with static arguments
	TaskDefConditionBasedDynamic = 6 // Condition based job with dynamic arguments
	TaskDefConditionBasedAgent   = 9 // Condition based job with agentic script
)

// Job Types
const (
	JobTypeFrontend = "frontend" // Scheduled from Web UI
	JobTypeSdk      = "sdk"      // Scheduled from SDK
	JobTypeTemplate = "template" // Template from frontend
	JobTypeContract = "contract" // Created from contract call (not yet implemented)
)

// Job Status
const (
	JobStatusRunning   = "running"   // Job created and actively monitored by scheduler
	JobStatusCompleted = "completed" // Job finished (atleast 1 task completed)
	JobStatusExpired   = "expired"   // Job past expiration time with no tasks completed
	JobStatusFailed    = "failed"    // Job failed permanently (≥1/3rd tasks failed)
	JobStatusDeleted   = "deleted"   // Job deleted by user (was active when deleted)
)

// Task Status
const (
	TaskStatusCreated   = "created"   // Task created by schedulers
	TaskStatusDispatched = "dispatched" // Task dispatched to performer
	TaskStatusExecuted = "executed" // Task executed by performer
	TaskStatusCompleted = "completed" // Task completed successfully
	TaskStatusFailed    = "failed"    // Task failed execution
)

// Argument Types
const (
	ArgTypeNone    = 0 // No arguments
	ArgTypeStatic  = 1 // Static arguments provided at job creation
	ArgTypeDynamic = 2 // Dynamic arguments generated at execution time
)

// Chain Status (for job chaining)
const (
	ChainStatusNone  = 0 // Not part of a chain
	ChainStatusHead  = 1 // First job in chain
	ChainStatusBlock = 2 // Middle or last job in chain
)

// Schedule Types (for time-based jobs)
const (
	ScheduleTypeCron     = "cron"     // Cron expression based scheduling
	ScheduleTypeInterval = "interval" // Interval based scheduling (seconds)
	ScheduleTypeSpecific = "specific" // Specific timestamp scheduling
)

// Condition Types (for condition-based jobs)
const (
	ConditionTypeGreaterThan  = "greater_than"
	ConditionTypeLessThan     = "less_than"
	ConditionTypeBetween      = "between"
	ConditionTypeEquals       = "equals"
	ConditionTypeNotEquals    = "not_equals"
	ConditionTypeGreaterEqual = "greater_equal"
	ConditionTypeLessEqual    = "less_equal"
)

// Value Source Types (for condition-based jobs)
const (
	ValueSourceTypeAPI    = "api"       // API endpoint as value source
	ValueSourceTypeOracle = "oracle"    // Oracle as value source
	ValueSourceTypeStatic = "static"    // Static value as value source
	ValueSourceTypeWebSocket = "websocket" // Websocket as value source
)

// Agent Script Languages
const (
	AgentScriptLanguageTypeScript = "ts"         // TypeScript
	AgentScriptLanguageGo         = "go"         // Go
	AgentScriptLanguagePython     = "python"     // Python
	AgentScriptLanguageJavaScript = "javascript" // JavaScript
)

// Execution Status (for agent script executions)
const (
	ExecutionStatusSuccess     = "success"      // Execution succeeded
	ExecutionStatusFailed      = "failed"       // Execution failed
	ExecutionStatusNoExecution = "no_execution" // Script decided not to execute
)

// Verification Status (for agent script executions)
const (
	VerificationStatusPending    = "pending"    // Awaiting verification
	VerificationStatusVerified   = "verified"   // Verified as correct
	VerificationStatusChallenged = "challenged" // Challenged by validator
	VerificationStatusSlashed    = "slashed"    // Performer slashed
)

// Challenge Reasons (for execution challenges)
const (
	ChallengeReasonWrongOutput      = "wrong_output"      // Output hash doesn't match expected result
	ChallengeReasonMissingExecution = "missing_execution" // Execution should have happened but didn't
	ChallengeReasonInvalidCalldata  = "invalid_calldata"  // Calldata is malformed or invalid
	ChallengeReasonWrongTrigger     = "wrong_trigger"     // Trigger condition was incorrectly evaluated
)

// Resolution Status (for execution challenges)
const (
	ResolutionStatusPending      = "pending"      // Challenge pending resolution
	ResolutionStatusApproved     = "approved"     // Challenge approved (performer was wrong)
	ResolutionStatusRejected     = "rejected"     // Challenge rejected (performer was correct)
	ResolutionStatusInconclusive = "inconclusive" // Resolution inconclusive
)

// Default Values
const (
	// Agent job defaults
	DefaultMaxExecutionTime = 50    // Default script timeout in seconds
	DefaultChallengePeriod  = 21600 // Default challenge period in seconds (6 hours)
)

// Keeper Network
type KeeperNetwork string

const (
	NetworkMainnet KeeperNetwork = "mainnet"
	NetworkSepolia KeeperNetwork = "sepolia"
	NetworkImua    KeeperNetwork = "imua"
)