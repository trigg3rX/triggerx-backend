package types

import (
	"time"
)

type ExecutionResult struct {
	Stats    ResourceStats `json:"stats"`
	Output   string        `json:"output"`
	Success  bool          `json:"success"`
	Error    error         `json:"error,omitempty"`
	Warnings []string      `json:"warnings,omitempty"`
}
