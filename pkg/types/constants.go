package types

const (
	// Time-based job task definitions
	TaskDefTimeBasedStart = 1
	TaskDefTimeBasedEnd   = 2

	// Event-based job task definitions
	TaskDefEventBasedStart = 3
	TaskDefEventBasedEnd   = 4

	// Condition-based job task definitions
	TaskDefConditionBasedStart = 5
	TaskDefConditionBasedEnd   = 6
)

type JobStatus string

const (
	JobStatusPending JobStatus = "pending"
	JobStatusInQueue JobStatus = "in-queue"
	JobStatusRunning JobStatus = "running"
)

const (
	TGPerPoint = 1 // 1 TG = 1 Point
	WeiPerTG = 1000000000000000 // 10^15
	EtherPerTG = 0.001 // 10^-3
)