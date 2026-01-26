package types

import "time"

// ApiKeyDataDTO represents the API key data transfer object.
// Mirrors ApiKeyDataEntity but uses json tags for API serialization.
type ApiKeyDataDTO struct {
	Key          string    `json:"key"`           // API key (primary key, UUID)
	Owner        string    `json:"owner"`         // Wallet address of owner
	IsActive     bool      `json:"is_active"`     // Active or revoked
	RateLimit    int       `json:"rate_limit"`    // Requests per minute
	SuccessCount int64     `json:"success_count"` // Successful requests
	FailedCount  int64     `json:"failed_count"`  // Failed requests
	LastUsed     time.Time `json:"last_used"`     // Last request time
	CreatedAt    time.Time `json:"created_at"`    // Key creation time
}

// UserDataDTO represents the user data transfer object.
// Mirrors UserDataEntity but uses json tags for API serialization.
type UserDataDTO struct {
	UserAddress   string    `json:"user_address"`    // Wallet address (primary key)
	EmailID       string    `json:"email_id"`        // Email for notifications
	JobIDs        []string  `json:"job_ids"`         // List of job IDs owned by user
	UserPoints    string    `json:"user_points"`     // Total points (Wei-based, sum of JobCostActual)
	TotalJobs     int64     `json:"total_jobs"`      // Total jobs created
	TotalTasks    int64     `json:"total_tasks"`     // Total tasks executed
	CreatedAt     time.Time `json:"created_at"`      // Account creation timestamp
	LastUpdatedAt time.Time `json:"last_updated_at"` // Last profile update
}

// SafeAddressDataDTO represents the Safe address data transfer object.
// Mirrors SafeAddressDataEntity but uses json tags for API serialization.
type SafeAddressDataDTO struct {
	UserAddress string    `json:"user_address"` // Wallet address of owner
	SafeAddress string    `json:"safe_address"` // Safe address
	SafeName    string    `json:"safe_name"`    // Safe name
	CreatedAt   time.Time `json:"created_at"`   // Creation time
}

// JobDataDTO represents the job data transfer object.
// Mirrors JobDataEntity but uses json tags for API serialization.
type JobDataDTO struct {
	JobID             string    `json:"job_id"`              // Unique job identifier (same as Job Registry)
	JobTitle          string    `json:"job_title"`           // User-defined job name
	TaskDefinitionID  TaskDefinitionID       `json:"task_definition_id"`  // Maps to task type (1-9)
	CreatedChainID    string    `json:"created_chain_id"`    // Chain where job was created
	UserAddress       string    `json:"user_address"`        // Job owner
	LinkJobID         string    `json:"link_job_id"`         // Linked job ID (for chaining)
	ChainStatus       ChainStatus `json:"chain_status"`        // 0=None, 1=Chain Head, 2=Chain Block
	SafeAddress       string    `json:"safe_address"`        // Safe address
	Timezone          string    `json:"timezone"`            // User's timezone (e.g., "America/New_York")
	JobType           JobType   `json:"job_type"`            // "frontend", "sdk", "template", "contract"
	TimeFrame         int64     `json:"time_frame"`          // Job validity duration (seconds)
	Recurring         bool      `json:"recurring"`           // Recurring job or one-time
	Status            JobStatus `json:"status"`              // "created", "running", "completed", "failed", "expired", "deleted"
	JobCostPrediction string    `json:"job_cost_prediction"` // Estimated cost (Wei, for a single task)
	JobCostActual     string    `json:"job_cost_actual"`     // Actual cost (Wei, sum of task costs)
	TaskIDs           []int64   `json:"task_ids"`            // Tasks executed for this job
	CreatedAt         time.Time `json:"created_at"`          // Job creation time
	UpdatedAt         time.Time `json:"updated_at"`          // Last update time
	LastExecutedAt    time.Time `json:"last_executed_at"`    // Last task execution time
}

