package types

import "time"

// ApiKeyDataEntity represents the apikeys table for API key management.
// Primary Key: key
type ApiKeyDataEntity struct {
	Key          string    `cql:"key"`           // API key (primary key, UUID)
	Owner        string    `cql:"owner"`         // Wallet address of owner
	IsActive     bool      `cql:"is_active"`     // Active or revoked
	RateLimit    int       `cql:"rate_limit"`    // Requests per minute
	SuccessCount int64     `cql:"success_count"` // Successful requests
	FailedCount  int64     `cql:"failed_count"`  // Failed requests
	LastUsed     time.Time `cql:"last_used"`     // Last request time
	CreatedAt    time.Time `cql:"created_at"`    // Key creation time
}

// UserDataEntity represents the user_data table for user profiles.
// Primary Key: user_address
type UserDataEntity struct {
	UserAddress   string    `cql:"user_address"`    // Wallet address (primary key)
	EmailID       string    `cql:"email_id"`        // Email for notifications
	JobIDs        []string  `cql:"job_ids"`         // List of job IDs owned by user
	UserPoints    string    `cql:"user_points"`     // Total points (Wei-based, sum of JobCostActual)
	TotalJobs     int64     `cql:"total_jobs"`      // Total jobs created
	TotalTasks    int64     `cql:"total_tasks"`     // Total tasks executed
	CreatedAt     time.Time `cql:"created_at"`      // Account creation timestamp
	LastUpdatedAt time.Time `cql:"last_updated_at"` // Last profile update
}

// SafeAddressDataEntity represents the safe_addresses table for Safe address management.
// Primary Key: (user_address, safe_address)
type SafeAddressDataEntity struct {
	UserAddress string    `cql:"user_address"` // Wallet address of owner
	SafeAddress string    `cql:"safe_address"` // Safe address
	SafeName    string    `cql:"safe_name"`    // Safe name
	CreatedAt   time.Time `cql:"created_at"`   // Creation time
}

// JobDataEntity represents the job_data table for all job types.
// Primary Key: job_id
type JobDataEntity struct {
	JobID             string    `cql:"job_id"`              // Unique job identifier (same as Job Registry)
	JobTitle          string    `cql:"job_title"`           // User-defined job name
	TaskDefinitionID  int       `cql:"task_definition_id"`  // Maps to task type (1-9)
	CreatedChainID    string    `cql:"created_chain_id"`    // Chain where job was created
	UserAddress       string    `cql:"user_address"`        // Job owner
	LinkJobID         string    `cql:"link_job_id"`         // Linked job ID (for chaining)
	ChainStatus       int       `cql:"chain_status"`        // 0=None, 1=Chain Head, 2=Chain Block
	SafeAddress       string    `cql:"safe_address"`        // Safe address
	Timezone          string    `cql:"timezone"`            // User's timezone (e.g., "America/New_York")
	JobType           string    `cql:"job_type"`            // "frontend", "sdk", "template", "contract"
	TimeFrame         int64     `cql:"time_frame"`          // Job validity duration (seconds)
	Recurring         bool      `cql:"recurring"`           // Recurring job or one-time
	Status            string    `cql:"status"`              // "created", "running", "completed", "failed", "expired", "deleted"
	JobCostPrediction string    `cql:"job_cost_prediction"` // Estimated cost (Wei, for a single task)
	JobCostActual     string    `cql:"job_cost_actual"`     // Actual cost (Wei, sum of task costs)
	TaskIDs           []int64   `cql:"task_ids"`            // Tasks executed for this job
	CreatedAt         time.Time `cql:"created_at"`          // Job creation time
	UpdatedAt         time.Time `cql:"updated_at"`          // Last update time
}

