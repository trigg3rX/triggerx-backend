package types

import "time"

type UpdateJobRequest struct {
	JobID          int64     `json:"job_id"`
	Recurring      bool      `json:"recurring"`
	TimeFrame      int64     `json:"time_frame"`
	UpdatedAt      time.Time `json:"updated_at"`
	LastExecutedAt time.Time `json:"last_executed_at"`
}