// TimeJobDataDTO represents the time job data transfer object.
// Mirrors TimeJobDataEntity but uses json tags for API serialization.
type TimeJobDataDTO struct {
	JobID                  string    `json:"job_id"`                   // Unique job identifier (varint in CQL, string in Go)
	TaskDefinitionID       TaskDefinitionID       `json:"task_definition_id"`       // Task type (1, 2, or 7)
	Network                Network   `json:"network"`                  // Network the job is running on
	ScheduleType           ScheduleType    `json:"schedule_type"`            // "cron", "interval", "specific"
	TimeInterval           int64     `json:"time_interval"`            // Interval in seconds (for interval type)
	CronExpression         string    `json:"cron_expression"`          // Cron expression (for cron type)
	SpecificSchedule       string    `json:"specific_schedule"`        // Specific timestamp (for specific type)
	Timezone               string    `json:"timezone"`                 // Timezone selected by user
	NextExecutionTimestamp time.Time `json:"next_execution_timestamp"` // Calculated next run time

	// Target data fields
	TargetChainID             string   `json:"target_chain_id"`              // Chain to execute on
	TargetContractAddress     string   `json:"target_contract_address"`      // Contract to call
	TargetFunction            string   `json:"target_function"`              // Function to invoke
	ABI                       string   `json:"abi"`                          // Contract ABI (JSON)
	ArgType                   ArgType  `json:"arg_type"`                     // 0=None, 1=Static, 2=Dynamic
	Arguments                 []string `json:"arguments"`                    // Static arguments (if ArgType=1, list<text> in CQL)
	ExecutionScriptURL        string   `json:"execution_script_url"`         // IPFS URL of the agent script
	ExecutionScriptLanguage   ScriptLanguage   `json:"execution_script_language"`    // Script language: 'ts', 'go', 'python', 'javascript'
	ExecutionScriptHash       string   `json:"execution_script_hash"`        // keccak256(scriptCode) for verification
	MaxExecutionTime        int    `json:"max_execution_time"`        // Script timeout in seconds (default 60)
	ChallengePeriod         int64  `json:"challenge_period"`          // Challenge period in seconds (default 21600 = 6 hours)

	// Common status fields
	IsActive       bool      `json:"is_active"`        // Job is active (true on creation, false when expiration time is reached)
	LastExecutedAt time.Time `json:"last_executed_at"` // Last execution time
	ExpirationTime time.Time `json:"expiration_time"`  // Job expiration
}

// EventJobDataDTO represents the event job data transfer object.
// Mirrors EventJobDataEntity but uses json tags for API serialization.
type EventJobDataDTO struct {
	JobID            string `json:"job_id"`             // Unique job identifier (varint in CQL, string in Go)
	TaskDefinitionID TaskDefinitionID    `json:"task_definition_id"` // Task type (3, 4, or 8)
	Network          Network `json:"network"`          // Network the job is running on
	Recurring        bool   `json:"recurring"`          // Trigger on every event in time frame or once

	// Event trigger fields (common for all)
	TriggerChainID         string `json:"trigger_chain_id"`         // Chain to monitor
	TriggerContractAddress string `json:"trigger_contract_address"` // Contract to watch
	TriggerEvent           string `json:"trigger_event"`            // Event signature (e.g., "Transfer(address,address,uint256)")
	EventFilterParaName    string `json:"event_filter_para_name"`   // Indexed parameter to filter (e.g., "to")
	EventFilterValue       string `json:"event_filter_value"`       // Filter value (e.g., specific address)

	// Target data fields
	TargetChainID             string   `json:"target_chain_id"`              // Chain to execute on
	TargetContractAddress     string   `json:"target_contract_address"`      // Contract to call
	TargetFunction            string   `json:"target_function"`              // Function to invoke
	ABI                       string   `json:"abi"`                          // Contract ABI (JSON)
	ArgType                   ArgType  `json:"arg_type"`                     // 0=None, 1=Static, 2=Dynamic
	Arguments                 []string `json:"arguments"`                    // Static arguments (if ArgType=1, list<text> in CQL)
	ExecutionScriptURL        string   `json:"execution_script_url"`         // IPFS URL of the agent script
	ExecutionScriptLanguage   ScriptLanguage   `json:"execution_script_language"`    // Script language: 'ts', 'go', 'python', 'javascript'
	ExecutionScriptHash       string   `json:"execution_script_hash"`        // keccak256(scriptCode) for verification
	MaxExecutionTime          int      `json:"max_execution_time"`           // Script timeout in seconds (default 60)
	ChallengePeriod           int64    `json:"challenge_period"`             // Challenge period in seconds (default 21600 = 6 hours)

	// Common status fields
	IsActive       bool      `json:"is_active"`        // Job is active (true on creation, false when expiration time is reached)
	LastExecutedAt time.Time `json:"last_executed_at"` // Last execution time
	ExpirationTime time.Time `json:"expiration_time"`  // Job expiration
}