// TimeJobDataEntity represents the time_job_data table for time-based scheduled jobs.
// Supports both traditional jobs (TDI 1, 2) and agent jobs (TDI 7).
// Primary Key: job_id
type TimeJobDataEntity struct {
	JobID                  string    `cql:"job_id"`                   // Unique job identifier (varint in CQL, string in Go)
	TaskDefinitionID       int       `cql:"task_definition_id"`       // Task type (1, 2, or 7)
	Network                string    `cql:"network"`                  // Network the job is running on
	ScheduleType           string    `cql:"schedule_type"`            // "cron", "interval", "specific"
	TimeInterval           int64     `cql:"time_interval"`            // Interval in seconds (for interval type)
	CronExpression         string    `cql:"cron_expression"`          // Cron expression (for cron type)
	SpecificSchedule       string    `cql:"specific_schedule"`        // Specific timestamp (for specific type)
	Timezone               string    `cql:"timezone"`                 // Timezone selected by user
	NextExecutionTimestamp time.Time `cql:"next_execution_timestamp"` // Calculated next run time

	// Target data fields
	TargetChainID             string   `cql:"target_chain_id"`              // Chain to execute on
	// Nullable for TDI 7
	TargetContractAddress     string   `cql:"target_contract_address"`      // Contract to call
	TargetFunction            string   `cql:"target_function"`              // Function to invoke
	ABI                       string   `cql:"abi"`                          // Contract ABI (JSON)
	ArgType                   int      `cql:"arg_type"`                     // 0=None, 1=Static, 2=Dynamic
	Arguments                 []string `cql:"arguments"`                    // Static arguments (if ArgType=1, list<text> in CQL)
	// Nullable for TDI 1
	ExecutionScriptURL      string   `cql:"execution_script_url"`         // IPFS CID for script
	ExecutionScriptLanguage string   `cql:"execution_script_language"`    // Script language: 'ts', 'go', 'python', 'javascript'
	ExecutionScriptHash     string   `cql:"execution_script_hash"`        // keccak256(scriptCode) for verification
	MaxExecutionTime    int    `cql:"max_execution_time"`    // Script timeout in seconds (default 60)
	ChallengePeriod     int64  `cql:"challenge_period"`      // Challenge period in seconds (default 21600 = 6 hours)

	// Common status fields
	IsActive       bool      `cql:"is_active"`        // Job is active (true on creation, false when expiration time is reached)
	LastExecutedAt time.Time `cql:"last_executed_at"` // Last execution time
	ExpirationTime time.Time `cql:"expiration_time"`  // Job expiration
}

// EventJobDataEntity represents the event_job_data table for blockchain event-triggered jobs.
// Supports both traditional jobs (TDI 3, 4) and agent jobs (TDI 8).
// Primary Key: job_id
type EventJobDataEntity struct {
	JobID            string `cql:"job_id"`             // Unique job identifier (varint in CQL, string in Go)
	TaskDefinitionID int    `cql:"task_definition_id"` // Task type (3, 4, or 8)
	Network          string `cql:"network"`          // Network the job is running on
	Recurring        bool   `cql:"recurring"`          // Trigger on every event in time frame or once

	// Event trigger fields (common for all)
	TriggerChainID         string `cql:"trigger_chain_id"`         // Chain to monitor
	TriggerContractAddress string `cql:"trigger_contract_address"` // Contract to watch
	TriggerEvent           string `cql:"trigger_event"`            // Event signature (e.g., "Transfer(address,address,uint256)")
	EventFilterParaName    string `cql:"event_filter_para_name"`   // Indexed parameter to filter (e.g., "to")
	EventFilterValue       string `cql:"event_filter_value"`       // Filter value (e.g., specific address)

	// Target data fields
	TargetChainID             string   `cql:"target_chain_id"`              // Chain to execute on
	// Nullable for TDI 8
	TargetContractAddress     string   `cql:"target_contract_address"`      // Contract to call
	TargetFunction            string   `cql:"target_function"`              // Function to invoke
	ABI                       string   `cql:"abi"`                          // Contract ABI (JSON)
	ArgType                   int      `cql:"arg_type"`                     // 0=None, 1=Static, 2=Dynamic
	Arguments                 []string `cql:"arguments"`                    // Static arguments (if ArgType=1, list<text> in CQL)
	// Nullable for TDI 1
	ExecutionScriptURL      string   `cql:"execution_script_url"`         // IPFS CID for script
	ExecutionScriptLanguage string   `cql:"execution_script_language"`    // Script language: 'ts', 'go', 'python', 'javascript'
	ExecutionScriptHash     string   `cql:"execution_script_hash"`        // keccak256(scriptCode) for verification
	MaxExecutionTime    int    `cql:"max_execution_time"`    // Script timeout in seconds (default 60)
	ChallengePeriod     int64  `cql:"challenge_period"`      // Challenge period in seconds (default 21600 = 6 hours)

	// Common status fields
	IsActive       bool      `cql:"is_active"`        // Job is active (true on creation, false when expiration time is reached)
	LastExecutedAt time.Time `cql:"last_executed_at"` // Last execution time
	ExpirationTime time.Time `cql:"expiration_time"`  // Job expiration
}

