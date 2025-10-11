package types

import "time"

// HealthCheckResponse represents the response from the health check endpoint across all services
type HealthCheckResponse struct {
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
	Service   string    `json:"service"`
	Version   string    `json:"version"`
	Error     string    `json:"error,omitempty"`
}