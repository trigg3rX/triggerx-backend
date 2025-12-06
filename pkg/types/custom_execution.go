package types

import "time"

// CustomScriptExecution tracks each execution of a custom script job
type CustomScriptExecution struct {
	ExecutionID      string    `json:"execution_id" db:"execution_id"`
	JobID            *BigInt   `json:"job_id" db:"job_id"`
	TaskID           int64     `json:"task_id" db:"task_id"`
	ScheduledTime    time.Time `json:"scheduled_time" db:"scheduled_time"`
	ActualTime       time.Time `json:"actual_time" db:"actual_time"`
	PerformerAddress string    `json:"performer_address" db:"performer_address"`

	// Inputs (for deterministic re-execution)
	InputTimestamp int64  `json:"input_timestamp" db:"input_timestamp"`
	InputStorage   string `json:"input_storage" db:"input_storage"` // JSON snapshot
	InputHash      string `json:"input_hash" db:"input_hash"`

	// Outputs
	ShouldExecute  bool   `json:"should_execute" db:"should_execute"`
	TargetContract string `json:"target_contract" db:"target_contract"`
	Calldata       string `json:"calldata" db:"calldata"`
	OutputHash     string `json:"output_hash" db:"output_hash"`

	// Metadata (API calls, contract calls, block numbers)
	ExecutionMetadata string `json:"execution_metadata" db:"execution_metadata"` // JSON

	// Proof
	ScriptHash string `json:"script_hash" db:"script_hash"`
	Signature  string `json:"signature" db:"signature"`

	// Execution Result
	TxHash          string `json:"tx_hash" db:"tx_hash"`
	ExecutionStatus string `json:"execution_status" db:"execution_status"` // 'success', 'failed', 'no_execution'
	ExecutionError  string `json:"execution_error" db:"execution_error"`

	// Verification
	VerificationStatus string    `json:"verification_status" db:"verification_status"` // 'pending', 'verified', 'challenged', 'slashed'
	ChallengeDeadline  time.Time `json:"challenge_deadline" db:"challenge_deadline"`
	IsChallenged       bool      `json:"is_challenged" db:"is_challenged"`
	ChallengeCount     int       `json:"challenge_count" db:"challenge_count"`

	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// ExecutionMetadata contains all API/contract call metadata
type ExecutionMetadata struct {
	Timestamp   int64         `json:"timestamp"`
	Reason      string        `json:"reason"`
	GasEstimate uint64        `json:"gasEstimate,omitempty"`
	APICalls    []APICallInfo `json:"apiCalls,omitempty"`
	ContractCalls []ContractCallInfo `json:"contractCalls,omitempty"`
}

// APICallInfo records non-deterministic API calls
type APICallInfo struct {
	URL          string      `json:"url"`
	BlockNumber  uint64      `json:"blockNumber,omitempty"` // Block at time of call
	Response     interface{} `json:"response"`
	StatusCode   int         `json:"statusCode"`
	Timestamp    int64       `json:"timestamp"`
}

// ContractCallInfo records deterministic contract/oracle calls
type ContractCallInfo struct {
	Contract    string      `json:"contract"`
	Function    string      `json:"function"`
	BlockNumber uint64      `json:"blockNumber"` // CRITICAL: for re-execution
	Response    interface{} `json:"response"`
	ChainID     string      `json:"chainId"`
}

// ScriptStorage stores persistent key-value pairs for scripts
type ScriptStorage struct {
	JobID        *BigInt   `json:"job_id" db:"job_id"`
	StorageKey   string    `json:"storage_key" db:"storage_key"`
	StorageValue string    `json:"storage_value" db:"storage_value"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

// CustomScriptOutput is the expected output format from user scripts
type CustomScriptOutput struct {
	ShouldExecute  bool                       `json:"shouldExecute"`
	TargetContract string                     `json:"targetContract,omitempty"`
	Calldata       string                     `json:"calldata,omitempty"`
	Metadata       CustomScriptOutputMetadata `json:"metadata"`
	StorageUpdates map[string]string          `json:"storageUpdates,omitempty"` // Phase 1: Storage updates from script
}

// CustomScriptOutputMetadata contains execution information
type CustomScriptOutputMetadata struct {
	Timestamp     int64             `json:"timestamp"`
	Reason        string            `json:"reason"`
	GasEstimate   uint64            `json:"gasEstimate,omitempty"`
	APICalls      []APICallInfo     `json:"apiCalls,omitempty"`
	ContractCalls []ContractCallInfo `json:"contractCalls,omitempty"`
}

// ExecutionProof represents cryptographic proof of script execution
type ExecutionProof struct {
	ExecutionID      string `json:"execution_id"`
	JobID            string `json:"job_id"`
	Timestamp        int64  `json:"timestamp"`
	ScriptHash       string `json:"script_hash"`
	InputHash        string `json:"input_hash"`
	OutputHash       string `json:"output_hash"`
	Signature        string `json:"signature"`
	PerformerAddress string `json:"performer_address"`
}

// ScheduleCustomTaskData is passed from scheduler to task dispatcher
type ScheduleCustomTaskData struct {
	TaskID           int64     `json:"task_id"`
	TaskDefinitionID int       `json:"task_definition_id"`
	JobID            *BigInt   `json:"job_id"`
	CustomScriptUrl  string    `json:"custom_script_url"`
	ScriptLanguage   string    `json:"script_language"`
	ScriptHash       string    `json:"script_hash"`
	ScheduledTime    time.Time `json:"scheduled_time"`
	TimeInterval     int64     `json:"time_interval"`
	ChallengePeriod  int64     `json:"challenge_period"`
	LastExecutedAt   time.Time `json:"last_executed_at"`
	ExpirationTime   time.Time `json:"expiration_time"`
}

// ExecutionChallenge represents a challenge to an execution
type ExecutionChallenge struct {
	ChallengeID      string    `json:"challenge_id" db:"challenge_id"`
	ExecutionID      string    `json:"execution_id" db:"execution_id"`
	ChallengerAddress string   `json:"challenger_address" db:"challenger_address"`
	ChallengeReason  string    `json:"challenge_reason" db:"challenge_reason"`

	// Challenger's claimed output
	ChallengerOutputHash     string `json:"challenger_output_hash" db:"challenger_output_hash"`
	ChallengerShouldExecute  bool   `json:"challenger_should_execute" db:"challenger_should_execute"`
	ChallengerTargetContract string `json:"challenger_target_contract" db:"challenger_target_contract"`
	ChallengerCalldata       string `json:"challenger_calldata" db:"challenger_calldata"`
	ChallengerSignature      string `json:"challenger_signature" db:"challenger_signature"`

	// Resolution
	ResolutionStatus string    `json:"resolution_status" db:"resolution_status"`
	ResolutionTime   time.Time `json:"resolution_time" db:"resolution_time"`
	ValidatorCount   int       `json:"validator_count" db:"validator_count"`
	ApproveCount     int       `json:"approve_count" db:"approve_count"`
	RejectCount      int       `json:"reject_count" db:"reject_count"`

	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// ValidationRequest is sent to validators for re-execution
type ValidationRequest struct {
	ExecutionID     string          `json:"execution_id"`
	JobID           *BigInt         `json:"job_id"`
	ScriptHash      string          `json:"script_hash"`
	ScriptURL       string          `json:"script_url"`
	InputTimestamp  int64           `json:"input_timestamp"`
	InputStorage    string          `json:"input_storage"`
	PerformerOutput PerformerOutput `json:"performer_output"`
	Metadata        ExecutionMetadata `json:"metadata"` // Includes API responses
}

// PerformerOutput contains the original performer's execution result
type PerformerOutput struct {
	ShouldExecute  bool   `json:"should_execute"`
	TargetContract string `json:"target_contract"`
	Calldata       string `json:"calldata"`
	OutputHash     string `json:"output_hash"`
}

// Attestation represents a validator's vote on an execution
type Attestation struct {
	ValidatorAddress string `json:"validator_address"`
	ExecutionID      string `json:"execution_id"`
	Approved         bool   `json:"approved"`
	Reason           string `json:"reason"`
	OutputHash       string `json:"output_hash"`
	Signature        string `json:"signature"`
}