// ConditionJobDataEntity represents the condition_job_data table for condition-based jobs.
// Supports both traditional jobs (TDI 5, 6) and agent jobs (TDI 9).
// Primary Key: job_id
type ConditionJobDataEntity struct {
	JobID            string `cql:"job_id"`             // Unique job identifier (varint in CQL, string in Go)
	TaskDefinitionID int    `cql:"task_definition_id"` // Task type (5, 6, or 9)
	Network          string `cql:"network"`          // Network the job is running on
	Recurring        bool   `cql:"recurring"`          // Trigger on every condition match or once

	// Condition trigger fields (common for all)
	ConditionType    string  `cql:"condition_type"`     // "balance", "state", "oracle"
	UpperLimit       float64 `cql:"upper_limit"`        // Upper threshold (double in CQL)
	LowerLimit       float64 `cql:"lower_limit"`        // Lower threshold (double in CQL)
	ValueSourceType  string  `cql:"value_source_type"`  // "api", "oracle", "websocket"
	ValueSourceURL   string  `cql:"value_source_url"`   // API URL or Websocket URL
	SelectedKeyRoute string  `cql:"selected_key_route"` // JSON path for API responses

	// Target data fields
	TargetChainID             string   `cql:"target_chain_id"`              // Chain to execute on
	// Nullable for TDI 9
	TargetContractAddress     string   `cql:"target_contract_address"`      // Contract to call
	TargetFunction            string   `cql:"target_function"`              // Function to invoke
	ABI                       string   `cql:"abi"`                          // Contract ABI (JSON)
	ArgType                   int      `cql:"arg_type"`                     // 0=None, 1=Static, 2=Dynamic
	Arguments                 []string `cql:"arguments"`                    // Static arguments (if ArgType=1, list<text> in CQL)
	// Nullable for TDI 1
	ExecutionScriptURL      string   `cql:"execution_script_url"`         // IPFS CID for script
	ExecutionScriptLanguage string   `cql:"execution_script_language"`    // Script language: 'ts', 'go', 'python', 'javascript'
	ExecutionScriptHash     string   `cql:"execution_script_hash"`        // keccak256(scriptCode) for verification
	MaxExecutionTime    int    `cql:"max_execution_time"`    // Script timeout in seconds (default 60)
	ChallengePeriod     int64  `cql:"challenge_period"`      // Challenge period in seconds (default 21600 = 6 hours)

	// Common status fields
	IsActive       bool      `cql:"is_active"`        // Job is active (true on creation, false when expiration time is reached)
	LastExecutedAt time.Time `cql:"last_executed_at"` // Last execution time
	ExpirationTime time.Time `cql:"expiration_time"`  // Job expiration
}

// TaskDataEntity represents the task_data table for individual task executions.
// Primary Key: task_id
type TaskDataEntity struct {
	TaskID               int64     `cql:"task_id"`                 // Unique task ID (bigint in CQL, auto-increment)
	TaskNumber           int64     `cql:"task_number"`             // Task sequence number on the contract
	TaskStatus           string    `cql:"task_status"`             // "created", "dispatched", "executed", "completed" , "failed"
	TaskError            string    `cql:"task_error"`              // Error message if task failed
	JobID                string    `cql:"job_id"`                  // Reference to job (varint in CQL, string in Go)
	TaskDefinitionID     int       `cql:"task_definition_id"`      // Task definition ID (1-9)
	CreatedAt            time.Time `cql:"created_at"`              // Task creation timestamp
	ExecutedAt           time.Time `cql:"executed_at"`             // Execution timestamp
	SubmittedAt          time.Time `cql:"submitted_at"`            // Submission timestamp
	TaskOpxPredictedCost string    `cql:"task_opx_predicted_cost"` // From JobCostPrediction (Wei)
	TaskOpxActualCost    string    `cql:"task_opx_actual_cost"`    // Actual cost including tx gas (Wei)
	ExecutionTxHash      string    `cql:"execution_tx_hash"`       // Transaction hash for action
	SubmissionTxHash     string    `cql:"submission_tx_hash"`      // Aggregator submission tx hash
	ConvertedArguments   []string  `cql:"converted_arguments"`     // Arguments used (JSON, list<text> in CQL)
	TaskPerformerAddress []string  `cql:"task_performer_address"`  // Keeper who executed (list<text> in CQL, multiple if first keeper failed)
	TaskAttesterAddress  []string  `cql:"task_attester_address"`   // Keepers who attested (list<text> in CQL)
	ProofOfTask          string    `cql:"proof_of_task"`           // Cryptographic proof (IPFS CID)
	IsSuccessful         bool      `cql:"is_successful"`           // Execution success
	IsAccepted           bool      `cql:"is_accepted"`             // Consensus accepted
	Network              string    `cql:"network"`                 // Network the task is running on
}