// ConditionJobDataDTO represents the condition job data transfer object.
// Mirrors ConditionJobDataEntity but uses json tags for API serialization.
type ConditionJobDataDTO struct {
	JobID            string `json:"job_id"`             // Unique job identifier (varint in CQL, string in Go)
	TaskDefinitionID TaskDefinitionID    `json:"task_definition_id"` // Task type (5, 6, or 9)
	Network          Network `json:"network"`          // Network the job is running on
	Recurring        bool   `json:"recurring"`          // Trigger on every condition match or once

	// Condition trigger fields (common for all)
	ConditionType    string  `json:"condition_type"`     // "balance", "state", "oracle"
	UpperLimit       float64 `json:"upper_limit"`        // Upper threshold (double in CQL)
	LowerLimit       float64 `json:"lower_limit"`        // Lower threshold (double in CQL)
	ValueSourceType  ValueSourceType  `json:"value_source_type"`  // "api", "oracle", "websocket"
	ValueSourceURL   string  `json:"value_source_url"`   // API URL or Websocket URL
	SelectedKeyRoute string  `json:"selected_key_route"` // JSON path for API responses

	// Target data fields
	TargetChainID             string   `json:"target_chain_id"`              // Chain to execute on
	TargetContractAddress     string   `json:"target_contract_address"`      // Contract to call
	TargetFunction            string   `json:"target_function"`              // Function to invoke
	ABI                       string   `json:"abi"`                          // Contract ABI (JSON)
	ArgType                   ArgType  `json:"arg_type"`                     // 0=None, 1=Static, 2=Dynamic
	Arguments                 []string `json:"arguments"`                    // Static arguments (if ArgType=1, list<text> in CQL)
	ExecutionScriptURL        string   `json:"execution_script_url"`         // IPFS URL of the agent script
	ExecutionScriptLanguage   ScriptLanguage   `json:"execution_script_language"`    // Script language: 'ts', 'go', 'python', 'javascript'
	ExecutionScriptHash       string   `json:"execution_script_hash"`        // keccak256(scriptCode) for verification
	MaxExecutionTime          int      `json:"max_execution_time"`           // Script timeout in seconds (default 60)
	ChallengePeriod           int64    `json:"challenge_period"`             // Challenge period in seconds (default 21600 = 6 hours)

	// Common status fields
	IsActive       bool      `json:"is_active"`        // Job is active (true on creation, false when expiration time is reached)
	LastExecutedAt time.Time `json:"last_executed_at"` // Last execution time
	ExpirationTime time.Time `json:"expiration_time"`  // Job expiration
}

// TaskDataDTO represents the task data transfer object.
// Mirrors TaskDataEntity but uses json tags for API serialization.
type TaskDataDTO struct {
	TaskID               int64     `json:"task_id"`                 // Unique task ID (bigint in CQL, auto-increment)
	TaskNumber           int64     `json:"task_number"`             // Task sequence number on the contract
	TaskStatus           TaskStatus    `json:"task_status"`             // "pending", "in-queue", "running", "completed", "failed", "expired", "deleted"
	TaskError            string    `json:"task_error"`              // Error message if task failed
	JobID                string    `json:"job_id"`                  // Reference to job (varint in CQL, string in Go)
	TaskDefinitionID     TaskDefinitionID       `json:"task_definition_id"`      // Task definition ID (1-9)
	CreatedAt            time.Time `json:"created_at"`              // Task creation timestamp
	ExecutedAt           time.Time `json:"executed_at"`             // Execution timestamp
	SubmittedAt          time.Time `json:"submitted_at"`            // Submission timestamp
	TaskOpxPredictedCost string    `json:"task_opx_predicted_cost"` // From JobCostPrediction (Wei)
	TaskOpxActualCost    string    `json:"task_opx_actual_cost"`    // Actual cost including tx gas (Wei)
	ExecutionTxHash      string    `json:"execution_tx_hash"`       // Transaction hash for action
	SubmissionTxHash     string    `json:"submission_tx_hash"`      // Aggregator submission tx hash
	ConvertedArguments   []string  `json:"converted_arguments"`     // Arguments used (JSON, list<text> in CQL)
	TaskPerformerAddress []string  `json:"task_performer_address"`  // Keeper who executed (list<text> in CQL, multiple if first keeper failed)
	TaskAttesterAddress  []string  `json:"task_attester_address"`   // Keepers who attested (list<text> in CQL)
	ProofOfTask          string    `json:"proof_of_task"`           // Cryptographic proof (IPFS CID)
	IsSuccessful         bool      `json:"is_successful"`           // Execution success
	IsAccepted           bool      `json:"is_accepted"`             // Consensus accepted
	Network              Network   `json:"network"`                 // Network the task is running on
}

