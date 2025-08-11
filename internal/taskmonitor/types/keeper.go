package types

import "time"

type TaskSubmissionData struct {
	TaskID int64
	TaskNumber int64
	TaskDefinitionID int
	IsAccepted bool
	TaskSubmissionTxHash string
	PerformerAddress string
	AttesterIds []int64
	ExecutionTxHash string
	ExecutionTimestamp time.Time
	TaskOpxCost float64
	ProofOfTask string
}