// KeeperDataEntity represents the keeper_data table for Keeper node registry.
// Primary Key: keeper_address
type KeeperDataEntity struct {
	KeeperAddress     string      `cql:"keeper_address"`     // Wallet address (primary key)
	KeeperName        string      `cql:"keeper_name"`        // Human-readable keeper name
	RewardsAddress    string      `cql:"rewards_address"`    // Address for reward payouts
	ConsensusAddress  string      `cql:"consensus_address"`  // Consensus layer address
	OperatorID        int         `cql:"operator_id"`        // Unique operator ID on the contract (int64 in CQL)
	VotingPower       string      `cql:"voting_power"`       // Voting power (Wei-based stake)
	Registered        bool        `cql:"registered"`         // Registered on-chain
	RegisteredAt      []time.Time `cql:"registered_at"`      // List of registration timestamps (list<timestamp> in CQL)
	Online            bool        `cql:"online"`             // Currently online
	Version           string      `cql:"version"`            // Keeper software version
	Network           string      `cql:"network"`            // Network the keeper is running on
	ConnectionAddress string      `cql:"connection_address"` // Connection address for P2P
	PeerID            string      `cql:"peer_id"`            // Peer ID for P2P
	Uptime            int64       `cql:"uptime"`             // Total uptime (seconds, bigint in CQL)
	LastCheckedIn     time.Time   `cql:"last_checked_in"`    // Last heartbeat time
	RewardsBooster    string      `cql:"rewards_booster"`    // Reward multiplier (for testnets, obsolete for mainnet)
	NoExecutedTasks   int64       `cql:"no_executed_tasks"`  // Total tasks executed (bigint in CQL)
	NoAttestedTasks   int64       `cql:"no_attested_tasks"`  // Total tasks attested (bigint in CQL)
	KeeperPoints      string      `cql:"keeper_points"`      // Total points (Wei, sum of TaskOpxCost executed/attested, daily rewards)
	ChatID            int64       `cql:"chat_id"`            // Telegram chat ID for alerts (int64 in CQL)
	EmailID           string      `cql:"email_id"`           // Email for alerts
}

