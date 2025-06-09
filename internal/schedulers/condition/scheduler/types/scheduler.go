package types

import (
	"time"
)

const (
	MaxRetries     = 3                // Max retries for failed operations
	RequestTimeout = 10 * time.Second // HTTP request timeout
	WorkerTimeout  = 30 * time.Second // Timeout for worker operations

	PollInterval = 1 * time.Second // Poll every 1 second as requested
)

// JobScheduleRequest represents the request to schedule a new condition job
type JobScheduleRequest struct {
	JobID                         int64    `json:"job_id" binding:"required"`
	TimeFrame                     int64    `json:"time_frame"`
	Recurring                     bool     `json:"recurring"`
	ConditionType                 string   `json:"condition_type" binding:"required"`
	UpperLimit                    float64  `json:"upper_limit"`
	LowerLimit                    float64  `json:"lower_limit"`
	ValueSourceType               string   `json:"value_source_type" binding:"required"`
	ValueSourceUrl                string   `json:"value_source_url" binding:"required"`
	TargetChainID                 string   `json:"target_chain_id" binding:"required"`
	TargetContractAddress         string   `json:"target_contract_address" binding:"required"`
	TargetFunction                string   `json:"target_function" binding:"required"`
	ABI                           string   `json:"abi"`
	ArgType                       int      `json:"arg_type"`
	Arguments                     []string `json:"arguments"`
	DynamicArgumentsScriptIPFSUrl string   `json:"dynamic_arguments_script_ipfs_url"`
}
