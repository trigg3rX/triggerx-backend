package types

import "time"

const (
	BlockConfirmations   = 3                // Wait for 3 block confirmations
	PollInterval         = 10 * time.Second // Poll every 10 seconds for new blocks
	WorkerTimeout        = 30 * time.Second // Timeout for worker operations
	MaxRetries           = 3                // Max retries for failed operations
	PerformerLockTTL     = 15 * time.Minute // Lock duration for job execution
	BlockCacheTTL        = 2 * time.Minute  // Cache TTL for block data
	EventCacheTTL        = 10 * time.Minute // Cache TTL for event data
	DuplicateEventWindow = 30 * time.Second // Window to prevent duplicate event processing
)
// JobScheduleRequest represents the request to schedule a new job
type JobScheduleRequest struct {
	JobID                         int64    `json:"job_id" binding:"required"`
	TimeFrame                     int64    `json:"time_frame"`
	Recurring                     bool     `json:"recurring"`
	TriggerChainID                string   `json:"trigger_chain_id" binding:"required"`
	TriggerContractAddress        string   `json:"trigger_contract_address" binding:"required"`
	TriggerEvent                  string   `json:"trigger_event" binding:"required"`
	TargetChainID                 string   `json:"target_chain_id" binding:"required"`
	TargetContractAddress         string   `json:"target_contract_address" binding:"required"`
	TargetFunction                string   `json:"target_function" binding:"required"`
	ABI                           string   `json:"abi"`
	ArgType                       int      `json:"arg_type"`
	Arguments                     []string `json:"arguments"`
	DynamicArgumentsScriptIPFSUrl string   `json:"dynamic_arguments_script_ipfs_url"`
}