// AgentScriptExecutionsEntity represents the agent_script_executions table for tracking agent script executions (TDI 7, 8, 9) with fraud-proof data.
// Primary Key: execution_id
type AgentScriptExecutionsEntity struct {
	ExecutionID      string    `cql:"execution_id"`       // Unique execution identifier (UUID, primary key)
	JobID            string    `cql:"job_id"`             // Reference to job (varint in CQL, string in Go)
	TaskID           int64     `cql:"task_id"`            // Link to task_data table (bigint in CQL)
	TaskDefinitionID int       `cql:"task_definition_id"` // 7, 8, or 9
	ScheduledTime    time.Time `cql:"scheduled_time"`     // When it was supposed to execute
	ActualTime       time.Time `cql:"actual_time"`        // When it actually executed
	PerformerAddress string    `cql:"performer_address"`  // Keeper who executed the script

	// Input data (for deterministic re-execution)
	InputTimestamp int64  `cql:"input_timestamp"` // Timestamp used as input (bigint in CQL)
	InputStorage   string `cql:"input_storage"`   // JSON snapshot of storage at execution time
	InputHash      string `cql:"input_hash"`      // Hash of inputs for verification

	// Trigger-specific input data (JSON)
	TriggerData string `cql:"trigger_data"` // TDI 7: {scheduled_time}
	// TDI 8: {event_data, block_number, tx_hash, event_params}
	// TDI 9: {condition_value, timestamp, condition_type}

	// Output data
	ShouldExecute  bool   `cql:"should_execute"`  // Whether to submit on-chain transaction
	TargetContract string `cql:"target_contract"` // Contract address to call
	Calldata       string `cql:"calldata"`        // Encoded function call (hex string)
	OutputHash     string `cql:"output_hash"`     // Hash of outputs for verification

	// Metadata (CRITICAL: API calls, contract calls, block numbers)
	ExecutionMetadata string `cql:"execution_metadata"` // JSON containing:
	// - API calls made: [{url, blockNumber, response}, ...]
	// - Contract calls: [{contract, blockNumber, function, response}, ...]
	// - Oracle calls: [{oracle, blockNumber, data}, ...]

	// Proof
	ScriptHash string `cql:"script_hash"` // Hash of the script code
	Signature  string `cql:"signature"`   // Performer's signature

	// Execution result
	TxHash          string `cql:"tx_hash"`          // Transaction hash if executed
	ExecutionStatus string `cql:"execution_status"` // 'success', 'failed', 'no_execution'
	ExecutionError  string `cql:"execution_error"`  // Error message if failed

	// Verification status
	VerificationStatus string    `cql:"verification_status"` // 'pending', 'verified', 'challenged', 'slashed'
	ChallengeDeadline  time.Time `cql:"challenge_deadline"`  // Deadline for challenges
	IsChallenged       bool      `cql:"is_challenged"`       // Whether execution has been challenged
	ChallengeCount     int       `cql:"challenge_count"`     // Number of challenges raised
	CreatedAt          time.Time `cql:"created_at"`          // Record creation timestamp
}

// ScriptStorageEntity represents the script_storage table for persistent key-value storage per agent job.
// Used by TDI 7, 8, 9 for maintaining state across executions.
// Primary Key: (job_id, storage_key)
type ScriptStorageEntity struct {
	JobID        string    `cql:"job_id"`        // Job identifier (varint in CQL, string in Go, part of composite primary key)
	StorageKey   string    `cql:"storage_key"`   // Storage key (text, part of composite primary key)
	StorageValue string    `cql:"storage_value"` // Storage value (text, JSON or string)
	UpdatedAt    time.Time `cql:"updated_at"`    // Last update timestamp
}

// ExecutionChallengesEntity represents the execution_challenges table for fraud-proof challenges against agent script executions.
// Primary Key: challenge_id
type ExecutionChallengesEntity struct {
	ChallengeID       string `cql:"challenge_id"`       // Unique challenge identifier (UUID, primary key)
	ExecutionID       string `cql:"execution_id"`       // Reference to agent_script_executions
	ChallengerAddress string `cql:"challenger_address"` // Address of the challenger
	ChallengeReason   string `cql:"challenge_reason"`   // 'wrong_output', 'missing_execution', 'invalid_calldata', 'wrong_trigger'

	// Challenger's claimed output
	ChallengerOutputHash     string `cql:"challenger_output_hash"`     // Hash of challenger's claimed output
	ChallengerShouldExecute  bool   `cql:"challenger_should_execute"`  // Challenger's claim: should execute?
	ChallengerTargetContract string `cql:"challenger_target_contract"` // Challenger's claimed target contract
	ChallengerCalldata       string `cql:"challenger_calldata"`        // Challenger's claimed calldata
	ChallengerSignature      string `cql:"challenger_signature"`       // Challenger's signature

	// Challenge bond (prevents spam)
	BondAmount string `cql:"bond_amount"` // Wei-based bond amount

	// Resolution
	ResolutionStatus string    `cql:"resolution_status"` // 'pending', 'approved', 'rejected', 'inconclusive'
	ResolutionTime   time.Time `cql:"resolution_time"`   // When challenge was resolved
	ValidatorCount   int       `cql:"validator_count"`   // Number of validators who voted
	ApproveCount     int       `cql:"approve_count"`     // Number of approving votes
	RejectCount      int       `cql:"reject_count"`      // Number of rejecting votes
	CreatedAt        time.Time `cql:"created_at"`        // Challenge creation timestamp
}