// KeeperDataDTO represents the keeper data transfer object.
// Mirrors KeeperDataEntity but uses json tags for API serialization.
type KeeperDataDTO struct {
	KeeperAddress     string      `json:"keeper_address"`     // Wallet address (primary key)
	KeeperName        string      `json:"keeper_name"`        // Human-readable keeper name
	RewardsAddress    string      `json:"rewards_address"`    // Address for reward payouts
	ConsensusAddress  string      `json:"consensus_address"`  // Consensus layer address
	OperatorID        int         `json:"operator_id"`        // Unique operator ID on the contract (int64 in CQL)
	VotingPower       string      `json:"voting_power"`       // Voting power (Wei-based stake)
	Registered        bool        `json:"registered"`         // Registered on-chain
	RegisteredAt      []time.Time `json:"registered_at"`      // List of registration timestamps (list<timestamp> in CQL)
	Online            bool        `json:"online"`             // Currently online
	Version           string      `json:"version"`            // Keeper software version
	Network           Network     `json:"network"`            // Network the keeper is running on
	ConnectionAddress string      `json:"connection_address"` // Connection address for P2P
	PeerID            string      `json:"peer_id"`            // Peer ID for P2P
	Uptime            int64       `json:"uptime"`             // Total uptime (seconds, bigint in CQL)
	LastCheckedIn     time.Time   `json:"last_checked_in"`    // Last heartbeat time
	RewardsBooster    string      `json:"rewards_booster"`    // Reward multiplier (for testnets, obsolete for mainnet)
	NoExecutedTasks   int64       `json:"no_executed_tasks"`  // Total tasks executed (bigint in CQL)
	NoAttestedTasks   int64       `json:"no_attested_tasks"`  // Total tasks attested (bigint in CQL)
	KeeperPoints      string      `json:"keeper_points"`      // Total points (Wei, sum of TaskOpxCost executed/attested, daily rewards)
	ChatID            int64       `json:"chat_id"`            // Telegram chat ID for alerts (int64 in CQL)
	EmailID           string      `json:"email_id"`           // Email for alerts
}

// AgentScriptExecutionsDTO represents the agent script executions data transfer object.
// Mirrors AgentScriptExecutionsEntity but uses json tags for API serialization.
type AgentScriptExecutionsDTO struct {
	ExecutionID      string    `json:"execution_id"`       // Unique execution identifier (UUID, primary key)
	JobID            string    `json:"job_id"`             // Reference to job (varint in CQL, string in Go)
	TaskID           int64     `json:"task_id"`            // Link to task_data table (bigint in CQL)
	TaskDefinitionID int       `json:"task_definition_id"` // 7, 8, or 9
	ScheduledTime    time.Time `json:"scheduled_time"`     // When it was supposed to execute
	ActualTime       time.Time `json:"actual_time"`        // When it actually executed
	PerformerAddress string    `json:"performer_address"`  // Keeper who executed the script

	// Input data (for deterministic re-execution)
	InputTimestamp int64  `json:"input_timestamp"` // Timestamp used as input (bigint in CQL)
	InputStorage   string `json:"input_storage"`   // JSON snapshot of storage at execution time
	InputHash      string `json:"input_hash"`      // Hash of inputs for verification

	// Trigger-specific input data (JSON)
	TriggerData string `json:"trigger_data"` // TDI 7: {scheduled_time}
	// TDI 8: {event_data, block_number, tx_hash, event_params}
	// TDI 9: {condition_value, timestamp, condition_type}

	// Output data
	ShouldExecute  bool   `json:"should_execute"`  // Whether to submit on-chain transaction
	TargetContract string `json:"target_contract"` // Contract address to call
	Calldata       string `json:"calldata"`        // Encoded function call (hex string)
	OutputHash     string `json:"output_hash"`     // Hash of outputs for verification

	// Metadata (CRITICAL: API calls, contract calls, block numbers)
	ExecutionMetadata string `json:"execution_metadata"` // JSON containing:
	// - API calls made: [{url, blockNumber, response}, ...]
	// - Contract calls: [{contract, blockNumber, function, response}, ...]
	// - Oracle calls: [{oracle, blockNumber, data}, ...]

	// Proof
	ScriptHash string `json:"script_hash"` // Hash of the script code
	Signature  string `json:"signature"`   // Performer's signature

	// Execution result
	TxHash          string `json:"tx_hash"`          // Transaction hash if executed
	ExecutionStatus string `json:"execution_status"` // 'success', 'failed', 'no_execution'
	ExecutionError  string `json:"execution_error"`  // Error message if failed

	// Verification status
	VerificationStatus VerificationStatus    `json:"verification_status"` // 'pending', 'verified', 'challenged', 'slashed'
	ChallengeDeadline  time.Time `json:"challenge_deadline"`  // Deadline for challenges
	IsChallenged       bool      `json:"is_challenged"`       // Whether execution has been challenged
	ChallengeCount     int       `json:"challenge_count"`     // Number of challenges raised
	CreatedAt          time.Time `json:"created_at"`          // Record creation timestamp
}

