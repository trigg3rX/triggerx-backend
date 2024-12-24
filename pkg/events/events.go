package events

import (
	"time"
)

type EventType string

const (
	JobCreated EventType = "JOB_CREATED"
	JobUpdated EventType = "JOB_UPDATED"

	TaskCreated EventType = "TASK_CREATED"

	UserCreated EventType = "USER_CREATED"
	UserUpdated EventType = "USER_UPDATED"

	KeeperRegistered EventType = "KEEPER_REGISTERED"
	KeeperDeregistered EventType = "KEEPER_DEREGISTERED"

	QuorumCreated EventType = "QUORUM_CREATED"
	QuorumUpdated EventType = "QUORUM_UPDATED"

	TaskHistoryCreated EventType = "TASK_HISTORY_CREATED"
	// Add other event types as needed
)

type JobCreatedEvent struct {
	JobID       int64     `json:"job_id"`
	JobType     int       `json:"jobType"`
	UserAddress string    `json:"user_address"`
	ChainID     int       `json:"chain_id"`
	Status      bool      `json:"status"`
	TimeCheck   time.Time `json:"time_check"`
}

type JobUpdatedEvent struct {
	JobID       int64     `json:"job_id"`
	JobType     int       `json:"jobType"`
	UserAddress string    `json:"user_address"`
	ChainID     int       `json:"chain_id"`
	Status      bool      `json:"status"`
	TimeCheck   time.Time `json:"time_check"`
}

type TaskCreatedEvent struct {
	TaskID int64 `json:"task_id"`
}

type UserCreatedEvent struct {
	UserID int64 `json:"user_id"`
}

type UserUpdatedEvent struct {
	UserID int64 `json:"user_id"`
}

type KeeperRegisteredEvent struct {
	KeeperID int64 `json:"keeper_id"`
}

type KeeperDeregisteredEvent struct {
	KeeperID int64 `json:"keeper_id"`
}

type QuorumCreatedEvent struct {
	QuorumID     int64 `json:"quorum_id"`
	QuorumNumber int   `json:"quorum_number"`
}

type QuorumUpdatedEvent struct {
	QuorumID     int64 `json:"quorum_id"`
	QuorumNumber int   `json:"quorum_number"`
}

type TaskHistoryCreatedEvent struct {
	TaskID int64 `json:"task_id"`
}

// Event represents a generic event in the system
type Event struct {
	Type    EventType   `json:"type"`
	Payload interface{} `json:"payload"`
}