// ScriptStorageDTO represents the script storage data transfer object.
// Mirrors ScriptStorageEntity but uses json tags for API serialization.
type ScriptStorageDTO struct {
	JobID        string    `json:"job_id"`        // Job identifier (varint in CQL, string in Go, part of composite primary key)
	StorageKey   string    `json:"storage_key"`   // Storage key (text, part of composite primary key)
	StorageValue string    `json:"storage_value"` // Storage value (text, JSON or string)
	UpdatedAt    time.Time `json:"updated_at"`    // Last update timestamp
}

// ExecutionChallengesDTO represents the execution challenges data transfer object.
// Mirrors ExecutionChallengesEntity but uses json tags for API serialization.
type ExecutionChallengesDTO struct {
	ChallengeID       string `json:"challenge_id"`       // Unique challenge identifier (UUID, primary key)
	ExecutionID       string `json:"execution_id"`       // Reference to agent_script_executions
	ChallengerAddress string `json:"challenger_address"` // Address of the challenger
	ChallengeReason   ChallengeReason `json:"challenge_reason"`   // 'wrong_output', 'missing_execution', 'invalid_calldata', 'wrong_trigger'

	// Challenger's claimed output
	ChallengerOutputHash     string `json:"challenger_output_hash"`     // Hash of challenger's claimed output
	ChallengerShouldExecute  bool   `json:"challenger_should_execute"`  // Challenger's claim: should execute?
	ChallengerTargetContract string `json:"challenger_target_contract"` // Challenger's claimed target contract
	ChallengerCalldata       string `json:"challenger_calldata"`        // Challenger's claimed calldata
	ChallengerSignature      string `json:"challenger_signature"`       // Challenger's signature

	// Challenge bond (prevents spam)
	BondAmount string `json:"bond_amount"` // Wei-based bond amount

	// Resolution
	ResolutionStatus ResolutionStatus    `json:"resolution_status"` // 'pending', 'approved', 'rejected', 'inconclusive'
	ResolutionTime   time.Time `json:"resolution_time"`   // When challenge was resolved
	ValidatorCount   int       `json:"validator_count"`   // Number of validators who voted
	ApproveCount     int       `json:"approve_count"`     // Number of approving votes
	RejectCount      int       `json:"reject_count"`      // Number of rejecting votes
	CreatedAt        time.Time `json:"created_at"`        // Challenge creation timestamp
}

// CompleteJobDataDTO is a special composite DTO that combines job data with job-type-specific data.
// Only one of the job-type-specific fields is populated based on JobType.
type CompleteJobDataDTO struct {
	JobDataDTO          JobDataDTO           `json:"job_data"`
	TimeJobDataDTO      *TimeJobDataDTO      `json:"time_job_data,omitempty"`
	EventJobDataDTO     *EventJobDataDTO     `json:"event_job_data,omitempty"`
	ConditionJobDataDTO *ConditionJobDataDTO `json:"condition_job_data,omitempty"`
}
